package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

const configFile = "jrnl.yml"

type Author struct {
	Email string
	Name  string
}

type Site struct {
	Title       string `yaml:",omitempty"`
	Description string `yaml:",omitempty"`
	Theme       string `yaml:",omitempty"`
	URL         string `yaml:",omitempty"`
	Atom        string
	RSS         string
	Blogroll    []string `yaml:",omitempty"`
}

type Config struct {
	Remote string `yaml:",omitempty"`

	Author Author `yaml:",omitempty"`
	Site   Site
}

func LoadConfig() (*Config, error) {
	f, err := os.Open(configFile)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var cfg Config

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) file() (*os.File, error) {
	return os.Create(configFile)
}

func (c *Config) Save() error {
	f, err := c.file()

	if err != nil {
		return err
	}

	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&c); err != nil {
		return err
	}
	return nil
}
