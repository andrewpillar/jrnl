package util

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

func Copy(src, dst string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(src, dst, info)
	}

	return CopyFile(src, dst, info)
}

func CopyDir(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(src)

	if err != nil {
		return err
	}

	for _, f := range files {
		fdst := filepath.Join(dst, f.Name())
		fsrc := filepath.Join(src, f.Name())

		if err := Copy(fsrc, fdst); err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(filepath.Dir(dst), info.Mode()); err != nil {
		return err
	}

	fdst, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer fdst.Close()

	if err = os.Chmod(fdst.Name(), info.Mode()); err != nil {
		return err
	}

	fsrc, err := os.Open(src)

	if err != nil {
		return err
	}

	defer fsrc.Close()

	_, err = io.Copy(fdst, fsrc)

	return err
}

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
