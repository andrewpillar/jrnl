package main

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/command"
	"github.com/andrewpillar/jrnl/post"
)

func main() {
	post.SourceDir = "_posts"
	post.SiteDir = "_site"

	c := cli.New()

	c.Main(command.Main)

	c.AddFlag(&cli.Flag{
		Name: "help",
		Long: "--help",
	})

	c.Command("init", command.Initialize)
	c.Command("title", command.Title)
	c.Command("tmpl", command.Template)

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

	remoteCmd.Command("ls", command.RemoteList)

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

	remoteCmd.Command("rm", command.RemoteRemove)

	publishCmd := c.Command("publish", command.Publish)

	publishCmd.AddFlag(&cli.Flag{
		Name:  "draft",
		Short: "-d",
		Long:  "--draft",
	})

	publishCmd.AddFlag(&cli.Flag{
		Name:  "remote",
		Short: "-r",
		Long:  "--remote",
	})

	assetCmd := c.Command("asset", command.Asset)

	assetCmd.Command("ls", command.AssetLs)

	assetAddCmd := assetCmd.Command("add", command.AssetAdd)

	assetAddCmd.AddFlag(&cli.Flag{
		Name:     "file",
		Short:    "-f",
		Long:     "--file",
		Argument: true,
		Default:  "",
	})

	assetAddCmd.AddFlag(&cli.Flag{
		Name:     "target",
		Short:    "-t",
		Long:     "--target",
		Argument: true,
		Default:  "",
	})

	assetCmd.Command("edit", command.AssetEdit)
	assetCmd.Command("rm", command.AssetRm)

	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
