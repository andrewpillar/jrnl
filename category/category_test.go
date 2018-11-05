package category

import (
	"testing"

	"github.com/andrewpillar/jrnl/meta"
)

var (
	categoryId = "some-category"
	subCategoryId = "parent/child"

	categoryHref = "/some-category"
	subCategoryHref = "/parent/child"

	categoryName = "Some Category"
	subCategoryName = "Parent / Child"
)

func init() {
	meta.PostsDir = "../testdata/_posts"
}

func TestAll(t *testing.T) {
	_, err := All()

	if err != nil {
		t.Errorf("failed to get categories: %s\n", err)
	}
}

func TestFind(t *testing.T) {
	c, err := Find(categoryId)

	if err != nil {
		t.Errorf("failed to find category %s: %s\n", categoryId, err)
	}

	if c.Name != categoryName {
		t.Errorf("expected category name to be %s it was %s\n", categoryName, c.Name)
	}

	if c.Href() != categoryHref {
		t.Errorf("expected category href to be %s it was %s\n", categoryHref, c.Href())
	}

	sc, err := Find(subCategoryId)

	if err != nil {
		t.Errorf("failed to find sub-category %s: %s\n", subCategoryId, err)
	}

	if sc.Href() != subCategoryHref {
		t.Errorf("expected category href to be %s it was %s\n", subCategoryHref, sc.Href())
	}

	if sc.Name != subCategoryName {
		t.Errorf("expected sub-category name to be %s it was %s\n", subCategoryName, sc.Name)
	}
}
