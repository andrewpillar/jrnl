package post

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"

	"github.com/mozillazg/go-slugify"
)

type Post struct {
	ID string

	Category string

	Title string

	SourcePath string

	SitePath string
}

func New(title, category string) *Post {
	categorySlug := bytes.Buffer{}

	parts := strings.Split(category, "/")

	for _, p := range parts {
		categorySlug.WriteString(slugify.Slugify(p) + string(os.PathSeparator))
	}

	titleSlug := slugify.Slugify(title)

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
			category.WriteString(util.Ucfirst(p))

			if i != len(tmp) - 1 {
				category.WriteString(" ")
			}
		}
	}

	title := util.Deslug(strings.Split(filepath.Base(sourcePath), ".")[0])

	return &Post{
		ID:         id,
		Category:   category.String(),
		Title:      title,
		SourcePath: sourcePath,
	}, nil
}

func (p *Post) Remove() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(meta.PostsDir, filepath.Dir(p.SourcePath))
}
