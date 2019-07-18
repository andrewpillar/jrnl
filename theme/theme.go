package theme

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/file"
)

type Theme struct {
	Name string
	Path string
}

func mktar(w io.Writer, src string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}

	gzw := gzip.NewWriter(w)

	defer gzw.Close()

	tw := tar.NewWriter(gzw)

	defer tw.Close()

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())

		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(
			strings.Replace(path, src, "", -1),
			string(os.PathSeparator),
		)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		defer f.Close()

		_, err = io.Copy(tw, f)

		return err
	}

	return filepath.Walk(src, walk)
}

func untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)

	if err != nil {
		return err
	}

	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
			case err == io.EOF:
				return nil
			case err != nil:
				return err
			case header == nil:
				continue
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
			case tar.TypeDir:
				_, err = os.Stat(target)

				if err != nil {
					if os.IsNotExist(err) {
						if err := os.MkdirAll(target, config.DirMode); err != nil {
							return err
						}

						continue
					}

					return err
				}
			case tar.TypeReg:
				flags := os.O_TRUNC|os.O_CREATE|os.O_RDWR

				f, err := os.OpenFile(target, flags, os.FileMode(header.Mode))

				if err != nil {
					return err
				}

				defer f.Close()

				if _, err = io.Copy(f, tr); err != nil {
					return err
				}
		}
	}
}

func All() ([]*Theme, error) {
	themes := make([]*Theme, 0)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		id := strings.Replace(path, config.ThemesDir + string(os.PathSeparator), "", 1)

		parts := strings.Split(id, ".")

		if parts[0] == "" {
			return nil
		}

		t, err := Find(parts[0])

		if err != nil {
			return err
		}

		themes = append(themes, t)

		return nil
	}

	err := filepath.Walk(config.ThemesDir, walk)

	return themes, err
}

func Find(name string) (*Theme, error) {
	path := filepath.Join(config.ThemesDir, name + ".tar.gz")

	_, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	return &Theme{
		Name:  name,
		Path:  path,
	}, nil
}

func New(name string) *Theme {
	return &Theme{
		Name:  name,
		Path:  filepath.Join(config.ThemesDir, name + ".tar.gz"),
	}
}

// Load the current theme. This will un-tar the current theme, and overwrite
// the existing _layouts, and _site/assets directories with the contents of the
// tarball.
func (t *Theme) Load() error {
	f, err := os.Open(filepath.Join(config.ThemesDir, t.Name + ".tar.gz"))

	if err != nil {
		return err
	}

	defer f.Close()

	if err := os.RemoveAll(config.LayoutsDir); err != nil {
		return err
	}

	if err := os.RemoveAll(config.AssetsDir); err != nil {
		return err
	}

	if err := untar(config.ThemesDir, f); err != nil {
		return err
	}

	assets := strings.Replace(filepath.Join(config.ThemesDir, config.AssetsDir), config.SiteDir, "", -1)
	layouts := filepath.Join(config.ThemesDir, config.LayoutsDir)

	if err := file.Copy(config.AssetsDir, assets); err != nil {
		return err
	}

	if err := file.Copy(config.LayoutsDir, layouts); err != nil {
		return err
	}

	if err := os.RemoveAll(assets); err != nil {
		return err
	}

	return os.RemoveAll(layouts)
}

// Save the current theme. This will re-tar the theme overwriting the current
// theme if it's the same.
func (t *Theme) Save() error {
	cfg, err := config.Open()

	if err != nil {
		return err
	}

	defer cfg.Close()

	assets := strings.Replace(filepath.Join(config.ThemesDir, t.Name, config.AssetsDir), config.SiteDir, "", -1)
	layouts := filepath.Join(config.ThemesDir, t.Name, config.LayoutsDir)

	if err := file.Copy(assets, config.AssetsDir); err != nil {
		return err
	}

	if err := file.Copy(layouts, config.LayoutsDir); err != nil {
		return err
	}

	f, err := os.OpenFile(t.Path, os.O_TRUNC|os.O_CREATE|os.O_RDWR, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	dir := filepath.Join(config.ThemesDir, t.Name)

	if err := mktar(f, dir); err != nil {
		return err
	}

	return os.RemoveAll(dir)
}
