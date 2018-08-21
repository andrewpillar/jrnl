package resolve

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/andrewpillar/jrnl/post"
)

type Resolver struct {
	wg *sync.WaitGroup

	dir string

	posts chan *post.Post
}

func New(dir string) *Resolver {
	return &Resolver{
		wg:      &sync.WaitGroup{},
		dir:     dir,
		posts:   make(chan *post.Post),
	}
}

func (r *Resolver) ResolvePost(id string) (*post.Post, error) {
	var p *post.Post

	ch := r.ResolvePosts()

	for found := range ch {
		if p == nil && found.ID == id {
			p = found
		}
	}

	if p == nil {
		return p, errors.New("post not found: " + id)
	}

	return p, nil
}

func (r *Resolver) ResolvePosts() chan *post.Post {
	filepath.Walk(r.dir, r.walk)

	go func() {
		r.wg.Wait()
		close(r.posts)
	}()

	return r.posts
}

func (r *Resolver) ResolvePostsToStore() post.Store {
	ch := r.ResolvePosts()

	posts := post.NewStore()

	for p := range ch {
		posts.Put(p)
	}

	return posts
}

func (r *Resolver) walk(path string, info os.FileInfo, err error) error {
	r.wg.Add(1)

	go func() {
		if info.Name() == r.dir || info.IsDir() {
			r.wg.Done()
			return
		}

		p, _ := post.NewFromPath(path)

		r.posts <- p
		r.wg.Done()
	}()

	return nil
}
