package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var RmCmd = &Command{
	Usage: "rm <page|post,...>",
	Short: "remove a page or post from the jrnl",
	Run:   rmCmd,
}

func pruneFile() (*os.File, error) {
	dir, err := cacheDir()

	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "prune")

	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.FileMode(0640))
}

func PrunedPaths() ([]string, error) {
	f, err := pruneFile()

	if err != nil {
		return nil, err
	}

	defer f.Close()

	paths := make([]string, 0)

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		if s := sc.Text(); s != "" {
			paths = append(paths, sc.Text())
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	return paths, nil
}

func FlushPrunedPaths() error {
	f, err := pruneFile()

	if err != nil {
		return err
	}
	return os.Remove(f.Name())
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

	if len(args) == 0 {
		return ErrUsage
	}

	f, err := pruneFile()

	if err != nil {
		return err
	}

	defer f.Close()

	for _, name := range args {
		p, path, err := loadPageOrPost(name)

		if err != nil {
			return err
		}

		if err := os.Remove(path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		sitePath := filepath.Join(siteDir, p.URL(), "index.html")

		if err := os.Remove(sitePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		remotePath := strings.TrimPrefix(sitePath, siteDir)

		if _, err := fmt.Fprintln(f, remotePath); err != nil {
			return err
		}
	}
	return nil
}
