package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

var (
	file = "jrnl.toml"
	root  = "."

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

type Config struct {
	f *os.File `toml:"-"`

	Site struct {
		Title       string   `toml:"title"`
		Description string   `toml:"description"`
		Link        string   `toml:"link"`
		Remote      string   `toml:"remote"`
		Theme       string   `toml:"theme"`
		Blogroll    []string `toml:"blogroll"`
	} `toml:"site"`

	Author struct {
		Name  string `toml:"name"`
		Email string `toml:"email"`
	} `toml:"author"`
}

// Check if jrnl has already been initialized in the given directory.
func Initialized(dir string) error {
	for _, f := range Dirs {
		info, err := os.Stat(filepath.Join(dir, f))

		if err != nil {
			return err
		}

		if !info.IsDir() {
			return errors.New("not a directory " + filepath.Join(dir, f))
		}
	}

	return nil
}

// Create the jrnl.toml file in the given directory. This is called during jrnl
// initialization which is why we pass it a directory whereby the jrnl would be
// initialized.
func Create(dir string) error {
	f, err := os.OpenFile(filepath.Join(dir, file), os.O_CREATE|os.O_RDWR, FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(stub))

	return err
}

// Open the jrnl.toml file for editing. It is expected for a subsequent call to
// Close to be made once the resource is no longer needed.
func Open() (*Config, error) {
	f, err := os.OpenFile(filepath.Join(root, file), os.O_RDWR, FileMode)

	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	dec := toml.NewDecoder(f)

	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}

	cfg.f = f

	return cfg, nil
}

func (c *Config) Save() error {
	info, err := c.f.Stat()

	if err != nil {
		return err
	}

	if info.Size() > 0 {
		if err := c.f.Truncate(0); err != nil {
			return err
		}
	}

	if _, err := c.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	enc := toml.NewEncoder(c.f)

	return enc.Encode(c)
}

func (c *Config) Close() error {
	return c.f.Close()
}
