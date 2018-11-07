package template

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/page"
	"github.com/andrewpillar/jrnl/post"
)

var (
	postId = "post"
	pageId = "page"
)

func init() {
	meta.PostsDir = "testdata/_posts"
	meta.PagesDir = "testdata/_pages"
	meta.LayoutsDir = "testdata/_layouts"
}

func TestPostRender(t *testing.T) {
	p, err := post.Find(postId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postId, err)
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load post %s: %s\n", p.ID, err)
	}

	p.Render()

	data := struct{
		Title string
		Post  post.Post
	}{
		Title: "Some Journal",
		Post:  p,
	}

	layout, err := ioutil.ReadFile("testdata/_layouts/" + p.Layout)

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	buf := &bytes.Buffer{}

	if err := Render(buf, "test-post", string(layout), data); err != nil {
		t.Errorf("failed to render template: %s\n", err)
	}

	b, err := ioutil.ReadFile("testdata/post.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if buf.String() != string(b) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", b)
		t.Errorf("recieved:\n%s\n", buf.String())
	}
}

func TestPageRender(t *testing.T) {
	p, err := page.Find(pageId)

	if err != nil {
		t.Errorf("failed to find page %s: %s\n", pageId, err)
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load page %s: %s\n", p.ID, err)
	}

	p.Render()

	data := struct{
		Title string
		Page  page.Page
	}{
		Title: "Some Page",
		Page:  p,
	}

	layout, err := ioutil.ReadFile("testdata/_layouts/" + p.Layout)

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	buf := &bytes.Buffer{}

	if err := Render(buf, "test-page", string(layout), data); err != nil {
		t.Errorf("failed to render template: %s\n", err)
	}

	b, err := ioutil.ReadFile("testdata/page.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if buf.String() != string(b) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", b)
		t.Errorf("recieved:\n%s\n", buf.String())
	}
}

func TestPrintCategories(t *testing.T) {
	c, err := category.All()

	if err != nil {
		t.Errorf("failed to get categories: %s\n", err)
	}

	b, err := ioutil.ReadFile("testdata/categories.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	b = b[:len(b) - 1]

	categories := printCategories(c)

	if categories != string(b) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", b)
		t.Errorf("recieved:\n%s\n", categories)
	}
}

func TestPartial(t *testing.T) {
	c, err := category.All()

	if err != nil {
		t.Errorf("failed to get categories: %s\n", err)
	}

	s, err := partial("partial", c)

	if err != nil {
		t.Errorf("failed to render partial: %s\n", err)
	}

	b, err := ioutil.ReadFile("testdata/partial.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

//	b = b[:len(b) - 1]

	if s != string(b) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", b)
		t.Errorf("recieved:\n%s\n", s)
	}
}
