package meta

import (
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	File = "_meta.yaml"

	PostsDir = "_posts"

	SiteDir = "_site"

	LayoutsDir = "_layouts"

	ThemesDir = "_themes"

	AssetsDir = filepath.Join(SiteDir, "assets")

	Dirs = []string{
		PostsDir,
		SiteDir,
		LayoutsDir,
		ThemesDir,
		AssetsDir,
	}
)

type Meta struct {
	*os.File `yaml:"-"`

	Title string

	Editor string

	Theme string

	Default string

	IndexLayout string

	DayIndexLayout string

	MonthIndexLayout string

	YearIndexLayout string

	CategoryIndexLayout string

	CategoryDayIndexLayout string

	CategoryMonthIndexLayout string

	CategoryYearIndexLayout string

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

	m := &Meta{
		File:   f,
		Editor: os.Getenv("EDITOR"),
	}

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

	m := &Meta{File: f}

	dec := yaml.NewDecoder(m)

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Meta) Save() error {
	info, err := m.Stat()

	if err != nil {
		return err
	}

	if info.Size() > 0 {
		if err := m.Truncate(0); err != nil {
			return err
		}
	}

	if _, err := m.Seek(0, io.SeekStart); err != nil {
		return err
	}

	enc := yaml.NewEncoder(m)

	if err := enc.Encode(m); err != nil {
		return err
	}

	enc.Close()

	return nil
}
