package command

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/template"
	"github.com/andrewpillar/jrnl/util"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

var (
	journalTitle string

	categories []category.Category

	postIndexes map[string][]post.Post

	yearPattern = "[_site]/[0-9]{4}"

	monthPattern = "[_site]/[0-9]{4}/[0-9]{2}"

	dayPattern = "[_site]/[0-9]{4}/[0-9]{2}/[0-9]{2}"

	categoryYearPattern = "[_site]/[-a-z0-9/]+/[0-9]{4}"

	categoryMonthPattern = "[_site]/[-a-z0-9/]+/[0-9]{4}/[0-9]{2}"

	categoryDayPattern = "[_site]/[-a-z0-9/]+/[0-9]{4}/[0-9]{2}/[0-9]{2}"

	categoryPattern = "[_site]/[-a-z0-9/]+"

	yearRegex = regexp.MustCompile(yearPattern)

	monthRegex = regexp.MustCompile(monthPattern)

	dayRegex = regexp.MustCompile(dayPattern)

	categoryYearRegex = regexp.MustCompile(categoryYearPattern)

	categoryMonthRegex = regexp.MustCompile(categoryMonthPattern)

	categoryDayRegex = regexp.MustCompile(categoryDayPattern)

	categoryRegex = regexp.MustCompile(categoryPattern)
)

func getHostKey(hostname string) (ssh.PublicKey, error) {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	s := bufio.NewScanner(f)

	var hostKey ssh.PublicKey

	for s.Scan() {
		fields := strings.Split(s.Text(), " ")

		if len(fields) != 3 {
			continue
		}

		if strings.Contains(fields[0], hostname) {
			var err error

			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(s.Bytes())

			if err != nil {
				return nil, err
			}

			break
		}
	}

	if hostKey == nil {
		return nil, errors.New("no key for host " + hostname)
	}

	return hostKey, nil
}

func indexPost(p post.Post, wg *sync.WaitGroup, m *sync.Mutex) {
	defer wg.Done()

	if !p.Index {
		return
	}

	parts := strings.Split(filepath.Dir(p.SitePath), string(os.PathSeparator))

	for i := range parts {
		path := filepath.Join(parts[:len(parts) - i - 1]...)

		if path == "" {
			continue
		}

		m.Lock()

		_, ok := postIndexes[path]

		if !ok {
			postIndexes[path] = make([]post.Post, 0)
		}

		postIndexes[path] = append(postIndexes[path], p)

		m.Unlock()
	}
}

func publishPost(
	p post.Post,
	wg *sync.WaitGroup,
	published chan<- post.Post,
	errs chan<- error,
) {
	defer wg.Done()

	if err := p.Load(); err != nil {
		errs <- err
		return
	}

	b := []byte{}

	if p.Layout != "" {
		f, err := os.Open(filepath.Join(meta.LayoutsDir, p.Layout))

		if err != nil {
			errs <- err
			return
		}

		defer f.Close()

		b, err = ioutil.ReadAll(f)

		if err != nil {
			errs <- err
			return
		}
	}

	p.Convert()

	if err := p.Publish(journalTitle, string(b), categories); err != nil {
		errs <- err
		return
	}

	published <- p
}

func publishPosts(posts []post.Post) (<-chan post.Post, <-chan error) {
	wg := &sync.WaitGroup{}

	published := make(chan post.Post)
	errs := make(chan error)

	for _, p := range posts {
		wg.Add(1)

		go publishPost(p, wg, published, errs)
	}

	go func() {
		wg.Wait()

		close(errs)
		close(published)
	}()

	return published, errs
}

