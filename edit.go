package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var EditCmd = &Command{
	Usage: "edit <page|post>",
	Short: "edit a page or post",
	Long: `Edit will open up the editor specified via the EDITOR environment variable for
editting the given page or post. This will search for the page to edit first,
then search for the post to edit, then error out if neither could be found.`,
	Run: editCmd,
}

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return errors.New("EDITOR not set")
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func editCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	page, ok, err := GetPage(args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to get page: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	var path string

	if ok {
		path = page.SourcePath
	} else {
		post, ok, err := GetPost(args[1])

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to get post: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}

		if !ok {
			fmt.Fprintf(os.Stderr, "%s %s: no such page or post\n", cmd.Argv0, args[0])
			os.Exit(1)
		}

		if err := post.Touch(); err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to touch post: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}

		path = post.SourcePath
	}

	if err := openInEditor(path); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open editor: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}
