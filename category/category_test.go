package category

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewpillar/jrnl/config"
)

func TestFind(t *testing.T) {
	c, err := Find(filepath.Join("category-one", "category-two"))

	if err != nil {
		t.Errorf("failed to find category: %s\n", err)
		return
	}

	expected := "Category One / Category Two"

	if c.Name != expected {
		t.Errorf("category name does not match: expected = '%s', actual = '%s'\n", expected, c.Name)
	}
}

func TestAll(t *testing.T) {
	categories, err := All()

	if err != nil {
		t.Errorf("failed to get categories: %s\n", err)
		return
	}

	expectedLen := 2

	if len(categories) != expectedLen {
		t.Errorf(
			"category count does not match: expected = '%d', actual = '%d'\n",
			expectedLen,
			len(categories),
		)
		return
	}

	expectedId := "category-one"

	if categories[0].ID != expectedId {
		t.Errorf(
			"category id does not match: expected = '%s', actual = '%s'\n",
			expectedId,
			categories[0].ID,
		)
	}
}

func TestMain(m *testing.M) {
	config.PostsDir = "testdata"

	code := m.Run()

	os.Exit(code)
}
