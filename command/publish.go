package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/resolve"
	"github.com/andrewpillar/jrnl/usage"
)

func publishPost(
	p *post.Post,
	draft bool,
	wg *sync.WaitGroup,
	errs chan error,
) {
	if err := p.Load(); err != nil {
		errs <- err
		wg.Done()

		return
	}

	dir := filepath.Dir(p.SitePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
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

	dst, err := os.OpenFile(p.SitePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		errs <- err
		wg.Done()

		return
	}

	defer dst.Close()

	t, err := template.New("post").Parse(string(tmpl))

	if err != nil {
		errs <- err
		wg.Done()

		return
	}

	p.Convert()

	if err = t.Execute(dst, p); err != nil {
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

			go publishPost(p, draft, wg, errs)

			break
		}

		if postId == "" {
			wg.Add(1)

			go publishPost(p, draft, wg, errs)
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

	if didErr {
		os.Exit(1)
	}
}
