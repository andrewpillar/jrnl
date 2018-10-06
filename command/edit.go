package command

import (
	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func Edit(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	p, err := post.Find(c.Args.Get(0))

	if err != nil {
		util.Exit("failed to find post", err)
	}

	util.OpenInEditor(m.Editor, p.SourcePath)
}
