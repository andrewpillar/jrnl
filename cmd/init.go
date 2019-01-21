package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/util"
)

func Init(c cli.Command) {
	target := c.Args.Get(0)

	if err := config.Initialized(target); err == nil {
		util.ExitError("journal already initialized", nil)
	}

	for _, dir := range config.Dirs {
		if err := os.MkdirAll(filepath.Join(target, dir), config.DirMode); err != nil {
			util.ExitError("failed to initialize", err)
		}
	}

	if err := config.Create(target); err != nil {
		util.ExitError("failed to initialize", err)
	}

	fmt.Println("journal initialized, set the title with 'jrnl title'")
}
