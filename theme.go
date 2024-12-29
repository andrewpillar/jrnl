package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var themeDirs = [...]string{
	assetDir,
	layoutDir,
}

var ThemeLsCmd = &Command{
	Usage: "ls",
	Short: "list installed themes",
	Run:   themeLsCmd,
}

func themeDir() (string, error) {
	dir, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}

	dir = filepath.Join(dir, filepath.Base(os.Args[0]), "themes")

	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", err
	}
	return dir, nil
}

func themeLsCmd(cmd *Command, args []string) error {
	dir, err := themeDir()

	if err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		name := info.Name()
		name = name[:len(name)-7]

		fmt.Println(name)
		return nil
	})
}

var ThemeRmCmd = &Command{
	Usage: "rm <name,...>",
	Short: "remove given themes",
	Run:   themeRmCmd,
}

func themeRmCmd(cmd *Command, args []string) error {
	dir, err := themeDir()

	if err != nil {
		return err
	}

	for _, name := range args {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}

var ThemeSaveCmd = &Command{
	Usage: "save <name>",
	Short: "save the current theme",
	Run:   themeSaveCmd,
}

func tardir(tw *tar.Writer, dir string) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, info.Name())

		if err != nil {
			return err
		}

		hdr.Name = filepath.ToSlash(path)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.IsDir() {
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

func untar(r io.Reader) error {
	gzr, err := gzip.NewReader(r)

	if err != nil {
		return err
	}

	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(hdr.Name, dirMode); err != nil {
				return err
			}
		case tar.TypeReg:
			info := hdr.FileInfo()

			err := func(r io.Reader, hdr *tar.Header) error {
				f, err := os.OpenFile(hdr.Name, os.O_TRUNC|os.O_CREATE|os.O_RDWR, info.Mode())

				if err != nil {
					return err
				}

				defer f.Close()

				if _, err := io.Copy(f, tr); err != nil {
					return err
				}
				return nil
			}(tr, hdr)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func themeSaveCmd(cmd *Command, args []string) error {
	cfg, err := LoadConfig()

	if err != nil {
		return err
	}

	name := cfg.Site.Theme

	if len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return errors.New("no theme name ")
	}

	dir, err := themeDir()

	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, name) + ".tar.gz")

	if err != nil {
		return err
	}

	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for _, dir := range themeDirs {
		if err := tardir(tw, dir); err != nil {
			return err
		}
	}

	return nil
}

var ThemeUseCmd = &Command{
	Usage: "use [name]",
	Short: "use the given jrnl theme",
	Run:   themeUseCmd,
}

func themeUseCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	cfg, err := LoadConfig()

	if err != nil {
		return err
	}

	dir, err := themeDir()

	if err != nil {
		return err
	}

	name := args[0]

	f, err := os.Open(filepath.Join(dir, name) + ".tar.gz")

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return errors.New("theme does not exist")
		}
		return err
	}

	defer f.Close()

	for _, dir := range themeDirs {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	if err := untar(f); err != nil {
		return err
	}

	cfg.Site.Theme = name

	if err := cfg.Save(); err != nil {
		return err
	}
	return nil
}

func ThemeCmd(argv0 string) *Command {
	cmd := &Command{
		Usage: "theme <command> [arguments]",
		Short: "manage jrnl themes",
		Run:   themeCmd,
		Commands: &CommandSet{
			Argv0: argv0 + " theme",
		},
	}

	cmd.Commands.Add("ls", ThemeLsCmd)
	cmd.Commands.Add("rm", ThemeRmCmd)
	cmd.Commands.Add("save", ThemeSaveCmd)
	cmd.Commands.Add("use", ThemeUseCmd)
	cmd.Commands.Add("rm", ThemeRmCmd)

	return cmd
}

func themeCmd(cmd *Command, args []string) error {
	if err := Initialized("."); err != nil {
		return err
	}

	if len(args) == 0 {
		cfg, err := LoadConfig()

		if err != nil {
			return err
		}
		fmt.Println(cfg.Site.Theme)
		return nil
	}

	if err := cmd.Commands.Parse(args); err != nil {
		return err
	}
	return nil
}
