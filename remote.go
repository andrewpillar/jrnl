package main

import (
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Remote interface {
	Connect() error

	Sync(local, remote string) error
}

type sshRemote struct {
	cli *ssh.Client
	url *url.URL
}

func (r *sshRemote) getPrivateKey() (ssh.Signer, error) {
	path := r.url.Query().Get("identity")

	if path == "" {
		return nil, errors.New("no identity specified")
	}

	b, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(b)

	if err != nil {
		return nil, err
	}
	return key, nil
}

func (r *sshRemote) Connect() error {
	signer, err := r.getPrivateKey()

	if err != nil {
		return err
	}

	dir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	callback, err := knownhosts.New(filepath.Join(dir, ".ssh", "known_hosts"))

	if err != nil {
		return err
	}

	cfg := ssh.ClientConfig{
		User: r.url.User.Username(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: callback,
	}

	cli, err := ssh.Dial("tcp", r.url.Host, &cfg)

	if err != nil {
		return err
	}

	r.cli = cli

	return nil
}

func (r *sshRemote) Sync(local, remote string) error {
	cli, err := sftp.NewClient(r.cli)

	if err != nil {
		return err
	}

	defer cli.Close()

	remote = cli.Join(r.url.Path, remote)

	if err := cli.MkdirAll(filepath.Dir(remote)); err != nil {
		return err
	}

	dst, err := cli.Create(remote)

	if err != nil {
		return err
	}

	defer dst.Close()

	src, err := os.Open(local)

	if err != nil {
		return err
	}

	defer src.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

type fileRemote struct {
	url *url.URL
}

func (r *fileRemote) Connect() error {
	return os.MkdirAll(r.url.Path, dirMode)
}

func (r *fileRemote) Sync(local, remote string) error {
	remote = filepath.Join(r.url.Host, r.url.Path, remote)

	if err := os.MkdirAll(filepath.Dir(remote), dirMode); err != nil {
		return err
	}

	dst, err := os.Create(remote)

	if err != nil {
		return err
	}

	defer dst.Close()

	src, err := os.Open(local)

	if err != nil {
		return err
	}

	defer src.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func ParseRemote(s string) (Remote, error) {
	url, err := url.Parse(s)

	if err != nil {
		return nil, err
	}

	var r Remote

	switch url.Scheme {
	case "ssh":
		r = &sshRemote{
			url: url,
		}
	case "file":
		r = &fileRemote{
			url: url,
		}
	default:
		return nil, errors.New("unknown remote scheme: " + url.Scheme)
	}
	return r, nil
}
