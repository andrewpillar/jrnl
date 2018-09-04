package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func isInitialized(target string) bool {
	for _, d := range meta.Dirs {
		f, err := os.Stat(filepath.Join(target, d))

		if os.IsNotExist(err) {
			return false
		}

		if !f.IsDir() {
			return false
		}
	}

	return true
}

func mustBeInitialized() {
	for _, d := range meta.Dirs {
		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			util.Error("not fully initialized", nil)
		}

		if !f.IsDir() {
			util.Error("unexpected non-directory file", errors.New(d))
		}
	}
}

func Initialize(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) > 1 {
		fmt.Println(usage.Initialize)
		return
	}

	target := c.Args.Get(0)

	if isInitialized(target) {
		util.Error("journal already initialized", nil)
	}

	for _, d := range meta.Dirs {
		d = filepath.Join(target, d)

		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				util.Error("failed to initialize journal", err)
			}

			continue
		}

		if !f.IsDir() {
			util.Error("unexpected non-directory file", errors.New(d))
		}
	}

	for l, s := range meta.Layouts {
		path := filepath.Join(target, meta.LayoutsDir, l)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)

		if err != nil {
			util.Error("failed to open layout file", err)
		}

		defer f.Close()

		_, err = f.Write([]byte(s))

		if err != nil {
			util.Error("failed to write layout file", err)
		}
	}

	m, err := meta.Init(target)

	if err != nil {
		util.Error("failed to create meta file", err)
	}

	m.Close()

	fmt.Println("journal initialized, set the title with 'jrnl title'")
}
