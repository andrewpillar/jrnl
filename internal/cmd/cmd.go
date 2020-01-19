package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func exitError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s:", os.Args[0])

	if msg != "" {
		fmt.Fprintf(os.Stderr, " %s", msg)

		if err != nil {
			fmt.Fprintf(os.Stderr, ":")
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, " %s", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func openInEditor(path string) {
	cmd := exec.Command(os.Getenv("EDITOR"), path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
