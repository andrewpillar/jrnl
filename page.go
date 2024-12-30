package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

type SitePage interface {
	URL() string

	Title() string

	Description() string

	Content() string

	Layout() string

	CreatedAt() time.Time
}

const (
	pageMask = os.O_TRUNC | os.O_RDWR | os.O_CREATE
	pagePerm = os.FileMode(0640)
)

var (
	reSlug = regexp.MustCompile("[^a-zA-Z0-9]")
	reDupe = regexp.MustCompile("-{2,}")

	PageCmd = &Command{
		Usage: "page <title>",
		Short: "create a new jrnl page",
		Long: `page will open up the editor specified via the EDITOR environment variable for
writing the page.

The -l flag can be given to specify a layout to use for the new page.`,
		Run: pageCmd,
	}
)

type Time struct {
	time.Time
}

func (t Time) MarshalYAML() (any, error) {
	return t.Format(time.DateTime), nil
}

func (t *Time) UnmarshalYAML(n *yaml.Node) error {
	v, err := time.Parse(time.DateTime, n.Value)

	if err != nil {
		return err
	}

	t.Time = v
	return nil
}

type MetaData struct {
	Title     string
	Layout    string
	CreatedAt Time `yaml:"createdAt,omitempty"`
	UpdatedAt Time `yaml:"updatedAt,omitempty"`
}

func (m *MetaData) Encode(w io.Writer) error {
	if _, err := io.WriteString(w, "---\n"); err != nil {
		return err
	}

	if err := yaml.NewEncoder(w).Encode(m); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "---\n"); err != nil {
		return err
	}
	return nil
}

// Decode decodes the front matter from the given reader. This will return the
// rest of the content that follows on from the front matter in a buffer.
func (m *MetaData) Decode(r io.Reader) (*bytes.Buffer, error) {
	buf := make([]byte, 0)
	br := bufio.NewReader(r)

	bounds := 0

loop:
	for bounds != 2 {
		b, err := br.ReadByte()

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		buf = append(buf, b)

		for b == '-' {
			b, err = br.ReadByte()

			if err != nil {
				if errors.Is(err, io.EOF) {
					break loop
				}
				return nil, err
			}

			buf = append(buf, b)

			if b == '\n' {
				bounds++
				break
			}
		}
	}

	var cont bytes.Buffer

	if _, err := io.Copy(&cont, br); err != nil {
		return nil, err
	}

	if err := yaml.NewDecoder(bytes.NewReader(buf)).Decode(m); err != nil {
		return nil, err
	}
	return &cont, nil
}

type Page struct {
	*MetaData

	Body string
}

func NewPage(title, layout string) *Page {
	return &Page{
		MetaData: &MetaData{
			Title:  title,
			Layout: layout,
		},
	}
}

func LoadPage(path string) (*Page, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var p Page

	if err := p.Decode(f); err != nil {
		return nil, err
	}
	return &p, nil
}

func Markdown(s string) string {
	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	md.Convert([]byte(s), &buf)

	return buf.String()
}

func (p *Page) Slug() string {
	s := strings.TrimSpace(p.MetaData.Title)
	s = reSlug.ReplaceAllString(s, "-")
	s = reDupe.ReplaceAllString(s, "-")

	return strings.ToLower(strings.TrimPrefix(strings.TrimSuffix(s, "-"), "-"))
}

func (p *Page) URL() string     { return "/" + p.Slug() }
func (p *Page) Title() string   { return p.MetaData.Title }
func (p *Page) Content() string { return Markdown(p.Body) }

func (p *Page) Description() string {
	if len(p.Body) > 4 {
		i := strings.Index(p.Body, "\n\n")

		if i < 0 {
			i = strings.Index(p.Body, "\n")
		}
		return Markdown(p.Body[:i])
	}
	return Markdown(p.Body)
}

func (p *Page) Layout() string       { return p.MetaData.Layout }
func (p *Page) CreatedAt() time.Time { return p.MetaData.CreatedAt.Time }

func (p *Page) Encode(w io.Writer) error {
	if err := p.MetaData.Encode(w); err != nil {
		return err
	}

	if _, err := io.WriteString(w, p.Body); err != nil {
		return err
	}
	return nil
}

func (p *Page) Decode(r io.Reader) error {
	p.MetaData = &MetaData{}

	buf, err := p.MetaData.Decode(r)

	if err != nil {
		return err
	}

	p.Body = buf.String()

	return nil
}

func (p *Page) Path() string {
	return filepath.Join(pageDir, p.Slug()) + ".md"
}

func (p *Page) File() (*os.File, error) {
	return os.OpenFile(p.Path(), pageMask, pagePerm)
}

func (p *Page) Touch() error {
	f, err := p.File()

	if err != nil {
		return err
	}

	defer f.Close()

	if err := p.Encode(f); err != nil {
		return err
	}
	return nil
}

func OpenInEditor(path string) error {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return errors.New("EDITOR not set")
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func pageCmd(cmd *Command, args []string) error {
	if err := Initialized("."); err != nil {
		return err
	}

	var layout string

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&layout, "l", "page", "the layout of the new post")
	fs.Parse(args)

	args = fs.Args()

	if len(args) == 0 {
		return ErrUsage
	}

	p := NewPage(args[0], layout)

	if err := p.Touch(); err != nil {
		return err
	}

	if err := OpenInEditor(p.Path()); err != nil {
		return err
	}
	return nil
}
