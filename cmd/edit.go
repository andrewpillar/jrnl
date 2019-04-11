package cmd

import (
	"errors"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func Edit(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	id := c.Args.Get(0)

	if id == "" {
		util.ExitError("failed to edit entry", errors.New("missing id"))
	}

	pg, err := page.Find(id)

	if err == nil {
		util.OpenInEditor(pg.SourcePath)
		return
	}

	if !os.IsNotExist(err) {
		util.ExitError("failed to edit page", err)
	}

	pt, err := post.Find(id)

	if err != nil {
		util.ExitError("failed to edit post", err)
	}

	if err := pt.Touch(); err != nil {
		util.ExitError("failed to edit post", err)
	}

	util.OpenInEditor(pt.SourcePath)
}
