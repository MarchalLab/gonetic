package types

import (
	"fmt"
	"strconv"
	"strings"
)

// CompactPath is a self-contained representation of a path in the network
type CompactPath struct {
	ID               int
	Probability      float64
	InteractionOrder []InteractionID
	ProbabilityOrder []float64
	Genes            GeneSet
	Direction        PathDirection
	StartGene        GeneID
	StartCondition   Condition
	EndGene          GeneID
	EndCondition     Condition
	FromScore        float64
	ToScore          float64
}

func EmptyCompactPath() *CompactPath {
	return NewCompactPath(
		0,
		0,
		nil,
		nil,
		nil,
		UndirectedPath,
		0,
		"",
		0,
		"",
		0,
		0,
	)
}

func NewCompactPath(
	pathID int,
	pathProbability float64,
	interactionOrder []InteractionID,
	edgeScores []float64,
	genes GeneSet,
	direction PathDirection,
	startGene GeneID,
	startSample Condition,
	endGene GeneID,
	endSample Condition,
	fromScore float64,
	toScore float64,
) *CompactPath {
	return &CompactPath{
		ID:               pathID,
		Probability:      pathProbability,
		InteractionOrder: interactionOrder,
		ProbabilityOrder: edgeScores,
		Genes:            genes,
		Direction:        direction,
		StartGene:        startGene,
		StartCondition:   startSample,
		EndGene:          endGene,
		EndCondition:     endSample,
		FromScore:        fromScore,
		ToScore:          toScore,
	}
}

func NewCompactPathWithInteractions(interactionOrder []InteractionID) *CompactPath {
	path := EmptyCompactPath()
	path.InteractionOrder = interactionOrder
	path.ProbabilityOrder = make([]float64, len(interactionOrder))
	return path
}

func (p *CompactPath) InteractionSet() InteractionIDSet {
	interactionSet := NewInteractionIDSet()
	for _, interactionID := range p.InteractionOrder {
		interactionSet.Set(interactionID)
	}
	return interactionSet
}

func (p *CompactPath) Length() int {
	return len(p.InteractionOrder)
}

// TxtString returns a string representation of the path in the format of the output file
func (p *CompactPath) TxtString(gim *GeneIDMap, from GeneID, condition Condition) string {
	steps := make([]string, 0, 2*p.Length()+1)
	steps = append(steps, string(gim.GetNameFromID(from)))
	interactionTypes := make([]InteractionTypeID, 0, p.Length())
	for _, interactionID := range p.InteractionOrder {
		dir := ""
		if interactionID.To() == from {
			from = interactionID.From()
			dir = "<-"
		} else if interactionID.From() == from {
			from = interactionID.To()
			dir = "->"
		} else {
			panic(fmt.Sprintf("Interaction does not contain previous gene %s in interaction %d->%d, %+v",
				gim.GetNameFromID(from),
				interactionID.From(),
				interactionID.To(),
				p,
			))
		}
		interactionTypes = append(interactionTypes, interactionID.Type())
		steps = append(steps, dir)
		steps = append(steps, string(gim.GetNameFromID(from)))
	}
	if condition == p.EndCondition || len(p.EndCondition) == 0 {
		return fmt.Sprintf("%s\t%s\t%s\t%s\t%+v\t%s\t%+v",
			condition,
			strconv.FormatFloat(p.Probability, 'f', -1, 64),
			strings.Join(steps, ""),
			strconv.FormatFloat(p.FromScore, 'f', -1, 64),
			p.ProbabilityOrder,
			strconv.FormatFloat(p.ToScore, 'f', -1, 64),
			interactionTypes,
		)

	}
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%+v\t%s\t%+v",
		condition,
		p.EndCondition,
		strconv.FormatFloat(p.Probability, 'f', -1, 64),
		strings.Join(steps, ""),
		strconv.FormatFloat(p.FromScore, 'f', -1, 64),
		p.ProbabilityOrder,
		strconv.FormatFloat(p.ToScore, 'f', -1, 64),
		interactionTypes,
	)
}

// Less compares two paths based on their edge IDs
func (p *CompactPath) Less(other *CompactPath) bool {
	if p.Length() != other.Length() {
		return p.Length() < other.Length()
	}
	for i := 0; i < p.Length(); i++ {
		if p.InteractionOrder[i] != other.InteractionOrder[i] {
			return p.InteractionOrder[i] < other.InteractionOrder[i]
		}
	}
	return false
}

type CompactPathList []*CompactPath

// Len returns the length of the path list
func (p CompactPathList) Len() int {
	return len(p)
}

// Less compares two paths based on their edge IDs
func (p CompactPathList) Less(i, j int) bool {
	return p[i].Less(p[j])
}

// Swap swaps two paths in the list
func (p CompactPathList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
