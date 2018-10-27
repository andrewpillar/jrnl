package page

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
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

func All() ([]Page, error) {
	pages := make([]Page, 0)

	err := Walk(func(p Page) error {
		pages = append(pages, p)

		return nil
	})

	return pages, err
}

func Find(id string) (Page, error) {
	fname := filepath.Join(meta.PagesDir, id + ".md")

	_, err := os.Stat(fname)

	if err != nil {
		return Page{}, err
	}

	f, err := os.Open(fname)

	if err != nil {
		return Page{}, err
	}

	defer f.Close()

	fm := frontMatter{}

	if err := util.UnmarshalFrontMatter(&fm, f); err != nil {
		return Page{}, err
	}

	site := filepath.Join(meta.SiteDir, filepath.Base(id), "index.html")

	return Page{
		ID:         id,
		Title:      fm.Title,
		Layout:     fm.Layout,
		SourcePath: fname,
		SitePath:   site,
	}, nil
}

func New(title string) Page {
	id := util.Slug(title)

	return Page{
		ID:         id,
		Title:      title,
		SourcePath: filepath.Join(meta.PagesDir, id + ".md"),
	}
}

func Walk(fn func(p Page) error) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == meta.PagesDir || info.IsDir() {
			return nil
		}

		id := strings.Replace(path, meta.PagesDir + string(os.PathSeparator), "", 1)

		p, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		if err := fn(p); err != nil {
			return err
		}

		return nil
	}

	return filepath.Walk(meta.PagesDir, walk)
}

func (p *Page) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := frontMatter{}

	if err := util.UnmarshalFrontMatter(&fm, f); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	p.Title = fm.Title
	p.Body = string(b)

	return nil
}

func (p *Page) Render() {
	p.Body = string(blackfriday.Run([]byte(p.Body)))
}

func (p *Page) Remove() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	if err := os.Remove(p.SitePath); err != nil {
		return err
	}

	p.ID = ""
	p.Title = ""
	p.Layout = ""
	p.SourcePath = ""
	p.SitePath = ""
	p.Body = ""

	return nil
}

func (p *Page) Touch() error {
	info, err := os.Stat(p.SourcePath)

	if err == nil {
		if err := p.Load(); err != nil {
			return err
		}

		if info.IsDir() {
			return errors.New("expected file, got directory for: " + p.SourcePath)
		}
	} else {
		if !os.IsNotExist(err) {
			return err
		}
	}

	f, err := os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := frontMatter{
		Title:  p.Title,
		Layout: p.Layout,
	}

	f.Write([]byte("---\n"))

	if err := util.MarshalFrontMatter(&fm, f); err != nil {
		return err
	}

	f.Write([]byte("---\n"))
	f.Write([]byte(p.Body))

	return nil
}
