package collections

import (
	"bytes"
	"testing"
)

// FuzzUniqueByAndGroupBy tests algebraic properties, not fixed expected examples.
func FuzzUniqueByAndGroupBy(f *testing.F) {
	f.Add([]byte{1, 2, 1, 3, 2})
	f.Add([]byte{})
	f.Add([]byte{255, 0, 7, 14, 21})

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 256 {
			return
		}
		before := bytes.Clone(input)
		key := func(v byte) byte { return v % 7 }
		unique := UniqueBy(input, key)
		groups := GroupBy(input, key)

		// Independently compute first occurrence using simple nested scans.
		var expected []byte
		for _, v := range input {
			found := false
			for _, prev := range expected {
				if key(prev) == key(v) {
					found = true
					break
				}
			}
			if !found {
				expected = append(expected, v)
			}
		}
		if !bytes.Equal(unique, expected) {
			t.Fatalf("UniqueBy(%v): got %v, want %v", input, unique, expected)
		}

		count := 0
		for k, group := range groups {
			for _, v := range group {
				if key(v) != k {
					t.Fatalf("element %d in wrong group %d", v, k)
				}
			}
			count += len(group)
		}
		if count != len(input) {
			t.Fatalf("lost or duplicated items: %d != %d", count, len(input))
		}
		if !bytes.Equal(input, before) {
			t.Fatal("generic algorithms modified their input")
		}
		if len(unique) > 0 {
			unique[0] ^= 0xff
			if !bytes.Equal(input, before) {
				t.Fatal("UniqueBy output aliases the input")
			}
		}
	})
}
