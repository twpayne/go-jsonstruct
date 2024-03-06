package jsonstruct

import (
	"cmp"
	"slices"
)

func allKeys[M map[K]V, K comparable, V any](m M) []K {
	result := make([]K, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	return result
}

func sortedKeys[M map[K]V, K cmp.Ordered, V any](m M) []K {
	result := allKeys(m)
	slices.Sort(result)
	return result
}
