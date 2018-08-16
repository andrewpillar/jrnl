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
		"post":     postTmpl,
		"index":    indexTmpl,
		"category": categoryTmpl,
	}

	CategoryTemplate = TemplatesDir + "/category"

	IndexTemplate = TemplatesDir + "/index"

	PostTemplate = TemplatesDir + "/post"
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
