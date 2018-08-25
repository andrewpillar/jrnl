package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func Deslug(s string) string {
	parts := strings.Split(s, "-")
	buf := bytes.Buffer{}

	for i, p := range parts {
		buf.WriteString(Ucfirst(p))

		if i != len(parts) - 1 {
			buf.WriteString(" ")
		}
	}

	return buf.String()
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
		fmt.Fprintf(os.Stderr, "  %s\n", err)
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

func Ucfirst(s string) string {
	for _, r := range []rune(s) {
		u := string(unicode.ToUpper(r))

		return u + s[len(u):]
	}

	return ""
}
