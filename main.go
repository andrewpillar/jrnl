package main

import (
	"fmt"
	"os"
)

func run(args []string) error {

	cmds := &CommandSet{
		Argv0: args[0],
		Long: `jrnl is a simple static site generator.

Usage:

    jrnl <command> [arguments]
`,
	}

	cmds.Add("cat", CatCmd)
	cmds.Add("config", ConfigCmd)
	cmds.Add("edit", EditCmd)
	cmds.Add("flush", FlushCmd)
	cmds.Add("init", InitCmd)
	cmds.Add("ls", LsCmd)
	cmds.Add("page", PageCmd)
	cmds.Add("post", PostCmd)
	cmds.Add("publish", PublishCmd)
	cmds.Add("rm", RmCmd)
	cmds.Add("theme", ThemeCmd(cmds.Argv0))
	cmds.Add("version", VersionCmd)

	cmds.Add("help", HelpCmd(cmds))

	return cmds.Parse(args[1:])
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
