package post

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/util"

	"github.com/russross/blackfriday"
)

var dateLayout = "2006-01-02T15:04"

type frontMatter struct {
	Title     string
	Index     bool
	Layout    string
	CreatedAt string `yaml:"createdAt"`
	UpdatedAt string `yaml:"updatedAt"`
}

type Post struct {
	page.Page

	Index     bool
	Preview   string
	Category  category.Category
	CreatedAt time.Time
	UpdatedAt time.Time
}

func All() ([]Post, error) {
	posts := make([]Post, 0)

	err := Walk(func(p Post) error {
		posts = append(posts, p)

		return nil
	})

	return posts, err
}

func Find(id string) (Post, error) {
	fname := filepath.Join(meta.PostsDir, id + ".md")

	_, err := os.Stat(fname)

	if err != nil {
		return Post{}, err
	}

	f, err := os.Open(fname)

	if err != nil {
		return Post{}, err
	}

	defer f.Close()

	parts := strings.Split(id, string(os.PathSeparator))

	categoryParts := []string{}
	postCategory := category.Category{}

	if len(parts) >= 2 {
		categoryParts = parts[:len(parts) - 1]
		categoryId := filepath.Join(categoryParts...)

		postCategory, err = category.Find(categoryId)
	}

	fm := frontMatter{}

	if err := util.UnmarshalFrontMatter(&fm, f); err != nil {
		return Post{}, err
	}

	createdAt, err := time.Parse(dateLayout, fm.CreatedAt)

	if err != nil {
		return Post{}, err
	}

	updatedAt, err := time.Parse(dateLayout, fm.UpdatedAt)

	date := string([]rune(fm.CreatedAt)[:10])

	if err != nil {
		return Post{}, err
	}

	categoryPath := filepath.Join(categoryParts...)
	datePath := filepath.Join(strings.Split(date, "-")...)

	site := filepath.Join(meta.SiteDir, categoryPath, datePath, filepath.Base(id), "index.html")

	return Post{
		Page: page.Page{
			ID:         id,
			Title:      fm.Title,
			Layout:     fm.Layout,
			SourcePath: filepath.Join(meta.PostsDir, id + ".md"),
			SitePath:   site,
		},
		Category:  postCategory,
		Index:     fm.Index,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func New(p *page.Page, category_ string) Post {
	now := time.Now()

	buf := bytes.Buffer{}
	parts := strings.Split(category_, "/")

	for i, prt := range parts {
		buf.WriteString(util.Slug(prt))

		if i != len(parts) - 1 {
			buf.WriteString(string(os.PathSeparator))
		}
	}

	categoryId := buf.String()
	titleSlug := util.Slug(p.Title)

	id := filepath.Join(categoryId, titleSlug)

	p.ID = id
	p.SourcePath = filepath.Join(meta.PostsDir, id + ".md")

	return Post{
		Page:     *p,
		Category: category.Category{
			ID:   categoryId,
			Name: category_,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func Walk(fn func(p Post) error) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == meta.PostsDir || info.IsDir() {
			return nil
		}

		id := strings.Replace(path, meta.PostsDir + string(os.PathSeparator), "", 1)

		p, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		if err := fn(p); err != nil {
			return err
		}

		return nil
	}

	return filepath.Walk(meta.PostsDir, walk)
}

func (p Post) HasCategory() bool {
	return p.Category.ID != "" && p.Category.Name != ""
}

func (p Post) Href() string {
	r := []rune(p.SitePath)

	return filepath.Dir(string(r[len(meta.SiteDir):]))
}

func (p *Post) Load() error {
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

	if len(b) > 2 {
		i := bytes.IndexByte(b, '\n')

		p.Preview = string(b[:i])
	}

	createdAt, err := time.Parse(dateLayout, fm.CreatedAt)

	if err != nil {
		return err
	}

	updatedAt, err := time.Parse(dateLayout, fm.UpdatedAt)

	if err != nil {
		return err
	}

	p.Title = fm.Title
	p.Body = string(b)
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return nil
}

func (p *Post) Render() {
	p.Page.Render()
	p.Preview = string(blackfriday.Run([]byte(p.Preview)))
}

func (p *Post) Remove() error {
	if err := p.Remove(); err != nil {
		return err
	}

	if err := util.RemoveEmptyDirs(meta.PostsDir, filepath.Dir(p.SourcePath)); err != nil {
		return err
	}

	p.Index = false
	p.Preview = ""
	p.Category = category.Category{}
	p.CreatedAt = time.Time{}
	p.UpdatedAt = time.Time{}

	return nil
}

func (p *Post) Touch() error {
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

	if err := os.MkdirAll(filepath.Dir(p.SourcePath), os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := frontMatter{
		Title:     p.Title,
		Index:     p.Index,
		Layout:    p.Layout,
		CreatedAt: p.CreatedAt.Format(dateLayout),
		UpdatedAt: time.Now().Format(dateLayout),
	}

	f.Write([]byte("---\n"))

	if err := util.MarshalFrontMatter(&fm, f); err != nil {
		return err
	}

	f.Write([]byte("---\n"))
	f.Write([]byte(p.Body))

	return nil
}
