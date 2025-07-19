package pathweight

import (
	"testing"

	"log/slog"
)

func TestFunctionalWeighting(t *testing.T) {
	logger := slog.Default()

	functionalScores := map[string]struct{}{
		"gene1\t0.8\tcondition1": {},
		"gene2\t0.5\tcondition1": {},
		"gene3\t0.3\tcondition1": {},
		"gene4\t0.8\tcondition2": {},
	}

	funcScore := true

	fw := NewFunctionalWeighting(logger, functionalScores, funcScore)

	tests := []struct {
		functionalData float64
		expected       float64
	}{
		{1.0, 4.0 / 4.0},
		{0.8, 4.0 / 4.0},
		{0.5, 2.0 / 4.0},
		{0.3, 1.0 / 4.0},
		{0.0, 0.0 / 4.0},
	}

	for _, test := range tests {
		result := fw.calculateProbabilityForFunctionalData(test.functionalData)
		if result != test.expected {
			t.Errorf("expected %f, got %f for functional data %f", test.expected, result, test.functionalData)
		}
	}
}
