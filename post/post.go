package post

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"

	"github.com/russross/blackfriday"

	"gopkg.in/yaml.v2"
)

var (
	DateLayout = "2006-01-02T15:04"

	frontMatterFmt = `---
title: %s
index: true
createdAt: %s
updatedAt: %s
---`
)

type frontMatter struct {
	Title string

	Layout string

	Index bool

	CreatedAt string `yaml:"createdAt"`

	UpdatedAt string `yaml:"updatedAt"`
}

type Post struct {
	ID string

	Category category.Category

	Title string

	Layout string

	Index bool

	Preview string

	Body string

	SourcePath string

	SitePath string

	CreatedAt time.Time

	UpdatedAt time.Time
}

func Find(id string) (Post, error) {
	sourcePath := filepath.Join(meta.PostsDir, id + ".md")

	_, err := os.Stat(sourcePath)

	if err != nil {
		return Post{}, err
	}

	parts := strings.Split(id, string(os.PathSeparator))

	buf := bytes.Buffer{}

	categoryParts := []string{}
	categoryId := ""

	if len(parts) >= 2 {
		categoryParts = parts[:len(parts) - 1]
		categoryId = filepath.Join(categoryParts...)

		for i, p := range categoryParts {
			buf.WriteString(p)

			if i != len(categoryParts) - 1 {
				buf.WriteString(" / ")
			}
		}
	}

	categoryName := buf.String()

	f, err := os.Open(sourcePath)

	if err != nil {
		return Post{}, err
	}

	defer f.Close()

	fm, err := unmarshalFrontMatter(f)

	if err != nil {
		return Post{}, err
	}

	createdAt, err := time.Parse(DateLayout, fm.CreatedAt)

	if err != nil {
		return Post{}, err
	}

	createdAtStr := string([]rune(fm.CreatedAt)[:10])

	updatedAt, err := time.Parse(DateLayout, fm.UpdatedAt)

	if err != nil {
		return Post{}, err
	}

	sitePath := filepath.Join(
		meta.SiteDir,
		filepath.Join(categoryParts...),
		filepath.Join(strings.Split(createdAtStr, "-")...),
		filepath.Base(id),
		"index.html",
	)

	return Post{
		ID:         id,
		Category:   category.Category{
			ID:     categoryId,
			Name:   categoryName,
		},
		Title:      fm.Title,
		Layout:     fm.Layout,
		Index:      fm.Index,
		SourcePath: sourcePath,
		SitePath:   sitePath,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func New(title, categoryName string) Post {
	date := time.Now()

	buf := bytes.Buffer{}

	parts := strings.Split(categoryName, "/")

	for i, p := range parts {
		buf.WriteString(util.Slug(p))

		if i != len(parts) - 1 {
			buf.WriteString(string(os.PathSeparator))
		}
	}

	categoryId := buf.String()
	titleSlug := util.Slug(title)

	id := filepath.Join(categoryId, titleSlug)
	sourcePath := filepath.Join(meta.PostsDir, categoryId, titleSlug + ".md")

	return Post{
		ID:         id,
		Category:   category.Category{
			ID:     categoryId,
			Name:   categoryName,
		},
		Title:      title,
		SourcePath: sourcePath,
		CreatedAt:  date,
		UpdatedAt:  date,
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

		root := meta.PostsDir + string(os.PathSeparator)
		id := strings.Replace(path, root, "", 1)

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

func ResolvePosts() ([]Post, error) {
	posts := make(map[string]Post)
	order := make([]string, 0)

	err := Walk(func(p Post) error {
		createdAt := p.CreatedAt.Format(DateLayout)

		posts[createdAt] = p

		order = append(order, createdAt)

		return nil
	})

	if err != nil {
		return []Post{}, err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(order)))

	ret := make([]Post, len(posts), len(posts))

	for i, key := range order {
		ret[i] = posts[key]
	}

	return ret, err
}

func marshalFrontMatter(fm *frontMatter, w io.Writer) error {
	enc := yaml.NewEncoder(w)

	if err := enc.Encode(fm); err != nil {
		return err
	}

	return nil
}

func unmarshalFrontMatter(r io.Reader) (frontMatter, error) {
	fm := frontMatter{}

	buf := bytes.Buffer{}
	tmp := make([]byte, 1)

	bounds := 0

	for {
		if bounds == 2 {
			break
		}

		_, err := r.Read(tmp)

		if err != nil {
			if err == io.EOF {
				break
			}

			return fm, err
		}

		buf.Write(tmp)

		for tmp[0] == '-' {
			_, err = r.Read(tmp)

			if err != nil {
				if err == io.EOF {
					break
				}

				return fm, err
			}

			buf.Write(tmp)

			if tmp[0] == '\n' {
				bounds++
				break
			}
		}
	}

	dec := yaml.NewDecoder(&buf)

	if err := dec.Decode(&fm); err != nil {
		return fm, err
	}

	return fm, nil
}

func (p *Post) Convert() {
	p.Body = string(blackfriday.Run([]byte(p.Body)))
	p.Preview = string(blackfriday.Run([]byte(p.Preview)))
}

func (p Post) HasCategory() bool {
	return p.Category.ID != "" && p.Category.Name != ""
}

func (p Post) Href() string {
	href := []rune(p.SitePath)

	return filepath.Dir(string(href[len(meta.SiteDir):]))
}

func (p *Post) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	fm, err := unmarshalFrontMatter(f)

	if err != nil {
		return err
	}

	p.Title = fm.Title
	p.Layout = fm.Layout
	p.Index = fm.Index

	createdAtTime, err := time.Parse(DateLayout, fm.CreatedAt)

	if err != nil {
		return err
	}

	updatedAtTime, err := time.Parse(DateLayout, fm.UpdatedAt)

	if err != nil {
		return err
	}

	p.CreatedAt = createdAtTime
	p.UpdatedAt = updatedAtTime

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	if len(b) > 2 {
		i := bytes.IndexByte(b, '\n')

		p.Preview = string(b[:i])
	}

	p.Body = string(b)

	return nil
}

func (p Post) Remove() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(meta.PostsDir, filepath.Dir(p.SourcePath))
}

func (p *Post) Touch() error {
	if err := p.Load(); err != nil {
		return err
	}

	p.UpdatedAt = time.Now()

	f, err := os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR, os.ModePerm)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := frontMatter{
		Title:     p.Title,
		Layout:    p.Layout,
		Index:     p.Index,
		CreatedAt: p.CreatedAt.Format(DateLayout),
		UpdatedAt: p.UpdatedAt.Format(DateLayout),
	}

	f.Write([]byte("---\n"))

	if err := marshalFrontMatter(&fm, f); err != nil {
		return err
	}

	f.Write([]byte("---\n"))
	f.Write([]byte(p.Body))

	return nil
}

func (p Post) WriteFrontMatter(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		frontMatterFmt,
		p.Title,
		p.CreatedAt.Format(DateLayout),
		p.UpdatedAt.Format(DateLayout),
	)

	return err
}
