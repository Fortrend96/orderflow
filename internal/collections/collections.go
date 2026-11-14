// Package collections contains reusable type-safe algorithms for slices and sets.
package collections

// Set contains unique values of a comparable type.
// Its zero value is ready to use. Set is not safe for concurrent mutation.
type Set[T comparable] struct {
	items map[T]struct{}
}

// Add inserts value and reports whether it was not already present.
func (s *Set[T]) Add(value T) bool {
	if s.items == nil {
		s.items = make(map[T]struct{})

	}

	if _, exists := s.items[value]; exists {
		return false
	}

	s.items[value] = struct{}{}
	return true
}

// Contains reports whether value belongs to the set.
func (s *Set[T]) Contains(value T) bool {
	_, exists := s.items[value]
	return exists
}

// Len reports the number of unique values.
func (s *Set[T]) Len() int {
	return len(s.items)
}

// Clear removes all values without requiring a new Set.
func (s *Set[T]) Clear() {
	clear(s.items)
}

// UniqueBy returns the first element for each distinct key, preserving order.
// It returns an independent slice and preserves nil input.
// The key function must be non-nil when the input is non-empty.
func UniqueBy[S ~[]E, E any, K comparable](values S, key func(E) K) S {
	if values == nil {
		return nil
	}

	result := make(S, 0, len(values))
	var seen Set[K]
	for _, value := range values {
		if seen.Add(key(value)) {
			result = append(result, value)
		}
	}

	return result
}

// GroupBy groups elements by key. Their order within each group is preserved.
// The returned map is independent of the input slice and is non-nil even for nil input.
// Iterating map keys has no guaranteed order.
func GroupBy[S ~[]E, E any, K comparable](values S, key func(E) K) map[K][]E {
	groups := make(map[K][]E)

	for _, value := range values {
		k := key(value)
		groups[k] = append(groups[k], value)
	}

	return groups
}
