package internal

import (
	"iter"
)

// Pair represents a tuple of two values
type NextPair[T1, T2 any] struct {
	First  T1
	Second T2
}

func IterZip[T1, T2 any](s1 []T1, s2 []T2) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		n := len(s1)
		if n2 := len(s2); n2 < n {
			n = n2
		}

		for i := 0; i < n; i++ {
			if !yield(s1[i], s2[i]) {
				return
			}
		}
	}
}
