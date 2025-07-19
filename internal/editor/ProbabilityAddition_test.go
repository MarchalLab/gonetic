package editor_test

import (
	"math"
	"testing"

	"log/slog"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/editor"
)

// Mock geneProbability map
var testGeneProbs = map[types.GeneID]float64{
	1: 0.4, // from
	2: 0.6, // to
}

type geneProbability func(types.GeneID) float64

func (gp geneProbability) Get(id types.GeneID) float64 {
	return gp(id)
}

type testCase struct {
	tag      string
	expected map[editor.WeightTarget]float64
}

func TestComposeWeighting(t *testing.T) {
	interaction := types.FromToToID(1, 2)
	original := 0.5

	probs := func(id types.GeneID) float64 {
		return testGeneProbs[id]
	}

	cases := []testCase{
		{
			tag: "none",
			expected: map[editor.WeightTarget]float64{
				editor.FromOnly:  original,
				editor.ToOnly:    original,
				editor.BothBayes: original,
			},
		},
		{
			tag: "mult",
			expected: map[editor.WeightTarget]float64{
				editor.FromOnly:  original * 0.4,
				editor.ToOnly:    original * 0.6,
				editor.BothBayes: original * (1 - (1-0.4)*(1-0.6)),
			},
		},
		{
			tag: "mean",
			expected: map[editor.WeightTarget]float64{
				editor.FromOnly:  (original + 0.4) / 2,
				editor.ToOnly:    (original + 0.6) / 2,
				editor.BothBayes: (original + (1 - (1-0.4)*(1-0.6))) / 2,
			},
		},
		{
			tag: "bayes",
			expected: map[editor.WeightTarget]float64{
				editor.FromOnly:  1 - (1-original)*(1-0.4),
				editor.ToOnly:    1 - (1-original)*(1-0.6),
				editor.BothBayes: 1 - (1-original)*(1-0.4)*(1-0.6),
			},
		},
	}

	for _, tc := range cases {
		for target, want := range tc.expected {
			t.Run(tc.tag+"_"+target.String(), func(t *testing.T) {
				weightFunc := editor.ComposeWeighting(tc.tag, target, slog.Default(), "test")
				got := weightFunc(interaction, original, probs)

				if !almostEqual(got, want, 1e-6) {
					t.Errorf("tag=%s, target=%s: got %f, want %f", tc.tag, target, got, want)
				}
			})
		}
	}
}

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}
