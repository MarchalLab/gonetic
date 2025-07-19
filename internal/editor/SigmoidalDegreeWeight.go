package editor

import (
	"math"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/powerlaws"
)

// Calculate the sigmoidal function for the degree of each gene. For the reweighting, this value is passed to the edges.
func calculateSigmoidal(x, inflectionValue, dampingFactor float64) float64 {
	return 1 / (1 + math.Exp((x-inflectionValue)/dampingFactor))
}

type SigmoidalDegreeWeight struct {
	scoreMap map[types.GeneID]float64
}

func (s SigmoidalDegreeWeight) Score(gene types.GeneID) float64 {
	return s.scoreMap[gene]
}

func NewSigmoidalDegreeWeight(network *graph.Network, cutoff float64) SigmoidalDegreeWeight {
	scoreMap := make(map[types.GeneID]float64)
	// Power law
	collection := make([]int64, 0, len(network.Genes()))
	for gene := range network.Genes() {
		outdegree := int64(network.OutDegree(gene))
		if outdegree == 0 {
			continue
		}
		collection = append(collection, outdegree)
	}
	counts := make(map[int64]int)
	geneSum := 0
	interactionSum := int64(0)
	for _, entry := range collection {
		if _, ok := counts[entry]; !ok {
			counts[entry] = 0
		}
		counts[entry]++
		geneSum++
		interactionSum += entry
	}
	inflectionValue := float64(powerlaws.DiscretePowerLawFit(collection).CdfInv(0.1))
	dampingFactor := inflectionValue / 3
	for gene := range network.Genes() {
		degree := float64(network.OutDegree(gene))
		scoreMap[gene] = math.Max(cutoff, calculateSigmoidal(degree, inflectionValue, dampingFactor))
	}
	return SigmoidalDegreeWeight{scoreMap}
}
