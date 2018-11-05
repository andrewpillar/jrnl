package template

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/post"
)

var postId = "some-post"

func init() {
	meta.PostsDir = "../testdata/_posts"
}

func TestRender(t *testing.T) {
	p, err := post.Find(postId)

	if err != nil {
		t.Errorf("failed to find post %s:\n", postId)
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

	buf := &bytes.Buffer{}

	b, err := ioutil.ReadFile("../testdata/_layouts/" + p.Layout)

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if err := Render(buf, "test-template", string(b), data); err != nil {
		t.Errorf("failed to render template: %s\n", err)
	}

	rendered, err := ioutil.ReadFile("../testdata/render.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if buf.String() != string(rendered) {
		t.Errorf("rendered output did not match\n")
		t.Errorf("expected:\n%s\n", rendered)
		t.Errorf("recevied:\n%s\n", buf.String())
	}
}

func TestPrintCategories(t *testing.T) {
	c, err := category.All()

	if err != nil {
		t.Errorf("failed to get categories: %s\n", err)
	}

	expected, err := ioutil.ReadFile("../testdata/categories.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	// Remove new-line
	expected = expected[:len(expected) - 1]

	categories := printCategories(c)

	if categories != string(expected) {
		t.Errorf("rendered output did not match\n")
		t.Errorf("expected:\n%s\n", expected)
		t.Errorf("received:\n%s\n", categories)
	}
}
