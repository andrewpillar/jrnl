package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/feed"
	"github.com/andrewpillar/jrnl/file"
	"github.com/andrewpillar/jrnl/index"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/site"
	"github.com/andrewpillar/jrnl/template"

	"github.com/gorilla/feeds"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

func getHostkey(hostname string) (ssh.PublicKey, error) {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	s := bufio.NewScanner(f)

	var hostkey ssh.PublicKey

	for s.Scan() {
		fields := strings.Split(s.Text(), " ")

		if len(fields) != 3 {
			continue
		}

		if strings.Contains(fields[0], hostname) {
			var err error

			hostkey, _, _, _, err = ssh.ParseAuthorizedKey(s.Bytes())

			if err != nil {
				return nil, err
			}

			break
		}
	}

	if hostkey == nil {
		return nil, errors.New("no key for host " + hostname)
	}

	return hostkey, nil
}

func publishPage(p *page.Page, data interface{}) error {
	if p.Layout == "" {
		return errors.New("no layout for " + p.ID)
	}

	b, err := ioutil.ReadFile(filepath.Join(config.LayoutsDir, p.Layout))

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p.SitePath), config.DirMode); err != nil {
		return err
	}

	f, err := os.OpenFile(p.SitePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	return template.Render(f, p.ID, string(b), data)
}

func publishPages(s site.Site) chan error {
	errs := make(chan error)

	wg := &sync.WaitGroup{}

	for _, p := range s.Pages {
		wg.Add(1)

		go func(p *page.Page) {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			p.Render()

			data := struct{
				Site site.Site
				Page *page.Page
			}{
				Site: s,
				Page: p,
			}

			if err := publishPage(p, data); err != nil {
				errs <- err
			}
		}(p)
	}

	go func() {
		wg.Wait()

		close(errs)
	}()

	return errs
}

func publishPosts(s site.Site, posts []*post.Post) (chan *post.Post, chan error) {
	published := make(chan *post.Post)
	errs := make(chan error)

	wg := &sync.WaitGroup{}

	for _, p := range posts {
		wg.Add(1)

		go func(p *post.Post) {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			p.Render()

			data := struct{
				Site site.Site
				Post *post.Post
			}{
				Site: s,
				Post: p,
			}

			if err := publishPage(p.Page, data); err != nil {
				errs <- err
				return
			}

			published <- p
		}(p)
	}

	go func() {
		wg.Wait()

		close(published)
		close(errs)
	}()

	return published, errs
}

func publishFeed(fd feed.Feed, posts []*post.Post, rss, atom string) error {
	mask := os.O_TRUNC|os.O_CREATE|os.O_RDWR

	if rss != "" {
		dir := filepath.Dir(rss)

		if err := os.MkdirAll(dir, config.DirMode); err != nil {
			return err
		}

		f, err := os.OpenFile(rss, mask, config.FileMode)

		if err != nil {
			return err
		}

		defer f.Close()

		if err := fd.WriteRss(f, posts); err != nil {
			return err
		}
	}

	if atom != "" {
		dir := filepath.Dir(rss)

		if err := os.MkdirAll(dir, config.DirMode); err != nil {
			return err
		}

		f, err := os.OpenFile(atom, mask, config.FileMode)

		if err != nil {
			return err
		}

		defer f.Close()

		if err := fd.WriteAtom(f, posts); err != nil {
			return err
		}
	}

	return nil
}

func remoteCopy(cli *sftp.Client, dst, src string) error {
	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return remoteDirCopy(cli, dst, src, info)
	}

	return remoteFileCopy(cli, dst, src, info)
}

func remoteDirCopy(cli *sftp.Client, dst, src string, info os.FileInfo) error {
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

		if err := remoteCopy(cli, fdst, fsrc); err != nil {
			return err
		}
	}

	return nil
}

func remoteFileCopy(cli *sftp.Client, dst, src string, info os.FileInfo) error {
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

func Publish(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		exitError("failed to get config", err)
	}

	defer cfg.Close()

	categories, err := category.All()

	if err != nil {
		exitError("failed to get all categories", err)
	}

	pages, err := page.All()

	if err != nil {
		exitError("failed to get all pages", err)
	}

	s := site.Site{
		Title:      cfg.Site.Title,
		Link:       cfg.Site.Link,
		Categories: categories,
		Pages:      pages,
	}

	code := 0

	errs := publishPages(s)

	for err := range errs {
		code = 1
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}

	posts, err := post.All()

	if err != nil {
		exitError("failed to get all posts", err)
	}

	published, errs := publishPosts(s, posts)

	postIndex := index.New()

	for published != nil && errs != nil {
		select {
			case p, ok := <-published:
				if !ok {
					published = nil
				} else {
					postIndex.Put(p)
				}
			case err, ok := <-errs:
				if !ok {
					errs = nil
				} else {
					code = 1
					fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
				}
		}
	}

	fd := feed.Feed{
		Title:       cfg.Site.Title,
		Link:        cfg.Site.Link,
		Description: cfg.Site.Description,
		Author:      &feeds.Author{
			Name:  cfg.Author.Name,
			Email: cfg.Author.Email,
		},
	}

	if err := publishFeed(fd, posts, c.Flags.GetString("rss"), c.Flags.GetString("atom")); err != nil {
		exitError("failed to publish posts", err)
	}

	for key := range postIndex {
		if err := postIndex.Publish(key, s); err != nil && !os.IsNotExist(err) {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		}
	}

	if c.Flags.IsSet("draft") {
		return
	}

	if cfg.Site.Remote == "" {
		exitError("failed to publish posts", errors.New("remote not set"))
	}

	if filepath.IsAbs(cfg.Site.Remote) {
		if err := file.Copy(cfg.Site.Remote, config.SiteDir); err != nil {
			exitError("failed to publish posts", err)
		}

		os.Exit(code)
	}

	parts := strings.Split(cfg.Site.Remote, "@")
	i := 0

	user := os.Getenv("USER")

	if len(parts) > 1 {
		user = parts[0]
		i = 1
	}

	parts = strings.Split(parts[i], ":")

	hostname := parts[0]
	dir := parts[1]

	key, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))

	if err != nil {
		exitError("failed to publish posts", err)
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		exitError("failed to publish posts", err)
	}

	hostkey, err := getHostkey(hostname)

	if err != nil {
		exitError("failed to publish posts", err)
	}

	sshCfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostkey),
	}

	conn, err := ssh.Dial("tcp", hostname + ":22", sshCfg)

	if err != nil {
		exitError("failed to publish posts", err)
	}

	defer conn.Close()

	scp, err := sftp.NewClient(conn)

	if err != nil {
		exitError("failed to publish posts", err)
	}

	defer scp.Close()

	if err := remoteCopy(scp, dir, config.SiteDir); err != nil {
		exitError("failed to publish posts", err)
	}

	os.Exit(code)
}
