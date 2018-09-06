package post

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/template"
	"github.com/andrewpillar/jrnl/util"

	"github.com/mozillazg/go-slugify"

	"gopkg.in/russross/blackfriday.v2"
)

var dateSlug = "2006-01-02T15:04"

type Post struct {
	ID string

	Category category.Category

	Title string

	Preview string

	Body string

	SourcePath string

	SitePath string

	CreatedAt time.Time
}

func New(title, category string) Post {
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

	return Post{
		ID:         id,
		Title:      title,
		SourcePath: sourcePath,
		CreatedAt:  createdAt,
	}
}

func Find(id string) (Post, error) {
	sourcePath := filepath.Join(meta.PostsDir, id + ".md")

	_, err := os.Stat(sourcePath)

	if err != nil {
		return Post{}, err
	}

	parts := strings.Split(id, string(os.PathSeparator))

	categoryParts := []string{}
	categoryId := ""

	if len(parts) >= 2 {
		categoryParts = parts[:len(parts) - 1]
		categoryId = filepath.Join(categoryParts...)
	}

	postCategory := category.Category{}

	if categoryId != "" {
		postCategory, err = category.Find(categoryId)

		if err != nil {
			return Post{}, err
		}
	}

	titleSlug := []rune(filepath.Base(sourcePath))

	createdAt, err := time.Parse(dateSlug, string(titleSlug[:len(dateSlug)]))

	if err != nil {
		return Post{}, err
	}

	createdAtSlug := []rune(createdAt.Format(dateSlug))

	title := util.Deslug(
		string(titleSlug[len(dateSlug) + 1:len(titleSlug) - 3]), " ",
	)

	sitePath := filepath.Join(
		meta.SiteDir,
		filepath.Join(categoryParts...),
		filepath.Join(strings.Split(string(createdAtSlug[:10]), "-")...),
		string(titleSlug[len(dateSlug) + 1:len(titleSlug) - 3]),
		"index.html",
	)

	return Post{
		ID:         id,
		Category:   postCategory,
		Title:      title,
		SourcePath: sourcePath,
		SitePath:   sitePath,
		CreatedAt:  createdAt,
	}, nil
}

func ResolvePosts() ([]Post, error) {
	posts := make([]Post, 0)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == meta.PostsDir || info.IsDir() {
			return nil
		}

		id := strings.Replace(
			path,
			meta.PostsDir + string(os.PathSeparator),
			"",
			1,
		)

		p, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		posts = append(posts, p)

		return nil
	}

	err := filepath.Walk(meta.PostsDir, walk)

	return posts, err
}

func (p *Post) Convert() {
	body := blackfriday.Run([]byte(p.Body))
	preview := blackfriday.Run([]byte(p.Preview))

	p.Body = string(body)
	p.Preview = string(preview)
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

func (p Post) Publish(
	title string,
	layout string,
	categories []category.Category,
) error {
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

	page := struct{
		Title      string
		Post       Post
		Categories []category.Category
	}{
		Title:      title,
		Post:       p,
		Categories: categories,
	}

	t, err := template.New("post", layout, page)

	if err != nil {
		return err
	}

	return t.Execute(f, page)
}

func (p *Post) Remove() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(meta.PostsDir, filepath.Dir(p.SourcePath))
}
