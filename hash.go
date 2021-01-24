package main

import (
	"bytes"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
)

type Hash struct {
	f   *os.File
	set map[string][]byte
}

type Hasher interface {
	Hash() []byte
}

func OpenHash() (*Hash, error) {
	f, err := os.OpenFile(filepath.Join(dataDir, "hash"), os.O_CREATE|os.O_RDWR, os.FileMode(0644))

	if err != nil {
		return nil, err
	}

	set := make(map[string][]byte)

	if err := gob.NewDecoder(f).Decode(&set); err != nil {
		if err != io.EOF {
			return nil, err
		}
	}
	return &Hash{
		f:   f,
		set: set,
	}, nil
}

func (h *Hash) Get(key string) ([]byte, bool) {
	b, ok := h.set[key]
	return b, ok
}

func (h *Hash) Put(key string, hs Hasher) bool {
	b := hs.Hash()

	b0 := h.set[key]

	if !bytes.Equal(b0, b) {
		h.set[key] = b
		return true
	}
	return false
}

func (h *Hash) Delete(key string) {
	if _, ok := h.set[key]; ok {
		delete(h.set, key)
	}
}

func (h *Hash) Save() error {
	if err := h.f.Truncate(0); err != nil {
		return err
	}

	if _, err := h.f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return gob.NewEncoder(h.f).Encode(h.set)
}

func (h *Hash) Close() error { return h.f.Close() }
