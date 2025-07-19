package powerlaws_test

import (
	"math"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/compare"
	"github.com/MarchalLab/gonetic/internal/powerlaws"
)

type hzTestCase struct {
	x float64
	q float64
	v float64
}

var hzTestCases = []hzTestCase{
	// not implemented
	{-1, 1, math.NaN()},     // x < 1 is not implemented (-1/12)
	{0, 1, math.NaN()},      // x < 1 is not implemented (-1/2)
	{0, 2, math.NaN()},      // x < 1 is not implemented (-3/2)
	{0.1, 1e10, math.NaN()}, // x < 1 is not implemented (-1.11111e9)
	{1.2, -1.5, math.NaN()}, // x not integer and q < 0 is not implemented (10.16663)
	{2, -5, math.NaN()},     // q <= 0 integer is not implemented (π^2/6 + 5269/3600)
	{2, -1, math.NaN()},     // q <= 0 integer is not implemented (π^2/6 + 1)
	{2, 0, math.NaN()},      // q <= 0 integer is not implemented (π^2/6)

	// Riemann zeta
	{1, 1, math.MaxFloat64},
	{2, 1, 1.64493}, // π^2/6
	{2, 1, 1.64500},
	{2.5, 1, 1.34149},
	{3, 1, 1.202057},          // Apery's constant
	{4, 1, 1.082323233711138}, //  π^4/90

	// generalized zeta
	{1, 0, math.MaxFloat64},
	{1.1, 1e10, 1.000000000005},
	{1.1, 1e9, 1.25892},
	{1.5, 1.5, 1.94811},
	{1.5, 1e9, 0.0000632456},
	{2, -10000000.5, 9.8696},
	{2, -1.5, 9.37925},
	{2, 2, 0.64493}, // π^2/6 - 1
	{2, 3, 0.39493}, // π^2/6 - 5/4
	{2, 4, 0.28382}, // π^2/6 - 49/36
	{2, 1000, 0.0010005001},
	{1.5, 2.5, 1.40378},
	{2.5, 1.5, 0.59026},
	{10, 0.9, 2.86963},
	{10, 0.09, 2.86797e10},
	{1000, 1000, 0}, // ~10^-3000
}

var tolerance = compare.Tolerance(0.0001)

func TestHurwitzZeta(t *testing.T) {
	for testCaseIdx, testCase := range hzTestCases {
		value := powerlaws.HurwitzZeta(testCase.x, testCase.q)
		if !tolerance.FloatEqualWithinTolerance(value, testCase.v) {
			t.Errorf("case %d: zeta(%f, %f) %f should be %f", testCaseIdx, testCase.x, testCase.q, value, testCase.v)
		}
	}
}
