package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	_ "strings"
	"sync"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/resolve"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func copyToRemote(remote string) error {
	f, err := os.Open(meta.File)

	if err != nil {
		return err
	}

	defer f.Close()

	m, err := meta.Decode(f)

	if err != nil {
		return err
	}

	r, ok := m.Remotes[m.Default]

	if !ok && remote == "" {
		return errors.New("no default remote has been set")
	}

	if filepath.IsAbs(r.Target) {
		return util.Copy(SiteDir, r.Target + "/" + SiteDir)
	}

	return nil
}

func publishPost(p *post.Post, wg *sync.WaitGroup, errs chan error) {
	if err := p.Load(); err != nil {
		errs <- err
		wg.Done()

		return
	}

	f, err := os.Open(PostTemplate)

	if err != nil {
		errs <- err
		wg.Done()

		return
	}

	defer f.Close()

	tmpl, err := ioutil.ReadAll(f)

	if err != nil {
		errs <- err
		wg.Done()

		return
	}

	p.Convert()

	if err := p.Publish(string(tmpl)); err != nil {
		errs <- err
		wg.Done()

		return
	}

	wg.Done()
}

func publishPosts(id string) (post.Store, chan error) {
	r := resolve.New(PostsDir)

	ch := r.ResolvePosts()

	wg := &sync.WaitGroup{}
	posts := make([]*post.Post, 0)
	errs := make(chan error)

	for p := range ch {
		posts = append(posts, p)

		if id != "" && p.ID == id {
			wg.Add(1)

			go publishPost(p, wg, errs)
		}

		if id == "" {
			wg.Add(1)

			go publishPost(p, wg, errs)
		}
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	return post.NewStore(posts...), errs
}

func Publish(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Publish)
		return
	}

	mustBeInitialized()

	_, errs := publishPosts(c.Args.Get(0))

	didErr := false

	for err := range errs {
		if err != nil {
			didErr = true

			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}

//	createIndexes(posts)

	if !c.Flags.IsSet("draft") {
		if err := copyToRemote(c.Flags.GetString("remote")); err != nil {
			didErr = true

			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}

	if didErr {
		os.Exit(1)
	}
}
