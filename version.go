package main

import "fmt"

var (
	build   string
	version string

	VersionCmd = &Command{
		Usage: "version",
		Short: "display version information",
		Run: func(cmd *Command, args []string) error {
			fmt.Println(cmd.Argv0, version, build)
			return nil
		},
	}
)
