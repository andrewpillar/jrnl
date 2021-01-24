package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	dataDir      = "_data"
	postsDir     = "_posts"
	pagesDir     = "_pages"
	siteDir      = "_site"
	themesDir    = "_themes"
	layoutsDir   = "_layouts"
	assetsDir    = filepath.Join(siteDir, "assets")

	dirs = []string{
		dataDir,
		postsDir,
		pagesDir,
		siteDir,
		themesDir,
		layoutsDir,
		assetsDir,
	}

	ErrInitialized = errors.New("journal not initialized")
	ErrBadInit     = errors.New("journal not properly initialized")

	InitCmd = &Command{
		Usage: "init [directory]",
		Short: "initializes a new journal",
		Long:  `init will initialize a new journal. If directory is given to the command then a
new journal will be initialized in that directory, otherwise the current
directory is used.`,
		Run:   initCmd,
	}
)

func initialized(root string) error {
	for _, dir := range dirs {
		info, err := os.Stat(filepath.Join(root, dir))

		if err != nil {
			return ErrInitialized
		}

		if !info.IsDir() {
			return ErrBadInit
		}
	}
	return nil
}

func initCmd(cmd *Command, args []string) {
	target := "."

	if len(args) >= 2 {
		target = args[1]
	}

	if err := initialized(target); err == nil {
		fmt.Fprintf(os.Stderr, "%s %s: journal already initialized\n", cmd.Argv0, args[0])
		os.Exit(1)
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(target, dir), os.FileMode(0755)); err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to initialize journal: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
	}

	cfg, err := CreateConfig(target)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to initialize journal: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
	cfg.Close()
}
