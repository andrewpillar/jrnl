package command

import (
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/resolve"
	"github.com/andrewpillar/jrnl/usage"
)

func List(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Ls)
		return
	}

	mustBeInitialized()

	r := resolve.New(SiteDir, PostsDir)

	posts := r.ResolvePostsToStore()

	posts.Sort()

	category := c.Flags.GetString("category")

	for _, p := range posts {
		if category == "" {
			fmt.Println(p.ID)
			continue
		}

		if strings.ToLower(p.Category) == strings.ToLower(category) {
			fmt.Println(p.ID)
		}
	}
}
