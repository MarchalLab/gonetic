package readers

import "cmp"

// Quickselect is a generic quickselect algorithm
func Quickselect[T cmp.Ordered](a []T, k int) T {
	lo, hi := 0, len(a)-1
	for {
		if lo == hi {
			return a[lo]
		}
		pivotIndex := lo + (hi-lo)/2
		pivotIndex = partition(a, lo, hi, pivotIndex)
		if k == pivotIndex {
			return a[k]
		} else if k < pivotIndex {
			hi = pivotIndex - 1
		} else {
			lo = pivotIndex + 1
		}
	}
}

// partition is a helper function for Quickselect
func partition[T cmp.Ordered](a []T, lo, hi, pivotIndex int) int {
	pivot := a[pivotIndex]
	a[pivotIndex], a[hi] = a[hi], a[pivotIndex]
	storeIndex := lo
	for i := lo; i < hi; i++ {
		if a[i] < pivot {
			a[storeIndex], a[i] = a[i], a[storeIndex]
			storeIndex++
		}
	}
	a[hi], a[storeIndex] = a[storeIndex], a[hi]
	return storeIndex
}
