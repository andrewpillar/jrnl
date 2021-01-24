package main

import (
	"fmt"
	"io"
	"os"
)

var CatCmd = &Command{
	Usage: "cat <page|post,...>",
	Short: "display the contents of a page or post",
	Long:  `Cat will display the contents of the given pages or posts.`,
	Run:   catCmd,
}

func catCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	code := 0

	for _, id := range args[1:] {
		var path string

		page, ok, err := GetPage(id)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: failed to get page: %s\n", cmd.Argv0, args[0], err)
			continue
		}

		if ok {
			path = page.SourcePath
		} else {
			post, ok, err := GetPost(id)

			if err != nil {
				code = 1
				fmt.Fprintf(os.Stderr, "%s %s: failed to get post: %s\n", cmd.Argv0, args[0], err)
				continue
			}

			if !ok {
				code = 1
				fmt.Fprintf(os.Stderr, "%s %s: no such page or post\n", cmd.Argv0, args[0])
				continue
			}
			path = post.SourcePath
		}

		func(path string) {
			f, err := os.Open(path)

			if err != nil {
				code = 1
				fmt.Fprintf(os.Stderr, "%s %s: no such page or post\n", cmd.Argv0, args[0])
				return
			}

			defer f.Close()
			io.Copy(os.Stdout, f)
		}(path)
	}
	os.Exit(code)
}
