package cmd

import (
	"errors"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
)

func Edit(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	id := c.Args.Get(0)

	if id == "" {
		exitError("failed to edit entry", errors.New("missing id"))
	}

	pg, err := page.Find(id)

	if err == nil {
		openInEditor(pg.SourcePath)
		return
	}

	if !os.IsNotExist(err) {
		exitError("failed to edit page", err)
	}

	pt, err := post.Find(id)

	if err != nil {
		exitError("failed to edit post", err)
	}

	if err := pt.Touch(); err != nil {
		exitError("failed to edit post", err)
	}

	openInEditor(pt.SourcePath)
}
