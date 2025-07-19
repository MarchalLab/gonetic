package types

import (
	"testing"
)

func TestNewCNFHeader(t *testing.T) {
	startSample := Condition("condition1")
	startGene := GeneID(1)
	header := NewCNFHeader(startSample, startGene)

	if header.Gene != startGene {
		t.Errorf("expected gene %d, got %d", startGene, header.Gene)
	}
	if header.ConditionName != startSample {
		t.Errorf("expected condition %s, got %s", startSample, header.ConditionName)
	}
}

func TestNewCNFHeaderFromString(t *testing.T) {
	headerStr := "1;condition1"
	expectedGene := GeneID(1)
	expectedCondition := Condition("condition1")
	header := NewCNFHeaderFromString(headerStr)

	if header.Gene != expectedGene {
		t.Errorf("expected gene %d, got %d", expectedGene, header.Gene)
	}
	if header.ConditionName != expectedCondition {
		t.Errorf("expected condition %s, got %s", expectedCondition, header.ConditionName)
	}
}

func TestCNFHeader_Name(t *testing.T) {
	header := CNFHeader{
		Gene:          GeneID(1),
		ConditionName: Condition("condition1"),
	}
	expectedName := "1;condition1"
	name := header.Name()

	if name != expectedName {
		t.Errorf("expected name %s, got %s", expectedName, name)
	}
}
