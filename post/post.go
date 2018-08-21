package post

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/andrewpillar/jrnl/meta"
	"github.com/andrewpillar/jrnl/util"

	"github.com/mozillazg/go-slugify"

	"gopkg.in/russross/blackfriday.v2"
)

var (
	dateLen = 16

	DateSlug = "2006-01-02T15:04"

	DateDir = "2006/01/02"

	SourceDir string

	SiteDir string
)

type Post struct {
	ID string

	JournalTitle string

	Category string

	Title string

	Preview string

	Body string

	SourcePath string

	SitePath string

	Date time.Time
}

func getId(parts []string) string {
	if parts[0] == SourceDir {
		parts = parts[1:]
	}

	id := []rune(filepath.Join(parts...))

	return string(id[:len(id) - 3])
}

func getCategory(parts []string) string {
	if len(parts) >= 3 {
		return filepath.Join(parts[1:len(parts) - 2]...)
	}

	return ""
}

func getDate(fname string) string {
	r := []rune(fname)

	return string(r[:dateLen])
}

func getTitle(fname string) string {
	r := []rune(fname)

	return string(r[dateLen + 1:len(r) - 3])
}

func New(title, category string) *Post {
	date := time.Now()

	categorySlug := slugify.Slugify(category)
	dateSlug := date.Format(DateSlug)
	titleSlug := slugify.Slugify(title)

	id := filepath.Join(categorySlug, dateSlug + "-" + titleSlug)

	sourcePath := filepath.Join(
		SourceDir,
		categorySlug,
		dateSlug + "-" + titleSlug + ".md",
	)

	p := &Post{
		ID:         id,
		SourcePath: sourcePath,
		Category:   category,
		Title:      title,
		Date:       date,
	}

	return p
}

func NewFromPath(path string) (*Post, error) {
	_, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	f, err := os.Open(meta.File)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	m, err := meta.Decode(f)

	if err != nil {
		return nil, err
	}

	parts := strings.Split(path, string(os.PathSeparator))

	id := getId(parts)
	categorySlug := getCategory(parts)
	titleSlug := getTitle(parts[len(parts) - 1])
	dateSlug := getDate(parts[len(parts) - 1])

	date, err := time.Parse(DateSlug, dateSlug)

	if err != nil {
		return nil, err
	}

	sitePath := filepath.Join(
		SiteDir,
		categorySlug,
		date.Format(DateDir),
		titleSlug,
		"index.html",
	)

	p := &Post{
		ID:           id,
		JournalTitle: m.Title,
		SourcePath:   path,
		SitePath:     sitePath,
		Title:        util.Deslug(titleSlug),
		Category:     util.Deslug(categorySlug),
		Date:         date,
	}

	return p, nil
}

func (p *Post) Convert() {
	body := blackfriday.Run([]byte(p.Body))
	preview := blackfriday.Run([]byte(p.Preview))

	p.Body = string(body)
	p.Preview = string(preview)
}

func (p *Post) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	if len(b) > 2 {
		i := bytes.Index(b, []byte("\n"))

		p.Preview = string(b[:i])
	}

	p.Body = string(b)

	return nil
}

func (p *Post) Publish(tmpl string) error {
	dir := filepath.Dir(p.SitePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	dst, err := os.OpenFile(p.SitePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		return err
	}

	defer dst.Close()

	t, err := template.New("post").Parse(tmpl)

	if err != nil {
		return err
	}

	return t.Execute(dst, p)
}

func (p *Post) Remove() error {
	_, err := os.Stat(p.SitePath)

	if !os.IsNotExist(err) {
		if err := p.removeSitePath(); err != nil {
			return err
		}
	}

	return p.removeSourcePath()
}

func (p *Post) removeSourcePath() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(SourceDir, filepath.Dir(p.SourcePath))
}

func (p *Post) removeSitePath() error {
	if err := os.Remove(p.SitePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(SiteDir, filepath.Dir(p.SitePath))
}
