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

	PartialsDir string

	AssetsDir string

	ThemesDir string

	Dirs []string

	IndexLayout = "index.html"

	DayIndexLayout = "day_index.html"

	MonthIndexLayout = "month_index.html"

	YearIndexLayout = "year_index.html"

	CategoryIndexLayout = "category_index.html"

	CategoryDayIndexLayout = "category_day_index.html"

	CategoryMonthIndexLayout = "category_month_index.html"

	CategoryYearIndexLayout = "category_year_index.html"

	PostLayout = "post.html"

	Layouts = map[string]string{
		IndexLayout:              index,
		DayIndexLayout:           dayIndex,
		MonthIndexLayout:         monthIndex,
		YearIndexLayout:          yearIndex,
		CategoryIndexLayout:      categoryIndex,
		CategoryDayIndexLayout:   categoryDayIndex,
		CategoryMonthIndexLayout: categoryMonthIndex,
		CategoryYearIndexLayout:  categoryYearIndex,
		PostLayout:               post,
	}
)

type Meta struct {
	f *os.File `yaml:"-"`

	Title string

	Theme string

	Default string

	Remotes map[string]Remote
}

type Remote struct {
	Target string

	Port int

	Identity string
}

func Init(dir string) (*Meta, error) {
	f, err := os.OpenFile(filepath.Join(dir, File), os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	m := &Meta{f: f}

	if err := m.Save(); err != nil {
		return nil, err
	}

	return m, nil
}

func Open() (*Meta, error) {
	f, err := os.OpenFile(File, os.O_RDWR, 0660)

	if err != nil {
		return nil, err
	}

	m := &Meta{f: f}

	dec := yaml.NewDecoder(f)

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Meta) Save() error {
	info, err := m.f.Stat()

	if err != nil {
		return err
	}

	if info.Size() > 0 {
		if err := m.f.Truncate(0); err != nil {
			return err
		}
	}

	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	enc := yaml.NewEncoder(m.f)

	if err := enc.Encode(m); err != nil {
		return err
	}

	enc.Close()

	return nil
}

func (m *Meta) Close() error {
	return m.f.Close()
}
