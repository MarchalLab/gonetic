package editor

import "math"

// MeanVariance returns the sample mean and unbiased variance.
func MeanVariance(x []float64) (mean, variance float64) {
	n := len(x)
	if n == 0 {
		return math.NaN(), math.NaN()
	}

	// Uniform weights
	sum := 0.0
	for _, v := range x {
		sum += v
	}
	mean = sum / float64(n)

	for _, v := range x {
		d := v - mean
		variance += d * d
	}
	variance /= float64(n - 1) // unbiased
	return mean, variance
}

// MeanStdDev returns the sample mean and unbiased standard deviation.
func MeanStdDev(x []float64) (mean, std float64) {
	mean, variance := MeanVariance(x)
	return mean, math.Sqrt(variance)
}

// Normal represents a normal (Gaussian) distribution.
type Normal struct {
	Mu    float64 // Mean of the normal distribution
	Sigma float64 // Standard deviation
}

// CDF computes the cumulative distribution function at x.
func (n Normal) CDF(x float64) float64 {
	return 0.5 * math.Erfc(-(x-n.Mu)/(n.Sigma*math.Sqrt2))
}
