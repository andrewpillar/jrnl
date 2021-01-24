package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type checkFunc func(id int, cmd string, t *testing.T)

var (
	pageLayout = []byte(`<html lang="en">
	<head>
		<title>{{.Page.Title}} - {{.Site.Title}}</title>
	</head>
	<body>{{.Page.Body}}</body>
</html>`)

	postLayout = []byte(`<html lang="en">
	<head>
		<title>{{.Post.Title}} - {{.Site.Title}}</title>
	</head>
	<body>{{.Post.Body}}</body>
</html>`)

	indexLayout = []byte(`<html lang="en">
	<head>
		<title>{{.Site.Title}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<strong>{{$p.Title}}</strong>
			<div>{{$p.Description}}</div>
		{{end}}
	</body>
</html>`)

	categoryIndexLayout = []byte(`<html lang="en">
	<head>
		<title>{{.Site.Title}} - {{.Category.Name}}</title>
	</head>
	<body>
		{{range $i, $p := .Posts}}
			<strong>{{$p.Title}}</strong>
			<div>{{$p.Description}}</div>
		{{end}}
	</body>
</html>`)
)

func cleanup(tmpdir string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
	os.Remove("jrnl.toml")
	os.RemoveAll(tmpdir)
}

func checkInitDirs(id int, cmd string, t *testing.T) {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("tests[%d](%s)) - check failed, could not stat dir %s: %s\n", id, cmd, dir, err)
		}
	}

	page, err := os.Create(filepath.Join(layoutsDir, "page"))

	if err != nil {
		t.Fatal(err)
	}

	defer page.Close()

	if _, err := page.Write(pageLayout); err != nil {
		t.Fatal(err)
	}

	post, err := os.Create(filepath.Join(layoutsDir, "post"))

	if err != nil {
		t.Fatal(err)
	}

	defer post.Close()

	if _, err := post.Write(postLayout); err != nil {
		t.Fatal(err)
	}

	index, err := os.Create(filepath.Join(layoutsDir, "index"))

	if err != nil {
		t.Fatal(err)
	}

	if _, err := index.Write(indexLayout); err != nil {
		t.Fatal(err)
	}

	catindex, err := os.Create(filepath.Join(layoutsDir, "category-index"))

	if err != nil {
		t.Fatal(err)
	}

	if _, err := catindex.Write(categoryIndexLayout); err != nil {
		t.Fatal(err)
	}
}

func checkPublished(paths ...string) checkFunc {
	return func(id int, cmd string, t *testing.T) {
		for _, path := range paths {
			if _, err := os.Stat(filepath.Join(siteDir, path)); err != nil {
				t.Fatalf("tests[%d](%s) - failed to stat %q: %s\n", id, cmd, path, err)
			}
		}
	}
}

func checkPublishedRemote(remote string, paths ...string) checkFunc {
	return func(id int, cmd string, t *testing.T) {
		for _, path := range paths {
			if _, err := os.Stat(filepath.Join(remote, path)); err != nil {
				t.Fatalf("tests[%d](%s) - failed to stat %q: %s\n", id, cmd, path, err)
			}
		}
	}
}

func splitargs(argv string) []string {
	args := make([]string, 0)

	n := 0
	off := 0
	quote := false
	end := len(argv) - 1

	for i, r := range argv {
		if r == '\'' {
			quote = !quote
			off = 1
		}

		if r == ' ' || i == end {
			if i == end {
				i++
			}

			if !quote {
				args = append(args, argv[n+off:i-off])
				n = i + 1
				off = 0
			}
		}
	}
	return args
}

func Test_Cmd(t *testing.T) {
	cmd := os.Getenv("TEST_CMD")

	if cmd == "" {
		t.Skip("TEST_CMD not set, skipping...")
	}

	if err := run(splitargs(cmd)); err != nil {
		t.Fatalf("failed to run cmd %q: %s\n", cmd, err)
	}
}

func Test_Jrnl(t *testing.T) {
	if err := os.Chdir("testdata"); err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "jrnl-remote-*")

	if err != nil {
		t.Fatal(err)
	}

	defer cleanup(dir)

	now := time.Now()
	date := strings.Replace(now.Format("2006-01-02"), "-", string(os.PathSeparator), -1)

	tests := []struct {
		cmd       string
		shouldErr bool
		check     checkFunc
	}{
		{
			"jrnl init",
			false,
			checkInitDirs,
		},
		{
			"jrnl page -l page about",
			false,
			nil,
		},
		{
			"jrnl post -l post 'First Post'",
			false,
			nil,
		},
		{
			"jrnl post -l post -c Programming 'Go 101'",
			false,
			nil,
		},
		{
			"jrnl post -l post 'Second Post'",
			false,
			nil,
		},
		{
			"jrnl config site.title 'My blog'",
			false,
			nil,
		},
		{
			"jrnl config site.remote " + dir,
			false,
			nil,
		},
		{
			"jrnl publish",
			false,
			checkPublishedRemote(
				dir,
				filepath.Join(date, "first-post", "index.html"),
				filepath.Join("programming", date, "go-101", "index.html"),
				filepath.Join(date, "second-post", "index.html"),
			),
		},
		{
			"jrnl rm second-post",
			false,
			nil,
		},
		{
			"jrnl publish",
			false,
			checkPublishedRemote(
				dir,
				filepath.Join(date, "first-post", "index.html"),
				filepath.Join("programming", date, "go-101", "index.html"),
			),
		},
	}

	os.Setenv("EDITOR", "true")

	for i, test := range tests {
		cmd := exec.Command(os.Args[0], "-test.run=Test_Cmd")
		cmd.Env = append(os.Environ(), "TEST_CMD="+test.cmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			if !test.shouldErr {
				t.Fatalf("tests[%d](%s) - Failed to run test: %s\n", i, test.cmd, err)
			}
		}

		if test.check != nil {
			test.check(i, test.cmd, t)
		}
	}
}
