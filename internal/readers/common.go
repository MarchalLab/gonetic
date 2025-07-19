package readers

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func LoadConditionData(data FileData, header string) types.ConditionSet {
	result := make(types.ConditionSet)
	index := data.Headers[header]
	for _, entry := range data.Entries {
		line := types.Condition(entry[index])
		result[line] = struct{}{}
	}
	return result
}

func LoadData(data FileData, headers ...string) map[string]struct{} {
	return LoadDataGeneric(data, map[string]struct{}{"gene name": {}}, headers...)
}

func LoadDataGeneric(data FileData, geneHeaders map[string]struct{}, headers ...string) map[string]struct{} {
	result := make(map[string]struct{})
	indexes := make([]int, 0, len(headers))
	geneNameIndex := make(map[int]string)
	for _, header := range headers {
		if _, ok := geneHeaders[header]; ok {
			geneNameIndex[len(indexes)] = header
		}
		indexes = append(indexes, data.Headers[header])
	}
	for _, entry := range data.Entries {
		line := make([]string, 0, len(indexes))
		for idx, index := range indexes {
			if _, ok := geneNameIndex[idx]; ok {
				line = append(line, entry[index])
			} else {
				line = append(line, entry[index])
			}
		}
		result[strings.Join(line, "\t")] = struct{}{}
	}
	return result
}

func MakeExpressionMap(logger *slog.Logger, gim *types.GeneIDMap, expressionFileData FileData, tag string) map[types.Condition]map[types.GeneID]float64 {
	expressionDataPerCondition := make(map[types.Condition]map[types.GeneID]float64)
	if tag == "none" {
		// do not use any of the data for weighting
		return expressionDataPerCondition
	}
	for entry := range LoadData(expressionFileData, "condition", "gene name", tag) {
		split := strings.Split(entry, "\t")
		condition := types.Condition(split[0])
		gene := gim.SetName(split[1])
		tagValue, err := strconv.ParseFloat(split[2], 64)
		if err != nil {
			logger.Error("Failed to parse expression value", "gene", gene, "condition", condition)
			continue
		}
		if _, ok := expressionDataPerCondition[condition]; !ok {
			expressionDataPerCondition[condition] = make(map[types.GeneID]float64)
		}
		if _, ok := expressionDataPerCondition[condition][gene]; ok {
			logger.Error("Multiple expression entries for gene",
				"gene", gim.GetNameFromID(gene),
				"condition", condition,
				"old", expressionDataPerCondition[condition][gene],
				"new", tagValue,
			)
		}
		expressionDataPerCondition[condition][gene] = tagValue
	}
	return expressionDataPerCondition
}

func MakeGeneMap(gim *types.GeneIDMap, geneData FileData) map[types.Condition]types.GeneSet {
	genesPerCondition := make(map[types.Condition]types.GeneSet)
	for entry := range LoadData(geneData, "condition", "gene name") {
		split := strings.Split(entry, "\t")
		condition := types.Condition(split[0])
		gene := gim.SetName(split[1])
		if _, ok := genesPerCondition[condition]; !ok {
			genesPerCondition[condition] = make(types.GeneSet)
		}
		genesPerCondition[condition][gene] = struct{}{}
	}
	return genesPerCondition
}

func MakeConditionMap(gim *types.GeneIDMap, geneData FileData) types.GeneConditionMap[struct{}] {
	conditionsPerGene := make(types.GeneConditionMap[struct{}])
	for entry := range LoadData(geneData, "condition", "gene name") {
		split := strings.Split(entry, "\t")
		condition := types.Condition(split[0])
		gene := gim.SetName(split[1])
		if _, ok := conditionsPerGene[gene]; !ok {
			conditionsPerGene[gene] = make(map[types.Condition]struct{})
		}
		conditionsPerGene[gene][condition] = struct{}{}
	}
	return conditionsPerGene
}
