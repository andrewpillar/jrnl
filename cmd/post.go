package cmd

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/post"
)

func Post(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	page, err := createPage(c.Args.Get(0))

	if err != nil {
		exitError("failed to create post", err)
	}

	p := post.New(page, c.Flags.GetString("category"))

	if err := p.Touch(); err != nil {
		exitError("failed to create post", err)
	}

	openInEditor(p.SourcePath)
	fmt.Println("new post added", p.ID)
}
