package normalform

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
)

type leafType int

const (
	noLeaf leafType = iota
	leaf
	aux
	value
	negative
	positive
	erroneous
	core
)

type nnfNode struct {
	id              int
	name            string
	probability     float64
	children        []int
	underlyingValue bool
	or              bool
	leafType        leafType
	negated         bool
}

func (node nnfNode) Name() string {
	return node.name
}

func (node nnfNode) computeProbability(selectedNodeNames types.InteractionIDSet) float64 {
	switch node.leafType {
	case noLeaf:
		return 0.0
	case leaf | value:
		return 0.0
	case aux:
		// always 1.0, also for negated // TODO why
		return 1.0
	case erroneous:
		// always 0.0, also for negated
		return 0.0
	case core:
		// always 1.0, negated is classified as erroneous
		return 1.0
	case negative:
		interactionID := parseInteraction(node.name)
		if selectedNodeNames.Has(interactionID) {
			return 1 - node.probability
		}
		return 1
	case positive:
		interactionID := parseInteraction(node.name)
		if selectedNodeNames.Has(interactionID) {
			return node.probability
		}
		return 0
	}
	return 0.0
}

func (node nnfNode) childCount() int {
	return len(node.children)
}

func (node nnfNode) Children() []int {
	return node.children
}

func (node nnfNode) ID() int {
	return node.id
}

func (node nnfNode) hasUnderlyingValue() bool {
	return node.underlyingValue
}

func (node nnfNode) IsOrNode() bool {
	return node.or
}

func (node nnfNode) IsLeafNode() bool {
	return node.leafType != noLeaf
}

func (node nnfNode) IsNegated() bool {
	return node.negated
}

func newNNFNode(id int, name string, children []int) nnfNode {
	return nnfNode{
		id:              id,
		name:            name,
		probability:     0.0,
		children:        children,
		underlyingValue: false,
		or:              false,
		leafType:        noLeaf,
		negated:         false,
	}
}

func newORNNFNode(id int, children []int) nnfNode {
	node := newNNFNode(id, "", children)
	node.or = true
	return node
}

func newANDNNFNode(id int, children []int) nnfNode {
	node := newNNFNode(id, "", children)
	return node
}

func newLeafNNFNode(id int, name string) nnfNode {
	node := newNNFNode(id, name, make([]int, 0))
	node.leafType = leaf
	return node
}

func newAuxLeafNNFNode(id int, name string, negated bool) nnfNode {
	node := newLeafNNFNode(id, name)
	node.negated = negated
	node.leafType = aux
	return node
}

func newValueLeafNNFNode(id int, name string) nnfNode {
	node := newLeafNNFNode(id, name)
	node.underlyingValue = true
	node.leafType = value
	return node
}

func newNegativeLeafNNFNode(id int, name string, probability float64) nnfNode {
	node := newValueLeafNNFNode(id, name)
	node.negated = true
	node.leafType = negative
	node.probability = probability
	return node
}

func newPositiveLeafNNFNode(id int, name string, probability float64) nnfNode {
	node := newValueLeafNNFNode(id, name)
	node.leafType = positive
	node.probability = probability
	return node
}

type NodeValue string
