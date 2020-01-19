package blog

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/internal/config"
)

type Theme struct {
	Name string
	Path string
}

func copyDir(dst, src string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(src)

	if err != nil {
		return err
	}

	for _, f := range files {
		if err := copyAll(filepath.Join(dst, f.Name()), filepath.Join(src, f.Name())); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(dst, src string, info os.FileInfo) error {
	if err := os.MkdirAll(filepath.Dir(dst), info.Mode()); err != nil {
		return err
	}

	fdst, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer fdst.Close()

	if err = os.Chmod(fdst.Name(), info.Mode()); err != nil {
		return err
	}

	fsrc, err := os.Open(src)

	if err != nil {
		return err
	}

	defer fsrc.Close()

	_, err = io.Copy(fdst, fsrc)
	return err
}

func copyAll(dst, src string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(dst, src, info)
	}

	return copyFile(dst, src, info)
}

func mktar(w io.Writer, src string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}

	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())

		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(path, src, "", -1), string(os.PathSeparator))

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

	return filepath.Walk(src, fn)
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
			if _, err = os.Stat(target); err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(target, config.DirMode); err != nil {
						return err
					}
					continue
				}
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))

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

func resolveTheme(path string) (Theme, error) {
	var t Theme

	if _, err := os.Stat(path); err != nil {
		return t, err
	}

	t.Name = strings.Split(strings.Replace(path, config.ThemesDir + string(os.PathSeparator), "", 1), ".")[0]
	t.Path = path
	return t, nil
}

func GetTheme(name string) (Theme, error) {
	return resolveTheme(filepath.Join(config.ThemesDir, name + ".tar.gz"))
}

func NewTheme(name string) Theme {
	return Theme{
		Name: name,
		Path: filepath.Join(config.ThemesDir, name + ".tar.gz"),
	}
}

func Themes() ([]Theme, error) {
	tt := make([]Theme, 0)

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		t, err := resolveTheme(path)

		if err != nil {
			return err
		}

		tt = append(tt, t)
		return nil
	}

	err := filepath.Walk(config.ThemesDir, fn)
	return tt, err
}

func (t Theme) Load() error {
	f, err := os.Open(t.Path)

	if err != nil {
		return err
	}

	defer f.Close()

	for _, dir := range []string{config.LayoutsDir, config.AssetsDir} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	if err := untar(config.ThemesDir, f); err != nil {
		return err
	}

	assets := strings.Replace(filepath.Join(config.ThemesDir, config.AssetsDir), config.SiteDir, "", -1)
	layouts := filepath.Join(config.ThemesDir, filepath.Base(config.LayoutsDir))

	if err := copyAll(config.AssetsDir, assets); err != nil {
		return err
	}

	if err := copyAll(config.LayoutsDir, layouts); err != nil {
		return err
	}

	if err := os.RemoveAll(assets); err != nil {
		return err
	}

	return os.RemoveAll(layouts)
}

func (t *Theme) Remove() error {
	if err := os.Remove(t.Path); err != nil {
		return err
	}

	t.Name = ""
	t.Path = ""
	return nil
}

func (t Theme) Save() error {
	assets := strings.Replace(filepath.Join(config.ThemesDir, t.Name, config.AssetsDir), config.SiteDir, "", -1)
	layouts := filepath.Join(config.ThemesDir, t.Name, filepath.Base(config.LayoutsDir))

	if err := copyAll(assets, config.AssetsDir); err != nil {
		return err
	}

	if err := copyAll(layouts, config.LayoutsDir); err != nil {
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
