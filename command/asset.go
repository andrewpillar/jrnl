package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func Asset(c cli.Command) {
	fmt.Println(usage.Asset)
}

func AssetLs(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.AssetLs)
		return
	}
}

func AssetAdd(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.AssetAdd)
		return
	}

	asset := c.Args.Get(0)

	file := c.Flags.GetString("file")

	target := AssetsDir

	if c.Flags.GetString("target") != "" {
		target = AssetsDir + "/" + c.Flags.GetString("target")
	}

	if file == "" {
		if err := os.MkdirAll(target, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		util.OpenInEditor(target + "/" + asset)

		return
	}

	if err := util.Copy(file, target + "/" + asset); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func AssetEdit(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.AssetEdit)
		return
	}

	if err := util.OpenInEditor(AssetsDir + "/" + c.Args.Get(0)); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func AssetRm(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.AssetAdd)
		return
	}
}
