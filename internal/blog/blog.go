package blog

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/andrewpillar/jrnl/internal/config"
	"github.com/andrewpillar/jrnl/internal/render"

	"github.com/gorilla/feeds"

	"github.com/grokify/html-strip-tags-go"

	"github.com/mmcdole/gofeed"

	"github.com/russross/blackfriday"

	"gopkg.in/yaml.v2"
)

type byCreatedAt []Post

type pageFrontMatter struct {
	Title  string
	Layout string
}

type postTime struct {
	time.Time
}

type postFrontMatter struct {
	pageFrontMatter `yaml:",inline"`

	Index     bool
	CreatedAt postTime `yaml:"createdAt"`
	UpdatedAt postTime `yaml:"updatedAt"`
}

type rollErrors []error

type Category struct {
	ID         string
	Name       string
	Categories []Category
}

type Feed struct {
	Title       string
	Link        string
	Description string
	Author      feeds.Author
}

type Page struct {
	ID         string
	Title      string
	Layout     string
	SourcePath string
	SitePath   string
	Body       string
	Delete     bool
}

type Post struct {
	Page

	Index       bool
	Description string
	Category    Category
	CreatedAt   postTime
	UpdatedAt   postTime
}

type Site struct {
	Title      string
	Link       string
	Categories []Category
	Pages      []Page
	Blogroll   []Feed
}

var (
	iso8601 = "2006-01-02T15:04"

	reslug = regexp.MustCompile("[^a-zA-Z0-9]")
	redash = regexp.MustCompile("-")
	redup  = regexp.MustCompile("-{2,}")
)

func dateString(t time.Time) string {
	return string([]rune(t.Format(iso8601)[:10]))
}

func marshalFrontMatter(fm interface{}, w io.Writer) error {
	if _, err := w.Write([]byte("---\n")); err != nil {
		return err
	}

	enc := yaml.NewEncoder(w)

	if err := enc.Encode(fm); err != nil {
		return err
	}

	_, err := w.Write([]byte("---\n"))
	return err
}

func resolveCategory(path string) (Category, error) {
	var c Category

	if _, err := os.Stat(path); err != nil {
		return c, err
	}

	buf := bytes.Buffer{}

	id := strings.Replace(path, config.PostsDir + string(os.PathSeparator), "", 1)
	parts := strings.Split(redash.ReplaceAllString(id, " "), string(os.PathSeparator))
	end := len(parts)-1

	for i, p := range parts {
		buf.WriteString(strings.Title(p))

		if i != end {
			buf.WriteString(" / ")
		}
	}

	return Category{
		ID:   id,
		Name: buf.String(),
	}, nil
}

func resolvePage(path string) (Page, error) {
	p := Page{SourcePath: path}
	err := p.Load()
	return p, err
}

func resolvePost(path string) (Post, error) {
	p := Post{
		Page: Page{SourcePath: path},
	}
	err := p.Load()
	return p, err
}

func Slug(s string) string {
	s = strings.TrimSpace(s)
	s = reslug.ReplaceAllString(s, "-")
	s = redup.ReplaceAllString(s, "-")
	return strings.ToLower(strings.TrimPrefix(strings.TrimSuffix(s, "-"), "-"))
}

func unmarshalFrontMatter(fm interface{}, r io.Reader) error {
	buf := &bytes.Buffer{}
	bounds := 0

	for bounds != 2 {
		b := make([]byte, 1)
		n, err := r.Read(b)

		if err != nil {
			return err
		}

		buf.Write(b[:n])

		for b[0] == '-' {
			n, err = r.Read(b)

			if err != nil {
				return err
			}

			buf.Write(b[:n])

			if b[0] == '\n' {
				bounds++
				break
			}
		}
	}
	dec := yaml.NewDecoder(buf)
	return dec.Decode(fm)
}

func Categories() ([]Category, error) {
	m := make(map[string]*Category)
	ids := sort.StringSlice(make([]string, 0))

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == config.PostsDir || !info.IsDir() || strings.Contains(path, config.IndexDir) {
			return nil
		}

		c, err := resolveCategory(path)

		if err != nil {
			return err
		}

		parts := strings.Split(c.ID, string(os.PathSeparator))

		if len(parts) >= 2 {
			id := filepath.Join(parts[:len(parts)-1]...)

			parent, ok := m[id]

			if !ok {
				return errors.New("no parent found for " + path)
			}

			parent.Categories = append(parent.Categories, c)
			return nil
		}

		m[c.ID] = &c
		ids = append(ids, c.ID)
		return nil
	}

	err := filepath.Walk(config.PostsDir, fn)

	ids.Sort()

	cc := make([]Category, 0, len(m))

	for _, id := range ids {
		cc = append(cc, (*m[id]))
	}
	return cc, err
}

func NewPage(title string) Page {
	id := Slug(title)

	return Page{
		ID:         id,
		Title:      title,
		SourcePath: filepath.Join(config.PagesDir, id + ".md"),
		SitePath:   filepath.Join(config.SiteDir, filepath.Base(id), "index.html"),
	}
}

