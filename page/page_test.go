package page

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewpillar/jrnl/config"
)

func TestAll(t *testing.T) {
	pages, err := All()

	if err != nil {
		t.Errorf("failed to get all pages: %s\n", err)
		return
	}

	 expectedLen := 2

	if len(pages) != expectedLen {
		t.Errorf(
			"page count does not match: expected = '%d', actual = '%d'",
			expectedLen,
			len(pages),
		)
		return
	}

	expectedId := "about"

	if pages[0].ID != expectedId {
		t.Errorf(
			"page id does not match: expected = '%s', actual = '%s'",
			expectedId,
			pages[0].ID,
		)
	}
}

func TestFind(t *testing.T) {
	p, err := Find("contact")

	if err != nil {
		t.Errorf("failed to find page: %s\n", err)
		return
	}

	expectedTitle := "Contact"

	if p.Title != expectedTitle {
		t.Errorf(
			"page title does not match: expected = '%s', actual = '%s'",
			expectedTitle,
			p.Title,
		)
		return
	}

	expectedSourcePath := filepath.Join("testdata", "contact.md")

	if p.SourcePath != expectedSourcePath {
		t.Errorf(
			"page source path does not match: expected = '%s', actual = '%s'",
			expectedSourcePath,
			p.SourcePath,
		)
		return
	}

	expectedSitePath := filepath.Join(config.SiteDir, "contact", "index.html")

	if p.SitePath != expectedSitePath {
		t.Errorf(
			"page site path does not match: expected = '%s', actual = '%s'",
			expectedSitePath,
			p.SitePath,
		)
		return
	}

	expectedHref := "/contact"

	if p.Href() != expectedHref {
		t.Errorf(
			"page href does not match: expected = '%s', actual = '%s'",
			expectedHref,
			p.Href(),
		)
		return
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load page: %s\n", err)
		return
	}

	expectedHTML := "<p>Contact <strong>page</strong></p>\n"

	p.Render()

	if p.Body != expectedHTML {
		t.Errorf(
			"page body does not match: expected = '%s', actual = '%s'",
			expectedHTML,
			p.Body,
		)
	}
}

func TestTouch(t *testing.T) {
	p := New("some-page")

	if err := p.Touch(); err != nil {
		t.Errorf("failed to touch page: %s\n", err)
		return
	}

	if err := p.Remove(); err != nil {
		t.Errorf("failed to remove page: %s\n", err)
	}
}

func TestMain(m *testing.M) {
	config.PagesDir = "testdata"

	code := m.Run()

	os.Exit(code)
}
