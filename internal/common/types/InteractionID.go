package types

import (
	"fmt"
	"strconv"
)

type InteractionID uint64

const (
	geneIDBits          = 29
	interactionTypeBits = 6
	maxGeneID           = (1 << geneIDBits) - 1
	maxInteractionType  = (1 << interactionTypeBits) - 1
)

func (i InteractionID) Type() InteractionTypeID {
	return IDToType(i)
}

func (i InteractionID) From() GeneID {
	return IDToFrom(i)
}

func (i InteractionID) To() GeneID {
	return IDToTo(i)
}

func (i InteractionID) FromTo() (GeneID, GeneID) {
	return IDToFromTo(i)
}

func (i InteractionID) Reverse() InteractionID {
	return FromToTypeToID(i.To(), i.From(), i.Type())
}

func (i InteractionID) OtherEndGene(gene GeneID) (GeneID, error) {
	if gene == i.From() {
		return i.To(), nil
	}
	if gene == i.To() {
		return i.From(), nil
	}
	return 0, fmt.Errorf("gene %d is not part of interaction %d->%d", gene, i.From(), i.To())
}

func (i InteractionID) IsUndefined() bool {
	return i == 0
}

func (i InteractionID) StringMinimal() string {
	return fmt.Sprintf("%d;%d;%d", i.From(), i.To(), i.Type())
}

func (i InteractionID) StringWithProbability(p float64) string {
	return fmt.Sprintf("%d;%d;%d;%s",
		i.From(),
		i.To(),
		i.Type(),
		strconv.FormatFloat(p, 'f', -1, 64),
	)
}

// FromToTypeToID is a function that returns the ID of the interaction from the given gene IDs and interaction type ID
func FromToTypeToID(from, to GeneID, typ InteractionTypeID) InteractionID {
	return InteractionID(uint64(typ)<<(2*geneIDBits) | uint64(from)<<geneIDBits | uint64(to))
}

// FromToToID is a function that returns the ID of the interaction from the given gene IDs
// TODO: this should be removed, use the version with type instead
func FromToToID(from, to GeneID) InteractionID {
	return InteractionID(uint64(from<<geneIDBits) | uint64(to))
}

// IDToFromTo is a function that returns the gene IDs of the interaction from the given ID
func IDToFromTo(id InteractionID) (GeneID, GeneID) {
	return IDToFrom(id), IDToTo(id)
}

// IDToFrom is a function that returns the gene ID of the start point of the interaction with the given ID
func IDToFrom(id InteractionID) GeneID {
	return GeneID((id >> geneIDBits) & maxGeneID)
}

// IDToTo is a function that returns the gene ID of the end point of the interaction with the given ID
func IDToTo(id InteractionID) GeneID {
	return GeneID(id & maxGeneID)
}

// IDToType is a function that returns the interaction type of the interaction with the given ID
func IDToType(id InteractionID) InteractionTypeID {
	return InteractionTypeID((id >> (2 * geneIDBits)) & maxInteractionType)
}
