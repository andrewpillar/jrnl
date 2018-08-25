package meta

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	File = "_meta.yml"

	PostsDir string

	SiteDir string

	LayoutsDir string

	AssetsDir string

	Dirs []string

	Layouts = map[string]string{
		"index.html":          index,
		"day_index.html":      dayIndex,
		"month_index.html":    monthIndex,
		"year_index.html":     yearIndex,
		"category_index.html": categoryIndex,
		"post.html":           post,
	}
)

type Meta struct {
	Title string

	Default string

	Remotes map[string]Remote
}

type Remote struct {
	Target string

	Port int

	Identity string
}

func Decode(r io.Reader) (*Meta, error) {
	m := &Meta{}

	dec := yaml.NewDecoder(r)

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

func Init(dir string) (*Meta, error) {
	fname := filepath.Join(dir, File)

	f, err := os.OpenFile(fname, os.O_CREATE, os.ModePerm)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return &Meta{}, nil
}

func (m *Meta) Encode(w io.Writer) error {
	enc := yaml.NewEncoder(w)

	if err := enc.Encode(m); err != nil {
		return err
	}

	defer enc.Close()

	return nil
}
