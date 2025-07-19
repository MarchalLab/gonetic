package types

// GeneID is an integer that represents a gene ID
type GeneID uint64
type GeneSet map[GeneID]struct{}
type GeneGeneMap[T any] map[GeneID]map[GeneID]T

// GeneName is a string that represents a gene name
type GeneName string
type GeneTranslationMap = TranslationMap[GeneName]

// GeneIDMap is a map that stores the mapping between gene names and gene IDs
type GeneIDMap = IDMap[GeneID, GeneName]

func NewGeneIDMap() *GeneIDMap {
	return NewIDMap[GeneID, GeneName]()
}
