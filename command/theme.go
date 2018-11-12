package command

import (
	"errors"
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/theme"
	"github.com/andrewpillar/jrnl/util"
)

func Theme(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	if m.Theme == "" {
		fmt.Println("no theme being used")
		return
	}

	fmt.Println("current theme: " + m.Theme)
}

func ThemeLs(c cli.Command) {
	util.MustBeInitialized()

	themes, err := theme.All()

	if err != nil {
		util.Exit("failed to get all themes", err)
	}

	for _, t := range themes {
		fmt.Println(t.Name)
	}
}

func ThemeSave(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	defer m.Close()

	name := c.Args.Get(0)

	if name != "" {
		m.Theme = util.Slug(name)
	}

	if m.Theme == "" {
		util.Exit("failed to save theme", errors.New("no theme specified"))
	}

	t, err := theme.Find(m.Theme)

	if err != nil {
		if !os.IsNotExist(err) {
			util.Exit("failed to find theme", err)
		}

		t, err = theme.New(m.Theme)

		if err != nil {
			util.Exit("failed to create theme", err)
		}
	}

	if err := t.Save(); err != nil {
		util.Exit("failed to save theme", err)
	}

	defer t.Close()

	fmt.Println("saved theme: " + t.Name)
}

func ThemeUse(c cli.Command) {
	util.MustBeInitialized()

	name := util.Slug(c.Args.Get(0))

	t, err := theme.Find(name)

	if err != nil {
		if os.IsNotExist(err) {
			util.Exit("theme does not exist", errors.New(name))
		}

		util.Exit("failed to use theme", err)
	}

	if err := t.Use(); err != nil {
		util.Exit("failed to use theme", err)
	}

	fmt.Println("using theme: " + t.Name)
}

func ThemeRm(c cli.Command) {
	util.MustBeInitialized()

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
