package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/resolve"
	"github.com/andrewpillar/jrnl/usage"
	"github.com/andrewpillar/jrnl/util"
)

func copyToRemote() error {
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

	if !ok {
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

func Publish(c cli.Command) {
	if c.Flags.IsSet("help") {
		fmt.Println(usage.Publish)
		return
	}

	mustBeInitialized()

	postId := c.Args.Get(0)

	r := resolve.New(SiteDir, PostsDir)

	ch := r.ResolvePosts()

	draft := c.Flags.IsSet("draft")

	wg := &sync.WaitGroup{}
	errs := make(chan error)

	for p := range ch {
		if postId != "" && p.ID == postId {
			wg.Add(1)

			go publishPost(p, wg, errs)

			break
		}

		if postId == "" {
			wg.Add(1)

			go publishPost(p, wg, errs)
		}
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	didErr := false

	for err := range errs {
		if err != nil {
			didErr = true

			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}

	if !draft {
		if err := copyToRemote(); err != nil {
			didErr = true

			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}

	if didErr {
		os.Exit(1)
	}
}
