package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func templateExists(tmpl string) bool {
	for t, _ := range Templates {
		if t == tmpl {
			return true
		}
	}

	return false
}

func Template(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.Tmpl)
		return
	}

	mustBeInitialized()

	tmpl := c.Args.Get(0)

	if !templateExists(tmpl) {
		fmt.Fprintf(os.Stderr, "invalid template\n")
		fmt.Fprintf(os.Stderr, "pick one of: category, index, post\n")
		os.Exit(1)
	}

	fname := TemplatesDir + "/" + tmpl

	_, err := os.Stat(fname)

	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "could not find template %s\n", tmpl)
		os.Exit(1)
	}

	util.OpenInEditor(fname)
}
