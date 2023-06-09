package utils

import (
	"fmt"
)

type Number interface {
	float64 | int
}

// type Number float64

func Map[T any](s []T, f func(T) T) []T {
	result := make([]T, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func Mean[N Number](a, b []N) []N {
	l := len(a)
	c := []N{}
	for i := 0; i < l; i++ {
		c = append(c, (a[i]+b[i])/2)
	}
	return c
}

func Sum[N Number](a, b []N) []N {
	l := len(a)
	c := []N{}
	for i := 0; i < l; i++ {
		c = append(c, a[i]+b[i])
	}
	return c
}

func convertToSlice[T any](a []any) []T {
	result := make([]T, len(a))
	for i, v := range a {
		if n, ok := v.(T); ok {
			result[i] = n
		} else {
			panic(fmt.Sprintf("type assertion panicked, expected %T, got %T", n, v))
		}
	}
	return result
}

func ConvertToSlice[T any](a any) []T {
	if _, ok := a.([]T); ok {
		return a.([]T)
	}

	if _, ok := a.([]any); ok {
		return convertToSlice[T](a.([]any))
	}

	panic(fmt.Sprintf("type assertion panicked, expected %T, got %T", a, a))
}
