package blog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewpillar/jrnl/internal/config"
)

func Test_Index(t *testing.T) {
	pp, err := Posts()

	if err != nil {
		t.Fatal(err)
	}

	index := NewIndex()

	for _, p := range pp {
		index.Put(p)
	}

	s := Site{
		Title: "test site",
		Link:  "https://example.com",
	}

	expectedFiles := []string{
		filepath.Join(config.SiteDir, "category", "2006", "index.html"),
		filepath.Join(config.SiteDir, "category", "2006", "01", "index.html"),
		filepath.Join(config.SiteDir, "category", "2006", "01", "02", "index.html"),
	}

	for key := range index {
		if _, err := index.Write(key, s); err != nil {
			if err == ErrNoLayout {
				continue
			}
			t.Errorf("%q: %s\n", key, err)
		}
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("couldn't stat path: %s\n", err)
		}
	}
	os.RemoveAll(filepath.Join(config.SiteDir, "category"))
}
