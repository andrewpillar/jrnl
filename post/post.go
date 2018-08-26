package post

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"

	"github.com/mozillazg/go-slugify"

	"gopkg.in/russross/blackfriday.v2"
)

var (
	dateSlug = "2006-01-02T15:04"
)

type Post struct {
	ID string

	Category string

	Title string

	Preview string

	Body string

	SourcePath string

	SitePath string

	CreatedAt time.Time
}

func New(title, category string) *Post {
	createdAt := time.Now()

	categorySlug := bytes.Buffer{}

	parts := strings.Split(category, "/")

	for i, p := range parts {
		categorySlug.WriteString(slugify.Slugify(p))

		if i != len(parts) - 1 {
			categorySlug.WriteString(string(os.PathSeparator))
		}
	}

	titleSlug := createdAt.Format(dateSlug) + "-" + slugify.Slugify(title)

	id := filepath.Join(categorySlug.String(), titleSlug)
	sourcePath := filepath.Join(
		meta.PostsDir,
		categorySlug.String(),
		titleSlug + ".md",
	)

	return &Post{
		ID:         id,
		Category:   category,
		Title:      title,
		SourcePath: sourcePath,
		CreatedAt:  createdAt,
	}
}

func Find(id string) (*Post, error) {
	sourcePath := filepath.Join(meta.PostsDir, id + ".md")

	_, err := os.Stat(sourcePath)

	if err != nil {
		return nil, err
	}

	parts := strings.Split(sourcePath, string(os.PathSeparator))
	categoryParts := []string{}

	category := bytes.Buffer{}

	if len(parts) >= 3 {
		categoryParts = parts[1:len(parts) - 1]

		for i, p := range categoryParts {
			category.WriteString(util.Deslug(p))

			if i != len(categoryParts) - 1 {
				category.WriteString(" ")
			}
		}
	}

	titleSlug := []rune(filepath.Base(sourcePath))

	createdAt, err := time.Parse(dateSlug, string(titleSlug[:len(dateSlug)]))

	if err != nil {
		return nil, err
	}

	createdAtSlug := []rune(createdAt.Format(dateSlug))

	title := util.Deslug(
		string(titleSlug[len(dateSlug) + 1:len(titleSlug) - 3]),
	)

	sitePath := filepath.Join(
		meta.SiteDir,
		filepath.Join(categoryParts...),
		filepath.Join(strings.Split(string(createdAtSlug[:10]), "-")...),
		string(titleSlug[len(dateSlug) + 1:len(titleSlug) - 3]),
		"index.html",
	)

	return &Post{
		ID:         id,
		Category:   category.String(),
		Title:      title,
		SourcePath: sourcePath,
		SitePath:   sitePath,
		CreatedAt:  createdAt,
	}, nil
}

func (p *Post) Convert() {
	body := blackfriday.Run([]byte(p.Body))
	preview := blackfriday.Run([]byte(p.Preview))

	p.Body = string(body)
	p.Preview = string(preview)
}

func (p *Post) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	if len(b) > 2 {
		i := bytes.Index(b, []byte("\n"))

		p.Preview = string(b[:i])
	}

	p.Body = string(b)

	return nil
}

func (p Post) Publish(title, layout string) error {
	dir := filepath.Dir(p.SitePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(
		p.SitePath,
		os.O_TRUNC|os.O_RDWR|os.O_CREATE,
		os.ModePerm,
	)

	if err != nil {
		return err
	}

	defer f.Close()

	t, err := template.New("post").Parse(layout)

	if err != nil {
		return err
	}

	page := struct{
		Title string
		Post  Post
	}{
		Title: title,
		Post:  p,
	}

	return t.Execute(f, page)
}

func (p *Post) Remove() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(meta.PostsDir, filepath.Dir(p.SourcePath))
}
