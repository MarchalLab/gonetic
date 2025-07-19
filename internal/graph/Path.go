package graph

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

// Path struct represents a path in the network
type Path struct {
	probability  float64
	interactions []types.InteractionID
	genes        types.GeneSet
	Direction    types.PathDirection
	EndGene      types.GeneID
	Length       int
	isExtended   bool
	index        int
	EndCondition types.Condition
	fromScore    float64
	toScore      float64
}

func (p *Path) Genes() types.GeneSet {
	return p.genes
}

func (p *Path) IsExtended() bool {
	return p.isExtended
}

func (p *Path) Index() int {
	return p.index
}

func (p *Path) SetIndex(index int) {
	p.index = index
}

// Probability returns the probability of the path
func (p *Path) Probability() float64 {
	return p.probability * p.fromScore * p.toScore
}

// SetProbability sets the probability of the path
func (p *Path) SetProbability(probability float64) {
	p.probability = probability
}

// MultiplyProbability multiplies the probability of the path with the given value
func (p *Path) MultiplyProbability(value float64) {
	p.probability *= value
}

// Priority returns true if the path has a higher priority than the other path
func (p *Path) Priority(other *Path) bool {
	return p.probability > other.probability
}

// Interactions returns the interactions of the path
func (p *Path) Interactions() []types.InteractionID {
	return p.interactions
}

// AddInteraction adds an interaction to the path
func (p *Path) FirstInteraction() types.InteractionID {
	if len(p.interactions) == 0 {
		return types.InteractionID(0)
	}
	return p.interactions[0]

}

// LastInteraction returns the last interaction of the path
func (p *Path) LastInteraction() types.InteractionID {
	if len(p.interactions) == 0 {
		return types.InteractionID(0)
	}
	return p.interactions[len(p.interactions)-1]

}

// hasGene returns true if the path contains the given gene
func (p *Path) hasGene(gene types.GeneID) bool {
	_, has := p.genes[gene]
	return has
}

// TxtString returns a string representation of the path in the format of the output file
func (p *Path) TxtString(probabilityMap *types.ProbabilityMap, gim *types.GeneIDMap, from types.GeneID, condition types.Condition) string {
	steps := make([]string, 0, 2*p.Length+1)
	steps = append(steps, string(gim.GetNameFromID(from)))
	probabilities := make([]float64, 0, p.Length)
	interactionTypes := make([]types.InteractionTypeID, 0, p.Length)
	for _, interactionID := range p.Interactions() {
		dir := ""
		if interactionID.To() == from {
			from = interactionID.From()
			dir = "<-"
		} else if interactionID.From() == from {
			from = interactionID.To()
			dir = "->"
		} else {
			panic(fmt.Sprintf("Interaction does not contain previous gene %s in interactionID %s->%s",
				gim.GetNameFromID(from),
				gim.GetNameFromID(interactionID.From()),
				gim.GetNameFromID(interactionID.To()),
			))
		}
		steps = append(steps, dir)
		steps = append(steps, string(gim.GetNameFromID(from)))
		probabilities = append(probabilities, probabilityMap.GetProbability(interactionID))
		interactionTypes = append(interactionTypes, interactionID.Type())
	}
	if condition == p.EndCondition {
		return fmt.Sprintf("%s\t%s\t%s\t%s\t%+v\t%s\t%+v",
			condition,
			strconv.FormatFloat(p.Probability(), 'f', -1, 64),
			strings.Join(steps, ""),
			strconv.FormatFloat(p.fromScore, 'f', -1, 64),
			probabilities,
			strconv.FormatFloat(p.toScore, 'f', -1, 64),
			interactionTypes,
		)
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%+v\t%s\t%+v",
		condition,
		p.EndCondition,
		strconv.FormatFloat(p.Probability(), 'f', -1, 64),
		strings.Join(steps, ""),
		strconv.FormatFloat(p.fromScore, 'f', -1, 64),
		probabilities,
		strconv.FormatFloat(p.toScore, 'f', -1, 64),
		interactionTypes,
	)
}

func (p *Path) Clone() *Path {
	interactions := make([]types.InteractionID, len(p.interactions))
	copy(interactions, p.interactions)
	genes := make(types.GeneSet, len(p.genes))
	for g := range p.genes {
		genes[g] = struct{}{}
	}
	return &Path{
		probability:  p.probability,
		interactions: interactions,
		genes:        genes,
		Direction:    p.Direction,
		EndGene:      p.EndGene,
		Length:       p.Length,
		isExtended:   p.isExtended,
		index:        -1,
		EndCondition: p.EndCondition,
		fromScore:    p.fromScore,
		toScore:      p.toScore,
	}
}

// RootPath creates a new path with only the start gene
func RootPath(startGene types.GeneID, fromScore float64) *Path {
	return &Path{
		probability:  1.0,
		interactions: make([]types.InteractionID, 0),
		genes:        types.GeneSet{startGene: {}},
		Direction:    types.UndirectedPath,
		EndGene:      startGene,
		Length:       0,
		isExtended:   false,
		index:        -1,
		EndCondition: "",
		fromScore:    fromScore,
		toScore:      1,
	}
}

// ExtendPath creates a new path by extending the given path with the given interaction
func ExtendPath(parent *Path, interactionID types.InteractionID, probability float64, direction types.PathDirection) *Path {
	// determine which gene is the new end gene
	endGene := interactionID.To()
	if endGene == parent.EndGene {
		endGene = interactionID.From()
	} else if interactionID.From() != parent.EndGene {
		panic(fmt.Sprintf("Interaction %d->%d is not connected to end gene %d", interactionID.From(), interactionID.To(), parent.EndGene))
	}
	// clone the parent
	extended := parent.Clone()
	extended.isExtended = true
	extended.Length++
	extended.Direction = direction
	extended.probability = probability
	extended.EndCondition = ""
	// add the interaction to the path
	extended.interactions = append(extended.interactions, interactionID)
	// update the genes
	extended.genes[interactionID.From()] = struct{}{}
	extended.genes[interactionID.To()] = struct{}{}
	extended.EndGene = endGene

	return extended
}
