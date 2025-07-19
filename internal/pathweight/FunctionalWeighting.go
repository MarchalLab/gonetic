package pathweight

import (
	"log/slog"
	"strconv"
	"strings"
)

type FunctionalWeighting struct {
	funcScore                  bool
	meanFunctionalData         float64
	functionalDataAllMutations []float64
}

func NewFunctionalWeighting(logger *slog.Logger, functionalScores map[string]struct{}, funcScore bool) FunctionalWeighting {
	functionalDataAllMutations := make([]float64, 0)
	// Load functional data
	functionalDataSum := 0.0
	// If synonymous mutations are present, filter them out
	for entry := range functionalScores {
		if entry == "" {
			continue
		}
		split := strings.Split(entry, "\t")
		synonymous := "no"
		if len(split) > 4 {
			synonymous = split[3]
		}
		if split[1] == "" || split[1] == "NA" || synonymous == "yes" {
			continue
		}
		functionalSCore, err := strconv.ParseFloat(split[1], 64)
		if err != nil {
			logger.Error("could not parse functional score", "string", split[1])
			continue
		}
		functionalDataAllMutations = append(functionalDataAllMutations, functionalSCore)
		functionalDataSum += functionalSCore
	}
	numberOfValuesInFunctionalList := len(functionalDataAllMutations)
	return FunctionalWeighting{
		funcScore:                  funcScore,
		meanFunctionalData:         functionalDataSum / float64(numberOfValuesInFunctionalList),
		functionalDataAllMutations: functionalDataAllMutations,
	}
}

// Calculate probabilities for functional data (nonparameteric eCDF)
func (fw FunctionalWeighting) calculateProbabilityForFunctionalData(functionalData float64) float64 {
	if !fw.funcScore {
		return 1
	}
	numberOfValuesSmallerThanFunctionalScore := 0
	for _, element := range fw.functionalDataAllMutations {
		if element <= functionalData {
			numberOfValuesSmallerThanFunctionalScore++
		}
	}
	CDSScore := float64(numberOfValuesSmallerThanFunctionalScore) / float64(len(fw.functionalDataAllMutations))
	// The higher the functional score, the more deleterious.
	return CDSScore
}
