package command

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/template"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

var (
	journalTitle string

	categories []category.Category

	postIndexes = make(map[string][]post.Post)

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

func indexPost(p post.Post, wg *sync.WaitGroup, m *sync.Mutex) {
	defer wg.Done()

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

	f, err := os.Open(filepath.Join(meta.LayoutsDir, meta.PostLayout))

	if err != nil {
		errs <- err
		return
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)

	if err != nil {
		errs <- err
		return
	}

	p.Convert()

	if err := p.Publish(journalTitle, string(b), categories); err != nil {
		errs <- err
		return
	}

	published <- p
}

func publishPosts(posts []post.Post) (chan post.Post, chan error) {
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

func writeIndex(
	dir string,
	posts []post.Post,
	wg *sync.WaitGroup,
	errs chan<- error,
) {
	defer wg.Done()

	index := filepath.Join(dir, "index.html")
	layout := ""

	var data interface{}

	// The current directory we have is the top level _site directory.
	if dir == meta.SiteDir {
		layout = filepath.Join(meta.LayoutsDir, meta.IndexLayout)

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

	notCategory := false

	timeFormat := ""
	timeIndex := 0

	if dayRegex.Match(pattern) {
		notCategory = true

		layout = filepath.Join(meta.LayoutsDir, meta.DayIndexLayout)

		timeFormat = filepath.Join("2006", "01", "02")
		timeIndex = 3
	} else if monthRegex.Match(pattern) {
		notCategory = true

		layout = filepath.Join(meta.LayoutsDir, meta.MonthIndexLayout)

		timeFormat = filepath.Join("2006", "01")
		timeIndex = 2
	} else if yearRegex.Match(pattern) {
		notCategory = true

		layout = filepath.Join(meta.LayoutsDir, meta.YearIndexLayout)

		timeFormat = filepath.Join("2006")
		timeIndex = 1
	}

	// We aren't at a category directory, just an arbitrary date directory for
	// posts which don't have a category.
	if notCategory {
		t, err := time.Parse(
			timeFormat,
			filepath.Join(parts[len(parts) - timeIndex:]...),
		)

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

	categoryDate := false

	if categoryDayRegex.Match(pattern) {
		categoryDate = true

		layout = filepath.Join(meta.LayoutsDir, meta.CategoryDayIndexLayout)

		timeFormat = filepath.Join("2006", "01", "02")
		timeIndex = 3
	} else if categoryMonthRegex.Match(pattern) {
		categoryDate = true

		layout = filepath.Join(meta.LayoutsDir, meta.CategoryMonthIndexLayout)

		timeFormat = filepath.Join("2006", "01")
		timeIndex = 2
	} else if categoryYearRegex.Match(pattern) {
		categoryDate = true

		layout = filepath.Join(meta.LayoutsDir, meta.CategoryYearIndexLayout)

		timeFormat = filepath.Join("2006")
		timeIndex = 1
	}

	// We have a date directory for a category.
	if categoryDate {
		slug := strings.Join(parts[1:len(parts) - timeIndex], " ")
		name := util.Deslug(slug, " / ")

		t, err := time.Parse(
			timeFormat,
			filepath.Join(parts[len(parts) - timeIndex:]...),
		)

		if err != nil {
			errs <- err
			return
		}

		data = struct{
			Title      string
			Category   string
			Time       time.Time
			Posts      []post.Post
			Categories []category.Category
		}{
			Title:      journalTitle,
			Category:   name,
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
		slug := strings.Join(parts[1:len(parts) - timeIndex], " ")
		name := util.Deslug(slug, " / ")

		layout = filepath.Join(meta.LayoutsDir, meta.CategoryIndexLayout)

		data = struct{
			Title      string
			Category   string
			Posts      []post.Post
			Categories []category.Category
		}{
			Title:      journalTitle,
			Category:   name,
			Posts:      posts,
			Categories: categories,
		}

		if err := writeIndexFile(layout, index, data); err != nil {
			errs <- err
			return
		}

		return
	}

	errs <- errors.New("could not match pattern to dir " + dir)
}

func writeIndexes(wg *sync.WaitGroup) chan error {
	errs := make(chan error)

	for dir, posts := range postIndexes {
		wg.Add(1)

		go writeIndex(dir, posts, wg, errs)
	}

	go func() {
		wg.Wait()

		close(errs)
	}()

	return errs
}

func writeIndexFile(layout, index string, data interface{}) error {
	if layout == "" {
		return errors.New("no layout for index " + index)
	}

	if data == nil {
		return errors.New("no data for index " + index)
	}

	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return err
	}

	dst, err := os.OpenFile(index, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		return err
	}

	defer dst.Close()

	t, err := template.New(index, string(b), data)

	if err != nil {
		return err
	}

	if err = t.Execute(dst, data); err != nil {
		return err
	}

	return nil
}

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

func publishToRemote(remote meta.Remote) {
	// Cheap way of checking for a local path over an SSH host.
	if filepath.IsAbs(remote.Target) {
		if err := util.Copy(meta.SiteDir, remote.Target); err != nil {
			util.Error("failed to publish to " + remote.Target, err)
		}
	}

	parts := strings.Split(remote.Target, "@")
	index := 0

	user := os.Getenv("USER")

	if len(parts) > 1 {
		user = parts[0]
		index = 1
	}

	parts = strings.Split(parts[index], ":")

	hostname := parts[0]
	dir := parts[1]

	addr := fmt.Sprintf("%s:%d", hostname, remote.Port)

	if remote.Identity == "" {
		remote.Identity = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}

	key, err := ioutil.ReadFile(remote.Identity)

	if err != nil {
		util.Error("failed to get identity file", err)
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		util.Error("failed to parse private key", err)
	}

	hostKey, err := getHostKey(hostname)

	if err != nil {
		util.Error("failed to find host key in known_hosts", err)
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
		util.Error("failed to establish initial connection", err)
	}

	defer conn.Close()

	scp, err := sftp.NewClient(conn)

	if err != nil {
		util.Error("failed to establish remote connection", err)
	}

	defer scp.Close()

	if err := util.CopyToRemote(meta.SiteDir, dir, scp); err != nil {
		util.Error("failed to copy to remote", err)
	}
}

func Publish(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Publish)
		return
	}

	mustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	m.Close()

	journalTitle = m.Title

	tmp, err := category.ResolveCategories()

	if err != nil {
		util.Error("failed to resolve categories", err)
	}

	categories = tmp

	posts, err := post.ResolvePosts()

	if err != nil {
		util.Error("failed to resolve posts", err)
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
					fmt.Fprintf(os.Stderr, "jrnl: %s\n", err)
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

	errs = writeIndexes(wg)

	for err := range errs {
		code = 1

		fmt.Fprintf(os.Stderr, "jrnl: %s\n", err)
	}

	if !c.Flags.IsSet("draft") && code == 0 {
		alias := c.Flags.GetString("remote")

		if alias == "" {
			alias = m.Default
		}

		if alias == "" {
			util.Error("missing remote", nil)
		}

		remote, ok := m.Remotes[alias]

		if !ok {
			util.Error("remote '" + alias + "' does not exist", nil)
		}

		publishToRemote(remote)
	}

	os.Exit(code)
}
