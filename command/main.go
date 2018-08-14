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

	MetaFile = "_meta.yml"

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

type meta struct {
	Title string `yaml:",omitempty"`

	Remotes []remote `yaml:",omitempty"`
}

type remote struct {
	Alias string

	Target string
}

func Main(c cli.Command) {
	fmt.Println(usage.Main)
}
