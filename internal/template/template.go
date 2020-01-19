package template

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/andrewpillar/jrnl/internal/config"

	"github.com/grokify/html-strip-tags-go"
)

var funcs template.FuncMap

func init() {
	funcs = template.FuncMap{
		"partial": partial,
		"strip":   strip.StripTags,
	}
}

func partial(path string, data interface{}) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(config.LayoutsDir, path))

	if err != nil {
		return "", err
	}

	t, err := template.New(path).Funcs(funcs).Parse(string(b))

	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}

	err = t.Execute(buf, data)
	return buf.String(), err
}

func Render(w io.Writer, name, layout string, data interface{}) error {
	t, err := template.New(name).Funcs(funcs).Parse(layout)

	if err != nil {
		return err
	}

	return t.Execute(w, data)
}
