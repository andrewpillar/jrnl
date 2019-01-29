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

// Create a new Renderer with the blackfriday.HTMLRenderer as the underlying struct the rendering
// every other node type.
func New() *Renderer {
	params := blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	}

	r := blackfriday.NewHTMLRenderer(params)

	return &Renderer{
		HTMLRenderer: r,
		Formatter:    html.New(html.WithClasses()),
	}
}

func (r *Renderer) getLexer(b []byte) chroma.Lexer {
	var lxr chroma.Lexer

	if len(b) > 0 {
		i := bytes.IndexAny(b, "\t ")

		if i < 0 {
			i = len(b)
		}

		lang := string(b[:i])

		lxr = lexers.Get(lang)
	}

	if lxr == nil {
		lxr = lexers.Fallback
	}

	return lxr
}

// The RenderNode method will check for the blackfriday.CodeBlock node type so we can have syntax
// highlighting for blocks of code. By default it will then fallback to the underlying HTMLRenderer
// provided with blackfriday for rendering every other node we receive.
func (r *Renderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
		case blackfriday.CodeBlock:
			lxr := r.getLexer(node.Info)

			iterator, err := lxr.Tokenise(nil, string(node.Literal))

			if err != nil {
				return blackfriday.Terminate
			}

			if err := r.Formatter.Format(w, styles.Fallback, iterator); err != nil {
				return blackfriday.Terminate
			}

			return blackfriday.GoToNext
		default:
			return r.HTMLRenderer.RenderNode(w, node, entering)
	}
}
