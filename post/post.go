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
	// Date format for the source post file.
	dateFmt = "2006-01-02T15:04"

	// Date format for the site directory where the final published post will
	// reside.
	dateDirFmt = "2006/01/02"
)

type Post struct {
	srcDir string

	siteDir string

	ID string

	JournalTitle string

	Category string

	Title string

	Body string

	SourcePath string

	SitePath string

	Date time.Time
}

func New(siteDir, srcDir, category, title string) *Post {
	date := time.Now()

	categorySlug := slugify.Slugify(category)
	dateSlug := date.Format(dateFmt)
	titleSlug := slugify.Slugify(title)

	id := dateSlug + "-" + titleSlug

	if categorySlug != "" {
		id = categorySlug + "/" + dateSlug + "-" + titleSlug
	}

	p := &Post{
		srcDir:   srcDir,
		siteDir:  siteDir,
		ID:       id,
		Category: category,
		Title:    title,
		Date:     date,
	}

	p.setSitePath(categorySlug, titleSlug)
	p.setSrcPath(categorySlug, titleSlug)

	return p
}

func NewFromSource(siteDir, srcPath string) (*Post, error) {
	_, err := os.Stat(srcPath)

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

	parts := strings.Split(srcPath, "/")

	id := resolveId(parts[1:])
	category, categorySlug := resolveCategory(parts)
	title, titleSlug := resolveTitle(parts)

	date, err := resolveDate(parts)

	if err != nil {
		return nil, err
	}

	p := &Post{
		srcDir:       parts[0],
		siteDir:      siteDir,
		ID:           id,
		JournalTitle: m.Title,
		Category:     category,
		Title:        title,
		Date:         date,
	}

	p.setSitePath(categorySlug, titleSlug)
	p.setSrcPath(categorySlug, titleSlug)

	return p, nil
}

func resolveCategory(parts []string) (string, string) {
	if len(parts) >= 3 {
		slug := parts[len(parts) - 2]

		return util.Deslug(slug), slug
	}

	return "", ""
}

func resolveDate(parts []string) (time.Time, error) {
	fname := parts[len(parts) - 1]
	r := []rune(fname)

	return time.Parse(dateFmt, string(r[:16]))
}

func resolveId(parts []string) string {
	r := []rune(strings.Join(parts, "/"))

	return string(r[:len(r) - 3])
}

func resolveTitle(parts []string) (string, string) {
	fname := parts[len(parts) - 1]
	r := []rune(fname)

	slug := string(r[17:len(r) - 3])

	return util.Deslug(slug), slug
}

func (p *Post) Convert() {
	md := blackfriday.Run([]byte(p.Body))

	p.Body = string(md)
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
		if err := p.RemoveSitePath(); err != nil {
			return err
		}
	}

	return p.RemoveSourcePath()
}

func (p *Post) RemoveSourcePath() error {
	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(p.srcDir, filepath.Dir(p.SourcePath))
}

func (p *Post) RemoveSitePath() error {
	if err := os.Remove(p.SitePath); err != nil {
		return err
	}

	return util.RemoveEmptyDirs(p.siteDir, filepath.Dir(p.SitePath))
}

func (p *Post) setSitePath(categorySlug, titleSlug string) {
	path := bytes.NewBufferString(p.siteDir + "/")

	if categorySlug != "" {
		path.WriteString(categorySlug + "/")
	}

	path.WriteString(p.Date.Format(dateDirFmt) + "/")
	path.WriteString(titleSlug + "/index.html")

	p.SitePath = path.String()
}

func (p *Post) setSrcPath(categorySlug, titleSlug string) {
	path := bytes.NewBufferString(p.srcDir + "/")

	if categorySlug != "" {
		path.WriteString(categorySlug + "/")
	}

	path.WriteString(p.Date.Format(dateFmt) + "-")
	path.WriteString(titleSlug + ".md")

	p.SourcePath = path.String()
}
