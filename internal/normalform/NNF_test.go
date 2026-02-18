package normalform

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/compare"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

func init() {

}

type testCase struct {
	name         string
	interactions []string
	expected     float64
}

var testIW = map[string]float64{
	"1;2;1": 0.95,
	"1;3;1": 0.9,
	"2;3;1": 0.85,
	"2;4;1": 0.8,
	"3;4;1": 0.75,
	"4;5;1": 0.8,
}

var testCases = []testCase{
	{
		"2",
		[]string{
			"1;2;1",
			"2;3;1",
		},
		testIW["1;2;1"] * testIW["2;3;1"],
	},
	{
		"2-2",
		[]string{
			"1;2;1",
			"2;3;1",
			"2;4;1",
		},
		testIW["1;2;1"] * (1 - (1-testIW["2;3;1"])*(1-testIW["2;4;1"])),
	},
	{
		"2-2",
		[]string{
			"1;2;1",
			"2;3;1",
		},
		testIW["1;2;1"] * testIW["2;3;1"],
	},
	{
		"2-2",
		[]string{
			"1;2;1",
			"2;4;1",
		},
		testIW["1;2;1"] * testIW["2;4;1"],
	},
	{
		"2-2",
		[]string{
			"2;3;1",
			"2;4;1",
		},
		0,
	},
	{
		"4",
		[]string{
			"1;2;1",
			"2;3;1",
			"3;4;1",
			"4;5;1",
		},
		testIW["1;2;1"] * testIW["2;3;1"] * testIW["3;4;1"] * testIW["4;5;1"],
	},
	{
		"4",
		[]string{
			"1;2;1",
			"2;3;1",
			"3;4;1",
		},
		0,
	},
	{
		"2-2-common",
		[]string{
			"1;2;1",
			"2;4;1",
			"1;3;1",
			"3;4;1",
		},
		1 - (1-testIW["1;2;1"]*testIW["2;4;1"])*(1-testIW["1;3;1"]*testIW["3;4;1"]),
	},
	{
		"2-2-same", []string{
			"1;2;1",
			"2;3;1",
		},
		testIW["1;2;1"] * testIW["2;3;1"],
	},
}

func TestEvaluateIntersection(t *testing.T) {
	for _, compiler := range []string{"c2d", "d4"} {
		for _, tc := range testCases {
			if !t.Run(tc.name, func(t *testing.T) {
				interactions := types.NewInteractionIDSet()
				interactionIndex := make(map[types.InteractionID]int)
				for idx, i := range tc.interactions {
					interactionID := parseInteraction(i)
					interactions.Set(interactionID)
					interactionIndex[interactionID] = idx
				}
				// run readNNF
				nnf := ReadDDNF(slog.Default(),
					filepath.Join("testdata", tc.name),
					fmt.Sprintf("%s.%s.nnf", tc.name, compiler),
					testIW,
				)

				// compute derivative
				nnf.InteractionIndex = interactionIndex
				derivative := nnf.EvaluateIntersection(interactions)

				// compare actual to golden
				tolerance := compare.Tolerance(.00001)
				if !tolerance.FloatEqualWithinTolerance(tc.expected, derivative) {
					t.Errorf("%s computed %f expected %f: %+v", compiler, derivative, tc.expected, nnf)
				}
			}) {
				break
			}
		}
	}
}
