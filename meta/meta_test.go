package meta

import (
	"os"
	"testing"
)

var (
	title  = "meta file"
	editor = "editor"
	theme  = "theme"
)

func TestMeta(t *testing.T) {
	m, err := Init(".")

	if err != nil {
		t.Errorf("failed to initialize meta file: %s\n", err)
	}

	m.Close()

	m, err = Open()

	if err != nil {
		t.Errorf("failed to open meta file: %s\n", err)
	}

	m.Title = title
	m.Editor = editor
	m.Theme = theme

	if err := m.Save(); err != nil {
		t.Errorf("failed to save meta file: %s\n", err)
	}

	m.Close()

	m, err = Open()

	if err != nil {
		t.Errorf("failed to open meta file: %s\n", err)
	}

	if m.Title != title {
		t.Errorf("expected meta title to be %s it was %s\n", title, m.Title)
	}

	if m.Editor != editor {
		t.Errorf("expected meta editor to be %s it was %s\n", editor, m.Editor)
	}

	if m.Theme != theme {
		t.Errorf("expected meta theme to be %s it was %s\n", theme, m.Theme)
	}

	m.Close()

	os.Remove(File)
}
