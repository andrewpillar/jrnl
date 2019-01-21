package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/theme"
	"github.com/andrewpillar/jrnl/util"
)

func Theme(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to get config", err)
	}

	defer cfg.Close()

	if cfg.Theme == "" {
		fmt.Println("no theme being used")
		return
	}

	fmt.Println("current theme: " + cfg.Theme)
}

func ThemeLs(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	themes, err := theme.All()

	if err != nil {
		util.ExitError("failed to get all theme", err)
	}

	for _, t := range themes {
		fmt.Println(t.Name)
	}
}

func ThemeSave(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to get config", err)
	}

	defer cfg.Close()

	name := c.Args.Get(0)

	if name != "" {
		cfg.Theme = util.Slug(name)
	}

	if cfg.Theme == "" {
		util.ExitError("failed to save theme", errors.New("no theme name specified"))
	}

	t, err := theme.Find(cfg.Theme)

	if err != nil {
		if !os.IsNotExist(err) {
			util.ExitError("failed to save theme", err)
		}

		t = theme.New(cfg.Theme)
	}

	if err := t.Save(); err != nil {
		util.ExitError("failed to save theme", err)
	}

	if err := cfg.Save(); err != nil {
		util.ExitError("failed to save theme", err)
	}

	fmt.Println("saved theme: " + t.Name)
}

func ThemeUse(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to get config", err)
	}

	defer cfg.Close()

	name := util.Slug(c.Args.Get(0))

	if name == "" {
		util.ExitError("failed to use theme", errors.New("missing theme name"))
	}

	t, err := theme.Find(name)

	if err != nil {
		util.ExitError("failed to use theme", err)
	}

	if err := t.Load(); err != nil {
		util.ExitError("failed to use theme", err)
	}

	cfg.Theme = name

	if err := cfg.Save(); err != nil {
		util.ExitError("failed to save config", err)
	}
}

func ThemeRm(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	code := 0

	for _, name := range c.Args {
		t, err := theme.Find(name)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove theme %s: %s\n", os.Args[0], name, err)
			continue
		}

		if err := os.Remove(t.Path); err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove theme %s: %s\n", os.Args[0], name, err)
		}
	}

	os.Exit(code)
}
