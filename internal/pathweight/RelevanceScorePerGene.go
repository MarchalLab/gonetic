package pathweight

import (
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

// MakeGeneConditionMap creates a GeneConditionMap from a map of strings.
func MakeGeneConditionMap(gim *types.GeneIDMap, inputData map[string]struct{}) types.GeneConditionMap[float64] {
	outputMap := make(types.GeneConditionMap[float64])
	for entry := range inputData {
		if entry == "" {
			continue
		}
		split := strings.Split(entry, "\t")
		if len(split) < 3 || split[1] == "NA" || split[1] == "" {
			continue
		}
		gene := types.GeneName(split[0])
		condition := types.Condition(split[2])
		functionalScore, err := strconv.ParseFloat(split[1], 64)
		if err != nil {
			continue
		}
		outputMap.Add(gim.GetIDFromName(gene), condition, functionalScore)
	}
	return outputMap
}

// RelevanceScorePerGene calculates the total weight for each gene, based on the functional data, frequency data and mutational outliers.
func RelevanceScorePerGene(
	gim *types.GeneIDMap,
	functionalScoreData map[string]struct{},
	freqIncreaseData map[string]struct{},
	mutatedGenes map[string]struct{}, // map of mutated genes across all populations
	functionalDataWeighting FunctionalWeighting, // estimated non-parametric distributions used to calculate the functional score for each gene
	frequencyDataWeighting FrequencyWeighting, // estimated non-parametric distributions used to calculate the frequency score for each gene
	mutatorOutlierValues map[types.Condition]float64, // mutator outlier values for each population
) types.GeneConditionMap[float64] {
	// Calculate the weights of each gene for each line, based on the frequency, the functional data and the correction factor for each line
	// If a gene contains multiple mutations in 1 line, the mutation with the best (highest) weight is retained.
	// If a mutation does not have frequency increase data or functional data, the mean value for that data is assigned to it.

	functionalScores := MakeGeneConditionMap(gim, functionalScoreData)
	frequencyIncreases := MakeGeneConditionMap(gim, freqIncreaseData)

	weightsPerGene := make(types.GeneConditionMap[float64])
	for mutation := range mutatedGenes {
		split := strings.Split(mutation, "\t")
		geneString := split[0]
		if strings.ToLower(geneString) == "intergenic" {
			continue
		}
		gene := types.GeneName(geneString)
		condition := types.Condition(split[1])
		functionalScore := functionalDataWeighting.meanFunctionalData
		if val, ok := functionalScores.Get(gim.GetIDFromName(gene), condition); ok {
			functionalScore = val
		}
		frequencyIncrease := frequencyDataWeighting.meanFrequencyData
		if val, ok := frequencyIncreases.Get(gim.GetIDFromName(gene), condition); ok {
			frequencyIncrease = val
		}
		functionalWeight := functionalDataWeighting.calculateProbabilityForFunctionalData(functionalScore)
		frequencyWeight := frequencyDataWeighting.calculateProbabilityForFreqDataPerLine(condition, frequencyIncrease)
		correctionForLine, ok := mutatorOutlierValues[condition]
		if !ok {
			correctionForLine = 1
		}
		mutationWeight := functionalWeight * frequencyWeight * correctionForLine
		if _, ok := weightsPerGene.Get(gim.GetIDFromName(gene), condition); ok {
			if weight, _ := weightsPerGene.Get(gim.GetIDFromName(gene), condition); weight < mutationWeight {
				weightsPerGene.Add(gim.GetIDFromName(gene), condition, mutationWeight)
			}
		} else {
			weightsPerGene.Add(gim.GetIDFromName(gene), condition, mutationWeight)
		}
	}
	return weightsPerGene
}
