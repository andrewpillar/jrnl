package main

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/command"
	"github.com/andrewpillar/jrnl/usage"
)

func usageHandler(c cli.Command) {
	if c.Name == "" {
		fmt.Println(usage.Jrnl)
		return
	}

	fmt.Println(usage.Commands[c.FullName()])
}

func main() {
	c := cli.New()

	c.AddFlag(&cli.Flag{
		Name:      "help",
		Long:      "--help",
		Exclusive: true,
		Handler:   func(f cli.Flag, c cli.Command) {
			usageHandler(c)
		},
	})

	c.NilHandler(usageHandler)

	c.Main(nil)

	c.Command("init", command.Init)
	c.Command("title", command.Title)

	postCmd := c.Command("post", command.Post)

	postCmd.AddFlag(&cli.Flag{
		Name:     "category",
		Short:    "-c",
		Long:     "--category",
		Argument: true,
		Default:  "",
	})

	c.Command("page", command.Page)

	c.Command("edit", command.Edit)
	c.Command("rm", command.Rm)

	lsCmd := c.Command("ls", command.Ls)

	lsCmd.AddFlag(&cli.Flag{
		Name:     "category",
		Short:    "-c",
		Long:     "--category",
		Argument: true,
		Default:  "",
	})

	remoteCmd := c.Command("remote", nil)

	remoteLsCmd := remoteCmd.Command("ls", command.RemoteLs)

	remoteLsCmd.AddFlag(&cli.Flag{
		Name:  "verbose",
		Short: "-v",
		Long:  "--verbose",
	})

	remoteSetCmd := remoteCmd.Command("set", command.RemoteSet)

	remoteSetCmd.AddFlag(&cli.Flag{
		Name:  "default",
		Short: "-d",
		Long:  "--default",
	})

	remoteSetCmd.AddFlag(&cli.Flag{
		Name:     "identity",
		Short:    "-i",
		Long:     "--identity",
		Argument: true,
		Default:  "",
	})

	remoteSetCmd.AddFlag(&cli.Flag{
		Name:     "port",
		Short:    "-p",
		Long:     "--port",
		Argument: true,
		Default:  22,
	})

	remoteCmd.Command("rm", command.RemoteRm)

	publishCmd := c.Command("publish", command.Publish)

	publishCmd.AddFlag(&cli.Flag{
		Name:  "draft",
		Short: "-d",
		Long:  "--draft",
	})

	publishCmd.AddFlag(&cli.Flag{
		Name:     "remote",
		Short:    "-r",
		Long:     "--remote",
		Argument: true,
		Default:  "",
	})

	themeCmd := c.Command("theme", command.Theme)

	themeCmd.Command("ls", command.ThemeLs)
	themeCmd.Command("save", command.ThemeSave)
	themeCmd.Command("use", command.ThemeUse)
	themeCmd.Command("rm", command.ThemeRm)

	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
