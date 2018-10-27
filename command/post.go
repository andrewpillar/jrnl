package command

import "github.com/andrewpillar/cli"

func Post(c cli.Command) {
	createPage(c, true)
}
