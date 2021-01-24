package main

import "fmt"

var (
	build   string
	version string

	VersionCmd = &Command{
		Usage: "version",
		Short: "display version information",
		Run:   versionCmd,
	}
)

func versionCmd(cmd *Command, _ []string) {
	fmt.Println(cmd.Argv0, version, build)
}
