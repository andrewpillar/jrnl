package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
)

func isInitialized(dir string) bool {
	for _, d := range meta.Dirs {
		info, err := os.Stat(filepath.Join(dir, d))

		if err != nil {
			return false
		}

		if !info.IsDir() {
			return false
		}
	}

	return true
}

func Init(c cli.Command) {
	target := c.Args.Get(0)

	if isInitialized(target) {
		util.Exit("journal already initialized", nil)
	}

	for _, d := range meta.Dirs {
		d = filepath.Join(target, d)

		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				util.Exit("failed to initialize journal", err)
			}

			continue
		}

		if !f.IsDir() {
			util.Exit("unexpected non-directory file", errors.New(d))
		}
	}

	m, err := meta.Init(target)

	if err != nil {
		util.Exit("failed to create meta file", err)
	}

	m.Close()

	fmt.Println("journal initialized, set the title with 'jrnl title'")
}
