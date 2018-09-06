package template

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/andrewpillar/jrnl/category"
	"github.com/andrewpillar/jrnl/meta"
)

type partialFunc func(name string) (string, error)

func New(name, layout string, data interface{}) (*template.Template, error) {
	funcs := template.FuncMap{
		"printCategories":     printCategories,
		"printHrefCategories": printHrefCategories,
		"partial":             partial(data),
	}

	t, err := template.New(name).Funcs(funcs).Parse(layout)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func partial(data interface{}) partialFunc {
	return func(name string) (string, error) {
		path := filepath.Join(meta.PartialsDir, name + ".html")

		b, err := ioutil.ReadFile(path)

		if err != nil {
			return "", err
		}

		funcs := template.FuncMap{
			"printCategories":     printCategories,
			"printHrefCategories": printHrefCategories,
			"partial":             partial(data),
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
}

func printCategories(
	categories []category.Category,
	item string,
	nested string,
) string {
	buf := bytes.Buffer{}

	for _, c := range categories {
		buf.WriteString("<" + item + ">" + c.Name)

		if len(c.Categories) > 0 {
			buf.WriteString("<" + nested + ">")
			buf.WriteString(printCategories(c.Categories, item, nested))
			buf.WriteString("</" + nested + ">")
		}

		buf.WriteString("</" + item + ">")
	}

	return buf.String()
}

func printHrefCategories(
	categories []category.Category,
	item string,
	nested string,
) string {
	buf := bytes.Buffer{}

	for _, c := range categories {
		link := "<a href=\"" + c.Href() + "\">" + c.Name + "</a>"
		buf.WriteString("<" + item + ">" + link)

		if len(c.Categories) > 0 {
			buf.WriteString("<" + nested + ">")
			buf.WriteString(printHrefCategories(c.Categories, item, nested))
			buf.WriteString("</" + nested + ">")
		}

		buf.WriteString("</" + item + ">")
	}

	return buf.String()
}
