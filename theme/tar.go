package theme

import (
	atar "archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func tar(src string, w io.Writer) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}

	gzw := gzip.NewWriter(w)

	defer gzw.Close()

	tw := atar.NewWriter(gzw)

	defer tw.Close()

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := atar.FileInfoHeader(info, info.Name())

		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(path, src, "", -1), string(os.PathSeparator))

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
	}

	return filepath.Walk(src, walk)
}

func untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)

	if err != nil {
		return err
	}

	defer gzr.Close()

	tr := atar.NewReader(gzr)

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
			case atar.TypeDir:
				_, err = os.Stat(target)

				if err != nil && os.IsNotExist(err) {
					if err := os.MkdirAll(target, 0775); err != nil {
						return err
					}
				}
			case atar.TypeReg:
				mode := os.FileMode(header.Mode)

				f, err := os.OpenFile(target, os.O_TRUNC|os.O_CREATE|os.O_RDWR, mode)

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
