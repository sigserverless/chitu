package utils

type Option[T any] interface {
	IsNone() bool
	ToNilable() *T
	Unwrap() T
}

type Some[T any] struct {
	Val T
}

func NewSome[T any](val T) Option[T] {
	return &Some[T]{val}
}

type None[T any] struct{}

func NewNone[T any]() Option[T] {
	return &None[T]{}
}

func (*Some[T]) IsNone() bool {
	return false
}

func (*None[T]) IsNone() bool {
	return true
}

func (s *Some[T]) ToNilable() *T {
	return &s.Val
}

func (n *None[T]) ToNilable() *T {
	return nil
}

func (s *Some[T]) Unwrap() T {
	return s.Val
}

func (n *None[T]) Unwrap() T {
	panic("Unwrap a None.")
}
