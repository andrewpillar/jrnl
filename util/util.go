package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func Deslug(slug string) string {
	parts := strings.Split(slug, "-")
	buf := &bytes.Buffer{}

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

func OpenInEditor(fname string) {
	cmd := exec.Command(os.Getenv("EDITOR"), fname)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

// Remove all empty directories within the given path, excluding the root
// directory itself.
func RemoveEmptyDirs(root, path string) error {
	parts := strings.Split(path, "/")

	for i := 0; i < len(parts); i++ {
		dir := strings.Join(parts[:len(parts) - i], "/")

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
