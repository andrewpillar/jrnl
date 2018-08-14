package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"
)

var (
	PostTemplate = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title></title>
		<link rel="stylesheet" type="text/css" href="main.css"/>
	</head>
	<body>
	</body>
</html>`

	IndexTemplate = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title></title>
		<link rel="stylesheet" type="text/css" href="main.css"/>
	</head>
	<body>
	</body>
</html>`

	CategoryTemplate = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title></title>
		<link rel="stylesheet" type="text/css" href="main.css"/>
	</head>
	<body>
	</body>
</html>`
)

func isInitialized() bool {
	for _, d := range Dirs {
		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			return false
		}

		if !f.IsDir() {
			return false
		}
	}

	return true
}

func mustBeInitialized() {
	for _, d := range Dirs {
		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			fmt.Fprintf(
				os.Stderr,
				"journal not fully initialized, run 'jrnl init'\n",
			)
			os.Exit(1)
		}

		if !f.IsDir() {
			fmt.Fprintf(os.Stderr, "journal incorrectly initialized\n")
			os.Exit(1)
		}
	}
}

func Initialize(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Init)
		return
	}

	if isInitialized() {
		fmt.Fprintf(os.Stderr, "jrnl already initialized\n")
		os.Exit(1)
	}

	for _, d := range Dirs {
		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			if err := os.Mkdir(d, os.ModePerm); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			continue
		}

		if !f.IsDir() {
			fmt.Fprintf(
				os.Stderr,
				"jrnl already partially, or fully initialized\n",
			)
			os.Exit(1)
		}
	}

	for k, v := range Templates {
		path := TemplatesDir + "/" + k

		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		defer f.Close()

		_, err = f.Write([]byte(v))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stdout, "jrnl initialized\n")
}
