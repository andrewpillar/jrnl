package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Site   struct {
		Title       string
		Description string
		Link        string
		Remote      string
		Theme       string
		Blogroll    []string
	}
	Author struct {
		Name  string
		Email string
	}
}

var (
	file = "jrnl.toml"
	root = "."

	stub = `

[site]
title       = ""
description = ""
link        = ""
remote      = ""
theme       = ""
blogroll    = []

[author]
name  = ""
email = ""
`

	PostsDir   = "_posts"
	PagesDir   = "_pages"
	SiteDir    = "_site"
	ThemesDir  = "_themes"
	LayoutsDir = "_layouts"
	IndexDir   = "_index"
	AssetsDir  = filepath.Join(SiteDir, "assets")

	Dirs = []string{
		PostsDir,
		PagesDir,
		SiteDir,
		ThemesDir,
		LayoutsDir,
		AssetsDir,
	}

	DirMode  = os.FileMode(0755)
	FileMode = os.FileMode(0644)
)

func Create(dir string) error {
	f, err := os.OpenFile(filepath.Join(dir, file), os.O_CREATE|os.O_RDWR, FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(stub))
	return err
}

func Initialized(dir string) error {
	for _, d := range Dirs {
		info, err := os.Stat(filepath.Join(dir, d))

		if err != nil {
			return err
		}

		if !info.IsDir() {
			return errors.New("not a directory " + filepath.Join(dir, d))
		}
	}
	return nil
}

func Open() (Config, error) {
	f, err := os.OpenFile(filepath.Join(root, file), os.O_RDWR, FileMode)

	if err != nil {
		return Config{}, err
	}

	defer f.Close()

	cfg := Config{}
	dec := toml.NewDecoder(f)

	if err := dec.Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Save() error {
	f, err := os.OpenFile(filepath.Join(root, file), os.O_TRUNC|os.O_CREATE|os.O_RDWR, FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}
