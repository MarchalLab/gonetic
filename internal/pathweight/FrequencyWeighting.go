package pathweight

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

/**
 * This class provides the frequencyWeighting. It takes the frequency increase of all mutations per population individually.
 * Based on this data, a non-parametric distribution is estimated per population and this distribution is used to assign
 * scores to each mutation.
 * Calculate probabilities for frequency data (non parameteric eCDF)
 *
 * @param freqIncreasePopulations A set of strings, each string including the name of the mutated gene, the name of the line in which the mutation is involved and the frequency increase.
 * @param populations A set of the names of all involved populations
 * @param freqCutoff A double which is used as a cutoff, i.e. mutation with a freq increase below this cutoff are ignored
 * @param doFrequencyWeighting A boolean. If true frequency weighting is performed, otherwise it is not.
 */

type FrequencyWeighting struct {
	doFrequencyWeighting              bool
	frequencyDataAllMutationsPerLines map[types.Condition][]float64
	meanFrequencyData                 float64
}

func NewFrequencyWeighting(logger *slog.Logger, freqIncreasePopulations map[string]struct{}, populations types.ConditionSet, freqCutoff float64, doFrequencyWeighting bool) FrequencyWeighting {
	// !!!! Keep in mind the freqCutoff should be identical to the error of the mutation calling in %.
	// This will be use to discard only mutations which have a DECREASE in frequency larger than this number.

	// Load FrequencyData
	freqIncreaseByPopulation := make(map[types.Condition][]float64)
	frequencyDataAllMutations := make([]float64, 0, len(freqIncreasePopulations))
	for entry := range freqIncreasePopulations {
		if entry == "" {
			continue
		}
		split := strings.Split(entry, "\t")
		population := types.Condition(split[len(split)-1])
		if _, ok := populations[population]; !ok {
			logger.Warn("population not found in populations set", "population", population)
			continue
		}
		if _, ok := freqIncreaseByPopulation[population]; !ok {
			freqIncreaseByPopulation[population] = make([]float64, 0)
		}
		freqIncrease, err := strconv.ParseFloat(split[1], 64)
		if err != nil {
			logger.Error("could not parse frequency increase", "string", split[1])
			continue
		}
		freqIncreaseByPopulation[population] = append(freqIncreaseByPopulation[population], freqIncrease)
		frequencyDataAllMutations = append(frequencyDataAllMutations, freqIncrease)
	}

	frequencyDataAllMutationsPerLines := make(map[types.Condition][]float64)
	frequencyDataSum := float64(0)
	frequencyDataCount := 0
	for population, freqIncreases := range freqIncreaseByPopulation {
		frequencyDataSpecificLine := make([]float64, 0)
		for _, freqIncrease := range freqIncreases {
			// The multiplication by -1 means that genes which have an increase lower than the negative of the cutoff are discarded
			if !doFrequencyWeighting || freqIncrease+freqCutoff > 0 {
				frequencyDataSpecificLine = append(frequencyDataSpecificLine, freqIncrease)
				frequencyDataSum += freqIncrease
				frequencyDataCount++
			}
		}
		frequencyDataAllMutationsPerLines[population] = frequencyDataSpecificLine
	}

	meanFrequencyData := 0.0
	if frequencyDataCount > 0 {
		meanFrequencyData = frequencyDataSum / float64(frequencyDataCount)
	}
	return FrequencyWeighting{
		doFrequencyWeighting:              doFrequencyWeighting,
		frequencyDataAllMutationsPerLines: frequencyDataAllMutationsPerLines,
		meanFrequencyData:                 meanFrequencyData,
	}
}

/**
 * Calculates the probability (score) for a specific frequency increase, based on all mutations in the line where the frequency increase was observed.
 *
 * @param line The name of the line in which the frequency increase was observed
 * @param frequencyIncrease A double which is the value of the frequency increase
 */

func (fw FrequencyWeighting) calculateProbabilityForFreqDataPerLine(population types.Condition, frequencyIncrease float64) float64 {
	if !fw.doFrequencyWeighting {
		return 1
	}
	if frequencyData, ok := fw.frequencyDataAllMutationsPerLines[population]; !ok || len(frequencyData) == 0 {
		return 1
	}
	numberOfValuesSmallerThanFrequencyIncrease := 0
	for _, entry := range fw.frequencyDataAllMutationsPerLines[population] {
		if entry <= frequencyIncrease {
			numberOfValuesSmallerThanFrequencyIncrease++
		}
	}
	numberOfMutationsForLine := len(fw.frequencyDataAllMutationsPerLines[population])
	// The higher the frequency increase, the higher the score.
	return float64(numberOfValuesSmallerThanFrequencyIncrease) / float64(numberOfMutationsForLine)
}
