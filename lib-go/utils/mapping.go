package utils

func MakeMapping[T any, K comparable](getKey func(T) K, items []T) map[K]T {
	mapping := map[K]T{}
	for _, item := range items {
		mapping[getKey(item)] = item
	}
	return mapping
}
