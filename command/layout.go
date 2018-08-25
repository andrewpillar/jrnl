package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
	"github.com/andrewpillar/jrnl/usage"
)

func Layout(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Layout)
		return
	}

	fmt.Println(usage.Layout)
}

func LayoutLs(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.LayoutLs)
		return
	}

	mustBeInitialized()

	for l := range meta.Layouts {
		fmt.Println(strings.Split(l, ".")[0])
	}
}

func LayoutEdit(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.LayoutEdit)
		return
	}

	mustBeInitialized()

	layout := c.Args.Get(0)

	_, ok := meta.Layouts[layout + ".html"]

	if !ok {
		util.Error("invalid layout", errors.New(layout))
	}

	fname := filepath.Join(meta.LayoutsDir, layout + ".html")

	_, err := os.Stat(fname)

	if err != nil {
		util.Error("failed to stat layout file", err)
	}

	util.OpenInEditor(fname)
}
