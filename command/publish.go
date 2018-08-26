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

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
//	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

var (
	// Map used to store all of the different paths that can have posts indexed
	// under them. All the keys in the map will be different degraded paths
	// from the post's source, for example:
	//
	//   _site/2006/01/02/some-category
	//   _site/2006/01/02
	//   _site/2006/01
	//   _site/2006
	//   _site
	postIndexes map[string]*post.Store

	categoryPattern = "[-_a-zA-Z0-9]+/[-a-zA-Z0-9/]+/[0-9]{4}/[0-9]{2}/[0-9]{2}"

	dayPattern = "[-_a-zA-Z0-9]+/[0-9]{4}/[0-9]{2}/[0-9]{2}"

	monthPattern = "[-_a-zA-Z0-9]+/[0-9]{4}/[0-9]{2}"

	yearPattern = "[-_a-zA-Z0-9]+/[0-9]{4}"

	categoryRegex = regexp.MustCompile(categoryPattern)

	dayRegex = regexp.MustCompile(dayPattern)

	monthRegex = regexp.MustCompile(monthPattern)

	yearRegex = regexp.MustCompile(yearPattern)
)

func publishPost(
	p *post.Post,
	title string,
	wg *sync.WaitGroup,
	posts chan *post.Post,
	errs chan error,
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

	layout, err := ioutil.ReadAll(f)

	if err != nil {
		errs <- err
		return
	}

	p.Convert()

	if err := p.Publish(title, string(layout)); err != nil {
		errs <- err
		return
	}

	posts <- p
}

// Add the given post to all of the possible path indexes that are available.
func addPostToIndex(p *post.Post, wg *sync.WaitGroup, m *sync.Mutex) {
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
			s := post.NewStore()

			postIndexes[path] = &s
		}

		postIndexes[path].Put(p)

		m.Unlock()
	}
}

func writeIndex(dir, title string, posts *post.Store, wg *sync.WaitGroup, errs chan error) {
	defer wg.Done()

	indexPath := filepath.Join(dir, "index.html")

	layoutPath := ""

	var page interface{}

	pattern := []byte(dir)

	if dir == meta.SiteDir {
		layoutPath = filepath.Join(meta.LayoutsDir, meta.IndexLayout)

		page = struct{
			Title string
			Posts post.Store
		}{
			Title: title,
			Posts: *posts,
		}
	}

	if categoryRegex.Match(pattern) {
		layoutPath = filepath.Join(meta.LayoutsDir, meta.CategoryIndexLayout)

		page = struct{
			Title    string
			Category string
			Posts    post.Store
		}{
			Title:    title,
			Category: "category here",
			Posts:    *posts,
		}
	} else if dayRegex.Match(pattern) {
		layoutPath = filepath.Join(meta.LayoutsDir, meta.DayIndexLayout)

		page = struct{
			Title string
			Posts post.Store
		}{
			Title: title,
			Posts: *posts,
		}
	} else if monthRegex.Match(pattern) {
		layoutPath = filepath.Join(meta.LayoutsDir, meta.MonthIndexLayout)

		page = struct{
			Title string
			Posts post.Store
		}{
			Title: title,
			Posts: *posts,
		}
	} else if yearRegex.Match(pattern) {
		layoutPath = filepath.Join(meta.LayoutsDir, meta.YearIndexLayout)

		page = struct{
			Title string
			Posts post.Store
		}{
			Title: title,
			Posts: *posts,
		}
	}

	if layoutPath == "" {
		errs <- errors.New("layout path is empty for dir: " + dir)
		return
	}

	layoutFile, err := os.Open(layoutPath)

	if err != nil {
		errs <- err
		return
	}

	defer layoutFile.Close()

	layout, err := ioutil.ReadAll(layoutFile)

	if err != nil {
		errs <- err
		return
	}

	t, err := template.New("index").Parse(string(layout))

	if err != nil {
		errs <- err
		return
	}

	f, err := os.OpenFile(indexPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		errs <- err
		return
	}

	if err = t.Execute(f, page); err != nil {
		errs <- err
		return
	}
}

func Publish(c cli.Command) {
//	if c.Flags.IsSet("help") {
//		fmt.Println(usage.Publish)
//		return
//	}

	mustBeInitialized()

	f, err := os.Open(meta.File)

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	defer f.Close()

	m, err := meta.Decode(f)

	if err != nil {
		util.Error("failed to read meta file", err)
	}

	r := post.NewResolver()

	posts := r.Resolve()

	postIndexes = make(map[string]*post.Store)

	wg := &sync.WaitGroup{}

	published := make(chan *post.Post)
	errs := make(chan error)

	for _, p := range posts {
		wg.Add(1)

		go publishPost(p, m.Title, wg, published, errs)
	}

	go func() {
		wg.Wait()

		close(published)
	}()

	mut := &sync.Mutex{}

	for p := range published {
		wg.Add(1)

		go addPostToIndex(p, wg, mut)
	}

	for k, v := range postIndexes {
		wg.Add(1)

		go writeIndex(k, m.Title, v, wg, errs)
	}

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
