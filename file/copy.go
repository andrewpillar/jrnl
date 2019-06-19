package file

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func copyDir(dst, src string, info os.FileInfo) error {
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

		if err := Copy(fdst, fsrc); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(dst, src string, info os.FileInfo) error {
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

// Recursively copy the given src file to the given dst.
func Copy(dst, src string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(dst, src, info)
	}

	return copyFile(dst, src, info)
}
