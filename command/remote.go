package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/usage"
)

func Remote(c cli.Command) {
	fmt.Println(usage.Remote)
}

func RemoteList(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.RemoteLs)
		return
	}

	mustBeInitialized()

	f, err := os.Open(meta.File)

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

	for k, v := range m.Remotes {
		fmt.Printf("%s - %s", k, v.Target)

		if m.Default == k {
			fmt.Printf("    [default]")
		}

		fmt.Printf("\n")
	}
}

func RemoteSet(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.RemoteSet)
		return
	}

	alias := c.Args.Get(0)
	target := c.Args.Get(1)

	if target == "" {
		fmt.Fprintf(os.Stderr, "missing remote target\n")
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

	r := meta.Remote{Target: target}

	if c.Flags.GetString("identity") != "" {
		r.Identity = c.Flags.GetString("identity")
	}

	m.Remotes[alias] = r

	if c.Flags.IsSet("default") {
		m.Default = alias
	}

	if err := m.Encode(f); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func RemoteRemove(c cli.Command) {
	fmt.Println(usage.RemoteRm)
}
