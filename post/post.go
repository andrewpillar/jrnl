package post

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"

	"github.com/mozillazg/go-slugify"
)

var (
	dateSlug = "2006-01-02T15:04"
)

type Post struct {
	ID string

	Category string

	Title string

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

	category := bytes.Buffer{}

	if len(parts) >= 3 {
		tmp := parts[1:len(parts) - 1]

		for i, p := range tmp {
			category.WriteString(util.Deslug(p))

			if i != len(tmp) - 1 {
				category.WriteString(" ")
			}
		}
	}

	titleSlug := []rune(filepath.Base(sourcePath))

	createdAt, err := time.Parse(dateSlug, string(titleSlug[:len(dateSlug)]))

	if err != nil {
		return nil, err
	}

	title := util.Deslug(string(titleSlug[len(dateSlug) + 1:len(titleSlug) - 3]))

	return &Post{
		ID:         id,
		Category:   category.String(),
		Title:      title,
		SourcePath: sourcePath,
		CreatedAt:  createdAt,
	}, nil
}

func (p *Post) Remove() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(meta.PostsDir, filepath.Dir(p.SourcePath))
}
