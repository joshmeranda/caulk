package removeNotDefined

type Store[T comparable] struct {
	data []T
}

func (s *Store[T]) Add(a T) {
	s.data = append(s.data, a)
}

func (s *Store[T]) Remove(a T) {}
