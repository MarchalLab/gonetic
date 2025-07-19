package types

import "testing"

func TestProbabilityMap_AddMultiplicative(t *testing.T) {
	pm := NewProbabilityMap()
	interactionID := FromToTypeToID(1, 2, 0)
	baseProbability := 0.5
	pm.SetProbability(interactionID, baseProbability)
	additionalProbability := 0.2
	expectedProbability := baseProbability * additionalProbability

	pm.AddMultiplicative(interactionID, additionalProbability)
	probability := pm.GetProbability(interactionID)

	if probability != expectedProbability {
		t.Errorf("expected combined probability: %f, got: %f", expectedProbability, probability)
	}
}

func TestProbabilityMap_AddBayesian(t *testing.T) {
	pm := NewProbabilityMap()
	interactionID := FromToTypeToID(1, 2, 0)
	baseProbability := 0.5
	pm.SetProbability(interactionID, baseProbability)
	additionalProbability := 0.2
	expectedProbability := 1 - (1-baseProbability)*(1-additionalProbability)

	pm.AddBayesian(interactionID, additionalProbability)
	probability := pm.GetProbability(interactionID)

	if probability != expectedProbability {
		t.Errorf("expected combined probability: %f, got: %f", expectedProbability, probability)
	}
}

func TestProbabilityMap_AddMean(t *testing.T) {
	pm := NewProbabilityMap()
	interactionID := FromToTypeToID(1, 2, 0)
	baseProbability := 0.5
	pm.SetProbability(interactionID, baseProbability)
	additionalProbability := 0.2
	expectedProbability := (baseProbability + additionalProbability) / 2

	pm.AddMean(interactionID, additionalProbability)
	probability := pm.GetProbability(interactionID)

	if probability != expectedProbability {
		t.Errorf("expected combined probability: %f, got: %f", expectedProbability, probability)
	}
}
