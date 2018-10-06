package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
)

var (
	errPostFindFmt = "%s: failed to find post %s: %s\n"

	errPostRemoveFmt = "%s: failed to remove post %s: %s\n"
)

func Rm(c cli.Command) {
	code := 0

	for _, id := range c.Args {
		p, err := post.Find(id)

		if err != nil {
			fmt.Fprintf(os.Stderr, errPostFindFmt, os.Args[0], id, err)

			code = 1
			continue
		}

		if err = p.Remove(); err != nil {
			fmt.Fprintf(os.Stderr, errPostRemoveFmt, os.Args[0], id, err)

			code = 1
		}
	}

	os.Exit(code)
}
