package interpretation

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

type GenesOfInterest struct {
	Identifier string
	Genes      types.GeneConditionMap[struct{}]
}

func NewGenesOfInterest(identifier string, genes types.GeneConditionMap[struct{}]) GenesOfInterest {
	if genes == nil {
		genes = make(types.GeneConditionMap[struct{}])
	}
	return GenesOfInterest{
		Identifier: identifier,
		Genes:      genes,
	}
}

func NewGenesOfInterestMap(gim *types.GeneIDMap, data []readers.FileData) map[string]GenesOfInterest {
	goi := make(map[string]GenesOfInterest, len(data))
	for _, entry := range data {
		goi[entry.ID] = NewGenesOfInterest(entry.ID, readers.MakeConditionMap(gim, entry))
	}
	return goi
}

func (goi GenesOfInterest) ConditionsOfInterest(coi types.ConditionSet) types.ConditionSet {
	if coi == nil {
		coi = make(types.ConditionSet)
	}
	for _, conditions := range goi.Genes {
		for condition := range conditions {
			coi[condition] = struct{}{}
		}
	}
	return coi
}

func ConditionsOfInterest(gois map[string]GenesOfInterest) types.ConditionSet {
	coi := make(types.ConditionSet)
	for _, goi := range gois {
		coi = goi.ConditionsOfInterest(coi)
	}
	return coi
}
