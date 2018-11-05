package page

import (
	"io/ioutil"
	"testing"

	"github.com/andrewpillar/jrnl/meta"
)

var pageId = "some-page"

func init() {
	meta.PagesDir = "../testdata/_pages"
}

func TestAll(t *testing.T) {
	_, err := All()

	if err != nil {
		t.Errorf("failed to get all pages: %s\n", err)
	}
}

func TestFind(t *testing.T) {
	_, err := Find(pageId)

	if err != nil {
		t.Errorf("failed to find page %s: %s\n", pageId, err)
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
}

func TestTouchRemove(t *testing.T) {
	p := New("new page")

	if err := p.Touch(); err != nil {
		t.Errorf("failed to touch page %s: %s\n", p.ID, err)
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

	p.Render()

	md, err := ioutil.ReadFile("../testdata/some-page.golden")

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	if p.Body != string(md) {
		t.Errorf("rendered output did not match\n")
		t.Errorf("expected:\n\t%s\n", md)
		t.Errorf("received:\n\t%s\n", p.Body)
	}
}
