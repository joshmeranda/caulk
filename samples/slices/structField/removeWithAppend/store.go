package removeWithAppend

type Store[T comparable] struct {
	data []T
}

func (s *Store[T]) Add(a T) {
	s.data = append(s.data, a)
}

func (s *Store[T]) Remove(a T) {
	for i, v := range s.data {
		if v == a {
			s.data = append(s.data[:i], s.data[i+1:]...)
			return
		}
	}
}
