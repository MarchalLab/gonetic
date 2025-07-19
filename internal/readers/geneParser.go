package readers

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// geneParser is an interface for parsing genes
type geneParser interface {
	parseGene(gene string) types.GeneID
}

// initialGeneParser is a geneParser to read the input data
type initialGeneParser struct {
	*types.GeneIDMap
}

func (gp initialGeneParser) parseGene(gene string) types.GeneID {
	return gp.SetName(gene)
}

// intermediateGeneParser is a geneParser to read the intermediate data
type intermediateGeneParser struct{}

func (gp intermediateGeneParser) parseGene(gene string) types.GeneID {
	return types.ParseID[types.GeneID](gene)
}
