package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

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

	src := PostsDir + "/" + c.Args.Get(0) + ".md"

	p, err := post.NewFromSource(SiteDir, src)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	switch c.Name {
		case "edit":
			util.OpenInEditor(p.SourcePath)
		case "rm":
			if err := p.Remove(); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
	}
}