func NewPost(title, category string) Post {
	now := postTime{Time: time.Now()}

	parts := strings.Split(category, "/")
	end := len(parts)-1

	var c Category

	buf := &bytes.Buffer{}

	for i, p := range parts {
		buf.WriteString(Slug(p))

		if i != end {
			buf.WriteString(string(os.PathSeparator))
		}
	}

	c.ID = buf.String()
	c.Name = category

	id := Slug(title)

	return Post{
		Page:     Page{
			ID:         id,
			Title:      title,
			SourcePath: filepath.Join(config.PostsDir, id + ".md"),
			SitePath:   filepath.Join(
				config.SiteDir,
				c.ID,
				strings.Replace(dateString(now.Time), "-", string(os.PathSeparator), -1),
				filepath.Base(id),
				"index.html",
			),
		},
		Category:  c,
		Index:     true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func GetCategory(id string) (Category, error) {
	return resolveCategory(filepath.Join(config.PostsDir, id))
}

func GetPage(id string) (Page, error) {
	return resolvePage(filepath.Join(config.PagesDir, id + ".md"))
}

func GetPost(id string) (Post, error) {
	return resolvePost(filepath.Join(config.PostsDir, id + ".md"))
}

func GetRoll(urls ...string) ([]Feed, error) {
	roll := make([]Feed, 0, len(urls))
	errs := make([]error, 0, len(urls))

	items := make(chan Feed)
	errsCh := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(len(urls))

	go func() {
		wg.Wait()
		close(items)
		close(errsCh)
	}()

	for _, url := range urls {
		go func(url string) {
			defer wg.Done()

			p := gofeed.NewParser()

			feed, err := p.ParseURL(url)

			if err != nil {
				errsCh <- errors.New(url + ": " + err.Error())
				return
			}

			if len(feed.Items) == 0 {
				return
			}

			item := feed.Items[0]

			author := feeds.Author{}

			if item.Author != nil {
				author.Name = item.Author.Name
				author.Email = item.Author.Email
			}

			items <- Feed{
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				Author:      author,
			}
		}(url)
	}

	for errsCh != nil && items != nil {
		select {
		case err, ok := <-errsCh:
			if !ok {
				errs = nil
				break
			}
			errs = append(errs, err)
		case it, ok := <-items:
			if !ok {
				items = nil
				break
			}
			roll = append(roll, it)
		}
	}

	return roll, rollErrors(errs).err()
}

func (e rollErrors) err() error {
	isNil := true

	for _, err := range e {
		if err != nil {
			isNil = false
			break
		}
	}

	if isNil {
		return nil
	}

	return e
}

func (e rollErrors) Error() string {
	buf := &bytes.Buffer{}
	end := len(e)-1

	for i, err := range e {
		buf.WriteString(err.Error())

		if i != end {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

func Pages() ([]Page, error) {
	pp := make([]Page, 0)

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == config.PagesDir || info.IsDir() {
			return nil
		}

		p, err := resolvePage(path)

		if err != nil {
			return err
		}

		pp = append(pp, p)
		return nil
	}

	err := filepath.Walk(config.PagesDir, fn)

	return pp, err
}

func Posts() ([]Post, error) {
	pp := make([]Post, 0)

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.Contains(path, config.IndexDir) {
			return nil
		}

		p, err := resolvePost(path)

		if err != nil {
			return err
		}

		pp = append(pp, p)
		return nil
	}

	err := filepath.Walk(config.PostsDir, fn)

	sort.Sort(byCreatedAt(pp))

	return pp, err
}

func (c Category) Href() string {
	return "/" + c.ID
}

func (f Feed) Write(w io.Writer, kind string, pp []Post) error {
	items := make([]*feeds.Item, 0, len(pp))

	for _, p := range pp {
		items = append(items, &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: f.Link + p.Href()},
			Description: strip.StripTags(p.Description),
			Author:      &feeds.Author{
				Name:  f.Author.Name,
				Email: f.Author.Email,
			},
			Created:     p.CreatedAt.Time,
		})
	}

	fd := &feeds.Feed{
		Title:       f.Title,
		Link:        &feeds.Link{Href: f.Link},
		Description: f.Description,
		Author:      &feeds.Author{
			Name:  f.Author.Name,
			Email: f.Author.Email,
		},
	}
	fd.Items = items

	switch kind {
	case "atom":
		return fd.WriteAtom(w)
	case "rss":
		return fd.WriteRss(w)
	}
	return nil
}

func (p byCreatedAt) Len() int {
	return len(p)
}

func (p byCreatedAt) Less(i, j int) bool {
	return p[i].CreatedAt.After(p[j].CreatedAt.Time)
}

func (p byCreatedAt) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Page) Hash() []byte {
	h := sha256.New()

	buf := &bytes.Buffer{}
	buf.WriteString(p.Title)
	buf.WriteString(p.Body)

	io.Copy(h, buf)

	return h.Sum(nil)
}

func (p Page) Href() string {
	r := []rune(p.SitePath)
	return filepath.Dir(string(r[len(config.SiteDir):]))
}

