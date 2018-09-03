package category

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
)

var (
	errInvalid = errors.New("invalid category")

	pattern = "[_site]/[-a-zA-Z0-9/]+"

	regex = regexp.MustCompile(pattern)
)

type Category struct {
	ID string

	Name string

	Categories []Category
}

func ResolveCategories() ([]Category, error) {
	categories := make(map[string]*Category)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == meta.PostsDir || !info.IsDir() {
			return nil
		}

		parts := strings.Split(path, string(os.PathSeparator))

		id := filepath.Join(parts[1:]...)

		c := &Category{
			ID:         id,
			Name:       util.Deslug(strings.Join(parts[1:], " "), " / "),
			Categories: make([]Category, 0),
		}

		if len(parts) > 2 {
			parentId := filepath.Join(parts[1:len(parts) - 1]...)

			parent, ok := categories[parentId]

			if !ok {
				return errors.New("no parent found for " + path)
			}

			c.Name = util.Deslug(parts[len(parts) - 1], " / ")

			parent.Categories = append(parent.Categories, *c)

			return nil
		}

		categories[id] = c

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

func PrintCategories(categories []Category, item, nested string) string {
	buf := bytes.Buffer{}

	for _, c := range categories {
		buf.WriteString("<" + item + ">" + c.Name)

		if len(c.Categories) > 0 {
			buf.WriteString("<" + nested + ">")
			buf.WriteString(PrintCategories(c.Categories, item, nested))
			buf.WriteString("</" + nested + ">")
		}

		buf.WriteString("</" + item + ">")
	}

	return buf.String()
}

func PrintHrefCategories(categories []Category, item, nested string) string {
	buf := bytes.Buffer{}

	for _, c := range categories {
		link := "<a href=\"" + c.Href() + "\">" + c.Name + "</a>"
		buf.WriteString("<" + item + ">" + link)

		if len(c.Categories) > 0 {
			buf.WriteString("<" + nested + ">")
			buf.WriteString(PrintHrefCategories(c.Categories, item, nested))
			buf.WriteString("</" + nested + ">")
		}

		buf.WriteString("</" + item + ">")
	}

	return buf.String()
}

func (c Category) Href() string {
	return "/" + c.ID
}
