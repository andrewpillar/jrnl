package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func Remote(c cli.Command) {
	fmt.Println(usage.Remote)
}

func RemoteLs(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.RemoteLs)
		return
	}

	mustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
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
		util.Error("missing remote target", nil)
	}

	port, err := c.Flags.GetInt("port")

	if err != nil {
		util.Error("failed to get port number from flag", err)
	}

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	defer m.Close()

	r := meta.Remote{
		Target: target,
		Port:   port,
	}

	if c.Flags.GetString("identity") != "" {
		r.Identity = c.Flags.GetString("identity")
	}

	m.Remotes[alias] = r

	if c.Flags.IsSet("default") {
		m.Default = alias
	}

	if err := m.Save(); err != nil {
		util.Error("faild to save meta file", err)
	}
}

func RemoteRm(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.RemoteRm)
		return
	}

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	defer m.Close()

	code := 0

	for _, alias := range c.Args {
		_, ok := m.Remotes[alias]

		if !ok {
			fmt.Fprintf(os.Stderr, "jrnl: could not find remote: %s\n", alias)

			code = 1

			continue
		}

		if alias == m.Default {
			m.Default = ""
		}

		delete(m.Remotes, alias)
	}

	if err := m.Save(); err != nil {
		util.Error("faild to save meta file", err)
	}

	os.Exit(code)
}
