package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

type Config struct {
	f *os.File

	Site struct {
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
	configFile = "jrnl.toml"
	configStub = `

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

	ConfigCmd = &Command{
		Usage: "config <key> <value>",
		Short: "set a configuration value within jrnl.toml",
		Long:  `Config will set the given configuration key with the given value in the
jrnl.toml file. For all properties in that file, this will simply overwrite
what's already there, except for the site.blogroll property which will append
the given value to the pre-existing blogroll. If an empty string is given to
site.blogroll then this will clear down the list.`,
		Run:   configCmd,
	}
)

func CreateConfig(dir string) (*Config, error) {
	f, err := os.OpenFile(filepath.Join(dir, configFile), os.O_CREATE|os.O_RDWR, os.FileMode(0644))

	if err != nil {
		return nil, err
	}

	c := &Config{}

	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return nil, err
	}

	c.f = f
	return c, nil
}

func OpenConfig() (*Config, error) {
	f, err := os.OpenFile(configFile, os.O_RDWR, os.FileMode(0644))

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		f: f,
	}

	err = toml.NewDecoder(f).Decode(cfg)
	return cfg, err
}

func (c *Config) Set(key, val string) error {
	switch key {
	case "site.title":
		c.Site.Title = val
	case "site.description":
		c.Site.Description = val
	case "site.link":
		c.Site.Link = val
	case "site.remote":
		c.Site.Remote = val
	case "site.theme":
		c.Site.Theme = val
	case "site.blogroll":
		if val == "" {
			c.Site.Blogroll = []string{}
			break
		}
		c.Site.Blogroll = append(c.Site.Blogroll, val)
	case "author.name":
		c.Author.Name = val
	case "author.email":
		c.Author.Email = val
	default:
		return errors.New("unknown configuration key")
	}
	return nil
}

func (c *Config) Close() error { return c.f.Close() }

func (c *Config) Save() error {
//	f, err := os.OpenFile(configFile, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.FileMode(0644))
//
//	if err != nil {
//		return err
//	}
//
//	defer f.Close()

	if err := c.f.Truncate(0); err != nil {
		return err
	}

	if _, err := c.f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return toml.NewEncoder(c.f).Encode(c)
}

func configCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	cfg, err := OpenConfig()

	if err != nil {
		println("open config error")
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if err := cfg.Set(args[1], args[2]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}
