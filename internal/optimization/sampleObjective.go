package optimization

import (
	"math"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/normalform"
	"github.com/MarchalLab/gonetic/internal/ranking"
)

type sampleObjective struct {
	ranking.MaxObjective
	dDNNFs        [][]*normalform.NNF
	objectiveType string
}

func newSampleObjective(
	dDNNFs [][]*normalform.NNF,
	objectiveType string,
) sampleObjective {
	return sampleObjective{
		MaxObjective:  ranking.MaxObjective{},
		dDNNFs:        dDNNFs,
		objectiveType: objectiveType,
	}
}

// Compute computes the sample count score.
// This is the sum over all path types of the number of samples that are explained by paths of that type.
func (obj sampleObjective) Compute(sub subnetwork) float64 {
	if sub.subnetworkSize() == 0 {
		return math.Inf(-1)
	}

	score := 0.0

	for _, dDNNFs := range obj.dDNNFs {
		// determine the number of samples that are explained by this path type
		samples := make(map[types.Condition]int)
		for _, dDNNF := range dDNNFs {
			if _, ok := samples[dDNNF.Condition()]; !ok {
				// initialize the sample count
				samples[dDNNF.Condition()] = 0
			}
			// add 1 to the sample count if the sample is explained by the dDNNF on this subnetwork
			intersection := sub.intersect(*dDNNF)
			evaluation := dDNNF.EvaluateIntersection(intersection)
			if evaluation > 0 {
				samples[dDNNF.Condition()] += 1
			}
		}
		switch obj.objectiveType {
		case "entropy":
			score += obj.computeEntropyScore(samples)
		case "effective":
			score += obj.computeEffectiveSampleScore(samples)
		default:
			panic("unknown objective type")
		}
	}

	// Final score: summed scores per path type
	return score
}

func (obj sampleObjective) computeEntropyScore(samples map[types.Condition]int) float64 {
	total := 0
	for _, count := range samples {
		total += count
	}

	if total == 0 {
		return 0.0 // no information, max penalty
	}

	var entropy float64
	nonzero := 0
	for _, count := range samples {
		if count == 0 {
			continue
		}
		p := float64(count) / float64(total)
		entropy -= p * math.Log2(p)
		nonzero++
	}

	if nonzero <= 1 {
		return 0.0 // either all zero or concentrated in one → minimal similarity
	}

	maxEntropy := math.Log2(float64(len(samples)))
	score := entropy / maxEntropy
	return score
}

func (obj sampleObjective) computeEffectiveSampleScore(samples map[types.Condition]int) float64 {
	total := 0
	for _, count := range samples {
		total += count
	}

	if total == 0 {
		return 0.0
	}

	var sumSquares float64
	for _, count := range samples {
		if count == 0 {
			continue
		}
		p := float64(count) / float64(total)
		sumSquares += p * p
	}

	if sumSquares == 0 {
		return 0.0
	}

	effSamples := 1.0 / sumSquares
	maxSamples := float64(len(samples))
	return effSamples / maxSamples // normalize to [0, 1]
}
