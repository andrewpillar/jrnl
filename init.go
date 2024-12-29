package main

import (
	"embed"
	"errors"
	"os"
	"path/filepath"
)

//go:embed embed/*.tmpl
var embeds embed.FS

const (
	dirMode = os.FileMode(0750)

	assetDir  = "_assets"
	layoutDir = "_layouts"
	pageDir   = "_pages"
	postDir   = "_posts"
	siteDir   = "_site"
)

var (
	jrnlDirs = [...]string{
		assetDir,
		layoutDir,
		pageDir,
		postDir,
		siteDir,
	}

	ErrInitialized    = errors.New("already initialized")
	ErrNotInitialized = errors.New("not initialized")

	InitCmd = &Command{
		Usage: "init [directory]",
		Short: "initialize a new journal",
		Long: `init will initialize a new journal. If a direction is given to the comment then a
new journal will be initialized in that directory, otherwise the current
directory is used.`,
		Run: initCmd,
	}
)

func Initialized(root string) error {
	for _, dir := range jrnlDirs {
		info, err := os.Stat(filepath.Join(root, dir))

		if err != nil {
			return ErrNotInitialized
		}

		if !info.IsDir() {
			return errors.New(dir + " is not a directory")
		}
	}
	return nil
}

func initCmd(cmd *Command, args []string) error {
	target := "."

	if len(args) >= 1 {
		target = args[0]
	}

	if err := Initialized(target); err == nil {
		return nil
	}

	for _, dir := range jrnlDirs {
		if err := os.MkdirAll(filepath.Join(target, dir), dirMode); err != nil {
			return err
		}
	}

	ents, err := embeds.ReadDir("embed")

	if err != nil {
		return err
	}

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		b, err := embeds.ReadFile(filepath.Join("embed", ent.Name()))

		if err != nil {
			return err
		}

		err = func(b []byte) error {
			name := ent.Name()
			name = name[:len(name)-5]

			f, err := os.Create(filepath.Join(layoutDir, name))

			if err != nil {
				return err
			}

			defer f.Close()

			_, err = f.Write(b)
			return err
		}(b)

		if err != nil {
			return err
		}
	}

	var cfg Config

	cfg.Site.Atom = "atom.xml"
	cfg.Site.RSS = "rss.xml"

	if err := cfg.Save(); err != nil {
		return err
	}
	return nil
}
