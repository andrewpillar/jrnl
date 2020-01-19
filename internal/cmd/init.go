package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/config"
)

func Init(c cli.Command) {
	target := c.Args.Get(0)

	if err := config.Initialized(target); err == nil {
		exitError("journal already initialized", nil)
	}

	for _, dir := range config.Dirs {
		if err := os.MkdirAll(filepath.Join(target, dir), config.DirMode); err != nil {
			exitError("failed to initialize", err)
		}
	}

	if err := config.Create(target); err != nil {
		exitError("failed to initialize", err)
	}

	fmt.Println("journal initialized, set the title with 'jrnl title'")
}
