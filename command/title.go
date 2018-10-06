package command

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
)

func Title(c cli.Command) {
	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
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
		util.Exit("failed to save meta file", err)
	}

	m.Close()
}
