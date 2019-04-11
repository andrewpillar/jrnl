package post

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/page"
)

func TestAll(t *testing.T) {
	posts, err := All()

	if err != nil {
		t.Errorf("failed to get all posts: %s\n", err)
		return
	}

	expectedLen := 2

	if len(posts) != expectedLen {
		t.Errorf(
			"post count does not match: expected = '%d', actual = '%d'\n",
			expectedLen,
			len(posts),
		)
		return
	}

	expectedId := "category/post"

	if posts[0].ID != expectedId {
		t.Errorf(
			"post id does not match: expected = '%s', actual = '%s'\n",
			expectedId,
			posts[0].ID,
		)
	}
}

func TestFind(t *testing.T) {
	p, err := Find("some-post")

	if err != nil {
		t.Errorf("failed to find post: %s\n", err)
		return
	}

	expectedSourcePath := filepath.Join(config.PostsDir, "some-post.md")

	if p.SourcePath != expectedSourcePath {
		t.Errorf(
			"post source path does not match: expected = '%s', actual = '%s'\n",
			expectedSourcePath,
			p.SourcePath,
		)
		return
	}

	expectedSitePath := filepath.Join(config.SiteDir, "2006", "01", "02", "some-post", "index.html")

	if p.SitePath != expectedSitePath {
		t.Errorf(
			"post site path does not match: expected = '%s', actual = '%s'\n",
			expectedSitePath,
			p.SitePath,
		)
		return
	}

	if err := p.Load(); err != nil {
		t.Errorf("failed to load post: %s\n", err)
		return
	}

	p.Render()

	expectedHTML := "<p>Some <strong>post</strong></p>\n"

	if p.Body != expectedHTML {
		t.Errorf(
			"post body does not match: expected = '%s', actual = '%s'\n",
			expectedHTML,
			p.Body,
		)
	}
}

func TestTouch(t *testing.T) {
	p := New(page.New("new-post"), "")

	if err := p.Touch(); err != nil {
		t.Errorf("failed to touch post: %s\n", err)
		return
	}

	if err := p.Remove(); err != nil {
		t.Errorf("failed to remove post: %s\n", err)
	}
}

func TestMain(m *testing.M) {
	config.PostsDir = "testdata"

	code := m.Run()

	os.Exit(code)
}
