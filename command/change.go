package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
	"github.com/andrewpillar/jrnl/usage"
)

func editPost(id string) {
	p, err := post.Find(id)

	if err != nil {
		util.Error("failed to find post", err)
	}

	util.OpenInEditor(p.SourcePath)
}

func rmPost(ids []string) {
	code := 0

	for _, id := range ids {
		p, err := post.Find(id)

		if err != nil {
			fmt.Fprintf(os.Stderr, "jrnl: failed to find post\n  %s\n", err)

			code = 1

			continue
		}

		if err = p.Remove(); err != nil {
			fmt.Fprintf(os.Stderr, "jrnl: failed to remove post\n  %s\n", err)

			code = 1
		}
	}

	os.Exit(code)
}

func ChangePost(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		switch c.Name {
			case "edit":
				fmt.Println(usage.Edit)
				return
			case "rm":
				fmt.Println(usage.Rm)
				return
		}
	}

	mustBeInitialized()

	switch c.Name {
		case "edit":
			editPost(c.Args.Get(0))
		case "rm":
			rmPost(c.Args)
	}
}
