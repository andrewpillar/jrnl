package util

import (
	"bytes"
	"os"
	"testing"
)

var (
	slugInput    = "foo / bar   ###zap"
	slugOutput   = "foo-bar-zap"
	deslugOutput = "foo bar zap"

	titleOutput = "Foo Bar Zap"

	dir      = "testdata/dir"
	emptyDir = "testdata/empty-dir"

	tmpDir = "testdata/root/empty-1/empty-2/empty-3"

	copyDir = "testdata/copy/src"
	copyDst = "testdata/copy/dst"

	tarIn  = "testdata/tar"
	tarOut = "testdata/tar.gz"
)

func TestSlug(t *testing.T) {
	s := Slug(slugInput)

	if s != slugOutput {
		t.Errorf("expected string %s to be %s it was %s\n", slugInput, slugOutput, s)
	}
}

func TestDeslug(t *testing.T) {
	s := Slug(slugInput)
	d := Deslug(s)

	if d != deslugOutput {
		t.Errorf("expected string %s to be %s it was %s\n", slugInput, deslugOutput, s)
	}
}

func TestTitle(t *testing.T) {
	title := Title(deslugOutput)

	if title != titleOutput {
		t.Errorf("expected string %s to be %s it was %s\n", deslugOutput, titleOutput, title)
	}
}

func TestFrontMatter(t *testing.T) {
	f, err := os.Open("testdata/front-matter.input")

	if err != nil {
		t.Errorf("failed to open file: %s\n", err)
	}

	defer f.Close()

	fm := &struct{
		Title  string
		Layout string
	}{}

	if err := UnmarshalFrontMatter(fm, f); err != nil {
		t.Errorf("failed to unmarshal front matter: %s\n", err)
	}

	buf := &bytes.Buffer{}

	if err := MarshalFrontMatter(fm, buf); err != nil {
		t.Errorf("failed to marshal front matter: %s\n", err)
	}
}

func TestEmptyDir(t *testing.T) {
	if DirEmpty(dir) {
		t.Errorf("expected dir %s to not be empty, it was not\n", dir)
	}

	if !DirEmpty(emptyDir) {
		t.Errorf("expected dir %s to be empty, it was not\n", emptyDir)
	}
}

func TestRemoveEmptyDirs(t *testing.T) {
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		t.Errorf("failed to make directories: %s\n", err)
	}

	if err := RemoveEmptyDirs("testdata/root", tmpDir); err != nil {
		t.Errorf("failed to remove empty dirs: %s\n", err)
	}

	_, err := os.Stat(tmpDir)

	if err == nil {
		t.Errorf("expected dir %s to not exist, it does\n", tmpDir)
	}

	if err != nil {
		if !os.IsNotExist(err) {
			t.Errorf("failed to stat dir %s: %s\n", tmpDir, err)
		}
	}
}

func TestCopy(t *testing.T) {
	if err := Copy(copyDir, copyDst); err != nil {
		t.Errorf("failed to copy dir %s: %s\n", copyDir, err)
	}

	if err := os.RemoveAll(copyDst); err != nil {
		t.Errorf("failed to remove dir %s: %s\n", copyDst, err)
	}
}
