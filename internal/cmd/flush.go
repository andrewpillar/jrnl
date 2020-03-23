package cmd

import (
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/config"
)

func Flush(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	if err := os.RemoveAll("jrnl.hash"); err != nil {
		exitError("failed to remove jrnl hash", err)
	}
}
