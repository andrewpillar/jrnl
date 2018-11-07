package post

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/page"
)

var (
	postId            = "post"
	postCategoryId    = "category-one/post"
	postSubCategoryId = "category-one/category-two/post"

	postHref            = "/2006/01/02/post"
	postCategoryHref    = "/category-one/2006/01/02/post"
	postSubCategoryHref = "/category-one/category-two/2006/01/02/post"
)

func init() {
	meta.PostsDir = "testdata/_posts"
}

func TestAll(t *testing.T) {
	p, err := All()

	if err != nil {
		t.Errorf("failed to get posts: %s\n", err)
	}

	if len(p) != 3 {
		t.Errorf("expected 3 posts but found %d\n", len(p))
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

	pc, err := Find(postCategoryId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postCategoryId, err)
	}

	if pc.Href() != postCategoryHref {
		t.Errorf("expected post href to be %s it was %s\n", postCategoryHref, pc.Href())
	}

	if !pc.HasCategory() {
		t.Errorf("expected post %s to have category, it did not\n", pc.ID)
	}

	psc, err := Find(postSubCategoryId)

	if err != nil {
		t.Errorf("failed to find post %s: %s\n", postSubCategoryId, err)
	}

	if !psc.HasCategory() {
		t.Errorf("expected post %s to have category, it did not\n", psc.ID)
	}

	if psc.Href() != postSubCategoryHref {
		t.Errorf("expected post href to be %s it was %s\n", postSubCategoryHref, psc.Href())
	}
}

func TestTouchRemove(t *testing.T) {
	pg := page.New("new post")

	p := New(&pg, "")

	if err := p.Touch(); err != nil {
		t.Errorf("failed to touch new post %s: %s\n", p.ID, err)
	}

	_, err := os.Stat(p.SourcePath)

	if err != nil {
		t.Errorf("failed to stat file: %s\n", err)
	}

	if err := p.Remove(); err != nil {
		t.Errorf("failed to remove post %s: %s\n", p.ID, err)
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

	pb, err := ioutil.ReadFile("testdata/preview_plain.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	// Remove trailing new-line.
	pb = pb[:len(pb) - 1]

	if p.Preview != string(pb) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", pb)
		t.Errorf("recieved:\n%s\n", p.Preview)
	}

	bb, err := ioutil.ReadFile("testdata/body_plain.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Body != string(bb) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", bb)
		t.Errorf("recieved:\n%s\n", p.Body)
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

	pb, err := ioutil.ReadFile("testdata/preview.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Preview != string(pb) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", pb)
		t.Errorf("recieved:\n%s\n", p.Preview)
	}

	bb, err := ioutil.ReadFile("testdata/body.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Body != string(bb) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", bb)
		t.Errorf("recieved:\n%s\n", p.Body)
	}
}
