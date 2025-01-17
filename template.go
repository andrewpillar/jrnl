package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type Templates struct {
	dir string
	mu  sync.RWMutex
	tab map[string]*template.Template
}

func NewTemplates(dir string) *Templates {
	return &Templates{
		dir: dir,
	}
}

func (t *Templates) key(names []string) string { return strings.Join(names, ":") }

func (t *Templates) Get(names []string) (*template.Template, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.tab == nil {
		return nil, false
	}

	tmpl, ok := t.tab[t.key(names)]
	return tmpl, ok
}

func (t *Templates) Set(names []string, tmpl *template.Template) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.tab == nil {
		t.tab = make(map[string]*template.Template)
	}
	t.tab[t.key(names)] = tmpl
}

func (t *Templates) tmplFuncHas(arr []string, item string) bool {
	tab := make(map[string]struct{})

	for _, s := range arr {
		tab[s] = struct{}{}
	}

	_, ok := tab[item]
	return ok
}

func (t *Templates) tmplFuncInclude(name string) (string, error) {
	b, err := os.ReadFile(name)

	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *Templates) tmplFuncPartial(name string, data any) (string, error) {
	tmpl, err := t.Load(name, name)

	if err != nil {
		return "", err
	}

	tmpl.Funcs(t.funcs())

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func (t *Templates) funcs() template.FuncMap {
	return template.FuncMap{
		"has":     t.tmplFuncHas,
		"include": t.tmplFuncInclude,
		"partial": t.tmplFuncPartial,
	}
}

func (t *Templates) Load(name string, tmpls ...string) (*template.Template, error) {
	tmpl, ok := t.Get(tmpls)

	if ok {
		return tmpl, nil
	}

	tmpl = template.New(name)
	tmpl.Funcs(t.funcs())

	for _, name := range tmpls {
		b, err := os.ReadFile(filepath.Join(t.dir, name))

		if err != nil {
			return nil, err
		}

		tmpl, err = tmpl.Parse(string(b))

		if err != nil {
			return nil, err
		}
	}

	t.Set(tmpls, tmpl)

	return tmpl, nil
}
