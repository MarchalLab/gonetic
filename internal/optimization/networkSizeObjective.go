package optimization

import (
	"github.com/MarchalLab/gonetic/internal/ranking"
)

type networkSizeObjective struct {
	ranking.MaxObjective
}

func newNetworkSizeObjective() networkSizeObjective {
	return networkSizeObjective{
		MaxObjective: ranking.MaxObjective{},
	}
}

func (obj *networkSizeObjective) Compute(s subnetwork) float64 {
	return 1 / float64(s.subnetworkSize())
}
