package cmd

import (
	"os"

	"github.com/alecthomas/chroma/styles"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/config"
	"github.com/andrewpillar/jrnl/internal/render"
)

func GenStyle(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	style := styles.Get(c.Args.Get(0))

	if style == nil {
		style = styles.Fallback
	}

	r := render.New()
	r.Formatter.WriteCSS(os.Stdout, style)
}