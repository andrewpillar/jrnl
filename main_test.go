package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

//go:embed testdata/embed/*
var testdataEmbed embed.FS

type TestCommand struct {
	Command *Command
	Args    []string
	Checks  []func(*testing.T)
}

func (c *TestCommand) Test(t *testing.T) {
	if err := c.Command.Run(c.Command, c.Args); err != nil {
		cmdline := fmt.Sprintf("%s %s", c.Command.Argv0, c.Args)

		t.Fatalf("command %q failed\n\t%s\n", cmdline, err)
	}

	for _, check := range c.Checks {
		check(t)
	}
}

func checkPathExists(path string) func(*testing.T) {
	return func(t *testing.T) {
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	}
}

func writeToFile(path, embed string) error {
	dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(0600))

	if err != nil {
		return err
	}

	defer dst.Close()

	src, err := testdataEmbed.Open(filepath.Join("testdata", "embed", embed))

	if err != nil {
		return err
	}

	defer src.Close()

	if _, err := dst.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

func writePage(page string) error {
	return writeToFile(filepath.Join(pageDir, page), page)
}

func writePost(post string) error {
	slug := time.Now().Format("2006-01-02") + "-" + post

	return writeToFile(filepath.Join(postDir, slug), "post.md")
}

func Test_Jrnl(t *testing.T) {
	if err := os.Chdir("testdata"); err != nil {
		t.Fatal(err)
	}

	os.Setenv("EDITOR", "true")
	os.Setenv("HOME", ".")

	cmds := [...]TestCommand{
		{
			Command: InitCmd,
		},
		{
			Command: &Command{
				Argv0: "update jrnl config",
				Run: func(cmd *Command, args []string) error {
					cfg, err := LoadConfig()

					if err != nil {
						return err
					}

					cfg.Remote = "file://remote"
					cfg.Author.Name = "Joe Bloggs"
					cfg.Author.Email = "joe.bloggs@localhost"
					cfg.Site.Title = "Joe's Blog"
					cfg.Site.URL = "http://localhost:8080"
					cfg.Site.Atom = "atom.xml"
					cfg.Site.RSS = "rss.xml"

					return cfg.Save()
				},
			},
		},
		{
			Command: &Command{
				Argv0: "create assets",
				Run: func(cmd *Command, args []string) error {
					f, err := os.Create(filepath.Join(assetDir, "main.css"))

					if err != nil {
						return err
					}

					defer f.Close()

					_, err = io.WriteString(f, "* {\n\tmargin: 0:\n}\n")
					return err
				},
			},
		},
		{
			Command: PageCmd,
			Args:    []string{"-l", "page", "About"},
		},
		{
			Command: &Command{
				Argv0: "writePage",
				Run: func(cmd *Command, args []string) error {
					return writePage(args[0])
				},
			},
			Args: []string{"about.md"},
		},
		{
			Command: PostCmd,
			Args:    []string{"-l", "post", "Hello, world"},
		},
		{
			Command: &Command{
				Argv0: "writePost",
				Run: func(cmd *Command, args []string) error {
					return writePost(args[0])
				},
			},
			Args: []string{"hello-world.md"},
		},
		{
			Command: PostCmd,
			Args:    []string{"-l", "post", "-p", "programming", "Go 101"},
		},
		{
			Command: &Command{
				Argv0: "writePost",
				Run: func(cmd *Command, args []string) error {
					return writePost("go-101.md")
				},
			},
		},
		{
			Command: PageCmd,
			Args:    []string{"-l", "parent", "Programming"},
		},
		{
			Command: ThemeSaveCmd,
			Args:    []string{"default"},
		},
		{
			Command: ThemeLsCmd,
		},
		{
			Command: &Command{
				Argv0: "os.Remove",
				Run: func(cmd *Command, args []string) error {
					for _, path := range args {
						if err := os.Remove(path); err != nil {
							return err
						}
					}
					return nil
				},
			},
			Args: []string{
				filepath.Join(layoutDir, "parent"),
				filepath.Join(assetDir, "main.css"),
			},
		},
		{
			Command: ThemeUseCmd,
			Args:    []string{"default"},
			Checks: []func(*testing.T){
				checkPathExists(filepath.Join(layoutDir, "parent")),
				checkPathExists(filepath.Join(assetDir, "main.css")),
			},
		},
		{
			Command: ThemeRmCmd,
			Args:    []string{"default"},
		},
		{
			Command: PublishCmd,
			Args:    []string{"-d", "-v"},
		},
		{
			Command: FlushCmd,
		},
		{
			Command: PageCmd,
			Args:    []string{"-l", "page", "test"},
		},
		{
			Command: PublishCmd,
			Args:    []string{"-v"},
		},
		{
			Command: RmCmd,
			Args:    []string{"test"},
		},
	}

	for _, cmd := range cmds {
		cmd.Test(t)
	}
}
