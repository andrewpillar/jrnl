package command

import (
	"fmt"
	"io"
	"os"

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

	f, err := os.OpenFile(meta.File, os.O_RDWR, os.ModePerm)

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	defer f.Close()

	m, err := meta.Decode(f)

	if err != nil {
		util.Error("failed to read meta file", err)
	}

	if len(c.Args) == 0 {
		if m.Title == "" {
			fmt.Println("no journal title set, run 'jrnl title' to set one")
			return
		}

		fmt.Println(m.Title)
		return
	}

	if err := f.Truncate(0); err != nil {
		util.Error("failed to truncate meta file", err)
	}

	_, err = f.Seek(0, io.SeekStart)

	if err != nil {
		util.Error("failed to seek beginning of meta file", err)
	}

	m.Title = c.Args.Get(0)

	if err = m.Encode(f); err != nil {
		util.Error("failed to write to meta file", err)
	}
}
