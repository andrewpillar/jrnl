package post

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/render"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/util"

	"github.com/russross/blackfriday"
)

var iso8601 = "2006-01-02T15:04"

type frontMatter struct {
	Title     string
	Index     bool
	Layout    string
	CreatedAt string `yaml:"createdAt"`
	UpdatedAt string `yaml:"updatedAt"`
}

type ByCreatedAt []*Post

type Post struct {
	*page.Page

	Index     bool
	Preview   string
	Category  category.Category
	CreatedAt time.Time
	UpdatedAt time.Time
}

func All() ([]*Post, error) {
	posts := make([]*Post, 0)

	err := Walk(func(p *Post) error {
		posts = append(posts, p)
		return nil
	})

	sort.Sort(ByCreatedAt(posts))

	return posts, err
}

func Find(id string) (*Post, error) {
	path := filepath.Join(config.PostsDir, id + ".md")

	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	parts := strings.Split(id, string(os.PathSeparator))

	categoryParts := []string{}
	postCategory := category.Category{}

	if len(parts) >= 2 {
		categoryParts = parts[:len(parts) - 1]
		categoryId := filepath.Join(categoryParts...)

		postCategory, err = category.Find(categoryId)

		if err != nil {
			return nil, err
		}
	}

	fm := &frontMatter{}

	if err := util.UnmarshalFrontMatter(fm, f); err != nil {
		return nil, err
	}

	createdAt, err := time.Parse(iso8601, fm.CreatedAt)

	if err != nil {
		return nil, err
	}

	updatedAt, err := time.Parse(iso8601, fm.UpdatedAt)

	if err != nil {
		return nil, err
	}

	date := string([]rune(fm.CreatedAt)[:10])

	categoryPath := filepath.Join(categoryParts...)
	datePath := filepath.Join(strings.Split(date, "-")...)

	return &Post{
		Page: &page.Page{
			ID:         id,
			Title:      fm.Title,
			Layout:     fm.Layout,
			SourcePath: filepath.Join(config.PostsDir, id + ".md"),
			SitePath:   filepath.Join(config.SiteDir, categoryPath, datePath, filepath.Base(id), "index.html"),
		},
		Category:  postCategory,
		Index:     fm.Index,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func New(p *page.Page, categoryName string) *Post {
	now := time.Now()

	buf := bytes.Buffer{}
	parts := strings.Split(categoryName, "/")
	end := len(parts) - 1

	for i, p := range parts {
		buf.WriteString(util.Slug(p))

		if i != end {
			buf.WriteString(string(os.PathSeparator))
		}
	}

	p.ID = filepath.Join(buf.String(), util.Slug(p.Title))
	p.SourcePath = filepath.Join(config.PostsDir, p.ID + ".md")

	return &Post{
		Page: p,
		Category: category.Category{
			ID:   buf.String(),
			Name: categoryName,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func Walk(fn func(p *Post) error) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.Contains(path, config.IndexDir) {
			return nil
		}

		id := strings.Replace(path, config.PostsDir + string(os.PathSeparator), "", 1)

		p, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		return fn(p)
	}

	return filepath.Walk(config.PostsDir, walk)
}

func (p ByCreatedAt) Len() int {
	return len(p)
}

func (p ByCreatedAt) Less(i, j int) bool {
	return p[i].CreatedAt.After(p[j].CreatedAt)
}

func (p ByCreatedAt) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *Post) HasCategory() bool {
	return p.Category.ID != "" && p.Category.Name != ""
}

func (p *Post) Load() error {
	if err := p.Page.Load(); err != nil {
		return err
	}

	b := []byte(p.Body)

	if len(b) > 2 {
		i := bytes.IndexByte(b, '\n')
		r := render.New()

		md := blackfriday.Run(b[:i], blackfriday.WithRenderer(r))

		p.Preview = string(md)
	}

	return nil
}

func (p *Post) Remove() error {
	if err := p.Page.Remove(); err != nil {
		return err
	}

	parts := strings.Split(filepath.Dir(p.SourcePath), string(os.PathSeparator))

	for i := range parts {
		dir := filepath.Join(parts[:len(parts) - i]...)

		if dir == config.PostsDir {
			break
		}

		f, err := os.Open(dir)

		if err != nil {
			return err
		}

		defer f.Close()

		if _, err := f.Readdirnames(1); err == io.EOF {
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Post) Touch() error {
	if err := os.MkdirAll(filepath.Dir(p.SourcePath), config.DirMode); err != nil {
		return err
	}

	f, err := p.Page.Open()

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &frontMatter{
		Title:     p.Title,
		Index:     p.Index,
		Layout:    p.Layout,
		CreatedAt: p.CreatedAt.Format(iso8601),
		UpdatedAt: time.Now().Format(iso8601),
	}

	if err := util.MarshalFrontMatter(fm, f); err != nil {
		return err
	}

	_, err = f.Write([]byte(p.Body))

	return err
}
