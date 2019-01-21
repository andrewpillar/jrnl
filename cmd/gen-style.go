package cmd

import (
	"os"

	"github.com/alecthomas/chroma/styles"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/render"
	"github.com/andrewpillar/jrnl/util"
)

func GenStyle(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	style := styles.Get(c.Args.Get(0))

	if style == nil {
		style = styles.Fallback
	}

	r := render.New()
	r.Formatter.WriteCSS(os.Stdout, style)
}
