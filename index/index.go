package index

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/config"
	"github.com/andrewpillar/jrnl/post"
	"github.com/andrewpillar/jrnl/site"
	"github.com/andrewpillar/jrnl/template"
)

var (
	day   = "day"
	month = "month"
	year  = "year"
	all   = "all"

	reyear  = regexp.MustCompile("[0-9]{4}")
	remonth = regexp.MustCompile("[0-9]{4}/[0-9]{2}")
	reday   = regexp.MustCompile("[0-9]{4}/[0-9]{2}/[0-9]{2}")

	recat = regexp.MustCompile("[-a-z0-9/]+")

	daylayout   = filepath.Join("2006", "01", "02")
	monthlayout = filepath.Join("2006", "01")
	yearlayout  = "2006"
)

type Index map[string][]*post.Post

func writeIndexFile(dst, layout string, data interface{}) error {
	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return err
	}

	if data == nil {
		return errors.New("no data for index " + layout)
	}

	f, err := os.OpenFile(dst, os.O_TRUNC|os.O_CREATE|os.O_RDWR, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	return template.Render(f, layout, string(b), data)
}

func New() Index {
	return Index(make(map[string][]*post.Post))
}

func (i *Index) Put(p *post.Post) {
	if !p.Index {
		return
	}

	parts := strings.Split(p.SitePath, string(os.PathSeparator))

	for j := range parts {
		key := filepath.Join(parts[:len(parts) - j - 2]...)

		if key == "" {
			break
		}

		(*i)[key] = append((*i)[key], p)
	}
}

func extractCategoryId(key, tsubset string) string {
	id := strings.Replace(key, config.SiteDir + string(os.PathSeparator), "", 1)
	id = strings.Replace(id, tsubset, "", 1)


	return strings.TrimSuffix(id, string(os.PathSeparator))
}

func (i Index) Publish(key string, s site.Site) error {
	var data interface{}
	var path string

	posts := i[key]

	sort.Sort(post.ByCreatedAt(posts))

	if key == config.SiteDir {
		data = struct{
			Site  site.Site
			Posts []*post.Post
		}{
			Site:  s,
			Posts: posts,
		}

		path = filepath.Join(config.PostsDir, config.IndexDir, all)

		return writeIndexFile(filepath.Join(key, "index.html"), path, data)
	}

	var tlayout, tvalue, categoryId string

	peek := posts[0]
	dir := filepath.Join(filepath.Dir(peek.SourcePath), config.IndexDir)
	bkey := []byte(key)

	if reday.Match(bkey) {
		tlayout = daylayout
		tvalue = string(reday.Find(bkey))

		path = filepath.Join(dir, day)

		if recat.Match(bkey) {
			categoryId = extractCategoryId(key, tvalue)

			path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, day)
		}
	} else if remonth.Match(bkey) {
		tlayout = monthlayout
		tvalue = string(remonth.Find(bkey))

		path = filepath.Join(dir, month)

		if recat.Match(bkey) {
			categoryId = extractCategoryId(key, tvalue)

			path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, month)
		}
	} else if reyear.Match(bkey) {
		tlayout = yearlayout
		tvalue = string(reyear.Find(bkey))

		path = filepath.Join(dir, year)

		if recat.Match(bkey) {
			categoryId = extractCategoryId(key, tvalue)

			path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, year)
		}
	}

	// Assume we have a category index.
	if path == "" && recat.Match(bkey) {
		categoryId = extractCategoryId(key, "")

		path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, all)
	}

	t, err := time.Parse(tlayout, tvalue)

	if err != nil {
		return err
	}

	if categoryId == "" {
		data = struct{
			Site  site.Site
			Time  time.Time
			Posts []*post.Post
		}{
			Site:  s,
			Time:  t,
			Posts: posts,
		}

		return writeIndexFile(filepath.Join(key, "index.html"), path, data)
	}

	c, err := category.Find(categoryId)

	if err != nil {
		return err
	}

	tempty, _ := time.Parse("", "")

	if t == tempty {
		data = struct{
			Site     site.Site
			Category category.Category
			Posts    []*post.Post
		}{
			Site:     s,
			Category: c,
			Posts:    posts,
		}

		return writeIndexFile(filepath.Join(key, "index.html"), path, data)
	}

	data = struct{
		Site     site.Site
		Category category.Category
		Time     time.Time
		Posts    []*post.Post
	}{
		Site:     s,
		Category: c,
		Time:     t,
		Posts:    posts,
	}

	return writeIndexFile(filepath.Join(key, "index.html"), path, data)
}
