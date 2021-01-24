package main

import (
	"fmt"
	"os"
)

var RmCmd = &Command{
	Usage: "rm <page|post,...>",
	Short: "remove the given page or post",
	Long: `Rm will remove the given page or post. This will remove the generated site page
too if one exists. If a post is removed that would be the last post in a given
category, then that category will be removed too.`,
	Run: rmCmd,
}

func rmCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	cfg, err := OpenConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open config: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	code := 0

	rmpaths := make([]string, 0, len(args[1:]))

	for _, id := range args[1:] {
		page, ok, err := GetPage(id)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: failed to get page: %s\n", cmd.Argv0, args[0], err)
			continue
		}

		if !ok {
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

			if err := post.Remove(); err != nil {
				code = 1
				fmt.Fprintf(os.Stderr, "%s %s: failed to remove post %q: %s\n", cmd.Argv0, args[0], id, err)
			}
			rmpaths = append(rmpaths, post.SitePath)
			continue
		}

		if err := page.Remove(); err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s %s: failed to remove page %q: %s\n", cmd.Argv0, args[0], id, err)
		}
		rmpaths = append(rmpaths, page.SitePath)
	}

	rem, err := OpenRemote(cfg.Site.Remote)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open remote: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	defer rem.Close()

	for _, path := range rmpaths {
		if err := rem.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "%s %s: failed to remove %q from remote: %s\n", cmd.Argv0, args[0], path, err)
				code = 1
			}
		}
	}
	os.Exit(code)
}