func (p *Page) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &pageFrontMatter{}

	if err := unmarshalFrontMatter(&fm, f); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	p.ID = strings.Split(filepath.Base(p.SourcePath), ".")[0]
	p.Title = fm.Title
	p.Layout = fm.Layout
	p.SitePath = filepath.Join(config.SiteDir, filepath.Base(p.ID), "index.html")
	p.Body = string(b)
	return nil
}

func (p Page) open() (*os.File, error) {
	info, err := os.Stat(p.SourcePath)

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		if err := p.Load(); err != nil {
			return nil, err
		}
	}

	if info != nil && info.IsDir() {
		return nil, errors.New("expected file, got directory: " + p.SourcePath)
	}

	return os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, config.FileMode)
}

func (p *Page) Remove() error {
	if _, err := os.Stat(p.SitePath); err == nil {
		return os.Remove(p.SitePath)
	}
	return os.Remove(p.SourcePath)
}

func (p *Page) Render() {
	p.Body = string(blackfriday.Run([]byte(p.Body), blackfriday.WithRenderer(render.New())))
}

func (p Page) saveWithFrontMatter(fm interface{}) error {
	f, err := os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, config.FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	if err := marshalFrontMatter(fm, f); err != nil {
		return err
	}

	_, err = f.WriteString(p.Body)
	return err
}

func (p Page) Save() error {
	fm := &pageFrontMatter{
		Title:  p.Title,
		Layout: p.Layout,
	}
	return p.saveWithFrontMatter(fm)
}

func (p *Page) Touch() error {
	f, err := p.open()

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &pageFrontMatter{
		Title:  p.Title,
		Layout: p.Layout,
	}

	if err := marshalFrontMatter(fm, f); err != nil {
		return err
	}

	_, err = f.Write([]byte(p.Body))
	return err
}

func (p Post) HasCategory() bool {
	return p.Category.ID != "" && p.Category.Name != ""
}

func (p *Post) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &postFrontMatter{}

	if err := unmarshalFrontMatter(fm, f); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	if len(b) > 4 {
		i := strings.Index(string(b), "\n\n")

		if i == -1 {
			i = strings.Index(string(b), "\n")
		}

		p.Description = string(b[:i])
	}

	trimmed := strings.Replace(p.SourcePath, config.PostsDir+string(os.PathSeparator), "", 1)
	parts := strings.Split(trimmed, string(os.PathSeparator))

	if len(parts) >= 2 {
		parts = append([]string{
			config.PostsDir,
		}, parts[:len(parts)-1]...)

		p.Category, err = resolveCategory(filepath.Join(parts...))

		if err != nil {
			return err
		}
	}

	p.ID = strings.Split(filepath.Base(p.SourcePath), ".")[0]
	p.Title = fm.Title
	p.Layout = fm.Layout
	p.SitePath = filepath.Join(
		config.SiteDir,
		p.Category.ID,
		strings.Replace(dateString(fm.CreatedAt.Time), "-", string(os.PathSeparator), -1),
		p.ID,
		"index.html",
	)
	p.Index = fm.Index
	p.Body = string(b)
	p.CreatedAt = fm.CreatedAt
	p.UpdatedAt = fm.UpdatedAt
	return nil
}

func (p *Post) Remove() error {
	if err := p.Page.Remove(); err != nil {
		return err
	}

	parts := strings.Split(filepath.Dir(p.SourcePath), string(os.PathSeparator))

	for i := range parts {
		dir := filepath.Join(parts[:len(parts)-i]...)

		if dir == config.PostsDir {
			break
		}

		f, err := os.Open(dir)

		if err != nil {
			return err
		}

		defer f.Close()

		if _, err := f.Readdirnames(1); err == io.EOF {
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Post) Render() {
	opt := blackfriday.WithRenderer(render.New())

	p.Description = string(blackfriday.Run([]byte(p.Description), opt))
	p.Body = string(blackfriday.Run([]byte(p.Body), opt))
}

func (p Post) Save() error {
	fm := &postFrontMatter{
		pageFrontMatter: pageFrontMatter{
			Title:  p.Title,
			Layout: p.Layout,
		},
		Index:     p.Index,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return p.saveWithFrontMatter(fm)
}

func (p *Post) Touch() error {
	if err := os.MkdirAll(filepath.Dir(p.SourcePath), config.DirMode); err != nil {
		return err
	}

	f, err := p.Page.open()

	if err != nil {
		return err
	}

	defer f.Close()

	fm := &postFrontMatter{
		pageFrontMatter: pageFrontMatter{
			Title:  p.Title,
			Layout: p.Layout,
		},
		Index:     p.Index,
		CreatedAt: p.CreatedAt,
		UpdatedAt: postTime{Time: time.Now()},
	}

	if err := marshalFrontMatter(fm, f); err != nil {
		return err
	}

	_, err = f.Write([]byte(p.Body))
	return err
}

func (t postTime) MarshalYAML() (interface{}, error) {
	s := t.Time.Format(iso8601)
	return s, nil
}

func (t *postTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var (
		s   string
		err error
	)

	if err = unmarshal(&s); err != nil {
		return err
	}

	t.Time, err = time.Parse(iso8601, s)
	return err
}
