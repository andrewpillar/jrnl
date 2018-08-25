package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
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

	filepath.Walk(meta.AssetsDir, walk)
}

func AssetAdd(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.AssetAdd)
		return
	}

	asset := c.Args.Get(0)
	file := c.Flags.GetString("file")
	dir := filepath.Join(meta.AssetsDir, c.Flags.GetString("dir"))

	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			util.Error("failed to create asset directory", err)
		}
	}

	if file == "" {
		if asset == "" {
			util.Error("missing asset argument", nil)
		}

		util.OpenInEditor(filepath.Join(dir, asset))

		return
	}

	if asset == "" {
		asset = file
	}

	if err := util.Copy(file, filepath.Join(dir, asset)); err != nil {
		util.Error("failed to copy file to asset directory", err)
	}
}

func AssetEdit(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.AssetEdit)
		return
	}

	util.OpenInEditor(filepath.Join(meta.AssetsDir, c.Args.Get(0)))
}

func AssetRm(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) == 0 {
		fmt.Println(usage.AssetRm)
		return
	}
}

func walk(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}

	parts := strings.Split(path, string(os.PathSeparator))

	fmt.Println(filepath.Join(parts[2:]...))

	return nil
}
