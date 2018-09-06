package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Deslug(str, sep string) string {
	buf := bytes.Buffer{}

	slugs := strings.Split(str, " ")

	for i, s := range slugs {
		buf.WriteString(strings.Replace(s, "-", " ", -1))

		if i != len(slugs) - 1 {
			buf.WriteString(sep)
		}
	}

	return strings.Title(buf.String())
}

func DirEmpty(dir string) bool {
	f, err := os.Open(dir)

	if err != nil {
		return false
	}

	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF {
		return true
	}

	return false
}

func Error(msg string, err error) {
	fmt.Fprintf(os.Stderr, "jrnl: %s\n", msg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s\n", err)
	}

	os.Exit(1)
}

func OpenInEditor(fname string) {
	cmd := exec.Command(os.Getenv("EDITOR"), fname)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func RemoveEmptyDirs(root, path string) error {
	parts := strings.Split(path, string(os.PathSeparator))

	for i := range parts {
		dir := filepath.Join(parts[:len(parts) - i]...)

		if dir == root {
			break
		}

		if DirEmpty(dir) {
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}

	return nil
}
