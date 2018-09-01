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
	"text/template"
	"time"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
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

func indexPosts(posts <-chan post.Post) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	m := &sync.Mutex{}

	for p := range posts {
		wg.Add(1)

		go indexPost(p, wg, m)
	}

	return wg
}

func publishPost(
	p post.Post,
	wg *sync.WaitGroup,
	published chan post.Post,
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

	if err = p.Publish(journalTitle, string(b)); err != nil {
		errs <- err
		return
	}

	published <- p
}

func publishPosts(posts []post.Post, errs chan<- error) chan post.Post {
	wg := &sync.WaitGroup{}
	published := make(chan post.Post)

	for _, p := range posts {
		wg.Add(1)

		go publishPost(p, wg, published, errs)
	}

	go func() {
		wg.Wait()
		close(published)
	}()

	return published
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

func writeIndexes(wg *sync.WaitGroup, errs chan error) {
	for dir, posts := range postIndexes {
		wg.Add(1)

		go writeIndex(dir, posts, wg, errs)
	}
}

func writeIndexFile(layout, index string, data interface{}) error {
	if layout == "" {
		return errors.New("no layout for index " + index)
	}

	if data == nil {
		return errors.New("no data for index " + index)
	}

	src, err := os.Open(layout)

	if err != nil {
		return err
	}

	defer src.Close()

	dst, err := os.OpenFile(index, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		return err
	}

	defer dst.Close()

	b, err := ioutil.ReadAll(src)

	if err != nil {
		return err
	}

	t, err := template.New("index").Parse(string(b))

	if err != nil {
		return err
	}

	if err = t.Execute(dst, data); err != nil {
		return err
	}

	return nil
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

	errs := make(chan error)

	published := publishPosts(posts, errs)

	wg := indexPosts(published)
	wg.Wait()

	writeIndexes(wg, errs)

	go func() {
		wg.Wait()
		close(errs)
	}()

	code := 0

	for err := range errs {
		code = 1

		fmt.Fprintf(os.Stderr, "jrnl: %s\n", err)
	}

	os.Exit(code)
}
