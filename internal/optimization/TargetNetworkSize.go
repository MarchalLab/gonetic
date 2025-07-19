package optimization

import (
	"log/slog"
	"math/rand"
	"sort"
)

type NetworkSizeOptimizer struct {
	*slog.Logger
	ObjectiveTypes []objectiveType
	CurrentMin     int
	CurrentMax     int
	FocusMin       int
	FocusMax       int
	TrueMax        int
	ScoreExponent  float64
	FocusFraction  float64
}

// NewNetworkSizeOptimizer initializes a NetworkSizeOptimizer object
func NewNetworkSizeOptimizer(
	logger *slog.Logger,
	objectiveTypes []objectiveType,
	initialMin, initialMax, trueMax int,
	focusFraction float64,
) *NetworkSizeOptimizer {
	return &NetworkSizeOptimizer{
		Logger:         logger,
		ObjectiveTypes: objectiveTypes,
		CurrentMin:     initialMin,
		CurrentMax:     initialMax,
		FocusMin:       initialMin,
		FocusMax:       initialMax,
		TrueMax:        trueMax,
		ScoreExponent:  0,
		FocusFraction:  focusFraction,
	}
}

// SampleNetworkSize samples from the global range [CurrentMin, CurrentMax] with focus range [FocusMin, FocusMax]
func (nso *NetworkSizeOptimizer) SampleNetworkSize() int {
	if rand.Float64() < nso.FocusFraction {
		// Sample from focus range
		return rand.Intn(nso.FocusMax-nso.FocusMin+1) + nso.FocusMin
	}
	// Sample from global range
	return rand.Intn(nso.CurrentMax-nso.CurrentMin+1) + nso.CurrentMin
}

// ComputeTargetNetworkSize updates min and max network size based on population
func (nso *NetworkSizeOptimizer) ComputeTargetNetworkSize(subnetworks []subnetwork) float64 {
	sizes := nso.collectRelevantSizes(subnetworks)
	if len(sizes) == 0 || nso.ScoreExponent == 0 {
		return nso.ScoreExponent
	}

	minSize, maxSize := computeSizeBounds(sizes)
	nso.FocusMin = smoothMinSize(nso.FocusMin, minSize)

	if nso.FocusMin < nso.CurrentMin {
		nso.CurrentMin = nso.FocusMin
	}

	nso.FocusMax = smoothMaxSize(nso.FocusMax, maxSize, nso.FocusMin)

	if nso.FocusMax > nso.TrueMax {
		nso.ScoreExponent = adjustScoreExponent(nso.ScoreExponent)
		nso.FocusMax = nso.TrueMax
		nso.FocusMin = 1
	}

	if nso.FocusMax > nso.CurrentMax {
		nso.CurrentMax = nso.FocusMax
	}

	nso.Info("ComputeTargetNetworkSize",
		"CurrentMin", nso.CurrentMin,
		"FocusMin", nso.FocusMin,
		"FocusMax", nso.FocusMax,
		"CurrentMax", nso.CurrentMax,
		"ScoreExponent", nso.ScoreExponent,
	)
	return nso.ScoreExponent
}

// collectRelevantSizes extracts sizes from top networks based on objectives
func (nso *NetworkSizeOptimizer) collectRelevantSizes(subnetworks []subnetwork) []int {
	var sizes []int
	for scoreIdx, objectiveType := range nso.ObjectiveTypes {
		if objectiveType != dDNNFObjectiveType {
			continue // Skip non-dDNNF objectives
		}
		indices := sortedIndices(subnetworks, scoreIdx)
		nTop := max(1, len(indices)/100)
		for i := 0; i < nTop; i++ {
			sizes = append(sizes, subnetworks[indices[i]].subnetworkSize())
		}
	}
	sort.Ints(sizes)
	return sizes
}

// computeSizeBounds determines min and max network sizes from sorted list
func computeSizeBounds(sizes []int) (int, int) {
	minSize := percentile(sizes, 0.10)
	maxSize := percentile(sizes, 0.90)
	return minSize, maxSize
}

// smoothMinSize adjusts min size to prevent sudden changes
func smoothMinSize(currentMin, computedMin int) int {
	newMin := limitedChange(currentMin, int(float64(computedMin)*0.9)-10)
	if newMin < 1 {
		newMin = 1
	}
	return newMin
}

// smoothMaxSize adjusts max size to prevent sudden changes
func smoothMaxSize(currentMax, computedMax, newMin int) int {
	newMax := limitedChange(currentMax, int(float64(computedMax)*1.1)+10)
	if newMax <= newMin {
		newMax = newMin + 1
	}
	return newMax
}

// adjustScoreExponent modifies the gene exponent to control network growth
func adjustScoreExponent(scoreExponent float64) float64 {
	return scoreExponent + 0.005
}

// sortedIndices sorts networks based on scores for a given objective index
func sortedIndices(subnetworks []subnetwork, scoreIdx int) []int {
	scores := make([][]float64, len(subnetworks))
	for i, subnetwork := range subnetworks {
		scores[i] = subnetwork.Scores()
	}
	indices := make([]int, len(scores))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return scores[a][scoreIdx] < scores[b][scoreIdx]
	})
	return indices
}

// percentile computes the p-th percentile value from a sorted slice
func percentile(data []int, p float64) int {
	index := int(float64(len(data)-1) * p)
	return data[index]
}
