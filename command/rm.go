package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
)

func Rm(c cli.Command) {
	code := 0

	for _, id := range c.Args {
		p, err := page.Find(id)

		if err != nil {
			if os.IsNotExist(err) {
				pst, err := post.Find(id)

				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: failed to find post %s: %s\n", os.Args[0], id, err)

					code = 1
					continue
				}

				if err := pst.Remove(); err != nil {
					fmt.Fprintf(os.Stderr, "%s: failed to remove post %s: %s\n", os.Args[0], id, err)

					code = 1
					continue
				}
			}

			fmt.Fprintf(os.Stderr, "%s: failed to find page %s: %s\n", os.Args[0], id, err)

			code = 1
			continue
		}

		if err := p.Remove(); err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to remove page %s: %s\n", os.Args[0], id, err)

			code = 1
		}
	}

	os.Exit(code)
}
