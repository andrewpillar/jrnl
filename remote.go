package main

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pkgsftp "github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

type File interface {
	Read(p []byte) (int, error)

	Write(p []byte) (int, error)

	Close() error
}

type FS interface {
	Open(path string) (File, error)

	Remove(path string) error

	Close() error
}

type disk struct {
	path string
}

type sftp struct {
	cli  *pkgsftp.Client
	path string
}

type Remote struct {
	fs FS
}

func getPublicKey(host string) (ssh.PublicKey, error) {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	var pubkey ssh.PublicKey

	for sc.Scan() {
		parts := strings.Split(sc.Text(), " ")

		if len(parts) != 3 {
			continue
		}

		if strings.Contains(parts[0], host) {
			pubkey, _, _, _, err = ssh.ParseAuthorizedKey(sc.Bytes())

			if err != nil {
				return nil, err
			}
			break
		}
	}

	if pubkey == nil {
		return nil, errors.New("no key for host: " + host)
	}
	return pubkey, nil
}

func OpenRemote(remote string) (*Remote, error) {
	if filepath.IsAbs(remote) {
		return &Remote{
			fs: &disk{path: remote},
		}, nil
	}

	// Assume SFTP.
	var (
		n int

		user string
		host string
		path string
	)

	for i, r := range remote {
		if r == '@' {
			user = remote[n:i]
			n = i + 1
		}

		if r == ':' {
			host = remote[n:i]
			path = remote[i + 1:]
		}
	}

	if host == "" {
		return nil, errors.New("missing host in remote url")
	}

	if path == "" {
		path = "/home/" + user
	}

	privkey, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))

	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(privkey)

	if err != nil {
		return nil, err
	}

	pubkey, err := getPublicKey(host)

	if err != nil {
		return nil, err
	}

	conn, err := ssh.Dial("tcp", host + ":22", &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(pubkey),
	})

	if err != nil {
		return nil, err
	}

	cli, err := pkgsftp.NewClient(conn)

	if err != nil {
		return nil, err
	}

	return &Remote{
		fs: &sftp{
			cli:  cli,
			path: path,
		},
	}, nil
}

func (d *disk) filepath(path string) string { return filepath.Join(d.path, path) }

func (d *disk) Open(path string) (File, error) {
	path = d.filepath(path)

	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0755)); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(0644))
}

func (d *disk) Remove(path string) error {
	path = d.filepath(path)

	if err := os.Remove(path); err != nil {
		return err
	}

	parts := strings.Split(filepath.Dir(path), string(os.PathSeparator))

	for i := range parts {
		dir := string(os.PathSeparator) + filepath.Join(parts[:len(parts)-i]...)

		if dir == d.path {
			break
		}

		info, err := ioutil.ReadDir(dir)

		if err != nil {
			return err
		}

		if len(info) == 0 {
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *disk) Close() error { return nil }

func (s *sftp) filepath(path string) string { return filepath.Join(s.path, path) }

func (s *sftp) Open(path string) (File, error) {
	path = s.filepath(path)

	if err := s.cli.MkdirAll(filepath.Dir(path)); err != nil {
		return nil, err
	}

	f, err := s.cli.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR)

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if err := s.cli.MkdirAll(path); err != nil {
			return nil, err
		}
		f, err = s.cli.Create(path)
	}
	return f, err
}

func (s *sftp) Remove(path string) error {
	path = s.filepath(path)

	if err := s.cli.Remove(path); err != nil {
		return err
	}

	parts := strings.Split(filepath.Dir(path), string(os.PathSeparator))

	for i := range parts {
		dir := string(os.PathSeparator) + filepath.Join(parts[:len(parts)-i]...)

		if dir == s.path {
			break
		}

		info, err := s.cli.ReadDir(dir)

		if err != nil {
			return err
		}

		if len(info) == 0 {
			if err := s.cli.Remove(dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *sftp) Close() error { return s.cli.Close() }

func (r *Remote) Copy(path string) error {
	src, err := os.Open(path)

	if err != nil {
		return err
	}

	defer src.Close()

	path = strings.Replace(path, siteDir, "", 1)

	dst, err := r.fs.Open(path)

	if err != nil {
		return err
	}

	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (r *Remote) Remove(path string) error {
	path = strings.Replace(path, siteDir, "", 1)
	return r.fs.Remove(path)
}

func (r *Remote) Close() error { return r.fs.Close() }
