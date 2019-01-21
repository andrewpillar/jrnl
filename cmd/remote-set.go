package cmd

import (
	"errors"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/util"
)

func RemoteSet(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to set remote", err)
	}

	defer cfg.Close()

	url := c.Args.Get(0)

	if url == "" {
		util.ExitError("failed to set remote", errors.New("missing remote target"))
	}

	cfg.Remote = url

	if err := cfg.Save(); err != nil {
		util.ExitError("failed to set remote", err)
	}
}
