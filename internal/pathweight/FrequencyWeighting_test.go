package pathweight

import (
	"testing"

	"log/slog"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func TestFrequencyWeighting(t *testing.T) {
	logger := slog.Default()

	freqIncreasePopulations := map[string]struct{}{
		"gene1\t0.8\tcondition1": {},
		"gene2\t0.5\tcondition1": {},
		"gene3\t0.3\tcondition1": {},
		"gene4\t0.8\tcondition2": {},
	}

	populations := types.ConditionSet{
		"condition1": {},
		"condition2": {},
	}

	freqCutoff := 0.1
	doFrequencyWeighting := true

	fw := NewFrequencyWeighting(logger, freqIncreasePopulations, populations, freqCutoff, doFrequencyWeighting)

	tests := []struct {
		population        types.Condition
		frequencyIncrease float64
		expected          float64
	}{
		{"condition1", 1.0, 3.0 / 3.0},
		{"condition1", 0.8, 3.0 / 3.0},
		{"condition1", 0.5, 2.0 / 3.0},
		{"condition1", 0.3, 1.0 / 3.0},
		{"condition1", 0.0, 0.0 / 3.0},
		{"condition2", 0.8, 1.0 / 1.0},
	}

	for _, test := range tests {
		result := fw.calculateProbabilityForFreqDataPerLine(test.population, test.frequencyIncrease)
		if result != test.expected {
			t.Errorf("expected %f, got %f for population %s and frequency increase %f", test.expected, result, test.population, test.frequencyIncrease)
		}
	}
}
