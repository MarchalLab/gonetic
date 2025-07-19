package compare

import (
	"math"
	"testing"
)

func TestFloatEqualWithinToleranceMatrix(t *testing.T) {
	type testCase struct {
		tolerance Tolerance
		values    []float64
		expected  [][]bool
	}

	testCases := []testCase{
		{
			tolerance: 0.01,
			values:    []float64{1.0, 1.01, 1.02, -1.0, math.NaN(), math.Inf(1), math.Inf(-1)},
			expected: [][]bool{
				{true, true, false, false, false, false, false},
				{true, true, true, false, false, false, false},
				{false, true, true, false, false, false, false},
				{false, false, false, true, false, false, false},
				{false, false, false, false, true, false, false},
				{false, false, false, false, false, true, false},
				{false, false, false, false, false, false, true},
			},
		},
	}

	for _, tc := range testCases {
		for i := 0; i < len(tc.values); i++ {
			for j := i; j < len(tc.values); j++ {
				t.Run("", func(t *testing.T) {
					result := tc.tolerance.FloatEqualWithinTolerance(tc.values[i], tc.values[j])
					if result != tc.expected[i][j] {
						t.Errorf("FloatEqualWithinTolerance(%v, %v, %v) = %v; want %v", tc.tolerance, tc.values[i], tc.values[j], result, tc.expected[i][j])
					}
				})
			}
		}
	}
}
