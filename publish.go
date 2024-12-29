package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/feeds"

	"github.com/mmcdole/gofeed"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

var PublishCmd = &Command{
	Usage: "publish [-d] [-v]",
	Short: "publish the jrnl",
	Long: `Publish will generate the HTML for each modified page and post in the _site
directory. The contents of the _site directory will then be copied to the
configured remote.

The -d flag will not copy the contents of the _site directory to the configured
remote.

The -v flag will print out the site paths that have been created.`,
	Run: publishCmd,
}

func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"partial": tmplFuncPartial,
	}
}

func tmplFuncPartial(name string, data any) (string, error) {
	b, err := os.ReadFile(filepath.Join(layoutDir, name))

	if err != nil {
		return "", err
	}

	tmpl := template.New(name)
	tmpl.Funcs(templateFunctions())

	if _, err := tmpl.Parse(string(b)); err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func templatePage(p SitePage, data *PageData) (string, error) {
	sitePath := filepath.Join(siteDir, p.URL(), "index.html")

	if err := os.MkdirAll(filepath.Dir(sitePath), dirMode); err != nil {
		return "", err
	}

	f, err := os.Create(sitePath)

	if err != nil {
		return "", err
	}

	defer f.Close()

	layout := p.Layout()

	b, err := os.ReadFile(filepath.Join(layoutDir, layout))

	if err != nil {
		return "", err
	}

	tmpl := template.New(filepath.Join(layoutDir, p.Layout()))
	tmpl.Funcs(templateFunctions())

	if _, err := tmpl.Parse(string(b)); err != nil {
		return "", err
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepQuotes:       true,
	})

	pr, pw := io.Pipe()
	defer pr.Close()

	errCh := make(chan error)
	done := make(chan struct{})

	go func() {
		defer pw.Close()

		if err := tmpl.Execute(pw, data); err != nil {
			errCh <- err
			return
		}
	}()

	go func() {
		if err := m.Minify("text/html", f, pr); err != nil {
			errCh <- err
			return
		}
		done <- struct{}{}
	}()

	select {
	case err := <-errCh:
		return "", err
	case <-done:
	}
	return sitePath, nil
}

type homePage struct{}

func (p homePage) URL() string          { return "/" }
func (p homePage) Title() string        { return "Home" }
func (p homePage) Content() string      { return "" }
func (p homePage) Description() string  { return "" }
func (p homePage) Layout() string       { return "home" }
func (p homePage) CreatedAt() time.Time { return time.Now() }

