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

	title := c.Args.Get(0)

	if title == "" {
		if m.Title == "" {
			fmt.Println("title not set, set the title with 'jrnl title'")
			return
		}

		fmt.Println(m.Title)
		return
	}

	m.Title = title

	if err := m.Save(); err != nil {
		util.Error("failed to save meta file", err)
	}

	m.Close()
}
