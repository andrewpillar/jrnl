package blog

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/internal/config"
	"github.com/andrewpillar/jrnl/internal/template"
)

type Index map[string][]Post

var (
	day   = "day"
	month = "month"
	year  = "year"
	all   = "all"

	reyear  = regexp.MustCompile("[0-9]{4}")
	remonth = regexp.MustCompile("[0-9]{4}/[0-9]{2}")
	reday   = regexp.MustCompile("[0-9]{4}/[0-9]{2}/[0-9]{2}")

	recat = regexp.MustCompile("[-a-z0-9/]+")

	dateLayouts = map[string]string{
		day:   filepath.Join("2006", "01", "02"),
		month: filepath.Join("2006", "01"),
		year:  "2006",
	}

	optionalLayouts = map[string]struct{}{
		filepath.Join(config.PostsDir, config.IndexDir, day):   {},
		filepath.Join(config.PostsDir, config.IndexDir, month): {},
		filepath.Join(config.PostsDir, config.IndexDir, year):  {},
	}

	ErrNoLayout = errors.New("no layout")
)

func isTimeZero(t time.Time) bool {
	empty, _ := time.Parse("", "")
	return t == empty
}

func NewIndex() Index {
	return Index(make(map[string][]Post))
}

func (i *Index) Put(p Post) {
	if !p.Index {
		return
	}

	parts := strings.Split(p.SitePath, string(os.PathSeparator))

	for j := range parts {
		key := filepath.Join(parts[:len(parts)-j-2]...)

		if key == "" || key == config.SiteDir {
			break
		}
		(*i)[key] = append((*i)[key], p)
	}
}

func (i Index) getIndexDataAndLayout(key string, pp []Post, s Site) (interface{}, string, error) {
	if key == config.SiteDir {
		return struct{
			Site  Site
			Posts []Post
		}{
			Site:  s,
			Posts: pp,
		}, filepath.Join(config.PostsDir, config.IndexDir, all), nil
	}

	var (
		dateKind   string = "all"
		dateLayout string
		dateValue  string
		categoryId string
	)

	b := []byte(key)

	if reday.Match(b) {
		dateKind = day
		dateLayout = "day"
		dateValue = string(reday.Find(b))
	} else if remonth.Match(b) {
		dateKind = month
		dateLayout = "month"
		dateValue = string(remonth.Find(b))
	} else if reyear.Match(b) {
		dateKind = year
		dateLayout = "year"
		dateValue = string(reyear.Find(b))
	}

	if recat.Match(b) {
		categoryId = strings.Replace(key, config.SiteDir + string(os.PathSeparator), "", 1)
		categoryId = strings.Replace(categoryId, dateValue, "", 1)
		categoryId = strings.TrimSuffix(categoryId, string(os.PathSeparator))
	}

	path := filepath.Join(config.PostsDir, config.IndexDir, dateKind)

	t, err := time.Parse(dateLayouts[dateLayout], dateValue)

	if err != nil {
		return nil, "", err
	}

	var c Category

	if categoryId != "" {
		c, err = GetCategory(categoryId)

		if err != nil {
			return nil, "", err
		}
	}

	if isTimeZero(t) {
		return struct{
			Site     Site
			Category Category
			Posts    []Post
		}{
			Site:     s,
			Category: c,
			Posts:    pp,
		}, path, nil
	}

	return struct{
		Site     Site
		Category Category
		Time     time.Time
		Posts    []Post
	}{
		Site:     s,
		Category: c,
		Time:     t,
		Posts:    pp,
	}, path, nil
}

func (i Index) Write(key string, s Site) (string, error) {
	pp := i[key]

	index := filepath.Join(key, "index.html")
	dir := filepath.Dir(index)

	sort.Sort(byCreatedAt(pp))

	if key == config.SiteDir {
		data := struct{
			Site  Site
			Posts []Post
		}{Site: s, Posts: pp}

		layout := filepath.Join(config.PostsDir, config.IndexDir, "all")

		f, err := os.OpenFile(index, os.O_TRUNC|os.O_CREATE|os.O_RDWR, config.FileMode)

		if err != nil {
			return "", err
		}

		defer f.Close()

		b, err := ioutil.ReadFile(layout)

		if err != nil {
			return "", err
		}
		return filepath.Join(key, "index.html"), template.Render(f, layout, string(b), data)
	}

	data, layout, err := i.getIndexDataAndLayout(key, pp, s)

	if err != nil {
		return "", err
	}

	if _, err := os.Stat(layout); err != nil {
		if os.IsNotExist(err) {
			if _, ok := optionalLayouts[layout]; ok {
				return "", nil
			}
			return "", ErrNoLayout
		}
		return "", err
	}

	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		if err := os.MkdirAll(dir, config.DirMode); err != nil {
			return "", err
		}
	}

	f, err := os.OpenFile(index, os.O_TRUNC|os.O_CREATE|os.O_RDWR, config.FileMode)

	if err != nil {
		return "", err
	}

	defer f.Close()

	b, err := ioutil.ReadFile(layout)

	if err != nil {
		return "", err
	}
	return index, template.Render(f, layout, string(b), data)
}
