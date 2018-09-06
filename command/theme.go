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

	"github.com/mozillazg/go-slugify"
)

func themeWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	parts := strings.Split(path, string(os.PathSeparator))

	if len(parts) == 2 {
		theme := strings.Split(parts[1], ".")[0]

		fmt.Println(theme)
	}

	return nil
}

func Theme(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Theme)
		return
	}

	mustBeInitialized()

	m, err := meta.Open()
	m.Close()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	if m.Theme == "" {
		fmt.Println("no theme being used")
		return
	}

	fmt.Println("current theme: " + m.Theme)
}

func ThemeLs(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.ThemeLs)
	}

	mustBeInitialized()

	if err := filepath.Walk(meta.ThemesDir, themeWalk); err != nil {
		util.Error("error whilst walking themes", err)
	}
}

func ThemeSave(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.ThemeSave)
		return
	}

	mustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	defer m.Close()

	assetsDir := strings.TrimPrefix(
		strings.Replace(meta.AssetsDir, meta.SiteDir, "", -1),
		string(os.PathSeparator),
	)

	theme := c.Args.Get(0)

	if theme != "" {
		m.Theme = slugify.Slugify(theme)
	}

	if m.Theme == "" {
		util.Error("no theme specified", nil)
	}

	path := filepath.Join(meta.ThemesDir, m.Theme)

	tmp := filepath.Join(path, assetsDir)

	if err := util.Copy(meta.AssetsDir, tmp); err != nil {
		util.Error("failed to copy dir", err)
	}

	tmp = filepath.Join(path, meta.LayoutsDir)

	if err := util.Copy(meta.LayoutsDir, tmp); err != nil {
		util.Error("failed to copy dir", err)
	}

	f, err := os.OpenFile(
		path + ".tar.gz",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		0660,
	)

	if err != nil {
		util.Error("failed to open tarball", err)
	}

	defer f.Close()

	if err := util.Tar(path, f); err != nil {
		util.Error("failed to create tarball", err)
	}

	if err := os.RemoveAll(path); err != nil {
		util.Error("failed to remove dir", err)
	}

	if err := m.Save(); err != nil {
		util.Error("failed to save meta file", err)
	}

	fmt.Println("saved theme: " + m.Theme)
}

func ThemeUse(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) != 1 {
		fmt.Println(usage.ThemeUse)
		return
	}

	mustBeInitialized()

	m, err := meta.Open()

	if err != nil {
		util.Error("failed to open meta file", err)
	}

	defer m.Close()

	m.Theme = slugify.Slugify(c.Args.Get(0))

	path := filepath.Join(meta.ThemesDir, m.Theme + ".tar.gz")

	_, err = os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			util.Error("theme does not exist", nil)
		}

		util.Error("error stating tar", err)
	}

	f, err := os.Open(path)

	if err != nil {
		util.Error("failed to open tar", err)
	}

	defer f.Close()

	if err := util.Untar(meta.ThemesDir, f); err != nil {
		util.Error("failed to untar theme", err)
	}

	assetsDir := strings.TrimPrefix(
		strings.Replace(meta.AssetsDir, meta.SiteDir, "", -1),
		string(os.PathSeparator),
	)

	tmp := filepath.Join(meta.ThemesDir, assetsDir)

	if err := util.Copy(tmp, meta.AssetsDir); err != nil {
		util.Error("failed to copy dir", err)
	}

	tmp = filepath.Join(meta.ThemesDir, meta.LayoutsDir)

	if err := util.Copy(tmp, meta.LayoutsDir); err != nil {
		util.Error("failed to copy dir", err)
	}

	if err = m.Save(); err != nil {
		util.Error("failed to save meta file", err)
	}

	fmt.Println("using theme: " + m.Theme)
}

func ThemeRm(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.ThemeRm)
		return
	}

	mustBeInitialized()

	code := 0

	for _, theme := range c.Args {
		fname := filepath.Join(meta.ThemesDir, theme + ".tar.gz")

		if err := os.Remove(fname); err != nil {
			code = 1

			fmt.Fprintf(
				os.Stderr,
				"jrnl: failed to remove theme %s: %s\n",
				theme,
				err,
			)
		}
	}

	os.Exit(code)
}
