package cmd

import (
	"errors"
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/page"
)

func createPage(title string) (*page.Page, error) {
	if title == "" {
		return nil, errors.New("missing title")
	}

	cfg, err := config.Open()

	if err != nil {
		return nil, err
	}

	defer cfg.Close()

	p := page.New(title)

	return p, nil
}

func Page(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	p, err := createPage(c.Args.Get(0))

	if err != nil {
		exitError("failed to create page", err)
	}

	if err := p.Touch(); err != nil {
		exitError("failed to create page", err)
	}

	openInEditor(p.SourcePath)
	fmt.Println("new page added", p.ID)
}
