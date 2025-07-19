package editor

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
)

/**
 * Class which takes a network weighting method which is based on the nodes of the network (the genes).
 * @param networkWeight Method stating how the weight for a specific string should be calculated
 * @param addition Method stating how probabilities on edges should be integrated when more than one probability should be attributed to the edge.
 * @return The weighted network
 */

type geneProbability func(types.GeneID) float64

type additionType func(types.InteractionID, float64, geneProbability) float64

func NetworkWeighting(networkWeight geneProbability, addition additionType, network *graph.Network) *graph.Network {
	probabilities := types.NewProbabilityMap()
	for id := range *network.Probabilities() {
		originalProbability := network.Probabilities().GetProbability(id)
		adjustedProbability := addition(id, originalProbability, networkWeight)
		probabilities.SetProbability(id, adjustedProbability)
	}
	return graph.NewNetwork(network.InteractionStore, probabilities, network.InteractionTypes(), nil)
}
