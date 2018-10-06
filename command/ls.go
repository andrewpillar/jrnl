package command

import (
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func printPost(p post.Post, verbose bool) {
	if verbose {
		fmt.Println("---")
		fmt.Println("ID:       ", p.ID)
		fmt.Println("Title:    ", p.Title)
		fmt.Println("Category: ", p.Category.Name)
		fmt.Println("Source:   ", p.SourcePath)
		fmt.Println("Site:     ", p.SitePath)
		fmt.Println("Created:  ", p.CreatedAt.Format(post.DateLayout))
		fmt.Println("Updated:  ", p.UpdatedAt.Format(post.DateLayout))
		return
	}

	fmt.Println(p.ID)
}

func Ls(c cli.Command) {
	util.MustBeInitialized()

	posts, err := post.ResolvePosts()

	if err != nil {
		util.Exit("failed to resolve posts", err)
	}

	category := c.Flags.GetString("category")
	verbose := c.Flags.IsSet("verbose")

	for _, p := range posts {
		if category == "" {
			printPost(p, verbose)
			continue
		}

		if strings.ToLower(category) == strings.ToLower(p.Category.Name) {
			printPost(p, verbose)
		}
	}
}
