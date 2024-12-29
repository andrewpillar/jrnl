package main

import (
	"errors"
	"os"
)

var FlushCmd = &Command{
	Usage: "flush",
	Short: "flush out the jrnl's cached content",
	Run: func(cmd *Command, args []string) error {
		if err := Initialized("."); err != nil {
			return err
		}

		dir, err := cacheDir()

		if err != nil {
			return err
		}

		if err := os.RemoveAll(dir); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	},
}
