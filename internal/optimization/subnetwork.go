package optimization

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/normalform"
	"github.com/MarchalLab/gonetic/internal/ranking"
)

// subnetwork is an interface for path-based subnetworks
// It extends the Solution interface from the ranking package
// It is aimed at effective crossover and mutation operators for NSGA2
type subnetwork interface {
	ranking.Solution

	// the following methods are not defined in genericSubnetwork
	expansion()
	reduction()

	// the following methods are called in other methods of genericSubnetwork
	// be careful when extending genericSubnetwork
	subnetworkSize() int
	intersect(nnf normalform.NNF) types.InteractionIDSet

	// the following methods are defined but not called in genericSubnetwork
	NonDominationLevel() int
	setNonDominationLevel(level int)
	CrowdingDistance() float64
	setCrowdingDistance(dist float64)
	SelectedPaths() map[PathID]types.InteractionIDSet
	Interactions() types.InteractionIDSet
	interactionCount() int
	addSelectedPath(pathID PathID, interactionSet types.InteractionIDSet)
	addInteraction(interaction types.InteractionID)
	String() string
	ParseString(string)
	geneCount() int
}
