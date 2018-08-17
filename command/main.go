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

	Templates = map[string]string{
		"post":     postTmpl,
		"index":    indexTmpl,
		"category": categoryTmpl,
	}

	CategoryTemplate = TemplatesDir + "/category"

	IndexTemplate = TemplatesDir + "/index"

	PostTemplate = TemplatesDir + "/post"
)

func Main(c cli.Command) {
	fmt.Println(usage.Main)
}
