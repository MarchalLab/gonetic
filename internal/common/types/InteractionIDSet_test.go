package types

import (
	"testing"
)

func createTestInteractionIDSet() (InteractionIDSet, []InteractionID) {
	set := NewInteractionIDSet()
	interactions := make([]InteractionID, 0, 2)

	interaction1 := FromToToID(1, 2)
	set.Set(interaction1)
	interactions = append(interactions, interaction1)

	interaction2 := FromToToID(3, 4)
	set.Set(interaction2)
	interactions = append(interactions, interaction2)

	return set, interactions
}

func TestNewInteractionIDSet(t *testing.T) {
	set := NewInteractionIDSet()
	if !set.Empty() {
		t.Errorf("expected new set to be empty")
	}
}

func TestInteractionIDSet_AddAndGet(t *testing.T) {
	set := NewInteractionIDSet()

	interaction := FromToToID(1, 2)
	set.Set(interaction)

	if !set.Has(interaction) {
		t.Errorf("expected set to contain interaction with ID: %d", interaction)
	}
}

func TestInteractionIDSet_Delete(t *testing.T) {
	set, interactions := createTestInteractionIDSet()

	interaction := interactions[0]
	set.Set(interaction)
	set.Delete(interaction)

	if set.Has(interaction) {
		t.Errorf("expected interaction with ID: %d to be deleted", interaction)
	}
}

func TestInteractionIDSet_Empty(t *testing.T) {
	_, interactions := createTestInteractionIDSet()
	set := NewInteractionIDSet()

	if !set.Empty() {
		t.Errorf("expected new set to be empty")
	}

	interaction := interactions[0]
	set.Set(interaction)

	if set.Empty() {
		t.Errorf("expected set to not be empty after adding an interaction")
	}
}

func TestInteractionIDSet_ToMap(t *testing.T) {
	set, interactions := createTestInteractionIDSet()

	result := set.Copy()
	if len(result) != len(interactions) {
		t.Errorf("expected map size: %d, got: %d", len(interactions), len(result))
	}

	for _, interaction := range interactions {
		if _, exists := result[interaction]; !exists {
			t.Errorf("expected map to contain interaction with ID: %d", interaction)
		}
	}
}

func TestInteractionIDSet_Equals(t *testing.T) {
	_, interactions := createTestInteractionIDSet()
	set1 := NewInteractionIDSet()
	set2 := NewInteractionIDSet()

	interaction := interactions[0]
	set1.Set(interaction)
	set2.Set(interaction)

	if !set1.Equals(set2) {
		t.Errorf("expected sets to be equal")
	}

	interaction2 := interactions[1]
	set2.Set(interaction2)

	if set1.Equals(set2) {
		t.Errorf("expected sets to be unequal")
	}
}

func TestInteractionIDSet_NetworkNodesSize(t *testing.T) {
	set, _ := createTestInteractionIDSet()

	expectedNodes := 4 // Nodes: 1, 2, 3, 4
	if set.NetworkNodesSize() != expectedNodes {
		t.Errorf("expected network nodes size: %d, got: %d", expectedNodes, set.NetworkNodesSize())
	}
}
