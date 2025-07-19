package editor

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
)

func RemoveLowScoringEdges(network *graph.Network, cutoff float64) *graph.Network {
	interactionTypes := make(types.InteractionTypeSet)
	probabilities := types.NewProbabilityMap()
	for id, p := range *network.Probabilities() {
		if p >= cutoff {
			probabilities.SetProbability(id, p)
		}
	}
	return graph.NewNetwork(network.InteractionStore, probabilities, interactionTypes, nil)
}
