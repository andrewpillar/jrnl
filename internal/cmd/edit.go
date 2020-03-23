package cmd

import (
	"errors"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
)

func Edit(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	id := c.Args.Get(0)

	if id == "" {
		exitError("failed to edit entry", errors.New("missing id"))
	}

	pg, err := blog.GetPage(id)

	if err == nil {
		openInEditor(pg.SourcePath)
		return
	}

	if !os.IsNotExist(err) {
		exitError("failed to edit page", err)
	}

	p, err := blog.GetPost(id)

	if err != nil {
		exitError("failed to edit post", err)
	}

	if err := p.Touch(); err != nil {
		exitError("failed to edit post", err)
	}

	openInEditor(p.SourcePath)
}
