package interpretation

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func pathInSubnetwork(interactions types.InteractionIDSet, subnetwork types.InteractionIDSet) bool {
	// check if path is contained in the subnetwork
	for interactionID := range interactions {
		if !subnetwork.Has(interactionID) {
			return false
		}
	}
	return true
}

// EdgesInCondition returns a map of interactions and their presence in the final subnetwork per condition
func EdgesInCondition(
	logger *slog.Logger,
	gim *types.GeneIDMap,
	orderedConditions types.Conditions,
	interactions types.InteractionIDSet,
	pathTypes []string,
	sldCutoff float64,
	pathsFile func(string) string,
	maxPaths int,
) (
	map[types.InteractionID][]float64, // summed probabilities for each interaction per condition
	map[string]map[types.Condition]map[types.GeneID][2]float64, // summed probabilities for start and end genes
	map[string]map[types.Condition]map[types.GeneID]struct{}, // genes of interest per condition per path type
) {
	edgesInCondition := make(map[types.InteractionID][]float64)
	pathsInCondition := make(map[string]map[types.Condition]map[types.GeneID][2]float64)
	goiInCondition := make(map[string]map[types.Condition]map[types.GeneID]struct{})
	goiInCondition["differential expression"] = make(map[types.Condition]map[types.GeneID]struct{})
	goiInCondition["mutation"] = make(map[types.Condition]map[types.GeneID]struct{})
	// map conditions to their index
	conditionIndexMap := make(map[types.Condition]int)
	for i, condition := range orderedConditions {
		conditionIndexMap[condition] = i
		goiInCondition["differential expression"][condition] = make(map[types.GeneID]struct{})
		goiInCondition["mutation"][condition] = make(map[types.GeneID]struct{})
	}

	for _, pathType := range pathTypes {
		pathsForType := make(map[types.Condition]map[types.GeneID][2]float64)
		// read the paths from the EQTL .paths file
		eqtlPaths := readers.ReadPathList(logger, gim, maxPaths, pathType, sldCutoff, pathsFile(pathType))
		// convert the weightedInteractions to a GeneGeneMap for easy access
		// and create the edgesInCondition map
		subnetwork := types.NewInteractionIDSet()
		for interaction := range interactions {
			subnetwork.Set(interaction)
			edgesInCondition[interaction] = make([]float64, len(orderedConditions))
		}
		// iterate over the paths and store the edges of present paths in the final subnetwork per condition
		for _, paths := range eqtlPaths {
			for _, path := range paths {
				condition := path.StartCondition
				if !pathInSubnetwork(path.InteractionSet(), subnetwork) {
					continue
				}
				// if the path is contained, add its probability to all its edges for this condition
				for interaction := range path.InteractionSet() {
					edgesInCondition[interaction][conditionIndexMap[condition]] += path.Probability
				}
				if _, ok := pathsForType[condition]; !ok {
					pathsForType[condition] = make(map[types.GeneID][2]float64)
				}
				// for each path, ensure the start and end genes are initialized in the pathsForType map
				for _, gene := range []types.GeneID{path.StartGene, path.EndGene} {
					if _, ok := pathsForType[condition][gene]; !ok {
						pathsForType[condition][gene] = [2]float64{0, 0}
					}
				}
				// if the path is contained, add its probability to the start and end gene for this condition
				val := pathsForType[condition][path.StartGene]
				pathsForType[condition][path.StartGene] = [2]float64{val[0] + path.Probability, val[1]}
				val = pathsForType[condition][path.EndGene]
				pathsForType[condition][path.EndGene] = [2]float64{val[0], val[1] + path.Probability}
			}
		}
		// store the paths for this path type
		pathsInCondition[pathType] = pathsForType
		// build the genes of interest for this path type
		switch pathType {
		case "eqtl":
			addGenesOfInterest(pathsForType, 0, "differential expression", goiInCondition)
			addGenesOfInterest(pathsForType, 1, "mutation", goiInCondition)
			break
		case "expression":
			addGenesOfInterest(pathsForType, 0, "differential expression", goiInCondition)
			addGenesOfInterest(pathsForType, 1, "differential expression", goiInCondition)
			break
		case "mutation":
			addGenesOfInterest(pathsForType, 0, "mutation", goiInCondition)
		}
		logger.Info("Condition-specific paths",
			slog.String("pathType", pathType),
			"goi", fmt.Sprintf("%+v", goiInCondition),
		)
	}

	return edgesInCondition, pathsInCondition, goiInCondition
}

