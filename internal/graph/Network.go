package graph

import (
	"fmt"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

type Network struct {
	*types.InteractionStore
	probabilities *types.ProbabilityMap

	interactionTypes types.InteractionTypeSet
	scores           []float64
}

func (n *Network) Probabilities() *types.ProbabilityMap {
	return n.probabilities
}

func (n *Network) Edges() map[int]float64 {
	edges := make(map[int]float64)
	for id, prob := range *n.Probabilities() {
		edges[int(id)] = prob
	}
	return edges
}

func (n *Network) Interactions() types.InteractionIDSet {
	interactions := make(types.InteractionIDSet)
	for id := range *n.Probabilities() {
		interactions.Set(id)
	}
	return interactions
}

func NewNetwork(
	store *types.InteractionStore,
	probabilityMap *types.ProbabilityMap,
	interactionTypes types.InteractionTypeSet,
	scores []float64,
) *Network {
	network := Network{
		InteractionStore: store,
		probabilities:    probabilityMap,
		interactionTypes: interactionTypes,
		scores:           scores,
	}
	return &network
}

func (n *Network) Scores() []float64 {
	return n.scores
}

func (n *Network) InteractionCount() int {
	return len(*n.Probabilities())
}

func (n *Network) Genes() types.GeneSet {
	genes := make(types.GeneSet)
	for nodes := range n.Outgoing() {
		genes[nodes] = struct{}{}
	}
	return genes
}

func (n *Network) Nodes() map[types.GeneID]struct{} {
	return n.Genes()
}

func (n *Network) String() string {
	s := ""
	for _, it := range n.interactionTypes {
		s += it.Name + "\n"
	}
	for id := range *n.Probabilities() {
		s += fmt.Sprintf("%d\t%d\n", id.From(), id.To())
	}
	return s
}

func (n *Network) OutDegree(gene types.GeneID) int {
	interactions, ok := n.Outgoing()[gene]
	if ok {
		return len(interactions)
	}
	return 0
}

func (n *Network) InDegree(gene types.GeneID) int {
	interactions, ok := n.Incoming()[gene]
	if ok {
		return len(interactions)
	}
	return 0
}

func (n *Network) Degree(gene types.GeneID) int {
	return n.InDegree(gene) + n.OutDegree(gene)
}

func (n *Network) IncomingInteractions(gene types.GeneID) types.InteractionIDSet {
	interactions, ok := n.Incoming()[gene]
	if ok {
		return interactions
	}
	return types.NewInteractionIDSet()
}

func (n *Network) OutgoingInteractions(gene types.GeneID) types.InteractionIDSet {
	interactions, ok := n.Outgoing()[gene]
	if ok {
		return interactions
	}
	return types.NewInteractionIDSet()
}

func (n *Network) InteractionTypeCount() int {
	return len(n.interactionTypes)
}

func (n *Network) InteractionTypes() types.InteractionTypeSet {
	return n.interactionTypes
}

func (n *Network) StringArray() [][]string {
	interactionTypes := make([]string, 0, n.InteractionTypeCount())
	for interactionType := range n.InteractionTypes() {
		interactionTypes = append(interactionTypes, fmt.Sprintf("# %s", interactionType))
	}
	interactions := make([]string, 0, n.InteractionCount())
	for interactionID, probability := range *n.Probabilities() {
		interactions = append(interactions, interactionID.StringWithProbability(probability))
	}
	return [][]string{interactionTypes, interactions}
}
