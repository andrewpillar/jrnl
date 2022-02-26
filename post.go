package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type postFrontMatter struct {
	pageFrontMatter `yaml:",inline"`

	CreatedAt postTime `yaml:"createdAt"`
	UpdatedAt postTime `yaml:"updatedAt"`
}

type postTime struct {
	time.Time
}

type Category struct {
	ID         string
	Name       string
	Categories []*Category
}

type Post struct {
	*Page

	Category    *Category
	Index       bool
	Description string
	CreatedAt   postTime
	UpdatedAt   postTime
}

var (
	iso8601 = "2006-01-02T15:04"

	redash = regexp.MustCompile("-")

	PostCmd = &Command{
		Usage: "post <title>",
		Short: "create a new journal post",
		Long: `Post will open up the editor specified via the EDITOR environment variable for
editting the new page.

The -c flag can be given to specify a category for the post being created.

The -l flag can be given to specify a layout to use for the new page. This will
be pre-populated in the front matter.`,
		Run: postCmd,
	}
)

func resolveCategory(path string) (*Category, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	id := strings.Replace(path, postsDir+string(os.PathSeparator), "", 1)
	parts := strings.Split(redash.ReplaceAllString(id, " "), string(os.PathSeparator))

	end := len(parts) - 1

	for i, p := range parts {
		buf.WriteString(strings.Title(p))

		if i != end {
			buf.WriteString(" / ")
		}
	}

	return &Category{
		ID:   id,
		Name: buf.String(),
	}, nil
}

func resolvePost(path string) (*Post, error) {
	p := &Post{
		Page: &Page{
			SourcePath: path,
		},
	}
	err := p.Load()
	return p, err
}

func GetCategory(id string) (*Category, bool, error) {
	c, err := resolveCategory(filepath.Join(postsDir, id))

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, false, err
		}
		return nil, false, nil
	}
	return c, true, nil
}

func Categories() ([]*Category, error) {
	m := make(map[string]*Category)
	ids := make([]string, 0)

	err := filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == postsDir || !info.IsDir() {
			return nil
		}

		category, err := resolveCategory(path)

		if err != nil {
			return err
		}

		parts := strings.Split(category.ID, string(os.PathSeparator))

		if len(parts) >= 2 {
			id := filepath.Join(parts[:len(parts)-1]...)

			parent, ok := m[id]

			if !ok {
				return errors.New("no parent for " + path)
			}

			parent.Categories = append(parent.Categories, category)
			return nil
		}

		m[category.ID] = category
		ids = append(ids, category.ID)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(ids)

	categories := make([]*Category, 0, len(m))

	for _, id := range ids {
		categories = append(categories, m[id])
	}
	return categories, nil
}

func GetPost(id string) (*Post, bool, error) {
	post, err := resolvePost(filepath.Join(postsDir, id+".md"))

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, false, err
		}
		return nil, false, nil
	}
	return post, true, nil
}

func Posts() ([]*Post, error) {
	posts := make([]*Post, 0)

	err := filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		post, err := resolvePost(path)

		if err != nil {
			return err
		}

		posts = append(posts, post)
		return nil
	})
	return posts, err
}

func WalkPosts(fn func(*Post) error) error {
	return filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		post, err := resolvePost(path)

		if err != nil {
			return err
		}
		return fn(post)
	})
}

func (t postTime) MarshalYAML() (interface{}, error) {
	return t.String(), nil
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

func (t *postTime) String() string {
	return t.Format(iso8601)
}

func (p *Post) HasCategory() bool { return p.Category.ID != "" }

func (p *Post) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	var fm postFrontMatter

	if err := unmarshalFrontMatter(&fm, f); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	if len(b) > 4 {
		i := strings.Index(string(b), "\n\n")

		if i < 0 {
			i = strings.Index(string(b), "\n")
		}
		p.Description = string(b[:i])
	}

	trimmed := strings.Replace(p.SourcePath, postsDir+string(os.PathSeparator), "", 1)
	parts := strings.Split(trimmed, string(os.PathSeparator))

	p.Category = &Category{}

	if len(parts) >= 2 {
		parts = append([]string{postsDir}, parts[:len(parts)-1]...)

		p.Category, err = resolveCategory(filepath.Join(parts...))

		if err != nil {
			return err
		}
	}

	p.ID = filepath.Join(p.Category.ID, strings.Split(filepath.Base(p.SourcePath), ".")[0])
	p.Title = fm.Title
	p.Layout = fm.Layout
	p.SitePath = filepath.Join(
		siteDir,
		p.Category.ID,
		strings.Replace(string(fm.CreatedAt.String()[:10]), "-", string(os.PathSeparator), -1),
		strings.Replace(p.ID, p.Category.ID, "", 1),
		"index.html",
	)
	p.Body = string(b)
	p.CreatedAt = fm.CreatedAt
	p.UpdatedAt = fm.UpdatedAt
	return nil
}

