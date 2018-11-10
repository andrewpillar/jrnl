package page

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/andrewpillar/jrnl/meta"
)

var (
	pageId = "page"

	pageHref = "/page"
)

func init() {
	meta.PagesDir = "testdata/_pages"
}

func TestAll(t *testing.T) {
	p, err := All()

	if err != nil {
		t.Errorf("failed to get pages: %s\n", err)
	}

	if len(p) != 1 {
		t.Errorf("expected 1 page got %d\n", len(p))
	}
}

func TestFind(t *testing.T) {
	p, err := Find(pageId)

	if err != nil {
		t.Errorf("failed to find page %s: %s\n", pageId, err)
	}

	if p.Href() != pageHref {
		t.Errorf("expected page href to be %s it was %s\n", pageHref, p.Href())
	}
}

func TestLoad(t *testing.T) {
	p, err := Find(pageId)

	if err != nil {
		t.Errorf("failed to find page %s: %s\n", pageId, err)
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load page %s: %s\n", p.ID, err)
	}

	b, err := ioutil.ReadFile("testdata/page_plain.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Body != string(b) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", b)
		t.Errorf("recieved:\n%s\n", p.Body)
	}
}

func TestTouchRemove(t *testing.T) {
	p := New("new page")

	if err := p.Touch(); err != nil {
		t.Errorf("failed to touch new page %s: %s\n", p.ID, err)
	}

	_, err := os.Stat(p.SourcePath)

	if err != nil {
		t.Errorf("failed to stat file: %s\n", err)
	}

	if err := p.Remove(); err != nil {
		t.Errorf("failed to remove page %s: %s\n", p.ID, err)
	}
}

func TestRender(t *testing.T) {
	p, err := Find(pageId)

	if err != nil {
		t.Errorf("failed to find page %s: %s\n", pageId, err)
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load page %s: %s\n", p.ID, err)
	}

	p.Render("")

	b, err := ioutil.ReadFile("testdata/page.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Body != string(b) {
		t.Errorf("value mismatch\n")
		t.Errorf("expected:\n%s\n", b)
		t.Errorf("recieved:\n%s\n", p.Body)
	}
}
