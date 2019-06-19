package config

import (
	"os"
	"path/filepath"
	"testing"
)

func Test(t *testing.T) {
	root = "testdata"

	if err := Create(root); err != nil {
		t.Errorf("failed to create file: %s\n", err)
		return
	}

	c, err := Open()

	if err != nil {
		t.Errorf("failed to open file: %s\n", err)
		return
	}

	c.Site.Title = "test"

	if err := c.Save(); err != nil {
		t.Errorf("failed to save file: %s\n", err)
		return
	}

	c.Close()

	os.Remove(filepath.Join(root, "jrnl.toml"))
}
