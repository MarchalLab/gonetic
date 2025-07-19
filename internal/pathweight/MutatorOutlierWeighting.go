package pathweight

import (
	"bufio"
	"log/slog"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

/**
 * A method for calculating the outlier weight of all populations, based on the MAD and the mutRateParam
 * It assesses which population(s) have a significantly higher mutation rate than the other provided populations.
 *
 * @param mutatedGenes A set of strings, which include the name of the mutated gene and the line in which the gene is mutated.
 * @param outlierPopulationsFileLocation A string, which represents the location of a file in which the user-defined outlier populations are present, if given.
 * @param populations A set of the names of all involved populations
 * @param mutRateParam A double which controls how harshly an outlier populations should be down weighted (smaller is more down weighted)
 * @param correction A boolean. If true mutator outlier weighting is performed, otherwise it is not.
 *
 * @return A map including the names of all populations and their outlier weight
 */

func CalculateMutatorOutlierWeighting(logger *slog.Logger, mutatedGenes map[string]struct{}, outlierPopulationsFileLocation string, populations types.ConditionSet, mutRateParam float64, correction bool) map[types.Condition]float64 {

	// Calculate the number of mutations per strain
	numberOfMutationsPerStrain := make(map[types.Condition]float64)
	for strain := range populations {
		numberOfMutationsInStrain := 0
		for entry := range mutatedGenes {
			if strings.Contains(entry, "\t") {
				if types.Condition(strings.Split(entry, "\t")[1]) == strain {
					numberOfMutationsInStrain++
				}
			}
		}
		numberOfMutationsPerStrain[strain] = float64(numberOfMutationsInStrain)
	}

	//Iglewicz and Hoaglin modified Z score for detecting outliers. (Boris Iglewicz and David Hoaglin (1993), "Volume 16: How to Detect and Handle Outliers", The ASQC Basic References in Quality Control: Statistical Techniques, Edward F. Mykytka, Ph.D., Editor.)
	mutationCounts := make([]float64, 0, len(numberOfMutationsPerStrain))
	for _, count := range numberOfMutationsPerStrain {
		mutationCounts = append(mutationCounts, count)
	}
	median := calculateMedian(mutationCounts)
	medianAdjusted := calculateMedianAdjusted(mutationCounts, median)
	AbsoluteValuesModifiedZScores := make(map[types.Condition]float64)
	for strain, count := range numberOfMutationsPerStrain {
		AbsoluteValuesModifiedZScores[strain] = 0.6745 * (count - median) / medianAdjusted
	}

	// calculate mutator outlier weighting
	if !correction {
		result := make(map[types.Condition]float64)
		for strain := range AbsoluteValuesModifiedZScores {
			result[strain] = 1.0
		}
		return result
	}

	// If no outlier populations file is given, determine outlier and weight
	if outlierPopulationsFileLocation == "" {
		result := make(map[types.Condition]float64)
		for strain, score := range AbsoluteValuesModifiedZScores {
			if score > 3.5 {
				result[strain] = mutRateParam / score
			} else {
				result[strain] = 1.0
			}
		}
		return result
	}

	// If the provided outlier populations file was empty, weight all lines by 1.
	file, err := os.Open(outlierPopulationsFileLocation)
	if err != nil {
		logger.Error("Could not open the outlier populations file.", "error", err)
		return CalculateMutatorOutlierWeighting(logger, mutatedGenes, outlierPopulationsFileLocation, populations, mutRateParam, correction)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	outlierLines := make(types.ConditionSet)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			outlierLines[types.Condition(line)] = struct{}{}
		}
	}
	if len(outlierLines) > 0 {
		result := make(map[types.Condition]float64)
		for strain := range AbsoluteValuesModifiedZScores {
			result[strain] = 1.0
		}
		return result

	}

	// If specific outlier populations were given, weight only them
	result := make(map[types.Condition]float64)
	for strain, score := range AbsoluteValuesModifiedZScores {
		if _, ok := outlierLines[strain]; ok && 1/score < 1 && score > 3.5 {
			result[strain] = mutRateParam / score
		} else {
			result[strain] = 1.0
		}
	}
	return result
}

// O(nlogn) median, use O(n) algorithm instead if this is performance critical? http://dx.doi.org/10.4230/LIPIcs.SEA.2017.24
func calculateMedian(s []float64) float64 {
	sort.Float64s(s) // sort the numbers
	mid := len(s) / 2
	if len(s)%2 == 1 {
		return s[mid]
	}
	return (s[mid-1] + s[mid]) / 2
}

func calculateMedianAdjusted(s []float64, medianValue float64) float64 {
	adjustedSequence := make([]float64, 0, len(s))
	for _, element := range s {
		adjustedSequence = append(adjustedSequence, math.Abs(element-medianValue))
	}
	return calculateMedian(adjustedSequence)
}
