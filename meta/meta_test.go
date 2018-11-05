package meta

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	_, err := Init(".")

	if err != nil {
		t.Errorf("failed to initialize meta file: %s\n", err)
	}

	os.Remove("_meta.yaml")
}

func TestOpen(t *testing.T) {
	File = "../testdata/_meta.yaml"

	m, err := Open()

	if err != nil {
		t.Errorf("failed to open meta file: %s\n", err)
	}

	defer m.Close()
}

func TestSave(t *testing.T) {
	File = "../testdata/_meta.yaml"

	m, err := Open()

	if err != nil {
		t.Errorf("failed to open meta file: %s\n", err)
	}

	defer m.Close()

	if err := m.Save(); err != nil {
		t.Errorf("failed to save meta file: %s\n", err)
	}
}
