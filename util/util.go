package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
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

func Deslug(s, sep string) string {
	buf := bytes.Buffer{}

	slugs := strings.Split(s, " ")

	for i, s := range slugs {
		parts := strings.Split(s, "-")

		for j, p := range parts {
			buf.WriteString(Ucfirst(p))

			if j != len(parts) - 1 {
				buf.WriteString(" ")
			}
		}

		if i != len(slugs) - 1 {
			buf.WriteString(sep)
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

func Tar(src string, w io.Writer) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}

	gzw := gzip.NewWriter(w)

	defer gzw.Close()

	tw := tar.NewWriter(gzw)

	defer tw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())

		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(
			strings.Replace(path, src, "", -1),
			string(os.PathSeparator),
		)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		defer f.Close()

		if _, err = io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}

func Untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)

	if err != nil {
		return err
	}

	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
			case err == io.EOF:
				return nil
			case err != nil:
				return err
			case header == nil:
				continue
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
			case tar.TypeDir:
				_, err = os.Stat(target)

				if err != nil && os.IsNotExist(err) {
					if err := os.MkdirAll(target, 0775); err != nil {
						return err
					}
				}
			case tar.TypeReg:
				f, err := os.OpenFile(
					target,
					os.O_TRUNC|os.O_CREATE|os.O_RDWR,
					os.FileMode(header.Mode),
				)

				if err != nil {
					return err
				}

				defer f.Close()

				if _, err = io.Copy(f, tr); err != nil {
					return err
				}
		}
	}
}

func Ucfirst(s string) string {
	for _, r := range []rune(s) {
		u := string(unicode.ToUpper(r))

		return u + s[len(u):]
	}

	return ""
}
