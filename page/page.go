package page

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/render"
	"github.com/andrewpillar/jrnl/util"

	"github.com/russross/blackfriday"
)

type frontMatter struct {
	Title  string
	Layout string
}

type Page struct {
	ID         string
	Title      string
	Layout     string
	SourcePath string
	SitePath   string
	Body       string
}

// Return all of the pages that can be found.
func All() ([]*Page, error) {
	pages := make([]*Page, 0)

	err := Walk(func(p *Page) error {
		pages = append(pages, p)

		return nil
	})

	return pages, err
}

// Find a page by the given id.
func Find(id string) (*Page, error) {
	return Resolve(filepath.Join(config.PagesDir, id + ".md"))
}

// Resolve a Page from the given filepath.
func Resolve(path string) (*Page, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	fm := &frontMatter{}

	if err := util.UnmarshalFrontMatter(fm, f); err != nil {
		return nil, err
	}

	id := strings.Split(filepath.Base(path), ".")[0]

	return &Page{
		ID:         id,
		Title:      fm.Title,
		Layout:     fm.Layout,
		SourcePath: path,
		SitePath:   filepath.Join(config.SiteDir, filepath.Base(id), "index.html"),
	}, nil
}

func New(title string) *Page {
	id := util.Slug(title)

	return &Page{
		ID:         id,
		Title:      title,
		SourcePath: filepath.Join(config.PagesDir, id + ".md"),
	}
}

// Walk over all of the pages in the _pages directory. Resolving each one we
// find, and passing it to the callback.
func Walk(fn func(p *Page) error) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == config.PagesDir || info.IsDir() {
			return nil
		}

		id := strings.Replace(path, config.PagesDir + string(os.PathSeparator), "", 1)

		p, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		return fn(p)
	}

	return filepath.Walk(config.PagesDir, walk)
}

func (p *Page) Href() string {
	r := []rune(p.SitePath)

	return filepath.Dir(string(r[len(config.SiteDir):]))
}

// This will parse the front matter from the page, and read in the rest of the
// file as the page's body.
func (p *Page) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &frontMatter{}

	if err := util.UnmarshalFrontMatter(fm, f); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	p.Title = fm.Title
	p.Layout = fm.Layout
	p.Body = string(b)

	return nil
}

// Open the underlying source file.
func (p *Page) Open() (*os.File, error) {
	info, err := os.Stat(p.SourcePath)

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		if err := p.Load(); err != nil {
			return nil, err
		}
	}

	if info != nil && info.IsDir() {
		return nil, errors.New("expected text file, got directory: " + p.SourcePath)
	}

	return os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, config.FileMode)
}

// Remove the underlying source file, and site path if it exists.
func (p *Page) Remove() error {
	if _, err := os.Stat(p.SitePath); err == nil {
		return os.Remove(p.SitePath)
	}

	return os.Remove(p.SourcePath)
}

// Convert the page's markdown to HTML.
func (p *Page) Render() {
	r := render.New()
	md := blackfriday.Run([]byte(p.Body), blackfriday.WithRenderer(r))

	p.Body = string(md)
}

func (p *Page) Touch() error {
	f, err := p.Open()

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &frontMatter{
		Title:  p.Title,
		Layout: p.Layout,
	}

	if err := util.MarshalFrontMatter(fm, f); err != nil {
		return err
	}

	_, err = f.Write([]byte(p.Body))

	return err
}
