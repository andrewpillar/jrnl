package command

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func Title(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Title)
		return
	}

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	m.Title = c.Args.Get(0)

	if err := m.Save(); err != nil {
		util.Error("failed to save meta file", err)
	}

	m.Close()
}
