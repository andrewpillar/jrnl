package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/sftp"

	"gopkg.in/yaml.v2"
)

var (
	redeslug = regexp.MustCompile("-")
	reslug   = regexp.MustCompile("[^a-zA-Z0-9]")
	redup    = regexp.MustCompile("-{2,}")
)

// Recursively copy the given src file to the given dst.
func Copy(dst, src string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(dst, src, info)
	}

	return CopyFile(dst, src, info)
}

// Copy the given src directory path, to the given dst.
func CopyDir(dst, src string, info os.FileInfo) error {
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

// Copy the given src file to the given dst.
func CopyFile(dst, src string, info os.FileInfo) error {
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

// CopyToRemote functions are the same as the prior ones, only it takes an *sftp.Client as the
// first argument, being the remote we want to copy to.
func CopyToRemote(cli *sftp.Client, dst, src string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyToRemoteDir(cli, dst, src, info)
	}

	return CopyToRemoteFile(cli, dst, src, info)
}

func CopyToRemoteDir(cli *sftp.Client, dst, src string, info os.FileInfo) error {
	if dst != "" {
		if err := cli.MkdirAll(dst); err != nil {
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

		if err := CopyToRemote(cli, fdst, fsrc); err != nil {
			return err
		}
	}

	return nil
}

func CopyToRemoteFile(cli *sftp.Client, dst, src string, info os.FileInfo) error {
	if err := cli.MkdirAll(filepath.Dir(dst)); err != nil {
		return err
	}

	fdst, err := cli.Create(dst)

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

	return err
}

// Exit the program with a non-zero exit status with the given msg, and err. This will format the
// exit error so that the program name will be prepended to the string written to stderr.
func ExitError(msg string, err error) {
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

// Open the given path in using the editor set in $EDITOR.
func OpenInEditor(path string) {
	cmd := exec.Command(os.Getenv("EDITOR"), path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func Deslug(s string) string {
	return redeslug.ReplaceAllString(s, " ")
}

func Slug(s string) string {
	s = strings.TrimSpace(s)

	s = reslug.ReplaceAllString(s, "-")
	s = redup.ReplaceAllString(s, "-")

	return strings.ToLower(s)
}

// Write the given front matter to the given io.Writer.
func MarshalFrontMatter(fm interface{}, w io.Writer) error {
	w.Write([]byte("---\n"))

	enc := yaml.NewEncoder(w)

	if err := enc.Encode(fm); err != nil {
		return err
	}

	_, err := w.Write([]byte("---\n"))

	return err
}

// Unmarshal the given io.Reader to front matter.
func UnmarshalFrontMatter(fm interface{}, r io.Reader) error {
	buf := &bytes.Buffer{}
	tmp := make([]byte, 1)

	bounds := 0

	for {
		if bounds == 2 {
			break
		}

		_, err := r.Read(tmp)

		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		buf.Write(tmp)

		for tmp[0] == '-' {
			_, err = r.Read(tmp)

			if err != nil {
				if err == io.EOF {
					break
				}

				return err
			}

			buf.Write(tmp)

			if tmp[0] == '\n' {
				bounds++
				break
			}
		}
	}

	dec := yaml.NewDecoder(buf)

	if err := dec.Decode(fm); err != nil {
		return err
	}

	return nil
}