func publishFeeds(cfg *Config, author *feeds.Author, items []*feeds.Item) error {
	feed := feeds.Feed{
		Title: cfg.Site.Title,
		Link: &feeds.Link{
			Href: cfg.Site.URL,
		},
		Description: cfg.Site.Description,
		Author:      author,
		Updated:     time.Now(),
		Items:       items,
	}

	if cfg.Site.Atom != "" {
		f, err := os.Create(filepath.Join(siteDir, cfg.Site.Atom))

		if err != nil {
			return err
		}

		defer f.Close()

		if err := feed.WriteAtom(f); err != nil {
			return err
		}
	}

	if cfg.Site.RSS != "" {
		f, err := os.Create(filepath.Join(siteDir, cfg.Site.RSS))

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

type PageData struct {
	Site     Site
	Author   Author
	Title    string
	Content  string
	Posts    []SitePage
	Blogroll []SitePage
}

type feedPage struct {
	item *gofeed.Item
}

func (p feedPage) URL() string          { return p.item.Link }
func (p feedPage) Title() string        { return p.item.Title }
func (p feedPage) Description() string  { return p.item.Description }
func (p feedPage) Content() string      { return "" }
func (p feedPage) Layout() string       { return "" }
func (p feedPage) CreatedAt() time.Time { return *p.item.PublishedParsed }

func publishCmd(cmd *Command, args []string) error {
	if err := Initialized("."); err != nil {
		return err
	}

	var draft, verbose bool

	flags := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	flags.BoolVar(&draft, "d", false, "don't upload jrnl to the remote")
	flags.BoolVar(&verbose, "v", false, "print out the files that will be uploaded")
	flags.Parse(args)

	pages := make([]SitePage, 0)
	pages = append(pages, homePage{})

	cfg, err := LoadConfig()

	if err != nil {
		return err
	}

	posts := make([]SitePage, 0)

	walk := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		p, err := LoadPage(path)

		if err != nil {
			return nil
		}

		// This is a post, not a page, so treat it as such.
		if strings.HasPrefix(path, postDir) {
			post := &Post{
				Page: p,
			}

			pages = append(pages, post)
			posts = append(posts, post)
			return nil
		}

		pages = append(pages, p)
		return nil
	}

	if err := filepath.Walk(pageDir, walk); err != nil {
		return err
	}

	if err := filepath.Walk(postDir, walk); err != nil {
		return err
	}

	p := gofeed.NewParser()

	blogroll := make([]SitePage, 0, len(cfg.Site.Blogroll))

	for _, url := range cfg.Site.Blogroll {
		feed, err := p.ParseURL(url)

		if err != nil {
			return err
		}

		if len(feed.Items) > 0 {
			blogroll = append(blogroll, feedPage{
				item: feed.Items[0],
			})
		}
	}

	items := make([]*feeds.Item, 0, len(pages))

	sem := make(chan struct{}, runtime.GOMAXPROCS(0)+10)
	errs := make(chan error)

	var wg sync.WaitGroup

	author := feeds.Author{
		Name:  cfg.Author.Name,
		Email: cfg.Author.Email,
	}

	for _, p := range pages {
		sem <- struct{}{}
		wg.Add(1)

		if _, ok := p.(*Post); ok {
			items = append(items, &feeds.Item{
				Title: p.Title(),
				Link: &feeds.Link{
					Href: cfg.Site.URL + p.URL(),
				},
				Description: p.Description(),
				Author:      &author,
				Created:     p.CreatedAt(),
			})
		}

		go func(p SitePage) {
			defer func() {
				wg.Done()
				<-sem
			}()

			data := PageData{
				Site:     cfg.Site,
				Author:   cfg.Author,
				Title:    p.Title(),
				Content:  p.Content(),
				Posts:    posts,
				Blogroll: blogroll,
			}

			sitePath, err := templatePage(p, &data)

			if err != nil {
				errs <- err
				return
			}

			if verbose {
				fmt.Println("built", sitePath)
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	erred := false

	for err := range errs {
		fmt.Fprintln(os.Stderr, err)
		erred = true
	}

	if erred {
		return errors.New("could not build jrnl")
	}

	if err := publishFeeds(cfg, &author, items); err != nil {
		return err
	}

	if !draft {
		sitePaths := make([]string, 0)

		err := filepath.Walk(siteDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			sitePaths = append(sitePaths, path)
			return nil
		})

		if err != nil {
			return err
		}

		remote, err := ParseRemote(cfg.Remote)

		if err != nil {
			return err
		}

		errs = make(chan error)

		for _, path := range sitePaths {
			sem <- struct{}{}
			wg.Add(1)

			go func(path string) {
				defer func() {
					wg.Done()
					<-sem
				}()

				b, err := CacheGet(path)

				if err != nil {
					if !errors.Is(err, fs.ErrNotExist) {
						errs <- err
						return
					}
				}

				hash, err := CacheHash(path)

				if err != nil {
					errs <- err
					return
				}

				if !bytes.Equal(b, hash) {
					remotePath := strings.TrimPrefix(path, siteDir)

					if err := remote.Sync(path, remotePath); err != nil {
						errs <- err
						return
					}

					if verbose {
						fmt.Println("published", path)
					}

					if err := CachePut(path); err != nil {
						errs <- err
						return
					}
				}
			}(path)
		}

		go func() {
			wg.Wait()
			close(errs)
		}()

		for err := range errs {
			fmt.Fprintln(os.Stderr, err)
			erred = true
		}

		if erred {
			return errors.New("failed to publish jrnl")
		}
	}
	return nil
}
