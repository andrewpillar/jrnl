package cmd

import (
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/util"
)

func Ls(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		util.ExitError("not initialized", err)
	}

	pages, err := page.All()

	if err != nil {
		util.ExitError("failed to get all pages", err)
	}

	posts, err := post.All()

	if err != nil {
		util.ExitError("failed to get all posts", err)
	}

	for _, p := range pages {
		fmt.Println(p.ID)
	}

	category := c.Flags.GetString("category")

	for _, p := range posts {
		if category == "" {
			fmt.Println(p.ID)
			continue
		}

		if strings.ToLower(category) == strings.ToLower(p.Category.Name) {
			fmt.Println(p.ID)
		}
	}
}
