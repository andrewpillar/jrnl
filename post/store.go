package post

import (
	"sort"
	"strings"
)

var (
	day = 2

	month = 1

	year = 0
)

type group map[string]*Store

type Store []*Post

func newGroup() *group {
	g := group(make(map[string]*Store))

	return &g
}

func NewStore(posts ...*Post) Store {
	if len(posts) == 0 {
		return Store(make([]*Post, 0))
	}

	return Store(posts)
}

func (s *Store) Put(post *Post) {
	(*s) = append((*s), post)
}

func (s *Store) Sort() {
	ids := make([]string, len(*s), len(*s))
	tmp := make(map[string]*Post)

	for i, p := range (*s) {
		ids[i] = p.ID
		tmp[p.ID] = p
	}

	sort.Strings(ids)

	(*s) = Store(make([]*Post, len(*s), len(*s)))

	for i, id := range ids {
		(*s)[i] = tmp[id]
	}
}

func (s Store) GroupByDay() *group {
	return s.groupByDate(day)
}

func (s Store) GroupByMonth() *group {
	return s.groupByDate(month)
}

func (s Store) GroupByYear() *group {
	return s.groupByDate(year)
}

func (s Store) groupByDate(i int) *group {
	g := newGroup()

	for _, p := range s {
		key := strings.Split(p.Date.Format(dateDirFmt), "/")[i]

		_, ok := (*g)[key]

		if !ok {
			tmp := NewStore()

			(*g)[key] = &tmp
		}

		(*g)[key].Put(p)
	}

	return g
}
