package graph_test

import (
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
)

func TestRootPath(t *testing.T) {
	// Create a gene
	gene := types.GeneID(1)

	// Call the RootPath function
	path := graph.RootPath(gene, 1)

	// Check that the properties of the path are correctly set
	if path.Probability() != 1.0 {
		t.Errorf("Expected probability to be 1.0, but got %f", path.Probability())
	}
	if path.EndGene != gene {
		t.Errorf("Expected end gene to be the same as the start gene, but got %v", path.EndGene)
	}
	if path.Direction != types.UndirectedPath {
		t.Errorf("Expected direction to be UndirectedPath, but got %v", path.Direction)
	}
	if path.IsExtended() != false {
		t.Errorf("Expected isExtended to be false, but got %v", path.IsExtended())
	}

	// Check that the genes map contains the start gene
	_, ok := path.Genes()[gene]
	if !ok {
		t.Error("Expected genes map to contain the start gene")
	}
}

func TestExtendPath(t *testing.T) {
	// Create a gene and an interaction
	startGene := types.GeneID(1)
	endGene := types.GeneID(2)
	typ := types.InteractionTypeID(3)
	interactionID := types.FromToTypeToID(startGene, endGene, typ)

	// Create a root path and extend it
	rootPath := graph.RootPath(startGene, 1)
	extendedPath := graph.ExtendPath(rootPath, interactionID, 0.5, types.UndirectedPath)

	// Check that the properties of the extended path are correctly set
	if extendedPath.Probability() != 0.5 {
		t.Errorf("Expected probability to be 0.5, but got %f", extendedPath.Probability())
	}
	if extendedPath.EndGene != endGene {
		t.Errorf("Expected end gene to be the same as the end gene of the interaction, but got %v", extendedPath.EndGene)
	}
	if extendedPath.Direction != types.UndirectedPath {
		t.Errorf("Expected direction to be UndirectedPath, but got %v", extendedPath.Direction)
	}
	if extendedPath.IsExtended() != true {
		t.Errorf("Expected isExtended to be true, but got %v", extendedPath.IsExtended())
	}
	if extendedPath.Length != rootPath.Length+1 {
		t.Errorf("Expected length to be one more than the length of the parent path, but got %v", extendedPath.Length)
	}

	// Check that the genes map contains the start gene and the end gene
	_, ok := extendedPath.Genes()[startGene]
	if !ok {
		t.Error("Expected genes map to contain the start gene")
	}
	_, ok = extendedPath.Genes()[endGene]
	if !ok {
		t.Error("Expected genes map to contain the end gene")
	}

	// Check that the interactions are correct
	interactions := extendedPath.Interactions()
	if len(interactions) != 1 || interactions[0] != interactionID {
		t.Errorf("Expected interactions %+v to contain the new interaction %d", interactions, interactionID)
	}
}
