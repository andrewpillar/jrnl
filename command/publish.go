package command

import (
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
	"github.com/andrewpillar/jrnl/util"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

var (
	yearPattern = "[" + meta.SiteDir + "]/[0-9]{4}"

	monthPattern = yearPattern + "/[0-9]{2}"

	dayPattern = monthPattern + "/[0-9]{2}"

	categoryPattern = "[" + meta.SiteDir + "]/[-a-z0-9/]+"

	categoryYearPattern = categoryPattern + "/[0-9]{4}"

	categoryMonthPattern = categoryYearPattern + "/[0-9]{2}"

	categoryDayPattern = categoryMonthPattern + "/[0-9]{2}"

	yearRegex = regexp.MustCompile(yearPattern)

	monthRegex = regexp.MustCompile(monthPattern)

	dayRegex = regexp.MustCompile(dayPattern)

	categoryRegex = regexp.MustCompile(categoryPattern)

	categoryYearRegex = regexp.MustCompile(categoryYearPattern)

	categoryMonthRegex = regexp.MustCompile(categoryMonthPattern)

	categoryDayRegex = regexp.MustCompile(categoryDayPattern)
)

type journal struct {
	meta meta.Meta

	mutex *sync.Mutex

	// Posts will be indexed under the directories they're stored, e.g.
	//     /2006/01/02  -> [post1]
	//     /2006/01     -> [post1, post2]
	//     /2006        -> [post1, post2, post3]
	indexes map[string][]post.Post

	Title      string
	Categories []category.Category
}

type postPage struct {
	journal

	Post post.Post
}

type indexPage struct {
	journal

	Posts      []post.Post
	Categories []category.Category
}

type timeIndexPage struct {
	indexPage

	Time time.Time
}

type categoryTimeIndexPage struct {
	timeIndexPage

	Category category.Category
}

type categoryIndexPage struct {
	indexPage

	Category category.Category
}

func (j *journal) index(p post.Post) {
	if !p.Index {
		return
	}

	parts := strings.Split(filepath.Dir(p.SitePath), string(os.PathSeparator))

	for i := range parts {
		key := filepath.Join(parts[:len(parts) - i - 1]...)

		if key == "" {
			continue
		}

		j.mutex.Lock()
		j.indexes[key] = append(j.indexes[key], p)
		j.mutex.Unlock()
	}
}

func (j journal) publish(p *post.Post) error {
	if p.Layout == "" {
		return errors.New("no layout for post " + p.ID)
	}

	if err := p.Load(); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(filepath.Join(meta.LayoutsDir, p.Layout))

	if err != nil {
		return err
	}

	p.Convert()

	dir := filepath.Dir(p.SitePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(p.SitePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		return err
	}

	defer f.Close()

	page := postPage{
		journal: j,
		Post:    *p,
	}

	return template.Render(f, "post-" + p.ID, string(b), page)
}

func (j journal) writeIndex(key string) error {
	posts, ok := j.indexes[key]

	if !ok {
		return errors.New("no posts for index " + key)
	}

	indexData := indexPage{
		journal:    j,
		Posts:      posts,
		Categories: j.Categories,
	}

	index := filepath.Join(key, "index.html")
	layout := ""

	if key == meta.SiteDir {
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.Index)

		return j.writeIndexFile(layout, index, indexData)
	}

	pattern := []byte(key)
	parts := strings.Split(key, string(os.PathSeparator))

	isDateIndex := false
	timeFmt := ""
	i := 0

	if dayRegex.Match(pattern) {
		isDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.Day)
		timeFmt = filepath.Join("2006", "01", "02")
		i = 3
	} else if monthRegex.Match(pattern) {
		isDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.Month)
		timeFmt = filepath.Join("2006", "01")
		i = 2
	} else if yearRegex.Match(pattern) {
		isDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.Year)
		timeFmt = filepath.Join("2006")
		i = 1
	}

	if isDateIndex {
		t, err := time.Parse(timeFmt, filepath.Join(parts[len(parts) - i:]...))

		if err != nil {
			return err
		}

		timeIndexData := timeIndexPage{
			indexPage: indexData,
			Time:      t,
		}

		return j.writeIndexFile(layout, index, timeIndexData)
	}

	isCategoryDateIndex := false

	if categoryDayRegex.Match(pattern) {
		isCategoryDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.CategoryDay)
		timeFmt = filepath.Join("2006", "01", "02")
		i = 3
	} else if categoryMonthRegex.Match(pattern) {
		isCategoryDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.CategoryMonth)
		timeFmt = filepath.Join("2006", "01")
		i = 2
	} else if categoryYearRegex.Match(pattern) {
		isCategoryDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, j.meta.IndexLayouts.CategoryYear)
		timeFmt = filepath.Join("2006")
		i = 1
	}

	if isCategoryDateIndex {
		id := strings.Join(parts[1:len(parts) - i], " ")

		c, err := category.Find(id)

		if err != nil {
			return err
		}

		t, err := time.Parse(timeFmt, filepath.Join(parts[len(parts) - i:]...))

		if err != nil {
			return err
		}

		timeIndexData := timeIndexPage{
			indexPage: indexData,
			Time:      t,
		}

		categoryTimeIndexData := categoryTimeIndexPage{
			timeIndexPage: timeIndexData,
			Category:      c,
		}

		return j.writeIndexFile(layout, index, categoryTimeIndexData)
	}

	if categoryRegex.Match(pattern) {
		id := strings.Join(parts[1:len(parts) - i], " ")

		c, err := category.Find(id)

		if err != nil {
			return err
		}

		categoryIndexData := categoryIndexPage{
			indexPage: indexData,
			Category:  c,
		}

		return j.writeIndexFile(layout, index, categoryIndexData)
	}

	return errors.New("could not match pattern to dir " + key)
}

