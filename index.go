package main

import (
	"time"

	"github.com/google/btree"
)

type indexItem struct {
	ID        string
	CreatedAt time.Time
}

type Index struct {
	tree *btree.BTree
}

func NewIndex() *Index {
	return &Index{
		tree: btree.New(3),
	}
}

func (i *Index) postItem(p *Post) indexItem {
	return indexItem{
		ID:        p.ID,
		CreatedAt: p.CreatedAt.Time,
	}
}

func (i *Index) Put(p *Post) {
	i.tree.ReplaceOrInsert(i.postItem(p))
}

func (i *Index) Walk(fn func(string)) {
	i.tree.Descend(func(it btree.Item) bool {
		fn(it.(indexItem).ID)
		return true
	})
}

func (a indexItem) Less(b btree.Item) bool {
	return !a.CreatedAt.After(b.(indexItem).CreatedAt)
}
