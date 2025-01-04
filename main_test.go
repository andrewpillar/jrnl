package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type commandTest struct {
	cmd   []string
	check func(t *testing.T, id int, args []string)
}

func (t *commandTest) args() []string {
	return append([]string{"jrnl"}, t.cmd...)
}

func checkDirsInitialized(t *testing.T, id int, args []string) {
	if err := Initialized("."); err != nil {
		t.Fatalf("tests[%d]: command %q failed: %s\n", id, args, err)
	}

	ents, err := embeds.ReadDir("embed")

	if err != nil {
		t.Fatalf("tests[%d]: failed to read dir: %s\n", id, err)
	}

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		name := ent.Name()

		b, err := embeds.ReadFile(filepath.Join("embed", name))

		if err != nil {
			t.Fatalf("tests[%d]: %s\n", id, err)
		}

		expected := sha256.New()
		expected.Write(b)

		b, err = os.ReadFile(filepath.Join(layoutDir, name[:len(name)-5]))

		if err != nil {
			t.Fatalf("tests[%d]: %s\n", id, err)
		}

		actual := sha256.New()
		actual.Write(b)

		if !bytes.Equal(expected.Sum(nil), actual.Sum(nil)) {
			t.Fatalf("tests[%d]: %s does not match what is expected\n", id, name)
		}
	}
}

func checkPage(dir, name string) func(*testing.T, int, []string) {
	return func(t *testing.T, id int, args []string) {
		_, err := LoadPage(filepath.Join(dir, name+".md"))

		if err != nil {
			t.Fatalf("tests[%d]: failed to load page: %s\n", id, err)
		}
	}
}

func checkTheme(name string) func(*testing.T, int, []string) {
	return func(t *testing.T, id int, args []string) {
		dir, err := themeDir()

		if err != nil {
			t.Fatalf("tests[%d]: failed to get theme dir: %s\n", id, err)
		}

		f, err := os.Open(filepath.Join(dir, name) + ".tar.gz")

		if err != nil {
			t.Fatalf("tests[%d]: failed to open theme: %s\n", id, err)
		}

		defer f.Close()

		gzr, err := gzip.NewReader(f)

		if err != nil {
			t.Fatalf("tests[%d]: failed to read theme: %s\n", id, err)
		}

		defer gzr.Close()

		tr := tar.NewReader(gzr)

		files := make(map[string]struct{})

		for {
			hdr, err := tr.Next()

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				t.Fatalf("tests[%d]: failed to read tar entry: %s\n", id, err)
			}

			if hdr.Typeflag == tar.TypeReg {
				files[hdr.Name] = struct{}{}
			}
		}

		expected := [...]string{
			"_layouts/home",
			"_layouts/page",
			"_layouts/post",
		}

		for _, path := range expected {
			if _, ok := files[path]; !ok {
				t.Fatalf("tests[%d]: could not find %q in %q\n", id, path, f.Name())
			}
		}
	}
}

func checkPageDeletedFromRemote(page string) func(*testing.T, int, []string) {
	return func(t *testing.T, id int, args []string) {
		path := filepath.Join("remote", page, "index.html")

		_, err := os.Stat(path)

		if err == nil {
			t.Fatalf("tests[%d]: expected path %q to be deleted\n", id, path)
		}
	}
}

func Test_Jrnl(t *testing.T) {
	if err := os.Chdir("testdata"); err != nil {
		t.Fatal(err)
	}

	tests := [...]commandTest{
		{
			cmd:   []string{"init"},
			check: checkDirsInitialized,
		},
		{
			cmd:   []string{"page", "-l", "page", "About"},
			check: checkPage(pageDir, "about"),
		},
		{
			cmd:   []string{"post", "-l", "post", "First post"},
			check: checkPage(postDir, time.Now().Format("2006-01-02")+"-first-post"),
		},
		{
			cmd:   []string{"post", "-l", "post", "Second post"},
			check: checkPage(postDir, time.Now().Format("2006-01-02")+"-second-post"),
		},
		{
			cmd: []string{"publish", "-d", "-v"},
		},
		{
			cmd: []string{"publish", "-v"},
		},
		{
			cmd:   []string{"theme", "save", "default"},
			check: checkTheme("default"),
		},
		{
			cmd: []string{"theme", "ls"},
		},
		{
			cmd: []string{"theme", "use", "default"},
		},
		{
			cmd: []string{"theme", "rm", "default"},
		},
		{
			cmd: []string{"flush"},
		},
		{
			cmd:   []string{"rm", "about"},
			check: checkPageDeletedFromRemote("about"),
		},
	}

	os.Setenv("EDITOR", "true")
	os.Setenv("HOME", ".")

	for i, test := range tests {
		t.Log("running command", test.args())

		if err := run(test.args()); err != nil {
			t.Fatalf("tests[%d]: command %q failed: %s\n", i, test.args(), err)
		}

		if test.check != nil {
			test.check(t, i, test.args())
		}

		// jrnl init passed so load the config and update it.
		if i == 0 {
			cfg, err := LoadConfig()

			if err != nil {
				t.Fatalf("tests[%d]: %s\n", i, err)
			}

			if err := os.MkdirAll("remote", dirMode); err != nil {
				t.Fatalf("tests[%d]: %s\n", i, err)
			}

			cfg.Remote = "file://remote"
			cfg.Author.Name = "Joe Bloggs"
			cfg.Author.Email = "joe.bloggs@localhost"
			cfg.Site.Title = "Joe's Blog"
			cfg.Site.URL = "http://localhost:8080"
			cfg.Site.Atom = "atom.xml"
			cfg.Site.RSS = "rss.xml"

			if err := cfg.Save(); err != nil {
				t.Fatalf("tests[%d]: %s\n", i, err)
			}

			err = func() error {
				f, err := os.Create(filepath.Join(assetDir, "main.css"))

				if err != nil {
					return err
				}

				defer f.Close()

				_, err = io.WriteString(f, "* {\n\tmargin: 0;\n}")
				return err
			}()

			if err != nil {
				t.Fatalf("tests[%d]: %s\n", i, err)
			}
		}
	}
}
