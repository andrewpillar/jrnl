package cmd

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
)

func Remove(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	h, err := getBlogHash()

	if err != nil {
		exitError("failed to load blog hash", err)
	}

	code := 0

	for _, id := range c.Args {
		pg, err := blog.GetPage(id)

		if err == nil {
			if err := pg.Remove(); err != nil {
				code = 1
				fmt.Fprintf(os.Stderr, "%s: failed to remove page %s: %s\n", os.Args[0], id, err)
			}

			h[id] = nil
			continue
		}

		if !os.IsNotExist(err) {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove page %s: %s\n", os.Args[0], id, err)
			continue
		}

		p, err := blog.GetPost(id)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove post %s: %s\n", os.Args[0], id, err)
		}

		if err := p.Remove(); err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to remove post %s: %s\n", os.Args[0], id, err)
		}

		h[id] = nil
	}

	if err := writeBlogHash(h); err != nil {
		exitError("failed to write blog hash", err)
	}

	if code != 0 {
		os.Exit(code)
	}
}
