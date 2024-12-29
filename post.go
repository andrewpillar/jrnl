package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"
)

var PostCmd = &Command{
	Usage: "post <title>",
	Short: "create a new journal post",
	Long: `post will open up the editor specified via the EDITOR environment variable for
writing a post.

The -l flag can be given to specify a layout to use for the new page.`,
	Run: postCmd,
}

type Post struct {
	*Page
}

func (p *Post) Slug() string { return p.MetaData.CreatedAt.Format("2006-01-02") + "-" + p.Page.Slug() }
func (p *Post) URL() string  { return "/" + p.Slug() }

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

	var layout string

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&layout, "l", "page", "the layout of the new post")
	fs.Parse(args)

	args = fs.Args()

	if len(args) == 0 {
		return ErrUsage
	}

	p := Post{
		Page: NewPage(args[0], layout),
	}

	p.MetaData.CreatedAt.Time = time.Now()

	if err := p.Touch(); err != nil {
		return err
	}

	if err := OpenInEditor(p.Path()); err != nil {
		return err
	}
	return nil
}
