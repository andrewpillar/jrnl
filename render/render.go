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

var (
	styleTag      = []byte("<style type=\"text/css\">")
	styleCloseTag = []byte("</style>")
)

type Renderer struct {
	*blackfriday.HTMLRenderer

	style     *chroma.Style
	formatter *html.Formatter

	Errs []error
}

func New() *Renderer {
	params := blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags,
	}

	r := blackfriday.NewHTMLRenderer(params)

	style := styles.Get("monokai")

	if style == nil {
		style = styles.Fallback
	}

	return &Renderer{
		HTMLRenderer: r,
		style:        style,
		formatter:    html.New(html.WithClasses()),
		Errs:         make([]error, 0),
	}
}

func (r *Renderer) getLexer(b []byte) chroma.Lexer {
	var lexer chroma.Lexer

	if len(b) > 0 {
		i := bytes.IndexAny(b, "\t ")

		if i < 0 {
			i = len(b)
		}

		lang := string(b[:i])

		lexer = lexers.Get(lang)
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}

	return lexer
}

func (r *Renderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
		case blackfriday.CodeBlock:
			lexer := r.getLexer(node.Info)

			w.Write(styleTag)
			r.formatter.WriteCSS(w, r.style)
			w.Write(styleCloseTag)

			iterator, err := lexer.Tokenise(nil, string(node.Literal))

			if err != nil {
				r.Errs = append(r.Errs, err)
			}

			if err := r.formatter.Format(w, r.style, iterator); err != nil {
				r.Errs = append(r.Errs, err)
			}

			break
		default:
			return r.HTMLRenderer.RenderNode(w, node, entering)
	}

	return blackfriday.GoToNext
}
