package compare

import (
	"cmp"
)

// Between returns the value if it falls between the lower or upper bound, otherwise the exceeded bound
func Between[T cmp.Ordered](lower, val, upper T) T {
	return min(max(lower, val), upper)
}
