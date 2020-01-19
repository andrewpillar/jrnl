package render

import (
	"bytes"
	"io"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"

	"github.com/russross/blackfriday"
)

type Renderer struct {
	*blackfriday.HTMLRenderer

	Formatter *html.Formatter
}

func New() *Renderer {
	r := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	})
	return &Renderer{
		HTMLRenderer: r,
		Formatter:    html.New(html.WithClasses(true)),
	}
}

func (r *Renderer) getLexer(b []byte) chroma.Lexer {
	var lxr chroma.Lexer

	if len(b) > 0 {
		i := bytes.IndexAny(b, "\t ")

		if i < 0 {
			i = len(b)
		}

		lxr = lexers.Get(string(b[:i]))
	}

	if lxr == nil {
		lxr = lexers.Fallback
	}

	return lxr
}

func (r *Renderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
	case blackfriday.CodeBlock:
		lxr := r.getLexer(node.Info)

		it, err := lxr.Tokenise(nil, string(node.Literal))

		if err != nil {
			return blackfriday.Terminate
		}

		if err := r.Formatter.Format(w, styles.Fallback, it); err != nil {
			return blackfriday.Terminate
		}
		return blackfriday.GoToNext
	default:
		return r.HTMLRenderer.RenderNode(w, node, entering)
	}
}
