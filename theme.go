package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Theme struct {
	Name string
	Path string
}

var (
	ThemeLsCmd = &Command{
		Usage: "ls",
		Short: "list the journal's themes",
		Run:   themeLsCmd,
	}

	ThemeRmCmd = &Command{
		Usage: "rm <name,...>",
		Short: "remove the given themes",
		Run:   themeRmCmd,
	}

	ThemeSaveCmd = &Command{
		Usage: "save <name>",
		Short: "save the current journal theme",
		Run:   themeSaveCmd,
	}

	ThemeUseCmd = &Command{
		Usage: "use <name>",
		Short: "use the given journal theme",
		Run:   themeUseCmd,
	}
)

func copydir(dst, src string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	infos, err := ioutil.ReadDir(src)

	if err != nil {
		return err
	}

	for _, info := range infos {
		dst1 := filepath.Join(dst, info.Name())
		src1 := filepath.Join(src, info.Name())

		if err := fscopy(dst1, src1); err != nil {
			return err
		}
	}
	return nil
}

func copyfile(dst, src string, info os.FileInfo) error {
	if err := os.MkdirAll(filepath.Dir(dst), info.Mode()); err != nil {
		return err
	}

	fdst, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer fdst.Close()

	if err := os.Chmod(fdst.Name(), info.Mode()); err != nil {
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

func fscopy(dst, src string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return copydir(dst, src, info)
	}
	return copyfile(dst, src, info)
}

func mktar(w io.Writer, src string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}

	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
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
	})
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
			if _, err := os.Stat(target); err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(target, os.FileMode(0755)); err != nil {
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

func resolveTheme(path string) (*Theme, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	return &Theme{
		Name: strings.Split(strings.Replace(path, themesDir + string(os.PathSeparator), "", 1), ".")[0],
		Path: path,
	}, nil
}

func GetTheme(name string) (*Theme, bool, error) {
	theme, err := resolveTheme(filepath.Join(themesDir, name + ".tar.gz"))

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, false, err
		}
		return nil, false, nil
	}
	return theme, true, nil
}

func Themes() ([]*Theme, error) {
	themes := make([]*Theme, 0)

	err := filepath.Walk(themesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		theme, err := resolveTheme(path)

		if err != nil {
			return err
		}

		themes = append(themes, theme)
		return nil
	})
	return themes, err
}

func (t *Theme) Load() error {
	f, err := os.Open(t.Path)

	if err != nil {
		return err
	}

	defer f.Close()

	for _, dir := range []string{layoutsDir, assetsDir} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	if err := untar(themesDir, f); err != nil {
		return err
	}

	assets := strings.Replace(filepath.Join(themesDir, assetsDir), siteDir, "", -1)
	layouts := filepath.Join(themesDir, filepath.Base(layoutsDir))

	if err := fscopy(assetsDir, assets); err != nil {
		return err
	}

	if err := fscopy(layoutsDir, layouts); err != nil {
		return err
	}

	if err := os.RemoveAll(assets); err != nil {
		return err
	}
	return os.RemoveAll(layouts)
}

func (t *Theme) Save() error {
	assets := strings.Replace(filepath.Join(themesDir, t.Name, assetsDir), siteDir, "", -1)
	layouts := filepath.Join(themesDir, t.Name, filepath.Base(layoutsDir))

	if err := fscopy(assets, assetsDir); err != nil {
		return err
	}

	if err := fscopy(layouts, layoutsDir); err != nil {
		return err
	}

	f, err := os.OpenFile(t.Path, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.FileMode(0644))

	if err != nil {
		return err
	}

	dir := filepath.Join(themesDir, t.Name)

	if err := mktar(f, dir); err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func ThemeCmd(argv0 string) *Command {
	cmd := &Command{
		Usage:    "theme <command> [arguments]",
		Short:    "manage the journal's themes",
		Run:      themeCmd,
		Commands: &CommandSet{
			Argv0: argv0 + " theme",
		},
	}

	cmd.Commands.Add("ls", ThemeLsCmd)
	cmd.Commands.Add("rm", ThemeRmCmd)
	cmd.Commands.Add("save", ThemeSaveCmd)
	cmd.Commands.Add("use", ThemeUseCmd)
	return cmd
}

func themeLsCmd(cmd *Command, args []string) {
	themes, err := Themes()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	for _, theme := range themes {
		fmt.Println(theme.Name)
	}
}

func themeSaveCmd(cmd *Command, args []string) {
	cfg, err := OpenConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) >=2 {
		name := args[1]

		if name != "" {
			cfg.Site.Theme = slug(name)
		}
	}

	if cfg.Site.Theme == "" {
		fmt.Fprintf(os.Stderr, "%s %s: no theme name specified\n", cmd.Argv0, args[0])
		os.Exit(1)
	}

	theme, ok, err := GetTheme(cfg.Site.Theme)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if !ok {
		theme = &Theme{
			Name: cfg.Site.Theme,
			Path: filepath.Join(themesDir, cfg.Site.Theme + ".tar.gz"),
		}
	}

	if err := theme.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to save theme: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to save theme: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}

func themeRmCmd(cmd *Command, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	code := 0

	for _, name := range args[1:] {
		theme, ok, err := GetTheme(name)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: failed to get theme %q: %s\n", cmd.Argv0, args[0], name, err)
			continue
		}

		if !ok {
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: no such theme %q\n", cmd.Argv0, args[0], name)
			continue
		}

		if err := os.Remove(theme.Path); err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: failed to remove theme %q: %s\n", cmd.Argv0, args[0], name, err)
		}
	}
	os.Exit(code)
}

func themeUseCmd(cmd *Command, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	cfg, err := OpenConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	name := args[1]

	theme, ok, err := GetTheme(name)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to get theme: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if !ok {
		fmt.Fprintf(os.Stderr, "%s %s: no such theme\n", cmd.Argv0, args[0])
		os.Exit(1)
	}

	if err := theme.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to load theme: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	cfg.Site.Theme = theme.Name

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to save config: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}

func themeCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) < 2 {
		cfg, err := OpenConfig()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to open config: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
		fmt.Println(cfg.Site.Theme)
		return
	}

	if err := cmd.Commands.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}
