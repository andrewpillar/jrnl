package index

import (
	"errors"
	"io"
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

// Index is a simple data structure for indexing posts based off of their
// creation date and category, if they belong to one. There are four types of
// indexes:
//
// * all
// * day
// * month
// * year
//
// And the layouts for these are stored in the _index directory. During
// publishing the index key being published against will be checked to
// determine the type of index being published.
type Index map[string][]*post.Post

// Take the given index layout and data, and write it to the given io.Writer.
func writeIndex(w io.Writer, layout string, data interface{}) error {
	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return err
	}

	if data == nil {
		return errors.New("no data for index " + layout)
	}

	return template.Render(w, layout, string(b), data)
}

// Create a new index.
func New() Index {
	return Index(make(map[string][]*post.Post))
}

// Posts will be indexed multiple times based off of the site path. This
// allows for indexing based off of day, month, year, category, and entire site
// posts. For example the given site path:
//
//   _site/2006/01/02/some-post/index.html
//
// Would generate multiple index keys:
//
// _site            => site index
// _site/2006       => year index
// _site/2006/01    => month index
// _site/2006/01/02 => day index
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

// Publish the index held in the given key. The given key will be checked to
// determine the type of index being published. This information is then used
// to determine which layout file should be used, and what data should be
// passed to that layout during templating.
//
// First we check to see if the index key is for the site wide index.
// Afterwards we check to see if the given key is either a day, month, or year
// key. Each time one of these checks happens, an additional check is performed
// to check if it's a category index too. If so we then set the categoryId for
// finding that category.
//
// Once we have got the necessary values from the key pertaining to the date
// information, primarily the layout of the time we want to parse, and the
// value itself; we check to see if we have a categoryId. If we don't we
// proceed with publishing the index, passing it the posts, and the parsed
// time.Time value.
//
// If we do have a categoryId, then we find the category, and pass it to the
// layout data along with the parsed time.Time value, if we have it.
func (i Index) Publish(key string, s site.Site) error {
	var data interface{}
	var path string

	flags := os.O_TRUNC|os.O_CREATE|os.O_RDWR
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

		f, err := os.OpenFile(filepath.Join(key, "index.html"), flags, config.FileMode)

		if err != nil {
			return err
		}

		defer f.Close()

		// Index layout for all site posts.
		path = filepath.Join(config.PostsDir, config.IndexDir, all)

		return writeIndex(f, path, data)
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

			// Index layout for posts on the day.
			path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, day)
		}
	} else if remonth.Match(bkey) {
		tlayout = monthlayout
		tvalue = string(remonth.Find(bkey))

		path = filepath.Join(dir, month)

		if recat.Match(bkey) {
			categoryId = extractCategoryId(key, tvalue)

			// Index layout for posts in the month.
			path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, month)
		}
	} else if reyear.Match(bkey) {
		tlayout = yearlayout
		tvalue = string(reyear.Find(bkey))

		path = filepath.Join(dir, year)

		if recat.Match(bkey) {
			categoryId = extractCategoryId(key, tvalue)

			// Index layout for posts in the year.
			path = filepath.Join(config.PostsDir, categoryId, config.IndexDir, year)
		}
	}

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

		f, err := os.OpenFile(filepath.Join(key, "index.html"), flags, config.FileMode)

		if err != nil {
			return err
		}

		defer f.Close()

		return writeIndex(f, path, data)
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

		f, err := os.OpenFile(filepath.Join(key, "index.html"), flags, config.FileMode)

		if err != nil {
			return err
		}

		defer f.Close()

		return writeIndex(f, path, data)
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

	f, err := os.OpenFile(filepath.Join(key, "index.html"), flags, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	return writeIndex(f, path, data)
}
