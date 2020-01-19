package cmd

import (
	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/config"
)

func Title(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		exitError("failed to get config", err)
	}

	title := c.Args.Get(0)

	if title != "" {
		cfg.Site.Title = title

		if err := cfg.Save(); err != nil {
			exitError("failed to save config", err)
		}
	}
}
