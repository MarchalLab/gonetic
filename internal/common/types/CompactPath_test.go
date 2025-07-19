package types

import (
	"testing"
)

func TestCompactPath_TxtString(t *testing.T) {
	// Mock data
	gim := NewGeneIDMap()
	gim.SetName("gene1")
	gim.SetName("gene2")
	gim.SetName("gene3")
	gim.SetName("gene4")

	interactions := NewInteractionIDSet()
	interaction1 := FromToToID(1, 2)
	interaction2 := FromToToID(2, 3)
	interaction3 := FromToToID(4, 3)
	interactions.Add([]InteractionID{
		interaction1,
		interaction2,
		interaction3,
	})

	path := CompactPath{
		ID:               1,
		Probability:      0.0336,
		InteractionOrder: []InteractionID{interaction1, interaction2, interaction3},
		ProbabilityOrder: []float64{0.5, 0.4, 0.3},
		Genes:            GeneSet{1: struct{}{}, 2: struct{}{}, 3: struct{}{}},
		Direction:        DownstreamPath,
		StartGene:        1,
		EndGene:          3,
		EndCondition:     "condition1",
		FromScore:        0.7,
		ToScore:          0.8,
	}

	expected := "condition1\t0.0336\tgene1->gene2->gene3<-gene4\t0.7\t[0.5 0.4 0.3]\t0.8\t[0 0 0]"
	result := path.TxtString(gim, 1, "condition1")

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
