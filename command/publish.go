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
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/template"
	"github.com/andrewpillar/jrnl/util"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

var (
	yearPattern  = "[" + meta.SiteDir + "]/[0-9]{4}"
	monthPattern = yearPattern + "/[0-9]{2}"
	dayPattern   = monthPattern + "/[0-9]{2}"

	categoryPattern      = "[" + meta.SiteDir + "]/[-a-z0-9/]+"
	categoryYearPattern  = categoryPattern + "/[0-9]{4}"
	categoryMonthPattern = categoryYearPattern + "/[0-9]{2}"
	categoryDayPattern   = categoryMonthPattern + "/[0-9]{2}"

	yearRegex  = regexp.MustCompile(yearPattern)
	monthRegex = regexp.MustCompile(monthPattern)
	dayRegex   = regexp.MustCompile(dayPattern)

	categoryRegex      = regexp.MustCompile(categoryPattern)
	categoryYearRegex  = regexp.MustCompile(categoryYearPattern)
	categoryMonthRegex = regexp.MustCompile(categoryMonthPattern)
	categoryDayRegex   = regexp.MustCompile(categoryDayPattern)
)

type indexData struct {
	Title      string
	Categories []category.Category
	Pages      []page.Page
	Posts      []post.Post
}

type timeIndexData struct {
	indexData

	Time time.Time
}

type categoryTimeIndexData struct {
	timeIndexData

	Category category.Category
}

type categoryIndexData struct {
	indexData

	Category category.Category
}

func indexPost(p post.Post, indexes map[string][]post.Post, wg *sync.WaitGroup, m *sync.Mutex) {
	defer wg.Done()

	if !p.Index {
		return
	}

	parts := strings.Split(filepath.Dir(p.SitePath), string(os.PathSeparator))

	for i := range parts {
		key := filepath.Join(parts[:len(parts) - i - 1]...)

		if key == "" {
			continue
		}

		m.Lock()
		indexes[key] = append(indexes[key], p)
		m.Unlock()
	}
}

func publishIndex(m *meta.Meta, key string, posts []post.Post, categories []category.Category, pages []page.Page) error {
	data := indexData{
		Title:      m.Title,
		Categories: categories,
		Pages:      pages,
		Posts:      posts,
	}

	index := filepath.Join(key, "index.html")
	layout := ""

	if key == meta.SiteDir {
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.Index)

		return writeIndexFile(layout, index, data)
	}

	b := []byte(key)
	parts := strings.Split(key, string(os.PathSeparator))

	isDateIndex := false
	timeFmt := ""
	i := 0

	if dayRegex.Match(b) {
		isDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.Day)
		timeFmt = filepath.Join("2006", "01", "02")
		i = 3
	} else if monthRegex.Match(b) {
		isDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.Month)
		timeFmt = filepath.Join("2006", "01")
		i = 2
	} else if yearRegex.Match(b) {
		isDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.Year)
		timeFmt = filepath.Join("2006", "01")
		i = 1
	}

	if isDateIndex {
		t, err := time.Parse(timeFmt, filepath.Join(parts[len(parts) - i:]...))

		if err != nil {
			return err
		}

		timeIndexData := timeIndexData{
			indexData: data,
			Time:      t,
		}

		return writeIndexFile(layout, index, timeIndexData)
	}

	isCategoryDateIndex := false

	if categoryDayRegex.Match(b) {
		isCategoryDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.CategoryDay)
		timeFmt = filepath.Join("2006", "01", "02")
		i = 3
	} else if categoryMonthRegex.Match(b) {
		isCategoryDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.CategoryMonth)
		timeFmt = filepath.Join("2006", "01")
		i = 2
	} else if categoryYearRegex.Match(b) {
		isCategoryDateIndex = true
		layout = filepath.Join(meta.LayoutsDir, m.IndexLayouts.CategoryYear)
		timeFmt = filepath.Join("2006")
		i = 1
	}

	if isCategoryDateIndex {
		id := strings.Join(parts[1:len(parts) - i], string(os.PathSeparator))

		c, err := category.Find(id)

		if err != nil {
			return err
		}

		t, err := time.Parse(timeFmt, filepath.Join(parts[len(parts) - i:]...))

		if err != nil {
			return err
		}

		categoryTimeIndexData := categoryTimeIndexData{
			timeIndexData: timeIndexData{
				indexData: data,
				Time:      t,
			},
			Category: c,
		}

		return writeIndexFile(layout, index, categoryTimeIndexData)
	}

	if categoryRegex.Match(b) {
		id := strings.Join(parts[1:len(parts) - i], string(os.PathSeparator))

		c, err := category.Find(id)

		if err != nil {
			return err
		}

		categoryIndexData := categoryIndexData{
			indexData: data,
			Category:  c,
		}

		return writeIndexFile(layout, index, categoryIndexData)
	}

	return errors.New("could not match pattern to dir: " + key)
}

