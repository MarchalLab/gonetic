package optimization

import (
	"math"

	"github.com/MarchalLab/gonetic/internal/normalform"
	"github.com/MarchalLab/gonetic/internal/ranking"
)

type dDNNFObjective struct {
	ranking.MaxObjective
	dDNNFs []*normalform.NNF
}

func newDDNNFObjective(dDNNFs []*normalform.NNF) dDNNFObjective {
	return dDNNFObjective{
		MaxObjective: ranking.MaxObjective{},
		dDNNFs:       dDNNFs,
	}
}

// Compute computes the score of a subnetwork, divided by the size of the subnetwork
func (obj dDNNFObjective) Compute(sub subnetwork) float64 {
	if sub.subnetworkSize() == 0 {
		return math.Inf(-1)
	}
	score := 0.0
	for _, dDNNF := range obj.dDNNFs {
		// compute score, this might update the dDNNF cache
		intersection := sub.intersect(*dDNNF)
		evaluation := dDNNF.EvaluateIntersection(intersection)
		score += evaluation
	}
	return score
}
