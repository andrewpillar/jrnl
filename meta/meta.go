package meta

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

var File = "_meta.yml"

type Meta struct {
	Title string

	Default string `yaml:",omitempty"`

	Remotes map[string]Remote `yaml:,omitempty"`
}

type Remote struct {
	Target string `yaml:",omitempty"`

	Identity string `yaml:",omitempty"`
}

func Decode(r io.Reader) (*Meta, error) {
	m := &Meta{}

	dec := yaml.NewDecoder(r)

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

func Init(dir string) (*Meta, error) {
	fname := File

	if dir != "" {
		fname = dir + "/" + fname
	}

	f, err := os.OpenFile(fname, os.O_CREATE, 0660)

	if err != nil {
		return nil, err
	}

	f.Close()

	return &Meta{}, nil
}

func (m *Meta) Encode(w io.Writer) error {
	enc := yaml.NewEncoder(w)

	if err := enc.Encode(m); err != nil {
		return err
	}

	enc.Close()

	return nil
}
