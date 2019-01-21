package cmd

import (
	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/util"
)

func Title(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to get config", err)
	}

	defer cfg.Close()

	title := c.Args.Get(0)

	if title != "" {
		cfg.Title = title

		if err := cfg.Save(); err != nil {
			util.ExitError("failed to save config", err)
		}
	}
}
