package template

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
)

func printCategories(categories []category.Category) string {
	buf := bytes.Buffer{}

	for _, c := range categories {
		link := "<a href=\"" + c.Href() + "\">" + c.Name + "</a>"

		buf.WriteString("<li>" + link)

		if len(c.Categories) > 0 {
			buf.WriteString("<ul>" + printCategories(c.Categories) + "</ul>")
		}

		buf.WriteString("</li>")
	}

	return buf.String()
}

func partial(name string, data interface{}) (string, error) {
	path := filepath.Join(meta.LayoutsDir, name)

	b, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}

	funcs := template.FuncMap{
		"partial":         partial,
		"printCategories": printCategories,
	}

	t, err := template.New("partial-" + name).Funcs(funcs).Parse(string(b))

	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}

	if err := t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func Render(w io.Writer, name, layout string, data interface{}) error {
	funcs := template.FuncMap{
		"partial":         partial,
		"printCategories": printCategories,
	}

	t, err := template.New(name).Funcs(funcs).Parse(layout)

	if err != nil {
		return err
	}

	return t.Execute(w, data)
}
