package category

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
)

type Category struct {
	ID string

	Name string

	Categories []Category
}

func Find(id string) (Category, error) {
	sourcePath := filepath.Join(meta.PostsDir, id)

	_, err := os.Stat(sourcePath)

	if err != nil {
		return Category{}, err
	}

	parts := strings.Split(util.Deslug(id), string(os.PathSeparator))
	name := bytes.Buffer{}

	for i, p := range parts {
		name.WriteString(util.Title(p))

		if i != len(parts) - 1 {
			name.WriteString(" / ")
		}
	}

	return Category{
		ID:         id,
		Name:       name.String(),
		Categories: make([]Category, 0),
	}, nil
}

func All() ([]Category, error) {
	categories := make(map[string]*Category)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == meta.PostsDir || !info.IsDir() {
			return nil
		}

		id := strings.Replace(path, meta.PostsDir + string(os.PathSeparator), "", 1)
		parts := strings.Split(id, string(os.PathSeparator))

		c, err := Find(id)

		if err != nil {
			return err
		}

		if len(parts) >= 2 {
			parentId := filepath.Join(parts[:len(parts) - 1]...)

			parent, ok := categories[parentId]

			if !ok {
				return errors.New("no parent found for " + path)
			}

			parent.Categories = append(parent.Categories, c)

			return nil
		}

		categories[id] = &c

		return nil
	}

	err := filepath.Walk(meta.PostsDir, walk)

	ret := make([]Category, len(categories), len(categories))
	i := 0

	for _, c := range categories {
		ret[i] = *c
		i++
	}

	return ret, err
}

func (c Category) Href() string {
	return "/" + c.ID
}
