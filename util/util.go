package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/andrewpillar/jrnl/meta"
)

var (
	redeslug = regexp.MustCompile("-")

	reslug = regexp.MustCompile("[^a-zA-Z0-9]")

	redup = regexp.MustCompile("-{2,}")
)

func Deslug(s string) string {
	return redeslug.ReplaceAllString(s, " ")
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

func Exit(msg string, err error) {
	code := 0
	w := os.Stdout

	if err != nil {
		code = 1
		w = os.Stderr
	}

	fmt.Fprintf(w, "%s: %s", os.Args[0], msg)

	if err != nil {
		fmt.Fprintf(w, ": %s", err)
	}

	fmt.Fprintf(w, "\n")

	os.Exit(code)
}

func MustBeInitialized() {
	for _, d := range meta.Dirs {
		info, err := os.Stat(d)

		if err != nil {
			Exit("not fully initialized", err)
		}

		if !info.IsDir() {
			Exit("unexpected non-directory file", errors.New(d))
		}
	}
}

func OpenInEditor(editor, fname string) {
	cmd := exec.Command(editor, fname)
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

func Slug(s string) string {
	s = reslug.ReplaceAllString(s, "-")
	s = redup.ReplaceAllString(s, "-")

	return strings.ToLower(s)
}

func Title(s string) string {
	t := bytes.Buffer{}

	parts := strings.Split(s, " ")

	for i, p := range parts {
		t.WriteString(Ucfirst(p))

		if i != len(p) - 1 {
			t.WriteString(" ")
		}
	}

	return strings.Trim(t.String(), " ")
}

func Ucfirst(s string) string {
	if len(s) == 0 {
		return ""
	}

	r := []rune(s)
	u := string(unicode.ToUpper(r[0]))

	return u + s[len(u):]
}
