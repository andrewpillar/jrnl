package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func Post(c cli.Command) {
	util.MustBeInitialized()

	title := c.Args.Get(0)

	if title == "" {
		util.Exit("missing title for post", nil)
	}

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	if m.Editor == "" {
		util.Exit(
			"could not find editor",
			errors.New("set editor in _meta.yml"),
		)
	}

	m.Close()

	p := post.New(title, c.Flags.GetString("category"))

	dir := filepath.Dir(p.SourcePath)

	info, err := os.Stat(dir)

	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				util.Exit("failed to create directory", err)
			}
		} else {
			util.Exit("failed to stat directory", err)
		}
	}

	if info != nil && !info.IsDir() {
		util.Exit("unexpected non-directory file", errors.New(dir))
	}

	f, err := os.OpenFile(p.SourcePath, os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		util.Exit("failed to open post file", err)
	}

	if err := p.WriteFrontMatter(f); err != nil {
		util.Exit("failed to write front matter", err)
	}

	util.OpenInEditor(m.Editor, p.SourcePath)

	f.Close()

	fmt.Println("new post added", p.ID)
}
