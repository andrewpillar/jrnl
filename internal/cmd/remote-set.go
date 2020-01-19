package cmd

import (
	"errors"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/config"
)

func RemoteSet(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		exitError("failed to set remote", err)
	}

	url := c.Args.Get(0)

	if url == "" {
		exitError("failed to set remote", errors.New("missing remote target"))
	}

	cfg.Site.Remote = url

	if err := cfg.Save(); err != nil {
		exitError("failed to set remote", err)
	}
}
