package command

import (
	"fmt"
	"os"
	"sort"
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

	posts := r.ResolvePostsToMap()

	ids := make([]string, len(posts), len(posts))

	i := 0

	for k := range posts {
		ids[i] = k

		i++
	}

	category := c.Flags.GetString("category")

	sort.Strings(ids)

	for _, id := range ids {
		p := posts[id]

		if category == "" {
			fmt.Fprintf(os.Stdout, "%s\n", p.ID)
			continue
		}

		if strings.ToLower(p.Category) == category {
			fmt.Fprintf(os.Stdout, "%s\n", p.ID)
		}
	}
}
