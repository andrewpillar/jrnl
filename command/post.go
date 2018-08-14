package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func Post(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.Post)
		return
	}

	mustBeInitialized()

	title := c.Args.Get(0)

	p := post.New(SiteDir, PostsDir, c.Flags.GetString("category"), title)

	dir := filepath.Dir(p.SourcePath)

	d, err := os.Stat(dir)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else {
		if !d.IsDir() {
			fmt.Fprintf(os.Stderr, "%s is not a directory\n", d.Name())
			os.Exit(1)
		}
	}

	f, err := os.OpenFile(p.SourcePath, os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	defer f.Close()

	f.Write([]byte("# " + title + "\n\n\n"))

	util.OpenInEditor(p.SourcePath)

	fmt.Fprintf(os.Stdout, "new post added: %s\n", p.ID)
}
