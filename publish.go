package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/gorilla/feeds"

	"github.com/grokify/html-strip-tags-go"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

type publishError struct {
	id   string
	kind string
	err  error
}

type Directory string

type Site struct {
	Title       string
	Description string
	Link        string
	Categories  []*Category
	Pages       []*Page
	Author      struct {
		Name  string
		Email string
	}
}

var PublishCmd = &Command{
	Usage: "publish [options]",
	Short: "publish the journal to the remote",
	Long:  ``,
	Run:   publishCmd,
}

func previewPost(id string, md goldmark.Markdown, buf *bytes.Buffer) (*Post, bool, error) {
	buf.Reset()

	p, ok, err := GetPost(id)

	if err != nil {
		return nil, false, err
	}

	if !ok {
		return nil, false, nil
	}

	if err := md.Convert([]byte(p.Description), buf); err != nil {
		return nil, false, err
	}

	p.Description = buf.String()
	return p, true, nil
}

func publishCategoryIndex(s Site, layout []byte, categoryidx map[string]*Index) ([]string, error) {
	posts := make([]*Post, 0)
	paths := make([]string, 0, len(categoryidx))

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	var buf bytes.Buffer

	for id, index := range categoryidx {
		var walkerr error

		posts = posts[0:0]

		cat, ok, err := GetCategory(id)

		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		index.Walk(func(id string) {
			p, ok, err := previewPost(id, md, &buf)

			if err != nil {
				walkerr = err
				return
			}

			if !ok {
				return
			}
			posts = append(posts, p)
		})

		if walkerr != nil {
			return nil, walkerr
		}

		data := struct {
			Site     Site
			Category *Category
			Posts    []*Post
		}{
			Site:     s,
			Category: cat,
			Posts:    posts,
		}

		err = func(id string) error {
			path := filepath.Join(siteDir, id, "index.html")

			f, err := os.Create(path)

			if err != nil {
				return err
			}

			defer f.Close()

			paths = append(paths, path)

			return executeTemplate(f, id + "-index", string(layout), data)
		}(id)

		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func publishFeed(s Site, index *Index, atom, rss string) error {
	items := make([]*feeds.Item, 0)

	author := &feeds.Author{
		Name:  s.Author.Name,
		Email: s.Author.Email,
	}

	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	var walkerr error

	index.Walk(func(id string) {
		p, ok, err := previewPost(id, md, &buf)

		if err != nil {
			walkerr = err
			return
		}

		if !ok {
			return
		}

		items = append(items, &feeds.Item{
			Title: s.Title,
			Link: &feeds.Link{
				Href: s.Link + p.Href(),
			},
			Description: strip.StripTags(p.Description),
			Author:      author,
			Created:     p.CreatedAt.Time,
		})
	})

	feed := &feeds.Feed{
		Title: s.Title,
		Link:  &feeds.Link{
			Href: s.Link,
		},
		Description: s.Description,
		Author:      author,
		Items:       items,
	}

	if atom != "" {
		f, err := os.Create(atom)

		if err != nil {
			return err
		}
		defer f.Close()

		if err :=  feed.WriteAtom(f); err != nil {
			return err
		}
	}

	if rss != "" {
		f, err := os.Create(rss)

		if err != nil {
			return err
		}
		defer f.Close()

		if err := feed.WriteRss(f); err != nil {
			return err
		}
	}
	return nil
}

func publishSiteIndex(s Site, layout []byte, index *Index) (string, error) {
	var walkerr error

	posts := make([]*Post, 0)

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	var buf bytes.Buffer

	index.Walk(func(id string) {
		p, ok, err := previewPost(id, md, &buf)

		if err != nil {
			walkerr = err
			return
		}

		if !ok {
			return
		}
		posts = append(posts, p)
	})

	if walkerr != nil {
		return "", walkerr
	}

	data := struct {
		Site  Site
		Posts []*Post
	}{
		Site:  s,
		Posts: posts,
	}

	path := filepath.Join(siteDir, "index.html")

	f, err := os.Create(path)

	if err != nil {
		return "", err
	}

	defer f.Close()

	if err := executeTemplate(f, "site-index", string(layout), data); err != nil {
		return "", err
	}
	return path, nil
}

func publishPages(s Site) (chan *Page, chan error) {
	pages := make(chan *Page)
	errs := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(s.Pages))

	for _, p := range s.Pages {
		go func(p *Page) {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			if err := p.Publish(s); err != nil {
				errs <- publishError{
					id:   p.ID,
					kind: "page",
					err:  err,
				}
				return
			}
			pages <- p
		}(p)
	}

	go func() {
		wg.Wait()
		close(pages)
		close(errs)
	}()

	return pages, errs
}

func publishPosts(s Site, set map[string]struct{}) (chan *Post, chan error) {
	sem := make(chan struct{}, runtime.GOMAXPROCS(0) + 10)

	posts := make(chan *Post)
	errs := make(chan error)

	var wg sync.WaitGroup

	WalkPosts(func(p *Post) error {
		if _, ok := set[p.ID]; !ok {
			return nil
		}

		wg.Add(1)

		go func() {
			sem <- struct{}{}
			defer func(){ <-sem }()
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			if err := p.Publish(s); err != nil {
				errs <- publishError{
					id:   p.ID,
					kind: "post",
					err:  err,
				}
				return
			}
			posts <- p
		}()
		return nil
	})

	go func() {
		wg.Wait()
		close(posts)
		close(errs)
	}()

	return posts, errs
}

func (d Directory) Hash() []byte {
	sha256 := sha256.New()

	filepath.Walk(string(d), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		defer f.Close()

		_, err = io.Copy(sha256, f)
		return err
	})
	return sha256.Sum(nil)
}

