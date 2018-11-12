package theme

import (
	"os"
	"testing"

	"github.com/andrewpillar/jrnl/meta"
)

var themeName = "test-theme"

func init() {
	meta.LayoutsDir = "testdata/_layouts"
	meta.SiteDir = "testdata/_site"
	meta.AssetsDir = "testdata/_site/assets"
	meta.ThemesDir = "testdata/_themes"
	meta.File = "testdata/_meta.yml"
}

func TestTheme(t *testing.T) {
	theme, err := New(themeName)

	if err != nil {
		t.Errorf("failed to create new theme %s: %s\n", themeName, err)
	}

	theme.Close()

	theme, err = Find(themeName)

	if err != nil {
		t.Errorf("failed to find theme %s: %s\n", themeName, err)
	}

	defer theme.Close()

	if err := theme.Save(); err != nil {
		t.Errorf("failed to save theme %s: %s\n", theme.Name, err)
	}

	os.Remove(theme.Path)
}
