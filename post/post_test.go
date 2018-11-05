package post

import (
	"io/ioutil"
	"testing"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/page"
)

var (
	postId = "some-post"
	postCategoryId = "some-category/some-post"
	postSubCategoryId = "parent/child/some-post"

	postHref = "/2006/01/02/some-post"
	categoryPostHref = "/some-category/2006/01/02/some-post"

	categoryName = "Some Category"
	subCategoryName = "Parent / Child"
)

func init() {
	meta.PostsDir = "testdata/_posts"
}

func TestAll(t *testing.T) {
	_, err := All()

	if err != nil {
		t.Errorf("expected to get posts but failed: %s\n", err)
	}
}

func TestFind(t *testing.T) {
	p, err := Find(postId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postId, err)
	}

	if p.Href() != postHref {
		t.Errorf("expected post href to be %s it was %s\n", postHref, p.Href())
	}

	cp, err := Find(postCategoryId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postCategoryId, err)
	}

	if !cp.HasCategory() {
		t.Errorf("expected post %s to have category it did not\n", cp.ID)
	}

	if cp.Category.Name != categoryName {
		t.Errorf("expected category name to be %s it was %s\n", categoryName, cp.Category.Name)
	}

	if cp.Href() != categoryPostHref {
		t.Errorf("expected post href to be %s it was %s\n", categoryPostHref, cp.Href())
	}

	scp, err := Find(postSubCategoryId)

	if err != nil {
		t.Errorf("expected to find post %s could not: %s\n", postSubCategoryId, err)
	}

	if scp.Category.Name != subCategoryName {
		t.Errorf("expected category name to be %s it was %s\n", subCategoryName, scp.Category.Name)
	}
}

func TestTouchRemove(t *testing.T) {
	pg := page.New("new post")

	p := New(&pg, "")

	if err := p.Touch(); err != nil {
		t.Errorf("failed to touch new post %s: %s\n", p.ID, err)
	}

	if err := p.Remove(); err != nil {
		t.Errorf("failed to remove new post %s: %s\n", p.ID, err)
	}
}

func TestLoad(t *testing.T) {
	p, err := Find(postId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postId, err)
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load post %s: %s\n", p.ID, err)
	}

	preview, err := ioutil.ReadFile("testdata/some-post-preview.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	// Remove the new-line ending.
	preview = preview[:len(preview) - 1]

	if p.Preview != string(preview) {
		t.Errorf("previews did not match\n")
		t.Errorf("expected:\n\t%s\n", string(preview))
		t.Errorf("received:\n\t%s\n", p.Preview)
	}
}

func TestRender(t *testing.T) {
	p, err := Find(postId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postId, err)
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load post %s: %s\n", p.ID, err)
	}

	p.Render()

	md, err := ioutil.ReadFile("testdata/some-post.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Body != string(md) {
		t.Errorf("rendered output did not match\n")
		t.Errorf("expected:\n\t%s\n", md)
		t.Errorf("received:\n\t%s\n", p.Body)
	}
}
