package command

import (
	"fmt"
	_ "os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"
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
}

func AssetEdit(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.AssetEdit)
		return
	}
}

func AssetRm(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.AssetAdd)
		return
	}
}
