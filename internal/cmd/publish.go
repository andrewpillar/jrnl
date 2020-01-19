package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
	"github.com/andrewpillar/jrnl/internal/hash"
	"github.com/andrewpillar/jrnl/internal/template"

	"github.com/gorilla/feeds"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

type errAgg []error

type publishPath struct {
	del  bool
	path string
}

var (
	fileMask = os.O_TRUNC|os.O_RDWR|os.O_CREATE
	hashFile = "jrnl.hash"
)

func (e errAgg) Err() error {
	isNil := true

	for _, err := range e {
		if err != nil {
			isNil = false
			break
		}
	}

	if isNil {
		return nil
	}

	return e
}

func (e errAgg) Error() string {
	buf := &bytes.Buffer{}
	end := len(e)-1

	for i, err := range e {
		buf.WriteString(err.Error())

		if i != end {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

func bytesEqual(a, b []byte) bool {
	return string(a) == string(b)
}

func getBlogHash() (hash.Hash, error) {
	var h hash.Hash

	if _, err := os.Stat(hashFile); err != nil {
		if os.IsNotExist(err) {
			return hash.New(), nil
		}
		return h, err
	}

	f, err := os.Open(hashFile)

	if err != nil {
		return h, err
	}

	defer f.Close()

	return hash.Decode(f)
}

func getHostkey(hostname string) (ssh.PublicKey, error) {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))

	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(f)

	var	hostkey ssh.PublicKey

	for s.Scan() {
		parts := strings.Split(s.Text(), " ")

		if len(parts) != 3 {
			continue
		}

		if strings.Contains(parts[0], hostname) {
			hostkey, _, _, _, err = ssh.ParseAuthorizedKey(s.Bytes())

			if err != nil {
				return nil, err
			}
			break
		}
	}

	if hostkey == nil {
		err = errors.New("no key for host " + hostname)
	}

	return hostkey, err
}

func getUpdatedBlog() ([]blog.Page, []blog.Post, error) {
	pages, err := blog.Pages()

	if err != nil {
		return pages, []blog.Post{}, err
	}

	posts, err := blog.Posts()

	if err != nil {
		return pages, posts, err
	}

	if _, err := os.Stat(hashFile); err != nil {
		if os.IsNotExist(err) {
			return pages, posts, nil
		}
		return pages, posts, err
	}

	f, err := os.Open(hashFile)

	if err != nil {
		return pages, posts, err
	}

	defer f.Close()

	h, err := hash.Decode(f)

	if err != nil {
		return pages, posts, err
	}

	updatedPages := make([]blog.Page, 0, len(pages))
	updatedPosts := make([]blog.Post, 0, len(posts))

	for _, p := range pages {
		b, ok := h[p.ID]

		if !ok {
			updatedPages = append(updatedPages, p)
			continue
		}

		if b == nil {
			p.Delete = true
			updatedPages = append(updatedPages, p)
			continue
		}

		if !bytesEqual(b, p.Hash()) {
			updatedPages = append(updatedPages, p)
		}
	}

	for _, p := range posts {
		b, ok := h[p.ID]

		if !ok {
			updatedPosts = append(updatedPosts, p)
			continue
		}

		if b == nil {
			p.Delete = true
			updatedPosts = append(updatedPosts, p)
			continue
		}

		if !bytesEqual(b, p.Hash()) {
			updatedPosts = append(updatedPosts, p)
		}
	}

	return updatedPages, updatedPosts, nil
}

func publishPage(p blog.Page, data interface{}) error {
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

	f, err := os.OpenFile(p.SitePath, fileMask, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()
	return template.Render(f, p.ID, string(b), data)
}

func publishPages(s blog.Site) (chan blog.Page, chan error) {
	published := make(chan blog.Page)
	errs := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(len(s.Pages))

	for _, p := range s.Pages {
		go func(p blog.Page) {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			p.Render()

			data := struct{
				Site blog.Site
				Page blog.Page
			}{Site: s, Page: p}

			if err := publishPage(p, data); err != nil {
				errs <- err
				return
			}
			published <- p
		}(p)
	}

	go func() {
		wg.Wait()
		close(errs)
		close(published)
	}()

	return published, errs
}

func publishPosts(s blog.Site, pp []blog.Post) (chan blog.Post, chan error) {
	published := make(chan blog.Post)
	errs := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(len(pp))

	for _, p := range pp {
		go func(p blog.Post) {
			defer wg.Done()

			if err := p.Load(); err != nil {
				errs <- err
				return
			}

			p.Render()

			data := struct{
				Site blog.Site
				Post blog.Post
			}{Site: s, Post: p}

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

func publishFeed(feed blog.Feed, pp []blog.Post, rss, atom string) error {
	m := map[string]string{
		"atom": atom,
		"rss":  rss,
	}

	fn := func(kind, fname string) error {
		f, err := os.OpenFile(fname, fileMask, config.FileMode)

		if err != nil {
			return err
		}

		defer f.Close()

		return feed.Write(f, kind, pp)
	}

	for k, v := range m {
		if v == "" {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(v), config.DirMode); err != nil {
			return err
		}

		if err := fn(k, v); err != nil {
			return err
		}
	}

	return nil
}

func publishLocal(path string, paths []publishPath) error {
	for _, p := range paths {
		dir := filepath.Dir(strings.Replace(p.path, config.SiteDir, path, 1))

		if err := os.MkdirAll(dir, config.DirMode); err != nil {
			return err
		}
	}

	errs := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(len(paths))

	for _, p := range paths {
		go func(p publishPath) {
			defer wg.Done()

			dst := strings.Replace(p.path, config.SiteDir, path, 1)

			if p.del {
				if err := os.Remove(dst); err != nil {
					errs <- err
					return
				}

				parts := append(
					[]string{string(os.PathSeparator)},
					strings.Split(filepath.Dir(dst), string(os.PathSeparator))...,
				)

				for i := range parts {
					dir := filepath.Join(parts[:len(parts)-i]...)

					if dir == path {
						break
					}

					if err := os.Remove(dir); err != nil {
						errs <- err
					}
				}
				return
			}

			fdst, err := os.OpenFile(dst, fileMask, config.FileMode)

			if err != nil {
				errs <- err
				return
			}

			defer fdst.Close()

			fsrc, err := os.Open(p.path)

			if err != nil {
				errs <- err
				return
			}

			if _, err := io.Copy(fdst, fsrc); err != nil {
				errs <- err
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	agg := errAgg(make([]error, 0, len(paths)))

	for err := range errs {
		agg = append(agg, err)
	}

	return agg.Err()
}

func publishRemote(uri string, paths []publishPath) error {
	user := os.Getenv("USER")

	parts := strings.Split(uri, "@")
	i := 0

	if len(parts) > 1 {
		user = parts[0]
		i = 1
	}

	parts = strings.Split(parts[i], ":")

	hostname, path := parts[0], parts[1]

	key, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))

	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		return err
	}

	hostkey, err := getHostkey(hostname)

	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(hostname, ":22"), &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.FixedHostKey(hostkey),
	})

	if err != nil {
		return err
	}

	defer conn.Close()

	sftp, err := sftp.NewClient(conn)

	if err != nil {
		return err
	}

	defer sftp.Close()

	for _, p := range paths {
		dir := filepath.Dir(strings.Replace(p.path, config.SiteDir, path, 1))

		if err := sftp.MkdirAll(dir); err != nil {
			return err
		}
	}

	errs := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(len(paths))

	for _, p := range paths {
		go func(p publishPath) {
			defer wg.Done()

			dst := strings.Replace(p.path, config.SiteDir, path, 1)

			if p.del {
				if err := sftp.Remove(dst); err != nil {
					errs <- err
					return
				}

				parts := strings.Split(filepath.Dir(dst), string(os.PathSeparator))

				for i := range parts {
					dir := filepath.Join(parts[:len(parts)-i]...)

					if dir == path {
						break
					}

					info, err := sftp.ReadDir(dir)

					if err != nil {
						errs <- err
						return
					}

					if len(info) == 0 {
						if err := sftp.Remove(dir); err != nil {
							errs <- err
							return
						}
					}
				}
				return
			}

			fdst, err := sftp.OpenFile(dst, fileMask)

			if err != nil {
				errs <- err
				return
			}

			defer fdst.Close()

			fsrc, err := os.Open(p.path)

			if err != nil {
				errs <- err
				return
			}

			defer fsrc.Close()

			if _, err := io.Copy(fdst, fsrc); err != nil {
				errs <- err
				return
			}
		}(p)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	agg := errAgg(make([]error, 0, len(paths)))

	for err := range errs {
		agg = append(agg, err)
	}

	return agg.Err()
}

func writeBlogHash(h hash.Hash) error {
	f, err := os.OpenFile(hashFile, fileMask, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	return h.Encode(f)
}

func Publish(c cli.Command) {
	if err := config.Initialized(""); err != nil {
		exitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		exitError("failed to get config", err)
	}

	cc, err := blog.Categories()

	if err != nil {
		exitError("failed to get all categories", err)
	}

	h, err := getBlogHash()

	if err != nil {
		exitError("failed to get blog hash file", err)
	}

	pages, posts, err := getUpdatedBlog()

	if err != nil {
		exitError("failed to get blog from hash file", err)
	}

	s := blog.Site{
		Title:      cfg.Site.Title,
		Link:       cfg.Site.Link,
		Categories: cc,
		Pages:      pages,
	}

	paths := make([]publishPath, 0, len(pages)+len(posts))

	code := 0
	publishedPages, errs := publishPages(s)

	for publishedPages != nil && errs != nil {
		select {
		case p, ok := <-publishedPages:
			if !ok {
				publishedPages = nil
				break
			}
			h[p.ID] = p.Hash()
			paths = append(paths, publishPath{del: p.Delete, path: p.SitePath})
		case err, ok := <-errs:
			if !ok {
				errs = nil
				break
			}
			code = 1
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		}
	}

	publishedPosts, errs := publishPosts(s, posts)

	index := blog.NewIndex()

	for publishedPosts != nil && errs != nil {
		select {
		case p, ok := <-publishedPosts:
			if !ok {
				publishedPosts = nil
				break
			}
			h[p.ID] = p.Hash()
			paths = append(paths, publishPath{del: p.Delete, path: p.SitePath})
			index.Put(p)
		case err, ok := <-errs:
			if !ok {
				errs = nil
				break
			}
			code = 1
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		}
	}

	if err := writeBlogHash(h); err != nil {
		exitError("failed to write blog hash file", err)
	}

	feed := blog.Feed{
		Title:       cfg.Site.Title,
		Link:        cfg.Site.Link,
		Description: cfg.Site.Description,
		Author:      feeds.Author{
			Name:  cfg.Author.Name,
			Email: cfg.Author.Email,
		},
	}

	if err := publishFeed(feed, posts, c.Flags.GetString("rss"), c.Flags.GetString("atom")); err != nil {
		code = 1
		fmt.Fprintf(os.Stderr, "failed to publish feed: %s\n", err)
	}

	for k := range index {
		path, err := index.Write(k, s)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		}

		// Index layouts are optional and can return empty strings.
		if path != "" {
			paths = append(paths, publishPath{del: false, path: path})
		}
	}

	if c.Flags.IsSet("draft") {
		goto exit
	}

	if cfg.Site.Remote == "" {
		exitError("failed to publish blog", errors.New("remote not set"))
	}

	if filepath.IsAbs(cfg.Site.Remote) {
		if err := publishLocal(cfg.Site.Remote, paths); err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		}
		goto exit
	}

	if err := publishRemote(cfg.Site.Remote, paths); err != nil {
		code = 1
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}

exit:
	if code != 0 {
		os.Exit(code)
	}
}
