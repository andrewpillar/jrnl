package command

import (
	"errors"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/page"
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

	id := c.Args.Get(0)

	if id == "" {
		util.Exit("failed to find page to edit", errors.New("missing id"))
	}

	p, err := page.Find(id)

	if err != nil {
		if os.IsNotExist(err) {
			pst, err := post.Find(id)

			if err != nil {
				util.Exit("failed to find post", err)
			}

			if err := pst.Touch(); err != nil {
				util.Exit("failed to touch post", err)
			}

			util.OpenInEditor(m.Editor, pst.SourcePath)
			return
		}

		util.Exit("failed to find page", err)
	}

	if err := p.Touch(); err != nil {
		util.Exit("failed to touch page", err)
	}

	util.OpenInEditor(m.Editor, p.SourcePath)
}
