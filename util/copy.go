package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
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

func CopyToRemote(src, dst string, conn *sftp.Client) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyToRemoteDir(src, dst, info, conn)
	}

	return CopyToRemoteFile(src, dst, info, conn)
}

func CopyToRemoteDir(src string, dst string, info os.FileInfo, conn *sftp.Client) error {
	if dst != "" {
		if err := conn.MkdirAll(dst); err != nil {
			return err
		}
	}

	files, err := ioutil.ReadDir(src)

	if err != nil {
		return err
	}

	for _, f := range files {
		fdst := filepath.Join(dst, f.Name())
		fsrc := filepath.Join(src, f.Name())

		if err := CopyToRemote(fsrc, fdst, conn); err != nil {
			return err
		}
	}

	return nil
}

func CopyToRemoteFile(src string, dst string, info os.FileInfo, conn *sftp.Client) error {
	dir := filepath.Dir(dst)

	if err := conn.MkdirAll(dir); err != nil {
		return err
	}

	fdst, err := conn.Create(dst)

	if err != nil {
		return err
	}

	defer fdst.Close()

	fsrc, err := os.Open(src)

	if err != nil {
		return err
	}

	defer fsrc.Close()

	_, err = io.Copy(fdst, fsrc)

	if err != nil {
		return err
	}

	return nil
}
