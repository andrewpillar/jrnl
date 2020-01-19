package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/jrnl/internal/blog"
	"github.com/andrewpillar/jrnl/internal/config"
	"github.com/andrewpillar/jrnl/internal/hash"
)

var (
	layouts = map[string]string{
		"page": `{{.Page.Title}} - {{.Site.Title}}
{{.Page.Body}}`,
		"post": `{{.Post.Title}} - {{.Site.Title}}
{{.Post.Body}}`,
	}
)

func writeLayout(t *testing.T, fname string, r io.Reader) {
	f, err := os.OpenFile(
		filepath.Join(config.LayoutsDir, fname),
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		config.FileMode,
	)

	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T) (*cli.Cli, func()) {
	c := setupCli()

	dir, err := ioutil.TempDir("", "jrnl-test-")

	if err != nil {
		t.Fatal(err)
	}

	remoteDir, err := ioutil.TempDir("", "jrnl-local-remote-")

	if err != nil {
		t.Fatal(err)
	}

	if err := c.Run([]string{"init", dir}); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("EDITOR", ""); err != nil {
		t.Fatal(err)
	}

	for fname, s := range layouts {
		writeLayout(t, fname, strings.NewReader(s))
	}

	if err := c.Run([]string{"remote-set", remoteDir}); err != nil {
		t.Fatal(err)
	}

	if err := c.Run([]string{"title", "Test Blog"}); err != nil {
		t.Fatal(err)
	}

	return c, func() {
		os.RemoveAll(dir)
		os.RemoveAll(remoteDir)
	}
}

func getBlogHash() (hash.Hash, error) {
	var h hash.Hash

	f, err := os.Open("jrnl.hash")

	if err != nil {
		return h, err
	}

	defer f.Close()

	h, err = hash.Decode(f)
	return h, err
}

func Test_Jrnl(t *testing.T) {
	c, cleanup := setup(t)
	defer cleanup()

	if err := c.Run([]string{"page", "Page One"}); err != nil {
		t.Fatal(err)
	}

	page, err := blog.GetPage("page-one")

	if err != nil {
		t.Fatal(err)
	}

	page.Layout = "page"

	if err := page.Save(); err != nil {
		t.Fatal(err)
	}

	if err := c.Run([]string{"post", "Post One"}); err != nil {
		t.Fatal(err)
	}

	post, err := blog.GetPost("post-one")

	if err != nil {
		t.Fatal(err)
	}

	post.Layout = "post"

	if err := post.Save(); err != nil {
		t.Fatal(err)
	}

	if err := c.Run([]string{"publish"}); err != nil {
		t.Fatal(err)
	}

	if err := c.Run([]string{"rm", "page-one"}); err != nil {
		t.Fatal(err)
	}

	h, err := getBlogHash()

	if err != nil {
		t.Fatal(err)
	}

	b, ok := h["page-one"]

	if !ok {
		t.Errorf("expected page-one to still be in blog hash\n")
	}

	if b != nil {
		t.Errorf("expected page-one to be nil in blog hash\n")
	}

	if err := c.Run([]string{"publish"}); err != nil {
		t.Fatal(err)
	}
}
