package feed

import (
	"io"

	"github.com/andrewpillar/jrnl/post"

	"github.com/gorilla/feeds"

	"github.com/mmcdole/gofeed"
)

type Feed struct {
	Title       string
	Link        string
	Description string
	Author      *feeds.Author
}

func Read(urls ...string) ([]Feed, error) {
	roll := make([]Feed, len(urls), len(urls))

	for i, url := range urls {
		p := gofeed.NewParser()

		feed, err := p.ParseURL(url)

		if err != nil {
			return roll, err
		}

		if len(feed.Items) == 0 {
			continue
		}

		item := feed.Items[0]

		roll[i] = Feed{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Author:      &feeds.Author{},
		}

		if item.Author != nil {
			roll[i].Author.Name = item.Author.Name
			roll[i].Author.Email = item.Author.Email
		}
	}

	return roll, nil
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

// Write the posts as an RSS feed to the given io.Writer.
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

// Write the posts as an Atom feed to the given io.Writer.
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
