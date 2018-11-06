package util

import (
	"bytes"
	"os"
	"testing"
)

var (
	slugInput    = "foo  -- bar //zap"
	slugOutput   = "foo-bar-zap"
	deslugOutput = "foo bar zap"

	titleOutput = "Foo Bar Zap"
)

type frontMatter struct {
	Title  string
	Layout string
}

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
	f, err := os.Open("../testdata/front-matter")

	if err != nil {
		t.Errorf("failed to open file: %s\n", err)
	}

	defer f.Close()

	fm := &frontMatter{}

	if err := UnmarshalFrontMatter(fm, f); err != nil {
		t.Errorf("failed to unmarshal front matter: %s\n", err)
	}

	buf := &bytes.Buffer{}

	if err := MarshalFrontMatter(fm, buf); err != nil {
		t.Errorf("failed to marshal front matter: %s\n", err)
	}

	println(buf.String())
}
