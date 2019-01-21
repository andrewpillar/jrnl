package category

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/util"
)

type Category struct {
	ID         string
	Name       string
	Categories []Category
}

func All() ([]Category, error) {
	categories := make(map[string]*Category)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == config.PostsDir || !info.IsDir() || strings.Contains(path, config.IndexDir) {
			return nil
		}

		id := strings.Replace(path, config.PostsDir + string(os.PathSeparator), "", 1)
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

	err := filepath.Walk(config.PostsDir, walk)

	ret := make([]Category, len(categories), len(categories))
	i := 0

	for _, c := range categories {
		ret[i] = *c
		i++
	}

	return ret, err
}

func Find(id string) (Category, error) {
	path := filepath.Join(config.PostsDir, id)

	_, err := os.Stat(path)

	if err != nil {
		return Category{}, err
	}

	parts := strings.Split(util.Deslug(id), string(os.PathSeparator))
	name := bytes.Buffer{}
	end := len(parts) - 1

	for i, p := range parts {
		name.WriteString(strings.Title(p))

		if i != end {
			name.WriteString(" / ")
		}
	}

	return Category{
		ID:         id,
		Name:       name.String(),
		Categories: make([]Category, 0),
	}, nil
}

func (c Category) Href() string {
	return ""
}

func (c Category) IsZero() bool {
	return	c.ID == ""   &&
			c.Name == "" &&
			len(c.Categories) == 0
}
