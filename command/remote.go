package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

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
}

func RemoteSet(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.RemoteSet)
		return
	}

	if c.Args.Get(1) == "" {
		fmt.Fprintf(os.Stderr, "missing remote target\n")
		os.Exit(1)
	}

	f, err := os.OpenFile(Remotes, os.O_CREATE|os.O_RDWR, 0660)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	defer f.Close()

	f.Write([]byte())
}

func RemoteRemove(c cli.Command) {
	fmt.Println(usage.RemoteRm)
}
