package command

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/usage"

	"gopkg.in/yaml.v2"
)

func Title(c cli.Command) {
	if c.Flags.IsSet("help") || len(c.Args) < 1 {
		fmt.Println(usage.Title)
		os.Exit(1)
	}

	f, err := os.OpenFile(MetaFile, os.O_RDWR, 0660)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	defer f.Close()

	dec := yaml.NewDecoder(f)
	m := &meta{}

	if err := dec.Decode(m); err != nil {
		fmt.Fprintf(os.Stderr, "dec: %s\n", err)
		os.Exit(1)
	}

	if err := f.Truncate(0); err != nil {
		fmt.Fprintf(os.Stderr, "trunc: %s\n", err)
		os.Exit(1)
	}

	_, err = f.Seek(0, 0)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	m.Title = c.Args.Get(0)

	enc := yaml.NewEncoder(f)

	if err := enc.Encode(m); err != nil {
		fmt.Fprintf(os.Stderr, "enc: %s\n", err)
		os.Exit(1)
	}

	enc.Close()
}