// addGenesOfInterest adds genes of interest to the condition-specific ranking
func addGenesOfInterest(
	pathsPerPathType map[types.Condition]map[types.GeneID][2]float64,
	scoreIndex int, // 0 for start gene, 1 for end gene
	goiType string, // "differential expression" or "mutation"
	goiInCondition map[string]map[types.Condition]map[types.GeneID]struct{},
) {
	if _, ok := goiInCondition[goiType]; !ok {
		goiInCondition[goiType] = make(map[types.Condition]map[types.GeneID]struct{})
	}
	for condition, genes := range pathsPerPathType {
		for geneID, scores := range genes {
			if scores[scoreIndex] != 0 {
				if _, ok := goiInCondition[goiType][condition]; !ok {
					goiInCondition[goiType][condition] = make(map[types.GeneID]struct{})
				}
				goiInCondition[goiType][condition][geneID] = struct{}{}
			}
		}
	}
}

// WriteConditionSpecificRanking writes the condition-specific ranking to a file
func (interpreter Interpreter) WriteConditionSpecificRanking(
	fileWriter *fileio.FileWriter,
	ranking map[string]map[types.Condition]map[types.GeneID]float64,
	conditions types.Conditions,
	directory string,
	geneMapping types.GeneTranslationMap,
) error {
	for identifier := range ranking {
		err := interpreter.writeConditionSpecificRanking(
			fileWriter,
			identifier,
			ranking[identifier],
			conditions,
			directory,
			geneMapping,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeConditionSpecificRanking writes the condition-specific ranking to a file for a specific identifier
func (interpreter Interpreter) writeConditionSpecificRanking(
	fileWriter *fileio.FileWriter,
	identifier string,
	ranking map[types.Condition]map[types.GeneID]float64,
	conditions types.Conditions,
	directory string,
	geneMapping types.GeneTranslationMap,
) error {
	linesToWrite := make([]string, 0, len(ranking))
	for _, condition := range conditions {
		sorted := sortKeysByValueDesc(ranking[condition])
		line := make([]string, 0, len(sorted)+1)
		line = append(line, condition.String())
		for _, gene := range sorted {
			name := interpreter.GetMappedName(gene, geneMapping)
			line = append(line, fmt.Sprintf("%s %f", name, ranking[condition][gene]))
		}
		linesToWrite = append(linesToWrite, strings.Join(line, "\t"))
	}
	return fileWriter.WriteLinesToNewFile(filepath.Join(
		directory,
		fmt.Sprintf("conditionSpecific%sRanking.txt", ToPascalCase(identifier)),
	), linesToWrite)
}

// ConditionSpecificRanking ranks genes based on their presence in the final subnetwork per condition
// Where top-level keys are "DE" and "mut" (the roles). Mapping logic:
// Create a combined ranking that sums appropriate scores across all relevant path types for each role.
func ConditionSpecificRanking(
	orderedConditions types.Conditions,
	pathsPerPathType map[string]map[types.Condition]map[types.GeneID][2]float64,
) map[string]map[types.Condition]map[types.GeneID]float64 {
	ranking := map[string]map[types.Condition]map[types.GeneID]float64{}
	for _, condition := range orderedConditions {
		deScores := make(map[types.GeneID]float64)
		mutScores := make(map[types.GeneID]float64)

		// Handle eqtl
		if eqtlPaths, ok := pathsPerPathType["eqtl"]; ok {
			for gene, scores := range eqtlPaths[condition] {
				// gather DE scores
				if scores[0] != 0 {
					deScores[gene] += scores[0]
				}
				// gather mutation scores
				if scores[1] != 0 {
					mutScores[gene] += scores[1]
				}
			}
		}

		// Handle expression
		if expPaths, ok := pathsPerPathType["expression"]; ok {
			for gene, scores := range expPaths[condition] {
				if scores[0]+scores[1] != 0 {
					deScores[gene] += scores[0] + scores[1]
				}
			}
		}

		// Handle mutation
		if mutPaths, ok := pathsPerPathType["mutation"]; ok {
			for gene, scores := range mutPaths[condition] {
				if scores[0] != 0 {
					mutScores[gene] += scores[0]
				}
			}
		}
		if len(deScores) > 0 {
			if _, exists := ranking["DE"]; !exists {
				ranking["DE"] = make(map[types.Condition]map[types.GeneID]float64)
			}
			ranking["DE"][condition] = deScores
		}
		if len(mutScores) > 0 {
			if _, exists := ranking["mutation"]; !exists {
				ranking["mutation"] = make(map[types.Condition]map[types.GeneID]float64)
			}
			ranking["mutation"][condition] = mutScores
		}
	}

	return ranking
}

// Function to sort map keys by their values in descending order
func sortKeysByValueDesc[K comparable](m map[K]float64) []K {
	// Create a slice of keys
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	// Sort the keys based on the values in the map in descending order
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})

	return keys
}
