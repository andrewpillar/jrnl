package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func createPage(c cli.Command, isPost bool) {
	util.MustBeInitialized()

	title := c.Args.Get(0)

	if title == "" {
		util.Exit("failed to create page", errors.New("missing title"))
	}

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	if m.Editor == "" {
		util.Exit("failed to find editor", errors.New("no editor set in _meta.yml"))
	}

	p := page.New(title)
	dir := filepath.Dir(p.SourcePath)

	info, err := os.Stat(dir)

	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				util.Exit("failed to create directory", err)
			}
		} else {
			util.Exit("failed to stat directory", err)
		}
	}

	if info != nil && !info.IsDir() {
		util.Exit("unexpected non-directory file", err)
	}

	if !isPost {
		if err := p.Touch(); err != nil {
			util.Exit("failed to touch page", err)
		}

		util.OpenInEditor(m.Editor, p.SourcePath)
		fmt.Println("new page added", p.ID)

		return
	}

	pst := post.New(&p, c.Flags.GetString("category"))

	if err := pst.Touch(); err != nil {
		util.Exit("failed to touch post", err)
	}

	util.OpenInEditor(m.Editor, p.SourcePath)
	fmt.Println("new post added", pst.ID)
}

func Page(c cli.Command) {
	createPage(c, false)
}
