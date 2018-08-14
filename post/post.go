package post

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewpillar/jrnl/util"

	"github.com/mozillazg/go-slugify"
)

var (
	// Date format for the source post file.
	dateFmt = "2006-01-02T15:04"

	// Date format for the site directory where the final published post will
	// reside.
	dateDirFmt = "2006/01/02/"
)

type Post struct {
	srcDir string

	siteDir string

	ID string

	Category string

	Title string

	Preview string

	SourcePath string

	SitePath string

	Date *time.Time
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
		Date:     &date,
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

	parts := strings.Split(srcPath, "/")

	rid := []rune(strings.Join(parts[1:], "/"))
	id := string(rid[:len(rid) - 3])

	fname := []rune(parts[len(parts) - 1])

	categorySlug := ""

	if len(parts) >= 3 {
		categorySlug = parts[len(parts) - 2]
	}

	titleSlug := string(fname[17:len(fname) - 3])
	dateSlug := string(fname[:16])

	date, err := time.Parse(dateFmt, dateSlug)

	if err != nil {
		return nil, err
	}

	p := &Post{
		srcDir:   parts[0],
		siteDir:  siteDir,
		ID:       id,
		Category: util.Deslug(categorySlug),
		Title:    util.Deslug(titleSlug),
		Date:     &date,
	}

	p.setSitePath(categorySlug, titleSlug)
	p.setSrcPath(categorySlug, titleSlug)

	return p, nil
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

	path.WriteString(p.Date.Format(dateDirFmt))
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
