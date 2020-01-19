package cmd

import (
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
)

func Ls(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	pages, err := blog.Pages()

	if err != nil {
		exitError("failed to get all pages", err)
	}

	posts, err := blog.Posts()

	if err != nil {
		exitError("failed to get all posts", err)
	}

	for _, p := range pages {
		fmt.Println(p.ID)
	}

	category := c.Flags.GetString("category")

	for _, p := range posts {
		shouldPrint := true

		if category != "" {
			shouldPrint = strings.ToLower(category) == strings.ToLower(p.Category.Name)
		}

		if shouldPrint {
			fmt.Println(p.ID)
		}
	}
}
