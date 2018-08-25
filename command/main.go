package command

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"
)

func Main(c cli.Command) {
	fmt.Println(usage.Main)
}