func (e publishError) Error() string {
	return "failed to publish " + e.kind + " " + e.id + ": " + e.err.Error()
}

func publishCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	var (
		atom    string
		draft   bool
		rss     string
		verbose bool
	)

	fs := flag.NewFlagSet(cmd.Argv0 + " " + args[0], flag.ExitOnError)
	fs.StringVar(&atom, "a", "", "the file to write the Atom feed to")
	fs.BoolVar(&draft, "d", false, "only publish the HTML, don't copy to the remote")
	fs.StringVar(&rss, "r", "", "the file to write the RSS feed to")
	fs.BoolVar(&verbose, "v", false, "display the files copied to the remote")
	fs.Parse(args[1:])

	cfg, err := OpenConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open config: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	categories, err := Categories()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to get all categories: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	s := Site{
		Title:       cfg.Site.Title,
		Description: cfg.Site.Description,
		Link:        cfg.Site.Link,
		Categories:  categories,
		Pages:       make([]*Page, 0),
	}
	s.Author.Name = cfg.Author.Name
	s.Author.Email = cfg.Author.Email

	hash, err := OpenHash()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open hash: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	defer hash.Close()

	err = WalkPages(func(p *Page) error {
		s.Pages = append(s.Pages, p)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed walk pages: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	index := NewIndex()
	categoryidx := make(map[string]*Index)

	for _, cat := range categories {
		categoryidx[cat.ID] = NewIndex()
	}

	postset := make(map[string]struct{}, 0)

	err = WalkPosts(func(p *Post) error {
		index.Put(p)

		if id := p.Category.ID; id != "" {
			categoryidx[id].Put(p)
		}

		if hash.Put(p.ID, p) {
			postset[p.ID] = struct{}{}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed walk posts: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	paths := make([]string, 0)

	if hash.Put(assetsDir, Directory(assetsDir)) {
		err := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				paths = append(paths, path)
			}
			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to walk assets directory: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
	}

	if err := publishFeed(s, index, atom, rss); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to publish feed: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if atom != "" {
		paths = append(paths, atom)
	}

	if rss != "" {
		paths = append(paths, rss)
	}

	code := 0

	pages, errs := publishPages(s)

	for pages != nil && errs != nil {
		select {
		case p, ok := <-pages:
			if !ok {
				pages = nil
				break
			}

			if hash.Put(p.ID, p) {
				paths = append(paths, p.SitePath)
			}
		case err, ok := <-errs:
			if !ok {
				errs = nil
				break
			}
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		}
	}

	posts, errs := publishPosts(s, postset)

	for posts != nil && errs != nil {
		select {
		case p, ok := <-posts:
			if !ok {
				posts = nil
				break
			}
			paths = append(paths, p.SitePath)
		case err, ok := <-errs:
			if !ok {
				errs = nil
				break
			}
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		}
	}

	layout, err := ioutil.ReadFile(filepath.Join(layoutsDir, "index"))

	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s %s: failed read site index layout: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
	}

	if len(layout) > 0 {
		path, err := publishSiteIndex(s, layout, index)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed publish site index: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
		paths = append(paths, path)
	}

	layout, err = ioutil.ReadFile(filepath.Join(layoutsDir, "category-index"))

	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s %s: failed read category index layout: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
	}

	if len(layout) > 0 {
		catpaths, err := publishCategoryIndex(s, layout, categoryidx)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed publish category index: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
		paths = append(paths, catpaths...)
	}

	if code == 0 {
		if err := hash.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to save hash: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
	}

	if draft {
		fmt.Println("published draft to", siteDir)
		os.Exit(code)
	}

	if cfg.Site.Remote == "" {
		fmt.Fprintf(os.Stderr, "%s %s: remote not set, set with '%s config site.remote'\n", cmd.Argv0, args[0], cmd.Argv0)
		os.Exit(1)
	}

	rem, err := OpenRemote(cfg.Site.Remote)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open remote: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	defer rem.Close()

	if verbose {
		fmt.Println("publishing to remote", cfg.Site.Remote)
	}

	for _, path := range paths {
		if verbose {
			fmt.Println(path)
		}

		if err := rem.Copy(path); err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to copy %q to remote: %s\n", cmd.Argv0, args[0], path, err)
			code = 1
		}
	}
	os.Exit(code)
}
