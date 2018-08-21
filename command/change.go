package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func editPost(id string) {
	p, err := post.NewFromPath(id + ".md")

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	util.OpenInEditor(p.SourcePath)
}

func rmPost(ids []string) {
	code := 0

	for _, id := range ids {
		p, err := post.NewFromPath(id + ".md")

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)

			code = 1
		}

		if err = p.Remove(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)

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
