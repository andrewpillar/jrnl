package template

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/andrewpillar/jrnl/config"
)

func partial(path string, data interface{}) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(config.LayoutsDir, path))

	if err != nil {
		return "", err
	}

	funcs := template.FuncMap{
		"partial": partial,
	}

	t, err := template.New(path).Funcs(funcs).Parse(string(b))

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
		"partial": partial,
	}

	t, err := template.New(name).Funcs(funcs).Parse(layout)

	if err != nil {
		return err
	}

	return t.Execute(w, data)
}
