package types

import "github.com/MarchalLab/gonetic/internal/common/compare"

// ProbabilityMap is a map that stores the mapping between interactions and probabilities
type ProbabilityMap map[InteractionID]float64

// NewProbabilityMap creates a new ProbabilityMap
func NewProbabilityMap() *ProbabilityMap {
	return &ProbabilityMap{}
}

// SetProbability sets the probability of the given interaction
func (pm *ProbabilityMap) SetProbability(interaction InteractionID, probability float64) {
	(*pm)[interaction] = compare.Between(0, probability, 1)
}

// GetProbability returns the probability of the given interaction
func (pm *ProbabilityMap) GetProbability(interaction InteractionID) float64 {
	return (*pm)[interaction]
}

func (pm *ProbabilityMap) AddMultiplicative(interaction InteractionID, probability float64) {
	newProbability := probability * pm.GetProbability(interaction)
	pm.SetProbability(interaction, newProbability)
}

func (pm *ProbabilityMap) AddBayesian(interaction InteractionID, probability float64) {
	newProbability := 1 - (1-probability)*(1-pm.GetProbability(interaction))
	pm.SetProbability(interaction, newProbability)
}

func (pm *ProbabilityMap) AddMean(interaction InteractionID, probability float64) {
	newProbability := (probability + pm.GetProbability(interaction)) / 2
	pm.SetProbability(interaction, newProbability)
}

func (pm *ProbabilityMap) Has(id InteractionID) bool {
	_, ok := (*pm)[id]
	return ok
}

func (pm *ProbabilityMap) Delete(id InteractionID) {
	delete(*pm, id)
}
