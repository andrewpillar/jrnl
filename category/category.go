package category

import (
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
}

func Find(id string) (Category, error) {
	sourcePath := filepath.Join(meta.PostsDir, id)

	if !regex.Match([]byte(sourcePath)) {
		return Category{}, errInvalid
	}

	_, err := os.Stat(sourcePath)

	if err != nil {
		return Category{}, err
	}

	name := strings.Replace(id, string(os.PathSeparator), " ", -1)

	return Category{
		ID:   id,
		Name: util.Deslug(name, " / "),
	}, nil
}

func ResolveCategories() ([]Category, error) {
	categories := make([]Category, 0)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == meta.PostsDir || !info.IsDir() {
			return nil
		}

		id := strings.Replace(
			path,
			meta.PostsDir + string(os.PathSeparator),
			"",
			1,
		)

		c, err := Find(id)

		if err != nil {
			if err == errInvalid {
				return nil
			}

			return err
		}

		categories = append(categories, c)

		return nil
	}

	err := filepath.Walk(meta.PostsDir, walk)

	return categories, err
}
