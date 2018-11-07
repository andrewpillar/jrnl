package category

import (
	"testing"

	"github.com/andrewpillar/jrnl/meta"
)

var (
	categoryId    = "category"
	subCategoryId = "category-one/category-two"

	categoryName    = "Category"
	subCategoryName = "Category One / Category Two"

	categoryHref    = "/category"
	subCategoryHref = "/category-one/category-two"
)

func init() {
	meta.PostsDir = "testdata/_posts"
}

func TestAll(t *testing.T) {
	c, err := All()

	if err != nil {
		t.Errorf("failed to get categories: %s\n", err)
	}

	if len(c) != 2 {
		t.Errorf("expected 2 categories but found %d\n", len(c))
	}
}

func TestFind(t *testing.T) {
	c, err := Find(categoryId)

	if err != nil {
		t.Errorf("failed to find cateogry %s: %s\n", categoryId, err)
	}

	if c.Name != categoryName {
		t.Errorf("expected category name to be %s it was %s\n", categoryName, c.Name)
	}

	if c.Href() != categoryHref {
		t.Errorf("expected category href to be %s it was %s\n", c.Href(), categoryHref)
	}

	sc, err := Find(subCategoryId)

	if err != nil {
		t.Errorf("failed to find category %s: %s\n", subCategoryId, err)
	}

	if sc.Name != subCategoryName {
		t.Errorf("expected category name to be %s it was %s\n", subCategoryName, sc.Name)
	}

	if sc.Href() != subCategoryHref {
		t.Errorf("expected category href to be %s it was %s\n", sc.Href(), subCategoryHref)
	}
}
