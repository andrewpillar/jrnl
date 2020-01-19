package cmd

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
)

func Page(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	p := blog.NewPage(c.Args.Get(0))

	if err := p.Touch(); err != nil {
		exitError("failed to create page", err)
	}

	openInEditor(p.SourcePath)

	fmt.Println("new page added", p.ID)
}
