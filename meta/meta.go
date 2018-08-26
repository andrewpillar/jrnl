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

	IndexLayout = "index.html"

	DayIndexLayout = "day_index.html"

	MonthIndexLayout = "month_index.html"

	YearIndexLayout = "year_index.html"

	CategoryIndexLayout = "category_index.html"

	PostLayout = "post.html"

	Layouts = map[string]string{
		IndexLayout:         index,
		DayIndexLayout:      dayIndex,
		MonthIndexLayout:    monthIndex,
		YearIndexLayout:     yearIndex,
		CategoryIndexLayout: categoryIndex,
		PostLayout:          post,
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
