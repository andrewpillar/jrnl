package cmd

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
)

func Post(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	p := blog.NewPost(c.Args.Get(0), c.Flags.GetString("category"))

	if err := p.Touch(); err != nil {
		exitError("failed to create post", err)
	}

	openInEditor(p.SourcePath)

	fmt.Println("new post added", p.ID)
}