func publishPage(p page.Page, data interface{}) error {
	if p.Layout == "" {
		return errors.New("no layout for " + p.ID)
	}

	b, err := ioutil.ReadFile(filepath.Join(meta.LayoutsDir, p.Layout))

	if err != nil {
		return err
	}

	dir := filepath.Dir(p.SitePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(p.SitePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		return err
	}

	defer f.Close()

	return template.Render(f, p.ID, string(b), data)
}

func publishPages(title string, categories []category.Category, pages []page.Page) (chan page.Page, chan error) {
	published := make(chan page.Page)
	errs := make(chan error)

	wg := &sync.WaitGroup{}

	err := page.Walk(func(p page.Page) error {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			p.Render()

			data := struct{
				Title      string
				Categories []category.Category
				Pages      []page.Page
				Page       page.Page
			}{
				Title:      title,
				Categories: categories,
				Pages:      pages,
				Page:       p,
			}

			if err := publishPage(p, data); err != nil {
				errs <- err
				return
			}

			published <- p
		}()

		return nil
	})

	if err != nil {
		go func() {
			errs <- err
		}()
	}

	go func() {
		wg.Wait()

		close(published)
		close(errs)
	}()

	return published, errs
}

func publishPosts(title string, categories []category.Category, pages []page.Page) (chan post.Post, chan error) {
	published := make(chan post.Post)
	errs := make(chan error)

	wg := &sync.WaitGroup{}

	err := post.Walk(func(p post.Post) error {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			p.Render()

			data := struct{
				Title      string
				Categories []category.Category
				Pages      []page.Page
				Post       post.Post
			}{
				Title:      title,
				Categories: categories,
				Pages:      pages,
				Post:       p,
			}

			if err := publishPage(p.Page, data); err != nil {
				errs <- err
				return
			}

			published <- p
		}()

		return nil
	})

	if err != nil {
		go func() {
			errs <- err
		}()
	}

	go func() {
		wg.Wait()

		close(published)
		close(errs)
	}()

	return published, errs
}

func publishToRemote(r meta.Remote) error {
	if filepath.IsAbs(r.Target) {
		if err := util.Copy(meta.SiteDir, r.Target); err != nil {
			return err
		}
	}

	parts := strings.Split(r.Target, "@")
	i := 0

	user := os.Getenv("USER")

	if len(parts) > 1 {
		user = parts[0]
		i = 1
	}

	parts = strings.Split(parts[i], ":")

	hostname := parts[0]
	dir := parts[1]

	addr := fmt.Sprintf("%s:%d", hostname, r.Port)

	if r.Identity == "" {
		r.Identity = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}

	key, err := ioutil.ReadFile(r.Identity)

	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		return err
	}

	hostKey, err := util.GetHostKey(hostname)

	if err != nil {
		return err
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
		return err
	}

	defer conn.Close()

	scp, err := sftp.NewClient(conn)

	if err != nil {
		return err
	}

	defer scp.Close()

	if err := util.CopyToRemote(meta.SiteDir, dir, scp); err != nil {
		return err
	}

	return nil
}

func writeIndexFile(layout, fname string, data interface{}) error {
	if data == nil {
		return errors.New("no data for index: " + fname)
	}

	if layout == meta.LayoutsDir || layout == "" {
		return errors.New("no layout for index: " + fname)
	}

	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return err
	}

	f, err := os.OpenFile(fname, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		return err
	}

	defer f.Close()

	return template.Render(f, fname, string(b), data)
}

func Publish(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	categories, err := category.All()

	if err != nil {
		util.Exit("failed to get all categories", err)
	}

	pages, err := page.All()

	if err != nil {
		util.Exit("failed to get all categories", err)
	}

	code := 0

	pagesCh, errs := publishPages(m.Title, categories, pages)

	for {
		select {
			case err, ok := <-errs:
				if !ok {
					errs = nil
				} else {
					code = 1
					fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
				}
			case _, ok := <-pagesCh:
				if !ok {
					pagesCh = nil
				}
		}

		if pagesCh == nil && errs == nil {
			break
		}
	}

	postsCh, errs := publishPosts(m.Title, categories, pages)

	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}

	indexes := make(map[string][]post.Post)

	for {
		select {
			case err, ok := <-errs:
				if !ok {
					errs = nil
				} else {
					code = 1
					fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
				}
			case p, ok := <-postsCh:
				if !ok {
					postsCh = nil
				} else {
					wg.Add(1)

					go indexPost(p, indexes, wg, mut)
				}
		}

		if postsCh == nil && errs == nil {
			break
		}
	}

	wg.Wait()

	errs = make(chan error)

	for key, posts := range indexes {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := publishIndex(m, key, posts, categories, pages); err != nil {
				errs <- err
				return
			}
		}()
	}

	go func() {
		wg.Wait()

		close(errs)
	}()

	for err := range errs {
		code = 1

		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}

	if !c.Flags.IsSet("draft") {
		remote := c.Flags.GetString("remote")

		if remote == "" {
			remote = m.Default
		}

		if remote == "" {
			util.Exit("failed to get remote", errors.New("no default set"))
		}

		r, ok := m.Remotes[remote]

		if !ok {
			util.Exit("failed to find remote", errors.New(remote))
		}

		if err := publishToRemote(r); err != nil {
			util.Exit("failed to publish to remote", err)
		}
	}

	os.Exit(code)
}
