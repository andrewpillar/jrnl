package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/usage"
)

func Title(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.Title)
		os.Exit(1)
	}

	f, err := os.OpenFile(meta.File, os.O_RDWR, 0660)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	defer f.Close()

	m, err := meta.Decode(f)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if err := f.Truncate(0); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	_, err = f.Seek(0, 0)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	m.Title = c.Args.Get(0)

	if err = m.Encode(f); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
