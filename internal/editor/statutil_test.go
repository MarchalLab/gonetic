package editor

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"

	"github.com/MarchalLab/gonetic/internal/common/compare"
)

var tol = compare.Tolerance(1e-10)

func TestMeanStdDevMatchesGonum(t *testing.T) {
	testCases := [][]float64{
		{},
		{1, 2, 3, 4, 5},
		{10, 10, 10, 10},
		{-3, 0, 3},
		{0.1, 0.2, 0.3, 0.4, 0.5},
		{1000, 1001, 999, 1002, 998},
	}

	for _, data := range testCases {
		wantMean, wantStd := stat.MeanStdDev(data, nil)

		gotMean, gotStd := MeanStdDev(data)

		if !tol.FloatEqualWithinTolerance(gotMean, wantMean) {
			t.Errorf("Mean mismatch\nData: %v\nGot: %.12f\nWant: %.12f", data, gotMean, wantMean)
		}
		if !tol.FloatEqualWithinTolerance(gotStd, wantStd) {
			t.Errorf("StdDev mismatch\nData: %v\nGot: %.12f\nWant: %.12f", data, gotStd, wantStd)
		}
	}
}

func TestNormalCDFMatchesGonum(t *testing.T) {
	testCases := []struct {
		mu, sigma float64
		points    []float64
	}{
		{0, 1, []float64{-2, -1, 0, 1, 2}},
		{5, 2, []float64{3, 4, 5, 6, 7}},
		{-3, 0.5, []float64{-4, -3.5, -3, -2.5, -2}},
		{100, 10, []float64{80, 90, 100, 110, 120}},
	}

	for _, tc := range testCases {
		ref := distuv.Normal{Mu: tc.mu, Sigma: tc.sigma, Src: rand.New(rand.NewSource(1))}
		n := Normal{Mu: tc.mu, Sigma: tc.sigma}

		for _, x := range tc.points {
			want := ref.CDF(x)
			got := n.CDF(x)

			if !tol.FloatEqualWithinTolerance(got, want) {
				t.Errorf("CDF mismatch\nMu: %.2f, Sigma: %.2f, x: %.2f\nGot: %.12f\nWant: %.12f",
					tc.mu, tc.sigma, x, got, want)
			}
		}
	}
}
