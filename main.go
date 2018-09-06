package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/command"
	"github.com/andrewpillar/jrnl/meta"
)

func main() {
	meta.PostsDir = "_posts"
	meta.SiteDir = "_site"
	meta.LayoutsDir = "_layouts"
	meta.AssetsDir = filepath.Join(meta.SiteDir, "assets")
	meta.ThemesDir = "_themes"

	meta.Dirs = []string{
		meta.PostsDir,
		meta.SiteDir,
		meta.LayoutsDir,
		meta.AssetsDir,
		meta.ThemesDir,
	}

	c := cli.New()

	c.Main(command.Main)

	c.AddFlag(&cli.Flag{
		Name: "help",
		Long: "--help",
	})

	c.Command("init", command.Initialize)
	c.Command("title", command.Title)

	postCmd := c.Command("post", command.Post)

	postCmd.AddFlag(&cli.Flag{
		Name:     "category",
		Short:    "-c",
		Long:     "--category",
		Argument: true,
		Default:  "",
	})

	c.Command("edit", command.ChangePost)
	c.Command("rm", command.ChangePost)

	listCmd := c.Command("ls", command.List)

	listCmd.AddFlag(&cli.Flag{
		Name:     "category",
		Short:    "-c",
		Long:     "--category",
		Argument: true,
		Default:  "",
	})

	remoteCmd := c.Command("remote", command.Remote)

	remoteCmd.Command("ls", command.RemoteLs)

	remoteSetCmd := remoteCmd.Command("set", command.RemoteSet)

	remoteSetCmd.AddFlag(&cli.Flag{
		Name:  "default",
		Short: "-d",
		Long:  "--default",
	})

	remoteSetCmd.AddFlag(&cli.Flag{
		Name:     "port",
		Short:    "-p",
		Long:     "--port",
		Argument: true,
		Default:  22,
	})

	remoteSetCmd.AddFlag(&cli.Flag{
		Name:     "identity",
		Short:    "-i",
		Long:     "--identity",
		Argument: true,
		Default:  "",
	})

	remoteCmd.Command("rm", command.RemoteRm)

	publishCmd := c.Command("publish", command.Publish)

	publishCmd.AddFlag(&cli.Flag{
		Name: "draft",
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
