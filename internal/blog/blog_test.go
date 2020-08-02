package blog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrewpillar/jrnl/internal/config"
)

func Test_GetCategory(t *testing.T) {
	tests := []struct{
		ID           string
		ExpectedID   string
		ExpectedName string
		ShouldFail   bool
	}{
		{ID: "category", ExpectedID: "category", ExpectedName: "Category", ShouldFail: false},
		{ID: "parent/child", ExpectedID: "parent/child", ExpectedName: "Parent / Child", ShouldFail: false},
		{ID: "foo", ShouldFail: true},
	}

	for _, test := range tests {
		c, err := GetCategory(test.ID)

		if err != nil {
			if test.ShouldFail {
				continue
			}
			t.Fatal(err)
		}

		if c.ID != test.ExpectedID {
			t.Errorf("category id mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedID, c.ID)
		}

		if c.Name != test.ExpectedName {
			t.Errorf("category name mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedID, c.ID)
		}
	}
}

func Test_GetPage(t *testing.T) {
	tests := []struct{
		ID             string
		ExpectedID     string
		ExpectedTitle  string
		ExpectedSource string
		ExpectedSite   string
		ShouldFail     bool
	}{
		{
			ID:             "about",
			ExpectedID:     "about",
			ExpectedTitle:  "About",
			ExpectedSource: filepath.Join(config.PagesDir, "about.md"),
			ExpectedSite:   filepath.Join(config.SiteDir, "about", "index.html"),
			ShouldFail:     false,
		},
		{
			ID:             "contact",
			ExpectedID:     "contact",
			ExpectedTitle:  "Contact",
			ExpectedSource: filepath.Join(config.PagesDir, "contact.md"),
			ExpectedSite:   filepath.Join(config.SiteDir, "contact", "index.html"),
			ShouldFail:     false,
		},
		{ID: "foo", ShouldFail: true},
	}

	for _, test := range tests {
		p, err := GetPage(test.ID)

		if err != nil {
			if test.ShouldFail {
				continue
			}
			t.Fatal(err)
		}

		if p.ID != test.ExpectedID {
			t.Errorf("page id mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedID, p.ID)
		}

		if p.Title != test.ExpectedTitle {
			t.Errorf("page title mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedTitle, p.Title)
		}

		if p.SourcePath != test.ExpectedSource {
			t.Errorf("page source path mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedSource, p.SourcePath)
		}

		if p.SitePath != test.ExpectedSite {
			t.Errorf("page site path mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedSite, p.SitePath)
		}
	}
}

func Test_GetPost(t *testing.T) {
	tests := []struct{
		ID             string
		ExpectedID     string
		ExpectedTitle  string
		ExpectedSource string
		ExpectedSite   string
		ExpectedIndex  bool
		ExpectedLayout string
		ShouldFail     bool
	}{
		{
			ID:             "some-post",
			ExpectedID:     "some-post",
			ExpectedTitle:  "Some Post",
			ExpectedSource: filepath.Join(config.PostsDir, "some-post.md"),
			ExpectedSite:   filepath.Join(config.SiteDir, "2006", "01", "02", "some-post", "index.html"),
			ExpectedIndex:  false,
			ExpectedLayout: "post",
			ShouldFail:     false,
		},
		{
			ID:             "category/post",
			ExpectedID:     "category/post",
			ExpectedTitle:  "Some Category Post",
			ExpectedSource: filepath.Join(config.PostsDir, "category", "post.md"),
			ExpectedSite:   filepath.Join(config.SiteDir, "category", "2006", "01", "02", "post", "index.html"),
			ExpectedIndex:  true,
			ExpectedLayout: "category-post",
			ShouldFail:     false,
		},
		{ID: "foo", ShouldFail: true},
	}

	for _, test := range tests {
		p, err := GetPost(test.ID)

		if err != nil {
			if test.ShouldFail {
				continue
			}
			t.Fatal(err)
		}

		if p.ID != test.ExpectedID {
			t.Errorf("post id mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedID, p.ID)
		}

		if p.Title != test.ExpectedTitle {
			t.Errorf("post title mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedTitle, p.Title)
		}

		if p.SourcePath != test.ExpectedSource {
			t.Errorf("post source path mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedSource, p.SourcePath)
		}

		if p.SitePath != test.ExpectedSite {
			t.Errorf("post site path mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedSite, p.SitePath)
		}

		if p.Index != test.ExpectedIndex {
			t.Errorf("post index mismatch\n\texpected = '%v'\n\tactual   = '%v'\n", test.ExpectedIndex, p.Index)
		}

		if p.Layout != test.ExpectedLayout {
			t.Errorf("post layout mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedLayout, p.Layout)
		}
	}
}

func Test_NewPageTouchRemove(t *testing.T) {
	tests := []struct{
		Title      string
		ExpectedID string
	}{
		{Title: "Page One", ExpectedID: "page-one"},
		{Title: "Page / Two", ExpectedID: "page-two"},
		{Title: "!$%Page*()Three\\\\", ExpectedID: "page-three"},
	}

	for _, test := range tests {
		p := NewPage(test.Title)

		if err := p.Touch(); err != nil {
			t.Fatal(err)
		}

		if p.ID != test.ExpectedID {
			t.Errorf("page id mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedID, p.ID)
		}

		if err := p.Remove(); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_NewPostTouchRemove(t *testing.T) {
	tests := []struct{
		Title      string
		Category   string
		ExpectedID string
	}{
		{Title: "Post One", ExpectedID: "post-one"},
		{Title: "Post Two", Category: "new-category", ExpectedID: "new-category/post-two"},
	}

	for _, test := range tests {
		p := NewPost(test.Title, test.Category)

		if err := p.Touch(); err != nil {
			t.Fatal(err)
		}

		if p.ID != test.ExpectedID {
			t.Errorf("post id mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedID, p.ID)
		}

		if err := p.Remove(); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_PageLoad(t *testing.T) {
	tests := []struct{
		Page          string
		ExpectedTitle string
		ExpectedBody  string
	}{
		{"about", "About", "About page.\n"},
		{"contact", "Contact", "Contact **page**\n"},
	}

	for _, test := range tests {
		p, err := GetPage(test.Page)

		if err != nil {
			t.Fatal(err)
		}

		if p.Title != test.ExpectedTitle {
			t.Errorf("page title mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedTitle, p.Title)
		}

		if p.Body != test.ExpectedBody {
			t.Errorf("page body mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedBody, p.Body)
		}
	}
}

func Test_PostLoad(t *testing.T) {
	tests := []struct{
		Post                string
		ExpectedTitle       string
		ExpectedDescription string
		ExpectedBody        string
	}{
		{"some-post", "Some Post", "Some **post**\nAnother line", "Some **post**\nAnother line\n\nAnother paragraph\n"},
		{"category/post", "Some Category Post", "Some category **post**", "Some category **post**\n"},
	}

	for _, test := range tests {
		p, err := GetPost(test.Post)

		if err != nil {
			t.Fatal(err)
		}

		if p.Title != test.ExpectedTitle {
			t.Errorf("post title mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedTitle, p.Title)
		}

		if p.Description != test.ExpectedDescription {
			t.Errorf("post description mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedDescription, p.Description)
		}

		if p.Body != test.ExpectedBody {
			t.Errorf("post body mismatch\n\texpected = '%s'\n\tactual   = '%s'\n", test.ExpectedBody, p.Body)
		}
	}
}

func Test_Roll(t *testing.T) {
	url := os.Getenv("ROLL_URLS")

	if url == "" {
		t.Skip("ROLL_URLS not set")
	}

	feed, err := GetRoll(strings.Split(url, ",")...)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range feed {
		t.Log(item.Title)
	}
}

func TestMain(m *testing.M) {
	config.PostsDir = filepath.Join("testdata", "_posts")
	config.PagesDir = filepath.Join("testdata", "_pages")
	config.SiteDir = filepath.Join("testdata", "_site")
	config.ThemesDir = filepath.Join("testdata", "_themes")
	config.LayoutsDir = filepath.Join("testdata", "_layouts")
	config.AssetsDir = filepath.Join(config.SiteDir, "assets")

	os.Exit(m.Run())
}
