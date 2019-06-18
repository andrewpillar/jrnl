package site

import (
	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/page"
)

// Simple struct that will be passed to all layout templates.
type Site struct {
	Title      string
	Site       string
	Categories []category.Category
	Pages      []*page.Page
}
