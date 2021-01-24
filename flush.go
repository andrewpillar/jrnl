package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var FlushCmd = &Command{
	Usage: "flush",
	Short: "clear the journal's hashed content",
	Long:  `Flush will remove the _hash directory that contains the generated hashes of
the journals pages and posts.`,
	Run:   flushCmd,
}

func flushCmd(cmd *Command, args []string) {
	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if err := os.Remove(filepath.Join(dataDir, "hash")); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
			os.Exit(1)
		}
	}
}
