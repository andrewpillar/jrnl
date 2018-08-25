package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func Post(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.Post)
		return
	}

	mustBeInitialized()

	p := post.New(c.Args.Get(0), c.Flags.GetString("category"))

	dir := filepath.Dir(p.SourcePath)

	d, err := os.Stat(dir)

	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				util.Error("failed to create post directory", err)
			}
		} else {
			util.Error("failed to stat post directory", err)
		}
	}

	if d != nil && !d.IsDir() {
		util.Error("unexpected non-directory file", errors.New(dir))
	}

	f, err := os.OpenFile(p.SourcePath, os.O_CREATE, os.ModePerm)

	if err != nil {
		util.Error("failed to open post file", err)
	}

	defer f.Close()

	util.OpenInEditor(p.SourcePath)

	fmt.Println("new post added", p.ID)
}
