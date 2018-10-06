package util

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/andrewpillar/jrnl/meta"
)

func partial(name string, data interface{}) (string, error) {
	path := filepath.Join(meta.LayoutsDir, name)

	b, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}

	funcs := template.FuncMap{
		"partial": partial,
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

func RenderTemplate(w io.Writer, name, layout string, data interface{}) error {
	funcs := template.FuncMap{
		"partial": partial,
	}

	t, err := template.New(name).Funcs(funcs).Parse(layout)

	if err != nil {
		return err
	}

	return t.Execute(w, data)
}
