package command

import (
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/usage"
)

func List(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Ls)
		return
	}

	mustBeInitialized()

	r := post.NewResolver()

	posts := r.Resolve()

	category := strings.Replace(c.Flags.GetString("category"), "/", " ", -1)

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
