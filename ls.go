package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
)

var LsCmd = &Command{
	Usage: "ls",
	Short: "list the pages and posts of the journal",
	Long:  `ls will list the pages and posts of the journal. The pages of the journal will
be listed first, followed by the posts.

The -c flag can be given to only display posts in the given category. The -h
flag can be given to hide pages, and display only posts. The -v flag can be
given to detail hash information about each page or post. This will display
whether or not the current item has been modified along with its current
hash.`,
	Run:   lsCmd,
}

func printHashInfo(hash *Hash, argv0, id string, h Hasher) {
	status := "unmodified "

	b, _ := hash.Get(id)

	if !bytes.Equal(b, h.Hash()) {
		status = "modified   "
	}

	hex := hex.EncodeToString(b)

	if hex == "" {
		hex = "000000000000"
	}
	fmt.Println(status, id, hex)
}

func lsCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	var (
		category string
		hide     bool
		verbose  bool
	)

	fs := flag.NewFlagSet(cmd.Argv0 + " " + args[0], flag.ExitOnError)
	fs.StringVar(&category, "c", "", "display only posts in the category")
	fs.BoolVar(&hide, "h", false, "don't display pages")
	fs.BoolVar(&verbose, "v", false, "display hash information about the page or post")
	fs.Parse(args[1:])

	pages, err := Pages()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to find pages: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	posts, err := Posts()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to find posts: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	hash, err := OpenHash()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open hash: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	defer hash.Close()

	if !hide {
		for _, page := range pages {
			if verbose {
				printHashInfo(hash, cmd.Argv0 + " " + args[0], page.ID, page)
				continue
			}
			fmt.Println(page.ID)
		}
	}

	category = strings.ToLower(category)

	for _, post := range posts {
		print_ := true

		if category != "" {
			print_ = category == strings.ToLower(post.Category.Name)
		}

		if print_ {
			if verbose {
				printHashInfo(hash, cmd.Argv0 + " " + args[0], post.ID, post)
				continue
			}
			fmt.Println(post.ID)
		}
	}
}
