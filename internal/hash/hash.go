package hash

import (
	"encoding/gob"
	"io"
)

type Hash map[string][]byte

func init() {
	gob.Register(New())
}

func New() Hash {
	return Hash(make(map[string][]byte))
}

func Decode(r io.Reader) (Hash, error) {
	h := New()

	dec := gob.NewDecoder(r)
	err := dec.Decode(&h)

	return h, err
}

func (h Hash) Encode(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(h)
}
