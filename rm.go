package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var RmCmd = &Command{
	Usage: "rm <page|post,...>",
	Short: "remove a page or post from the jrnl",
	Run:   rmCmd,
}

func loadPageOrPost(name string) (SitePage, string, error) {
	name += ".md"

	path := filepath.Join(pageDir, name)

	p, err := LoadPage(path)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
		goto loadPost
	}
	return p, path, nil

loadPost:
	path = filepath.Join(postDir, name)

	page, err := LoadPage(path)

	if err != nil {
		return nil, "", err
	}

	return &Post{
		Page: page,
	}, path, nil
}

func rmCmd(cmd *Command, args []string) error {
	if err := Initialized("."); err != nil {
		return err
	}

	cfg, err := LoadConfig()

	if err != nil {
		return err
	}

	if len(args) == 0 {
		return ErrUsage
	}

	var remote Remote

	if cfg.Remote != "" {
		remote, err = ParseRemote(cfg.Remote)

		if err != nil {
			return err
		}

		if err := remote.Connect(); err != nil {
			return err
		}
	}

	for _, name := range args {
		p, path, err := loadPageOrPost(name)

		if err != nil {
			return err
		}

		if err := os.Remove(path); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}

		sitePath := filepath.Join(siteDir, p.URL())

		if err := os.RemoveAll(sitePath); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}

		if remote != nil {
			sitePath = strings.TrimPrefix(sitePath, siteDir)

			if err := remote.Remove(sitePath); err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}
		}
	}
	return nil
}
