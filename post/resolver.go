package post

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewpillar/jrnl/meta"
)

func ResolvePosts() (Store, error) {
	posts := NewStore()

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == meta.PostsDir || info.IsDir() {
			return nil
		}

		id := strings.Replace(
			path,
			meta.PostsDir + string(os.PathSeparator),
			"",
			1,
		)

		p, err := Find(strings.Split(id, ".")[0])

		if err != nil {
			return err
		}

		posts.Put(p)

		return nil
	}

	err := filepath.Walk(meta.PostsDir, walk)

	return posts, err
}
