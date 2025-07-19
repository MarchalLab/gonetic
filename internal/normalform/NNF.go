package normalform

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

/**
 * The NNF class represents an NNF tree. The other classes defined in this file represent different possible nodes in the NNF tree.
 *
 * @param nodes The sequence of nodes in the NNF tree
 * @param rootNode The root node of the NNF tree
 * @param utility
 * @see The paper [[http://ai2-s2-pdfs.s3.amazonaws.com/115a/59f493bf9eb72cdea988a16e948d208c2d22.pdf]] about the NNF language
 */

type NNF struct {
	*slog.Logger
	nodes            []nnfNode
	pathNode         nnfNode
	startGene        string // TODO
	condition        types.Condition
	values           types.InteractionIDSet
	parentMap        [][]int
	InteractionIndex map[types.InteractionID]int
	fromScore        float64
	toScore          float64 // TODO this needs to be 1 float per path
}

func (nnf *NNF) Nodes() []nnfNode {
	return nnf.nodes
}

func (nnf *NNF) Condition() types.Condition {
	return nnf.condition
}

func (nnf *NNF) Name() string {
	return fmt.Sprintf("%s;%s", nnf.startGene, nnf.condition)
}

func (nnf *NNF) Values() types.InteractionIDSet {
	return nnf.values
}

func parseInteraction(line string) types.InteractionID {
	split := strings.Split(line, ";")
	// parse genes
	fromGene, _ := strconv.Atoi(split[0])
	toGene, _ := strconv.Atoi(split[1])
	typ, _ := strconv.Atoi(split[2])
	// make new interaction
	return types.FromToTypeToID(
		types.GeneID(fromGene),
		types.GeneID(toGene),
		types.InteractionTypeID(typ),
	)
}

func newNNF(logger *slog.Logger, nodes []nnfNode) NNF {
	var pathNode nnfNode
	for _, node := range nodes {
		if !types.IsPathStringFormat(node.name) {
			continue
		}
		if node.IsNegated() {
			continue
		}
		pathNode = node
		break
	}
	values := types.NewInteractionIDSet()
	for _, node := range nodes {
		if !node.hasUnderlyingValue() {
			// node is not an interaction
			continue
		}
		interaction := parseInteraction(node.name)
		values.Set(interaction)
	}
	parentMap := make([][]int, len(nodes))
	for _, node := range nodes {
		parentMap[node.id] = make([]int, 0)
	}
	for _, node := range nodes {
		for _, childId := range node.children {
			parentMap[childId] = append(parentMap[childId], node.id)
		}
	}
	split := strings.Split(pathNode.name, ";")
	return NNF{
		Logger:    logger,
		nodes:     nodes,
		pathNode:  pathNode,
		startGene: split[0],
		condition: types.Condition(split[1]),
		values:    values,
		parentMap: parentMap,
		fromScore: 1, // TODO
		toScore:   1, // TODO
	}
}

func (nnf *NNF) calculateValues(selectedNodeNames types.InteractionIDSet) []float64 {
	valueMap := make([]float64, len(nnf.nodes))
	for i, node := range nnf.nodes {
		// leaf node: get probability of node
		if node.IsLeafNode() {
			valueMap[i] = node.computeProbability(selectedNodeNames)
			continue
		}
		// or node: add child values
		if node.IsOrNode() {
			var sum = 0.0
			for _, child := range node.children {
				sum += valueMap[child]
			}
			valueMap[i] = sum
			continue
		}
		// and node: multiply child values
		var product = 1.0
		for _, child := range node.children {
			product *= valueMap[child]
			if product == 0 {
				break
			}
		}
		valueMap[i] = product
	}
	return valueMap
}

// EvaluateIntersection determines the probability of at least one valid path in the nnf that is in the intersection
// intersection is the intersection of a subnetwork and `nnf`
func (nnf *NNF) EvaluateIntersection(intersection types.InteractionIDSet) float64 {
	if intersection.Empty() {
		return 0
	}
	// compute score
	values := nnf.calculateValues(intersection)
	score := values[len(values)-1]
	if score > 1 && score < 1.00001 {
		// sometimes score is slightly off due to rounding errors in the computation
		// if the value is slightly larger than 1, then we assume this is a rounding error, so set it to 1
		score = 1
	}
	if score > 1 {
		nnf.Error("unexpected score",
			"score", score,
			"nnf.name", nnf.Name(),
			"values", values,
			"intersection", intersection,
		)
	}
	return score
}