func publishToRemote(remote meta.Remote) {
	if filepath.IsAbs(remote.Target) {
		if err := util.Copy(meta.SiteDir, remote.Target); err != nil {
			util.Exit("failed to publish to " + remote.Target, err)
		}
	}

	parts := strings.Split(remote.Target, "@")
	i := 0

	user := os.Getenv("USER")

	if len(parts) > 1 {
		user = parts[0]
		i = 1
	}

	parts = strings.Split(parts[i], ":")

	hostname := parts[0]
	dir := parts[1]

	addr := fmt.Sprintf("%s:%d", hostname, remote.Port)

	if remote.Identity == "" {
		remote.Identity = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}

	key, err := ioutil.ReadFile(remote.Identity)

	if err != nil {
		util.Exit("failed to get identity file", err)
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		util.Exit("failed to parse private key", err)
	}

	hostKey, err := getHostKey(hostname)

	if err != nil {
		util.Exit("failed to find host key in known_hosts", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	conn, err := ssh.Dial("tcp", addr, config)

	if err != nil {
		util.Exit("failed to establish initial connection", err)
	}

	defer conn.Close()

	scp, err := sftp.NewClient(conn)

	if err != nil {
		util.Exit("failed to establish scp connection", err)
	}

	defer scp.Close()

	if err := util.CopyToRemote(meta.SiteDir, dir, scp); err != nil {
		util.Exit("failed to copy to remote", err)
	}
}

func writeIndex(
	m meta.Meta,
	dir string,
	posts []post.Post,
	wg *sync.WaitGroup,
	errs chan<- error,
) {
	defer wg.Done()

	index := filepath.Join(dir, "index.html")
	layout := ""

	var data interface{}

	if dir == meta.SiteDir {
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayout)

		data = struct{
			Title      string
			Posts      []post.Post
			Categories []category.Category
		}{
			Title:      journalTitle,
			Posts:      posts,
			Categories: categories,
		}

		if err := writeIndexFile(layout, index, data); err != nil {
			errs <- err
			return
		}

		return
	}

	pattern := []byte(dir)

	parts := strings.Split(dir, string(os.PathSeparator))

	dateIndex := false

	timeFmt := ""
	i := 0

	if dayRegex.Match(pattern) {
		dateIndex = true

		layout = filepath.Join(meta.LayoutsDir, m.DayIndexLayout)

		timeFmt = filepath.Join("2006", "01", "02")
		i = 3
	} else if monthRegex.Match(pattern) {
		dateIndex = true

		layout = filepath.Join(meta.LayoutsDir, m.MonthIndexLayout)

		timeFmt = filepath.Join("2006", "01")
		i = 2
	} else if yearRegex.Match(pattern) {
		dateIndex = true

		layout = filepath.Join(meta.LayoutsDir, m.YearIndexLayout)

		timeFmt = filepath.Join("2006")
		i = 1
	}

	if dateIndex {
		t, err := time.Parse(timeFmt, filepath.Join(parts[len(parts) - i:]...))

		if err != nil {
			errs <- err
			return
		}

		data = struct{
			Title      string
			Time       time.Time
			Posts      []post.Post
			Categories []category.Category
		}{
			Title:      journalTitle,
			Time:       t,
			Posts:      posts,
			Categories: categories,
		}

		if err = writeIndexFile(layout, index, data); err != nil {
			errs <- err
			return
		}

		return
	}

	categoryDateIndex := false

	if categoryDayRegex.Match(pattern) {
		categoryDateIndex = true

		layout = filepath.Join(meta.LayoutsDir, m.CategoryDayIndexLayout)

		timeFmt = filepath.Join("2006", "01", "02")
		i = 3
	} else if categoryMonthRegex.Match(pattern) {
		categoryDateIndex = true

		layout = filepath.Join(meta.LayoutsDir, m.CategoryMonthIndexLayout)

		timeFmt = filepath.Join("2006", "01")
		i = 2
	} else if categoryYearRegex.Match(pattern) {
		categoryDateIndex = true

		layout = filepath.Join(meta.LayoutsDir, m.CategoryYearIndexLayout)

		timeFmt = filepath.Join("2006")
		i = 1
	}

	if categoryDateIndex {
		id := strings.Join(parts[1:len(parts) - i], " ")

		c, err := category.Find(id)

		if err != nil {
			errs <- err
			return
		}

		t, err := time.Parse(timeFmt, filepath.Join(parts[len(parts) - i:]...))

		if err != nil {
			errs <- err
			return
		}

		data = struct{
			Title      string
			Category   category.Category
			Time       time.Time
			Posts      []post.Post
			Categories []category.Category
		}{
			Title:      journalTitle,
			Category:   c,
			Time:       t,
			Posts:      posts,
			Categories: categories,
		}

		if err = writeIndexFile(layout, index, data); err != nil {
			errs <- err
			return
		}

		return
	}

	if categoryRegex.Match(pattern) {
		id := strings.Join(parts[1:len(parts) - i], " ")

		c, err := category.Find(id)

		if err != nil {
			errs <- err
			return
		}

		data = struct{
			Title      string
			Category   category.Category
			Posts      []post.Post
			Categories []category.Category
		}{
			Title:      journalTitle,
			Category:   c,
			Posts:      posts,
			Categories: categories,
		}

		if err = writeIndexFile(layout, index, data); err != nil {
			errs <- err
			return
		}

		return
	}

	errs <- errors.New("could not match pattern to dir " + dir)
}

func writeIndexes(m meta.Meta, wg *sync.WaitGroup) <-chan error {
	errs := make(chan error)

	for dir, posts := range postIndexes {
		wg.Add(1)

		go writeIndex(m, dir, posts, wg, errs)
	}

	go func() {
		wg.Wait()

		close(errs)
	}()

	return errs
}

func writeIndexFile(layout, fname string, data interface{}) error {
	if data == nil {
		return errors.New("no data for index " + fname)
	}

	if layout == meta.LayoutsDir {
		return errors.New("no layout for index " + fname)
	}

	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return err
	}

	dst, err := os.OpenFile(fname, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		return err
	}

	defer dst.Close()

	return template.Render(dst, fname, string(b), data)
}

func Publish(c cli.Command) {
	util.MustBeInitialized()

	postIndexes = make(map[string][]post.Post)

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	journalTitle = m.Title

	tmp, err := category.ResolveCategories()

	if err != nil {
		util.Exit("failed to resolve categories", err)
	}

	categories = tmp

	posts, err := post.ResolvePosts()

	if err != nil {
		util.Exit("failed to resolve posts", err)
	}

	code := 0

	published, errs := publishPosts(posts)

	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}

	for {
		select {
			case err, ok := <-errs:
				if !ok {
					errs = nil
				} else {
					code = 1
					fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
				}
			case p, ok := <-published:
				if !ok {
					published = nil
				} else {
					wg.Add(1)

					go indexPost(p, wg, mut)
				}
		}

		if errs == nil && published == nil {
			break
		}
	}

	wg.Wait()

	errs = writeIndexes(*m, wg)

	for err := range errs {
		code = 1

		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}

	if !c.Flags.IsSet("draft") {
		alias := c.Flags.GetString("remote")

		if alias == "" {
			alias = m.Default
		}

		if alias == "" {
			util.Exit("missing remote", errors.New("no default set"))
		}

		remote, ok := m.Remotes[alias]

		if !ok {
			util.Exit("failed to find remote", errors.New(alias))
		}

		publishToRemote(remote)
	}

	os.Exit(code)
}
