package cmd

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func Rm(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	code := 0

	for _, id := range c.Args {
		pg, err := page.Find(id)

		if err == nil {
			if err := pg.Remove(); err != nil {
				code = 1
				fmt.Fprintf(os.Stderr, "%s: failed to remove page %s: %s\n", os.Args[0], id, err)
			}

			continue
		}

		if !os.IsNotExist(err) {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove page %s: %s\n", os.Args[0], id, err)
			continue
		}

		pt, err := post.Find(id)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove post %s: %s\n", os.Args[0], id, err)
		}

		if err := pt.Remove(); err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove post %s: %s\n", os.Args[0], id, err)
		}
	}

	os.Exit(code)
}
