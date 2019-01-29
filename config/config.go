package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	file = "jrnl.yml"

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
	*os.File `yaml:"-"`

	Title  string
	Site   string
	Author struct {
		Name  string
		Email string
	}
	Description string
	Theme       string
	Remote      string
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

// Create the jrnl.yml file in the given directory. This is called during jrnl initialization
// which is why we pass it a directory whereby the jrnl would be initialized.
func Create(dir string) error {
	f, err := os.OpenFile(filepath.Join(dir, file), os.O_CREATE|os.O_RDWR, FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	cfg := &Config{
		File: f,
	}

	return cfg.Save()
}

// Open the jrnl.yml file.
func Open() (*Config, error) {
	f, err := os.OpenFile(file, os.O_RDWR, FileMode)

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		File: f,
	}

	dec := yaml.NewDecoder(cfg)

	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save the changes made to the jrnl.yml file.
func (c *Config) Save() error {
	info, err := c.Stat()

	if err != nil {
		return err
	}

	if info.Size() > 0 {
		if err := c.Truncate(0); err != nil {
			return err
		}
	}

	if _, err := c.Seek(0, io.SeekStart); err != nil {
		return err
	}

	enc := yaml.NewEncoder(c)

	if err := enc.Encode(c); err != nil {
		return err
	}

	enc.Close()

	return nil
}
