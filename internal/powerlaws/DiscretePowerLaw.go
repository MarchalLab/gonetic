package powerlaws

import (
	"math"
	"sort"
)

// DiscretePowerLaw implementation based on https://github.com/Data2Semantics/powerlaws (Java implementation)
type DiscretePowerLaw struct {
	xMin     int64
	exponent float64
	pdenum   float64
}

func (d DiscretePowerLaw) Exponent() float64 {
	return d.exponent
}

func (d DiscretePowerLaw) XMin() int64 {
	return d.xMin
}

// Step range
const alphaMin = 1.5
const alphaMax = 3.5
const alphaStep = 0.01

func FitDiscretePowerLaw(data []int64, xMin int64) DiscretePowerLaw {
	bestAlpha := -1.0
	maxLL := -math.MaxFloat64
	for alpha := alphaMin; alpha < alphaMax; alpha += alphaStep {
		ll := logLikelihood(data, alpha, float64(xMin))
		if ll > maxLL {
			bestAlpha = alpha
			maxLL = ll
		}
	}
	return newDiscretePowerLaw(xMin, bestAlpha)
}

func DiscretePowerLawFit(data []int64) DiscretePowerLaw {
	counts := CumulativeCounts(data)
	unique := make(map[int64]struct{})
	for _, datum := range data {
		unique[datum] = struct{}{}
	}
	best := DiscretePowerLaw{}
	bestDistance := math.MaxFloat64
	for datum := range unique {
		current := FitDiscretePowerLaw(data, datum)
		currentDistance := current.DiscreteKSTest(counts)
		if currentDistance < bestDistance {
			bestDistance = currentDistance
			best = current
		}
	}
	return best
}

func newDiscretePowerLaw(xMin int64, exponent float64) DiscretePowerLaw {
	return DiscretePowerLaw{
		xMin:     xMin,
		exponent: exponent,
		pdenum:   HurwitzZeta(exponent, float64(xMin)),
	}
}

// CdfInv computes a value x such that P(x) = q
func (d DiscretePowerLaw) CdfInv(q float64) int64 {
	x1, x2 := float64(d.xMin), float64(d.xMin)
	first := true
	for first || d.cdfComp(x2) >= q {
		x1 = x2
		x2 = 2 * x1
		first = false
	}
	return int64(d.binarySearch(x1, x2, q))
}

func CumulativeCounts(data []int64) []int {
	// sort data
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	// create sufficiently large count array
	xMax := data[len(data)-1]
	cumulativeCounts := make([]int, xMax+1)
	// populate cumulative counts
	dataIndex := 0
	for x := int64(0); x <= xMax; x++ {
		for dataIndex < len(data) && data[dataIndex] <= x {
			cumulativeCounts[x]++
			dataIndex++
		}
		if x > 0 {
			cumulativeCounts[x] += cumulativeCounts[x-1]
		}
	}
	return cumulativeCounts
}

// DiscreteKSTest performs the Kolmogorov Smirnov test on this power law and the distribution suggested by the data
// Note: this function assumes that data are sorted
// Note: unlike the continuous KS test, this version loops over all (integer) values of x between xMin and xMax rather than just the data points.
func (d DiscretePowerLaw) DiscreteKSTest(counts []int) float64 {
	xMax := int64(len(counts) - 1)
	shift := counts[d.xMin-1]
	copySize := float64(counts[xMax] - shift)
	plCDF := 0.0
	maxDiff := -math.MaxFloat64
	for x := d.xMin; x <= xMax; x++ {
		count := float64(counts[x] - shift)
		dataCDF := count / copySize
		plCDF += math.Pow(float64(x), -d.exponent) / HurwitzZeta(d.exponent, float64(d.xMin))
		diff := math.Abs(dataCDF - plCDF)
		maxDiff = math.Max(diff, maxDiff)
	}
	return maxDiff
}

func (d DiscretePowerLaw) binarySearch(lower, upper, target float64) float64 {
	// * stop recursion when the interval falls within a single integer
	if math.Floor(lower) == math.Floor(upper) {
		return lower
	}
	r := upper - lower
	midpoint := r/2.0 + lower
	pm := d.cdfComp(midpoint)
	if pm < target {
		return d.binarySearch(lower, midpoint, target)
	} else {
		return d.binarySearch(midpoint, upper, target)
	}
}

func (d DiscretePowerLaw) cdfComp(x float64) float64 {
	return HurwitzZeta(d.exponent, x) / d.pdenum
}

// logLikelihood computes the log likelihood of the data for given parameters
func logLikelihood(data []int64, alpha float64, xMin float64) float64 {
	sum := 0.0
	n := 0
	for _, datum := range data {
		if datum >= int64(xMin) {
			sum += math.Log(float64(datum))
			n++
		}
	}
	return -float64(n)*math.Log(HurwitzZeta(alpha, xMin)) - alpha*sum
}
