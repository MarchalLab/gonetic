package types

import (
	"math/rand"
	"testing"
)

func TestFromToTypeToID(t *testing.T) {
	from, to := GeneID(1), GeneID(2)
	expectedID := InteractionID(1<<29 + 2)

	id := FromToTypeToID(from, to, 0)
	if id != expectedID {
		t.Errorf("expected ID: %d, got: %d", expectedID, id)
	}
}

func TestIDToFromTo(t *testing.T) {
	expectedFrom, expectedTo := GeneID(1), GeneID(2)
	id := InteractionID(1<<29 + 2)

	from, to := IDToFromTo(id)
	if from != expectedFrom || to != expectedTo {
		t.Errorf("expected from: %d, to: %d; got from: %d, to: %d", expectedFrom, expectedTo, from, to)
	}
}

// TestRoundTripBasic tests a simple round-trip encoding/decoding
func TestRoundTripBasic(t *testing.T) {
	from, to := GeneID(1234), GeneID(5678)
	typ := InteractionTypeID(26)

	id := FromToTypeToID(from, to, typ)
	decodedFrom, decodedTo := IDToFromTo(id)
	decodedTyp := IDToType(id)

	if decodedFrom != from || decodedTo != to || decodedTyp != typ {
		t.Errorf("round-trip failed: expected from: %d, to: %d, typ: %d; got from: %d, to: %d, typ: %d",
			from, to, typ, decodedFrom, decodedTo, decodedTyp)
	}
}

// TestRoundTripPropertyBased tests a large number of random inputs to ensure correctness
func TestRoundTripPropertyBased(t *testing.T) {
	const numTests = 100000

	for i := 0; i < numTests; i++ {
		from := GeneID(rand.Uint32()) & maxGeneID
		to := GeneID(rand.Uint32()) & maxGeneID
		typ := InteractionTypeID(rand.Uint32()) & maxInteractionType

		id := FromToTypeToID(from, to, typ)
		decodedFrom, decodedTo := IDToFromTo(id)
		decodedTyp := IDToType(id)

		if decodedFrom != from || decodedTo != to || decodedTyp != typ {
			t.Fatalf("Property-based test failed on iteration %d:\nExpected: from=%d, to=%d, typ=%d\nGot:      from=%d, to=%d, typ=%d",
				i, from, to, typ, decodedFrom, decodedTo, decodedTyp)
		}
	}
}

// TestRoundTripEdgeCases tests boundary conditions (lowest and highest possible values)
func TestRoundTripEdgeCases(t *testing.T) {
	edgeCases := []struct {
		from GeneID
		to   GeneID
		typ  InteractionTypeID
	}{
		{0, 0, 0}, // all zeros
		{maxGeneID, maxGeneID, maxInteractionType}, // all max
		{maxGeneID, 0, 0},                          // from max
		{0, maxGeneID, 0},                          // to max
		{0, 0, maxInteractionType},                 // type max
	}

	for _, c := range edgeCases {
		id := FromToTypeToID(c.from, c.to, c.typ)
		decodedFrom, decodedTo := IDToFromTo(id)
		decodedTyp := IDToType(id)

		if decodedFrom != c.from || decodedTo != c.to || decodedTyp != c.typ {
			t.Fatalf("Edge case failed:\nExpected: from=%d, to=%d, typ=%d\nGot:      from=%d, to=%d, typ=%d",
				c.from, c.to, c.typ, decodedFrom, decodedTo, decodedTyp)
		}
	}
}

func TestInteractionID_StringMinimal(t *testing.T) {
	from, to := GeneID(1), GeneID(2)
	typ := InteractionTypeID(3)
	interactionID := FromToTypeToID(from, to, typ)

	expected := "1;2;3"
	if interactionID.StringMinimal() != expected {
		t.Errorf("expected StringMinimal: %s, got: %s", expected, interactionID.StringMinimal())
	}
}
