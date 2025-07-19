package powerlaws_test

import (
	"math"
	"sort"
	"testing"

	"github.com/MarchalLab/gonetic/internal/powerlaws"
)

type discreteTestCase struct {
	dataMap         map[int64]int
	inflectionValue int64
}

var dataTestCases = []discreteTestCase{
	{
		map[int64]int{1: 262, 2: 187, 3: 116, 4: 95, 5: 81, 6: 73, 7: 60, 8: 47, 9: 48, 10: 39, 11: 36, 12: 31, 13: 30, 14: 25, 15: 29, 16: 26, 17: 15, 18: 10, 19: 11, 20: 8, 21: 13, 22: 7, 23: 7, 24: 4, 25: 5, 26: 6, 27: 5, 28: 6, 29: 6, 30: 5, 31: 2, 32: 9, 33: 7, 34: 8, 35: 18, 36: 7, 37: 3, 38: 5, 39: 4, 40: 1, 41: 3, 42: 2, 43: 4, 44: 2, 45: 2, 46: 1, 47: 5, 48: 8, 49: 6, 50: 2, 51: 4, 52: 3, 53: 2, 54: 4, 55: 15, 56: 2, 57: 2, 58: 2, 59: 3, 75: 1, 76: 12, 81: 3, 87: 1, 90: 1, 106: 1, 109: 1, 116: 2, 117: 14, 118: 15, 119: 11, 120: 9, 121: 3, 122: 6, 123: 2, 124: 1, 126: 4, 127: 6, 128: 1, 129: 5, 130: 5, 131: 2, 132: 2, 133: 4, 137: 1, 138: 1, 148: 1, 156: 1, 157: 1, 159: 3, 163: 2},
		106,
	},
}

func caseToData(testCase discreteTestCase) []int64 {
	data := make([]int64, 0)
	for key, val := range testCase.dataMap {
		for idx := 0; idx < val; idx++ {
			data = append(data, key)
		}
	}
	return data
}

func TestCumulativeCounts(t *testing.T) {
	for testCaseIdx, testCase := range dataTestCases {
		data := make([]int64, 0)
		cumul := make(map[int]int)
		curr := 0

		maxKey := int64(0)
		for key := range testCase.dataMap {
			maxKey = max(key, maxKey)
		}
		for key := int64(0); key <= maxKey; key++ {
			val := testCase.dataMap[key]
			for idx := 0; idx < val; idx++ {
				data = append(data, key)
				curr++
			}
			count := curr
			cumul[int(key)] = count
		}
		counts := powerlaws.CumulativeCounts(data)
		if counts[len(counts)-1] != len(data) {
			t.Errorf("case %d: incorrect total count %d, expected %d", testCaseIdx, counts[len(counts)-1], len(data))
		}
		for val, count := range counts {
			if count != cumul[val] {
				t.Errorf("case %d: incorrect count %d at position %d, expected %d", testCaseIdx, count, val, cumul[val])
			}
		}
	}
}

func TestDiscretePowerLaw_CdfInv(t *testing.T) {
	for testCaseIdx, testCase := range dataTestCases {
		data := caseToData(testCase)
		inflectionValue := powerlaws.DiscretePowerLawFit(data).CdfInv(0.1)
		if inflectionValue != testCase.inflectionValue {
			t.Errorf("case %d: incorrect inflection value %d should be %d", testCaseIdx, inflectionValue, testCase.inflectionValue)
		}
	}
}

// originalDiscreteKSTest performs the Kolmogorov Smirnov test on a power law with parameters xMin and exponent
// This is the original implementation, iterating over the data on each call
// Note: that unlike the continuous KS test, this version loops over all (integer) values of x between xMin and xMax rather than just the data points.
func originalDiscreteKSTest(xMin int64, exponent float64, data []int64) float64 {
	xMax := int64(math.MinInt64)
	dataCopy := make([]int64, 0, len(data))
	for _, datum := range data {
		if datum >= xMin {
			dataCopy = append(dataCopy, datum)
			if datum > xMax {
				xMax = datum
			}
		}
	}
	sort.Slice(dataCopy, func(i, j int) bool { return dataCopy[i] < dataCopy[j] })
	counts := count(xMin, xMax, dataCopy)
	maxDiff := -math.MaxFloat64
	copySize := float64(len(dataCopy))
	plCDF := 0.0
	for x := xMin; x < xMax+1; x++ {
		count := float64(counts[x-xMin])
		dataCDF := count / copySize
		plCDF += math.Pow(float64(x), -exponent) / powerlaws.HurwitzZeta(exponent, float64(xMin))
		diff := math.Abs(dataCDF - plCDF)
		maxDiff = math.Max(diff, maxDiff)
	}
	return maxDiff
}

// count creates a cumulative histogram that counts for each integer from min to max (inclusive) the number of data points lower than or equal to that integer.
// Note: this function expects data to be sorted
func count(min, max int64, data []int64) []int64 {
	counts := make([]int64, 0, max+1-min)
	counts = append(counts, 0)
	i := min
	for _, datum := range data {
		if datum > max {
			break
		}
		for datum > i {
			counts = append(counts, counts[i-min])
			i++
		}
		counts[i-min] = counts[i-min] + 1
	}
	return counts
}

func TestNew(t *testing.T) {
	for testCaseIdx, testCase := range dataTestCases {
		data := caseToData(testCase)
		counts := powerlaws.CumulativeCounts(data)
		unique := make(map[int64]struct{})
		for _, datum := range data {
			unique[datum] = struct{}{}
		}
		for datum := 1; datum < len(counts); datum++ {
			if counts[datum] == counts[datum-1] {
				// no data with this value
				continue
			}
			current := powerlaws.FitDiscretePowerLaw(data, int64(datum))
			distance := originalDiscreteKSTest(current.XMin(), current.Exponent(), data)
			newDistance := current.DiscreteKSTest(counts)
			if newDistance != distance {
				t.Errorf("case %d, %d: incorrect distance value %f should be %f", datum, testCaseIdx, newDistance, distance)
			}
		}
	}
}
