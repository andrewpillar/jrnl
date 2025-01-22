package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"
)

var PostCmd = &Command{
	Usage: "post <title>",
	Short: "create a new jrnl post",
	Long: `post will open up the editor specified via the EDITOR environment variable for
writing a post.

The -l flag can be given to specify a layout to use for the new post.

The -p flag can be given to specify the parent for the new post.`,
	Run: postCmd,
}

type Post struct {
	*Page
}

func NewPost(title string) Post {
	return Post{
		Page: NewPage(title),
	}
}

func (p *Post) Slug() string { return p.MetaData.CreatedAt.Format("2006-01-02") + "-" + p.Page.Slug() }

func (p *Post) CreatedAt() time.Time { return p.MetaData.CreatedAt.Time }
func (p *Post) UpdatedAt() time.Time { return p.MetaData.UpdatedAt.Time }

func (p *Post) URL() string {
	url := "/" + p.Slug()

	if p.MetaData.Parent != "" {
		return p.MetaData.Parent + url
	}
	return url
}

func (p *Post) Path() string {
	return filepath.Join(postDir, p.Slug()) + ".md"
}

func (p *Post) File() (*os.File, error) {
	return os.OpenFile(p.Path(), pageMask, pagePerm)
}

func (p *Post) Touch() error {
	f, err := p.File()

	if err != nil {
		return err
	}

	defer f.Close()

	p.MetaData.UpdatedAt.Time = time.Now()

	if err := p.Encode(f); err != nil {
		return err
	}
	return nil
}

func postCmd(cmd *Command, args []string) error {
	if err := Initialized("."); err != nil {
		return err
	}

	var layout, parent string

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&layout, "l", "post", "the layout of the new post")
	fs.StringVar(&parent, "p", "", "the parent of the new post")
	fs.Parse(args)

	args = fs.Args()

	if len(args) == 0 {
		return ErrUsage
	}

	p := NewPost(args[0])
	p.MetaData.Layouts = append(p.MetaData.Layouts, layout)
	p.MetaData.Parent = parent

	p.MetaData.CreatedAt.Time = time.Now()

	if err := p.Touch(); err != nil {
		return err
	}

	if err := OpenInEditor(p.Path()); err != nil {
		return err
	}
	return nil
}
