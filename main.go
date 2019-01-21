package main

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/cmd"
)

func usageHandler(c cli.Command) {

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

	c.Command("init", cmd.Init)
	c.Command("title", cmd.Title)

	categoryFlag := &cli.Flag{
		Name:     "category",
		Short:    "-c",
		Long:     "--category",
		Argument: true,
	}

	c.Command("page", cmd.Page)
	c.Command("post", cmd.Post).AddFlag(categoryFlag)

	c.Command("ls", cmd.Ls).AddFlag(categoryFlag)

	c.Command("edit", cmd.Edit)
	c.Command("rm", cmd.Rm)

	c.Command("remote-set", cmd.RemoteSet)

	publishCmd := c.Command("publish", cmd.Publish)

	publishCmd.AddFlag(&cli.Flag{
		Name:  "draft",
		Short: "-d",
		Long:  "--draft",
	})

	themeCmd := c.Command("theme", cmd.Theme)

	themeCmd.Command("ls", cmd.ThemeLs)
	themeCmd.Command("save", cmd.ThemeSave)
	themeCmd.Command("use", cmd.ThemeUse)
	themeCmd.Command("rm", cmd.ThemeRm)

	c.Command("gen-style", cmd.GenStyle)

	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
