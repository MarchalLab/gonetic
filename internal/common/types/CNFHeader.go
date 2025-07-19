package types

import "fmt"

// CNFHeader is the start gene and start condition of a CNF
type CNFHeader struct {
	Gene          GeneID
	ConditionName Condition
}

func NewCNFHeader(
	startSample Condition,
	startGene GeneID,
) CNFHeader {
	return CNFHeader{
		Gene:          startGene,
		ConditionName: startSample,
	}
}

func NewCNFHeaderFromString(
	header string,
) CNFHeader {
	var gene GeneID
	var condition Condition
	fmt.Sscanf(header, "%d;%s", &gene, &condition)
	return CNFHeader{
		Gene:          gene,
		ConditionName: condition,
	}
}

func (h CNFHeader) Name() string {
	return fmt.Sprintf("%d;%s", h.Gene, h.ConditionName)
}
