package command

import (
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func List(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Ls)
		return
	}

	mustBeInitialized()

	posts, err := post.ResolvePosts()

	if err != nil {
		util.Error("failed to resolve posts", err)
	}

	category := strings.Replace(c.Flags.GetString("category"), "/", " ", -1)

	for _, p := range posts {
		if category == "" {
			fmt.Println(p.ID)
			continue
		}

		if strings.ToLower(p.Category.Name) == strings.ToLower(category) {
			fmt.Println(p.ID)
		}
	}
}
