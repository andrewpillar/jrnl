package site

import (
	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/feed"
	"github.com/andrewpillar/jrnl/page"
)

type Site struct {
	Title      string
	Link       string
	Categories []category.Category
	Pages      []*page.Page
	Blogroll   []feed.Feed
}
