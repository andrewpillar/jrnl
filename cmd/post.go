package cmd

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func Post(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	page, err := createPage(c.Args.Get(0))

	if err != nil {
		util.ExitError("failed to create post", err)
	}

	p := post.New(page, c.Flags.GetString("category"))

	if err := p.Touch(); err != nil {
		util.ExitError("failed to create post", err)
	}

	util.OpenInEditor(p.SourcePath)
	fmt.Println("new post added", p.ID)
}
