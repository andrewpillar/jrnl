package theme

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
)

type Theme struct {
	*os.File

	Name string
	Path string
}

func All() ([]*Theme, error) {
	themes := make([]*Theme, 0)

	err := Walk(func(t *Theme) error {
		themes = append(themes, t)

		return nil
	})

	return themes, err
}

func New(name string) (*Theme, error) {
	path := filepath.Join(meta.ThemesDir, name + ".tar.gz")

	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		return nil, err
	}

	return &Theme{
		File: f,
		Name: name,
		Path: path,
	}, nil
}

func Find(name string) (*Theme, error) {
	path := filepath.Join(meta.ThemesDir, name + ".tar.gz")

	f, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)

	if err != nil {
		return nil, err
	}

	return &Theme{
		File: f,
		Name: name,
		Path: path,
	}, nil
}

func Walk(fn func(t *Theme) error) error {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		id := strings.Replace(path, meta.ThemesDir + string(os.PathSeparator), "", 1)

		t, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		if err := fn(t); err != nil {
			return err
		}

		return nil
	}

	return filepath.Walk(meta.ThemesDir, walk)
}

func (t Theme) Save() error {
	m, err := meta.Open()

	if err != nil {
		return err
	}

	defer m.Close()

	m.Theme = t.Name

	assets := strings.Replace(meta.AssetsDir, meta.SiteDir, "", -1)
	assets = strings.TrimPrefix(assets, string(os.PathSeparator))

	path := filepath.Join(meta.ThemesDir, t.Name)
	tmp := filepath.Join(path, assets)

	if err := util.Copy(meta.AssetsDir, tmp); err != nil {
		return err
	}

	tmp = filepath.Join(path, meta.LayoutsDir)

	if err := util.Copy(meta.LayoutsDir, tmp); err != nil {
		return err
	}

	if err := tar(path, t); err != nil {
		return err
	}

	if err := os.RemoveAll(path); err != nil {
		return err
	}

	if err := m.Save(); err != nil {
		return err
	}

	return nil
}

func (t Theme) Use() error {
	m, err := meta.Open()

	if err != nil {
		return err
	}

	defer m.Close()

	m.Theme = t.Name

	if err := untar(meta.ThemesDir, t); err != nil {
		return err
	}

	assets := strings.Replace(meta.AssetsDir, meta.SiteDir, "", -1)
	assets = strings.TrimPrefix(assets, string(os.PathSeparator))

	tmp := filepath.Join(meta.ThemesDir, assets)

	if err := util.Copy(tmp, meta.AssetsDir); err != nil {
		return err
	}

	tmp = filepath.Join(meta.ThemesDir, meta.LayoutsDir)

	if err := util.Copy(tmp, meta.LayoutsDir); err != nil {
		return err
	}

	if err := m.Save(); err != nil {
		return err
	}

	return nil
}
