package blog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewpillar/jrnl/internal/config"
)

func Test_Theme(t *testing.T) {
	th := NewTheme("test-theme")

	if err := th.Save(); err != nil {
		t.Fatal(err)
	}

	layouts := []string{
		filepath.Join(config.LayoutsDir, "page"),
		filepath.Join(config.LayoutsDir, "post"),
		filepath.Join(config.AssetsDir, "style.css"),
	}

	for _, f := range layouts {
		os.Remove(f)
	}

	if err := th.Load(); err != nil {
		t.Fatal(err)
	}

	for _, f := range layouts {
		if _, err := os.Stat(f); err != nil {
			t.Fatal(err)
		}
	}

	if err := th.Remove(); err != nil {
		t.Fatal(err)
	}
}
