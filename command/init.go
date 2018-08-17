package command

import (
	"errors"
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/usage"
)

var (
	postTmpl = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title></title>
		<link rel="stylesheet" type="text/css" href="main.css"/>
	</head>
	<body>
		<h1>{{.Title}}</h1>
		<div>{{.Body}}</div>
	</body>
</html>`

	 indexTmpl = `<!DOCTYPE HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<title></title>
		<link rel="stylesheet" type="text/css" href="main.css"/>
	</head>
	<body>
	</body>
</html>`

	categoryTmpl = `<!DOCTYPE HTML>
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

func isInitialized(target string) bool {
	for _, d := range Dirs {
		if target != "" {
			d = target + "/" + d
		}

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

func initDirs(target string) error {
	for _, d := range Dirs {
		f, err := os.Stat(d)

		if os.IsNotExist(err) {
			if target != "" {
				d = target + "/" + d
			}

			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				return err
			}

			continue
		}

		if !f.IsDir() {
			return errors.New("journal already partially, or fully initialized")
		}
	}

	return nil
}

func initTemplates(target string) error {
	for k, v := range Templates {
		path := TemplatesDir + "/" + k

		if target != "" {
			path = target + "/" + TemplatesDir + "/" + k
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)

		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.Write([]byte(v))

		if err != nil {
			return err
		}
	}

	return nil
}

func Initialize(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) > 1 {
		fmt.Println(usage.Init)
		return
	}

	target := c.Args.Get(0)

	if isInitialized(target) {
		fmt.Fprintf(os.Stderr, "journal already initialized\n")
		os.Exit(1)
	}

	if err := initDirs(target); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if err := initTemplates(target); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	m, err := meta.Init(target)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fname := meta.File

	if target != "" {
		fname = target + "/" + meta.File
	}

	f, err := os.OpenFile(fname, os.O_RDWR, 0660)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	defer f.Close()

	if err := m.Encode(f); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("journal initialized, set the title with 'jrnl title'\n")
}
