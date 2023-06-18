package stack

type Stack[T any] struct {
	arr []T
}

func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Size() int {
	return len(s.arr)
}

func (s *Stack[T]) Push(item T) {
	s.arr = append(s.arr, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.Size() == 0 {
		var t T
		return t, false
	}
	result := s.arr[s.Size()-1]
	s.arr = s.arr[:s.Size()-1]
	return result, true
}
