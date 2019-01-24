package feed

import (
	"io"

	"github.com/andrewpillar/jrnl/post"

	"github.com/gorilla/feeds"
)

type Feed struct {
	Title       string
	Link        string
	Description string
	Author      *feeds.Author
}

func (f Feed) generateItems(posts []*post.Post) []*feeds.Item {
	items := make([]*feeds.Item, len(posts), len(posts))

	for i, p := range posts {
		items[i] = &feeds.Item{
			Title: p.Title,
			Link:  &feeds.Link{
				Href: f.Link + p.Href(),
			},
			Description: p.Preview,
			Author:      f.Author,
			Created:     p.CreatedAt,
		}
	}

	return items
}

func (f Feed) WriteRss(w io.Writer, posts []*post.Post) error {
	fd := &feeds.Feed{
		Title: f.Title,
		Link:  &feeds.Link{
			Href: f.Link,
		},
		Description: f.Description,
		Author:      f.Author,
	}

	fd.Items = f.generateItems(posts)

	return fd.WriteRss(w)
}

func (f Feed) WriteAtom(w io.Writer, posts []*post.Post) error {
	fd := &feeds.Feed{
		Title: f.Title,
		Link:  &feeds.Link{
			Href: f.Link,
		},
		Description: f.Description,
		Author:      f.Author,
	}

	fd.Items = f.generateItems(posts)

	return fd.WriteAtom(w)
}
