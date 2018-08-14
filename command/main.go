package command

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"
)

var (
	PostsDir = "_posts"

	SiteDir = "_site"

	TemplatesDir = "_templates"

	Dirs = []string{
		PostsDir,
		SiteDir,
		TemplatesDir,
	}

	Remotes = "_remotes"

	Templates = map[string]string{
		"post":     PostTemplate,
		"index":    IndexTemplate,
		"category": CategoryTemplate,
	}
)

func Main(c cli.Command) {
	fmt.Println(usage.Main)
}
