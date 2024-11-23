package removeWithSlices

import "slices"

type Store[T comparable] struct {
	data []T
}

func (s *Store[T]) Add(a T) {
	s.data = append(s.data, a)
}

func (s *Store[T]) Remove(a T) {
	s.data = slices.DeleteFunc(s.data, func(v T) bool {
		return v == a
	})
}
