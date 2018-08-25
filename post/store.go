package post

type Store []*Post

func NewStore(posts ...*Post) Store {
	if len(posts) == 0 {
		return Store(make([]*Post, 0))
	}

	return Store(posts)
}

func (s *Store) Put(posts ...*Post) {
	(*s) = append((*s), posts...)
}

func (s Store) Get(i int) (*Post, bool) {
	if i > len(s) {
		return nil, false
	}

	return s[i], true
}