func (j journal) writeIndexFile(layout, index string, data interface{}) error {
	if data == nil {
		return errors.New("no data for index " + index)
	}

	if layout == meta.LayoutsDir || layout == "" {
		return errors.New("no layout for index " + index)
	}

	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return err
	}

	dst, err := os.OpenFile(index, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		return err
	}

	defer dst.Close()

	return template.Render(dst, index, string(b), data)
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

	hostKey, err := util.GetHostKey(hostname)

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
		util.Exit("failed to establish connection", err)
	}

	defer scp.Close()

	if err := util.CopyToRemote(meta.SiteDir, dir, scp); err != nil {
		util.Exit("failed to copy to remote", err)
	}
}

func Publish(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	categories, err := category.ResolveCategories()

	if err != nil {
		util.Exit("failed to resolve all categories", err)
	}

	j := journal{
		meta:       *m,
		mutex:      &sync.Mutex{},
		indexes:    make(map[string][]post.Post),
		Title:      m.Title,
		Categories: categories,
	}

	code := 0

	wg := &sync.WaitGroup{}

	published := make(chan post.Post)
	errs := make(chan error)

	post.Walk(func(p post.Post) error {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := j.publish(&p); err != nil {
				errs <- err
				return
			}

			published <- p
		}()

		return nil
	})

	go func() {
		wg.Wait()

		close(published)
		close(errs)
	}()

	indexWg := &sync.WaitGroup{}

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
					indexWg.Add(1)

					go func() {
						defer indexWg.Done()

						j.index(p)
					}()
				}
		}

		if errs == nil && published == nil {
			break
		}
	}

	indexWg.Wait()

	errs = make(chan error)

	for key := range j.indexes {
		indexWg.Add(1)

		go func() {
			defer indexWg.Done()

			if err := j.writeIndex(key); err != nil {
				errs <- err
			}
		}()
	}

	go func() {
		indexWg.Wait()

		close(errs)
	}()

	for err := range errs {
		code = 1

		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}

	if !c.Flags.IsSet("draft") {
		remoteStr := c.Flags.GetString("remote")

		if remoteStr == "" {
			remoteStr = m.Default
		}

		if remoteStr == "" {
			util.Exit("missing remote", errors.New("no default set"))
		}

		remote, ok := m.Remotes[remoteStr]

		if !ok {
			util.Exit("failed to find remote", errors.New(remoteStr))
		}

		publishToRemote(remote)
	}

	os.Exit(code)
}
