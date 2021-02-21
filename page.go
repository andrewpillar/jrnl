package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/grokify/html-strip-tags-go"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"gopkg.in/yaml.v3"
)

type pageFrontMatter struct {
	Title  string
	Layout string
}

type Page struct {
	ID         string
	Title      string
	Layout     string
	Body       string
	SourcePath string
	SitePath   string
}

var (
	reslug = regexp.MustCompile("[^a-zA-Z0-9]")
	redup  = regexp.MustCompile("-{2,}")

	funcs template.FuncMap

	PageCmd = &Command{
		Usage: "page <title>",
		Short: "create a new journal page",
		Long: `Page will open up the editor specified via the EDITOR environment variable for
editting the new page.

The -l flag can be given to specify a layout to use for the new page. This will
be pre-populated in the front matter.`,
		Run: pageCmd,
	}
)

func init() {
	funcs = template.FuncMap{
		"partial": partial,
		"strip":   strip.StripTags,
	}
}

func partial(path string, data interface{}) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(layoutsDir, path))

	if err != nil {
		return "", err
	}

	t, err := template.New(path).Funcs(funcs).Parse(string(b))

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	return buf.String(), err
}

func executeTemplate(w io.Writer, name, text string, data interface{}) error {
	t, err := template.New(name).Funcs(funcs).Parse(text)

	if err != nil {
		return err
	}
	return t.Execute(w, data)
}

func render(s string) (string, error) {
	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	md.Renderer().AddOptions(html.WithUnsafe())

	if err := md.Convert([]byte(s), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func resolvePage(path string) (*Page, error) {
	p := &Page{
		SourcePath: path,
	}
	err := p.Load()
	return p, err
}

func slug(s string) string {
	if s == "" {
		return ""
	}

	s = strings.TrimSpace(s)
	s = reslug.ReplaceAllString(s, "-")
	s = redup.ReplaceAllString(s, "-")
	return strings.ToLower(strings.TrimPrefix(strings.TrimSuffix(s, "-"), "-"))
}

func readbyte(r io.Reader) (byte, error) {
	b := make([]byte, 1)

	if _, err := r.Read(b); err != nil {
		return 0, err
	}
	return b[0], nil
}

func marshalFrontMatter(v interface{}, w io.Writer) error {
	if _, err := w.Write([]byte("---\n")); err != nil {
		return err
	}

	if err := yaml.NewEncoder(w).Encode(v); err != nil {
		return err
	}

	_, err := w.Write([]byte("---\n"))
	return err
}

func unmarshalFrontMatter(v interface{}, r io.Reader) error {
	buf := make([]byte, 0)
	bounds := 0

	for bounds != 2 {
		b, err := readbyte(r)

		if err != nil {
			return err
		}
		buf = append(buf, b)

		for b == '-' {
			b, err = readbyte(r)

			if err != nil {
				return err
			}
			buf = append(buf, b)

			if b == '\n' {
				bounds++
				break
			}
		}
	}
	return yaml.Unmarshal(buf, v)
}

func GetPage(id string) (*Page, bool, error) {
	page, err := resolvePage(filepath.Join(pagesDir, id+".md"))

	if err != nil {
		if !os.IsNotExist(err) {
			return nil, false, err
		}
		return nil, false, nil
	}
	return page, true, nil
}

func Pages() ([]*Page, error) {
	pages := make([]*Page, 0)

	err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		page, err := resolvePage(path)

		if err != nil {
			return err
		}

		pages = append(pages, page)
		return nil
	})
	return pages, err
}

func WalkPages(fn func(*Page) error) error {
	return filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		page, err := resolvePage(path)

		if err != nil {
			return err
		}
		return fn(page)
	})
}

func (p *Page) readLayout() (string, error) {
	if p.Layout == "" {
		return "", errors.New("layout not set")
	}

	b, err := ioutil.ReadFile(filepath.Join(layoutsDir, p.Layout))

	if err != nil {
		return "", err
	}
	return string(b), err
}

func (p *Page) siteFile() (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p.SitePath), os.FileMode(0755)); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(p.SitePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0644))

	if err != nil {
		return nil, err
	}
	return f, nil
}

func (p *Page) Hash() []byte {
	sha256 := sha256.New()
	sha256.Write([]byte(p.Title))
	sha256.Write([]byte(p.Body))
	return sha256.Sum(nil)
}

func (p *Page) Href() string {
	l := len(siteDir)

	return filepath.Dir(p.SitePath[l:])
}

func (p *Page) Load() error {
	f, err := os.Open(p.SourcePath)

	if err != nil {
		return err
	}

	defer f.Close()

	var fm pageFrontMatter

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
	p.SitePath = filepath.Join(siteDir, filepath.Base(p.ID), "index.html")
	p.Body = string(b)
	return nil
}

func (p *Page) Publish(s Site) error {
	renderedBody, err := render(p.Body)

	if err != nil {
		return err
	}

	layout, err := p.readLayout()

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	data0 := struct {
		Site Site
	}{Site: s}

	if err := executeTemplate(&buf, p.ID, renderedBody, data0); err != nil {
		return err
	}

	f, err := p.siteFile()

	if err != nil {
		return err
	}

	defer f.Close()

	p1 := *p
	p1.Body = buf.String()

	data := struct {
		Site Site
		Page *Page
	}{
		Site: s,
		Page: &p1,
	}
	return executeTemplate(f, p.ID, layout, data)
}

func (p *Page) Touch() error {
	f, err := os.OpenFile(p.SourcePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.FileMode(0644))

	if err != nil {
		return err
	}

	defer f.Close()

	fm := pageFrontMatter{
		Title:  p.Title,
		Layout: p.Layout,
	}

	if err := marshalFrontMatter(&fm, f); err != nil {
		return err
	}

	_, err = f.Write([]byte(p.Body))
	return err
}

func (p *Page) Remove() error {
	if err := os.Remove(p.SitePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	parts := strings.Split(filepath.Dir(p.SitePath), string(os.PathSeparator))

	for i := range parts {
		dir := filepath.Join(parts[:len(parts)-i]...)

		if dir == siteDir {
			break
		}

		f, err := os.Open(dir)

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		defer f.Close()

		if _, err := f.Readdirnames(1); err == io.EOF {
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}

	if err := os.Remove(p.SourcePath); err != nil {
		return err
	}

	hash, err := OpenHash()

	if err != nil {
		return err
	}

	hash.Delete(p.ID)
	return hash.Save()
}

func pageCmd(cmd *Command, args []string) {
	var layout string

	fs := flag.NewFlagSet(cmd.Argv0+" "+args[0], flag.ExitOnError)
	fs.StringVar(&layout, "l", "", "the layout to use for new post")
	fs.Parse(args[1:])

	if err := initialized(""); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s %s: usage: %s\n", cmd.Argv0, args[0], cmd.Usage)
		os.Exit(1)
	}

	title := fs.Args()[0]

	id := slug(title)

	page := &Page{
		ID:         id,
		Title:      title,
		Layout:     layout,
		SourcePath: filepath.Join(pagesDir, id+".md"),
		SitePath:   filepath.Join(siteDir, id, "index.html"),
	}

	if err := page.Touch(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to create page: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if err := openInEditor(page.SourcePath); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open editor: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}
