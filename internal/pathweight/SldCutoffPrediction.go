package pathweight

import (
	"math"
	"sort"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

type network interface {
	Probabilities() *types.ProbabilityMap
}

// SldCutoffPrediction predicts the cutoff if it is not provided
// The prediction is made based on the interaction probabilities and taken as the probability of the interaction at the 1% best point which is first squared and then divided by 10
func SldCutoffPrediction(network network) float64 {
	edges := *network.Probabilities()
	if len(edges) == 0 {
		return 0.0
	}
	interactionProbabilities := make([]float64, len(edges))
	for _, probability := range edges {
		interactionProbabilities = append(interactionProbabilities, probability)
	}
	sort.Float64s(interactionProbabilities)
	cutOff := int(math.Round(float64(len(interactionProbabilities)) * 0.99))
	OnePercentBestPoint := interactionProbabilities[cutOff-1]
	return math.Pow(OnePercentBestPoint, 2) / 10.0
}
