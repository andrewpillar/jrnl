package post

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
)

type Resolver struct {
	dir string

	posts Store
}

func NewResolver() *Resolver {
	return &Resolver{
		dir:   meta.PostsDir,
		posts: NewStore(),
	}
}

func (r *Resolver) Resolve() Store {
	filepath.Walk(r.dir, r.walk)

	return r.posts
}

func (r *Resolver) walk(path string, info os.FileInfo, err error) error {
	if info.Name() == r.dir || info.IsDir() {
		return nil
	}

	id := strings.Replace(path, r.dir + "/", "", 1)

	p, _ := Find(strings.Split(id, ".")[0])

	r.posts.Put(p)

	return nil
}