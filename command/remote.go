package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"
)

var (
	paddedFmt = "%%%ds - %%%ds"

	errRemoteFindFmt = "%s: could not find remote: %s\n"
)

func RemoteLs(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	m.Close()

	verbose := c.Flags.IsSet("verbose")

	aliasLen := 0
	targetLen := 0

	for k, v := range m.Remotes {
		if len(k) > aliasLen {
			aliasLen = len(k)
		}

		if len(v.Target) > targetLen {
			targetLen = len(v.Target)
		}
	}

	for k, v := range m.Remotes {
		if verbose {
			fmt.Println("---")
			fmt.Printf("Alias:    %s", k)

			if m.Default == k {
				fmt.Printf("  [default]")
			}

			fmt.Printf("\n")

			fmt.Println("Target:  ", v.Target)

			if !filepath.IsAbs(v.Target) {
				fmt.Println("Port:    ", v.Port)
				fmt.Println("Identity:", v.Identity)
			}
		} else {
			fmt.Printf(fmt.Sprintf(paddedFmt, -aliasLen, -targetLen), k, v.Target)

			if m.Default == k {
				fmt.Printf("  [default]")
			}

			fmt.Printf("\n")
		}
	}
}

func RemoteSet(c cli.Command) {
	util.MustBeInitialized()

	alias := c.Args.Get(0)
	target := c.Args.Get(1)

	if alias == "" {
		util.Exit("missing alias", nil)
	}

	if target == "" {
		util.Exit("missing target", nil)
	}

	port, err := c.Flags.GetInt("port")

	if err != nil {
		util.Exit("failed to get port number", err)
	}

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	defer m.Close()

	r := meta.Remote{
		Target: target,
		Port:   port,
	}

	identity := c.Flags.GetString("identity")

	if identity != "" {
		r.Identity = identity
	}

	m.Remotes[alias] = r

	if c.Flags.IsSet("default") {
		m.Default = alias
	}

	if err := m.Save(); err != nil {
		util.Exit("failed to save meta file", err)
	}
}

func RemoteRm(c cli.Command) {
	util.MustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Exit("failed to open meta file", err)
	}

	code := 0

	for _, alias := range c.Args {
		_, ok := m.Remotes[alias]

		if !ok {
			fmt.Fprintf(os.Stderr, errRemoteFindFmt, os.Args[0], alias)

			code = 1
			continue
		}

		if alias == m.Default {
			m.Default = ""
		}

		delete(m.Remotes, alias)
	}

	if err := m.Save(); err != nil {
		util.Exit("failed to save meta file", err)
	}

	os.Exit(code)
}
