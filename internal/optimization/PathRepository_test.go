package optimization

import (
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func TestJoinPathRepositories(t *testing.T) {
	// Create some mock interactions
	interaction1 := types.FromToTypeToID(1, 2, 1)
	interaction2 := types.FromToTypeToID(2, 3, 1)
	interaction3 := types.FromToTypeToID(3, 4, 1)

	// Create InteractionSets
	set1 := types.NewInteractionIDSet()
	set1.Set(interaction1)
	set1.Set(interaction2)
	order1 := make([]types.InteractionID, 0)
	order1 = append(order1, interaction1)
	order1 = append(order1, interaction2)

	set2 := types.NewInteractionIDSet()
	set2.Set(interaction2)
	set2.Set(interaction3)
	order2 := make([]types.InteractionID, 0)
	order2 = append(order2, interaction2)
	order2 = append(order2, interaction3)

	// Create PathRepositories
	repo1 := PathRepository{
		pathType:           "type1",
		NumberOfPaths:      1,
		pathInteractionSet: map[int]types.CompactPathList{0: {types.NewCompactPathWithInteractions(order1)}},
	}

	repo2 := PathRepository{
		pathType:           "type2",
		NumberOfPaths:      1,
		pathInteractionSet: map[int]types.CompactPathList{0: {types.NewCompactPathWithInteractions(order2)}},
	}

	// Join the repositories
	joinedRepo := NewPathRepositories()
	joinedRepo.Add(repo1)
	joinedRepo.Add(repo2)

	// Check the number of paths
	expectedNumberOfPaths := repo1.NumberOfPaths + repo2.NumberOfPaths
	if joinedRepo.NumberOfPaths() != expectedNumberOfPaths {
		t.Errorf("expected %d paths, got %d", expectedNumberOfPaths, joinedRepo.NumberOfPaths())
	}

	// Check the pathInteractionSet of the joined repository
	expectedPathInteractionSet := map[PathID]types.InteractionIDSet{
		NewPathID(0, 0, 0): set1,
		NewPathID(1, 0, 0): set2,
	}

	for idx, expectedSet := range expectedPathInteractionSet {
		if !joinedRepo.pathInteractionSetFromId(idx).Equals(expectedSet) {
			t.Errorf("expected pathInteractionSet[%d] to be %v, got %v", idx, expectedSet, joinedRepo.pathInteractionSetFromId(idx))
		}
	}
}
