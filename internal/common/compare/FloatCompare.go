package compare

import (
	"math"
)

type Tolerance float64

func isPosInf(x float64) bool {
	return math.IsInf(x, +1)
}

func isNegInf(x float64) bool {
	return math.IsInf(x, -1)
}

var funcs = []func(float64) bool{math.IsNaN, isPosInf, isNegInf}

func (tolerance Tolerance) FloatEqualWithinTolerance(x, y float64) bool {
	// if x and y are equal, we are done
	if x == y {
		return true
	}
	// check for NaN, +Inf, and -Inf
	for _, f := range funcs {
		if f(x) || f(y) {
			return f(x) && f(y)
		}
	}
	// if they have different sign, return false
	if x*y < 0 {
		return false
	}
	// if they are finite and same sign, check if diff / mean is within tolerance
	diff := math.Abs(x - y)
	mean := math.Abs(x+y) / 2.0
	return (diff / mean) < float64(tolerance)
}