func (p *Post) Publish(s Site) error {
	renderedDesc, err := render(p.Description)

	if err != nil {
		return err
	}

	renderedBody, err := render(p.Body)

	if err != nil {
		return err
	}

	layout, err := p.readLayout()

	if err != nil {
		return err
	}

	f, err := p.siteFile()

	if err != nil {
		return err
	}

	defer f.Close()

	page := *p.Page
	page.Body = renderedBody

	p1 := *p
	p1.Page = &page
	p1.Description = renderedDesc

	data := struct {
		Site Site
		Post *Post
	}{
		Site: s,
		Post: &p1,
	}
	return executeTemplate(f, p.ID, layout, data)
}

func (p *Post) Remove() error {
	if err := p.Page.Remove(); err != nil {
		return err
	}

	parts := strings.Split(filepath.Dir(p.SourcePath), string(os.PathSeparator))

	for i := range parts {
		dir := filepath.Join(parts[:len(parts)-i]...)

		if dir == postsDir {
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

func (p *Post) Touch() error {
	if err := os.MkdirAll(filepath.Dir(p.SourcePath), os.FileMode(0755)); err != nil {
		return err
	}

	if _, err := os.Stat(p.SourcePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	f, err := os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.FileMode(0644))

	if err != nil {
		return err
	}

	defer f.Close()

	fm := postFrontMatter{
		pageFrontMatter: pageFrontMatter{
			Title:  p.Title,
			Layout: p.Layout,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: postTime{
			Time: time.Now(),
		},
	}

	if err := marshalFrontMatter(&fm, f); err != nil {
		return err
	}

	if _, err := f.Write([]byte(p.Body)); err != nil {
		return err
	}
	return nil
}

func (c Category) Href() string { return "/" + c.ID }

func postCmd(cmd *Command, args []string) {
	var (
		category string
		layout   string
	)

	fs := flag.NewFlagSet(cmd.Argv0+" "+args[0], flag.ExitOnError)
	fs.StringVar(&category, "c", "", "the category of the new post")
	fs.StringVar(&layout, "l", "", "the layout to use for new post")
	fs.Parse(args[1:])

	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	fsargs := fs.Args()

	if len(fsargs) < 1 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	title := fsargs[0]

	if title == "" {
		fmt.Fprintf(os.Stderr, "%s %s: missing post title", cmd.Argv0, args[0])
		os.Exit(1)
	}

	id := slug(title)

	now := postTime{
		Time: time.Now(),
	}

	parts := strings.Split(category, "/")
	end := len(parts) - 1

	var buf bytes.Buffer

	for i, p := range parts {
		buf.WriteString(slug(p))

		if i != end {
			buf.WriteString(string(os.PathSeparator))
		}
	}

	categoryId := buf.String()

	post := &Post{
		Page: &Page{
			ID:         filepath.Join(categoryId, id),
			Title:      title,
			Layout:     layout,
			SourcePath: filepath.Join(postsDir, categoryId, id+".md"),
			SitePath: filepath.Join(
				siteDir,
				categoryId,
				strings.Replace(now.String(), "-", string(os.PathSeparator), -1),
				id,
				"index.html",
			),
		},
		Category: &Category{
			ID:   categoryId,
			Name: category,
		},
		Index:     true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := post.Touch(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to create post: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if err := openInEditor(post.SourcePath); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open editor: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}
