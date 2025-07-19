package interpretation

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func TestPathInSubnetwork(t *testing.T) {
	tests := []struct {
		name         string
		interactions []types.InteractionID
		subnetwork   types.GeneGeneMap[struct{}]
		expected     bool
	}{
		{
			name: "Path in subnetwork",
			interactions: []types.InteractionID{
				types.FromToToID(1, 2),
				types.FromToToID(2, 3),
			},
			subnetwork: types.GeneGeneMap[struct{}]{
				1: {2: struct{}{}},
				2: {3: struct{}{}},
			},
			expected: true,
		},
		{
			name: "Path not in subnetwork",
			interactions: []types.InteractionID{
				types.FromToToID(1, 2),
				types.FromToToID(2, 4),
			},
			subnetwork: types.GeneGeneMap[struct{}]{
				1: {2: struct{}{}},
				2: {3: struct{}{}},
			},
			expected: false,
		},
		{
			name:         "Empty interactions",
			interactions: []types.InteractionID{},
			subnetwork: types.GeneGeneMap[struct{}]{
				1: {2: struct{}{}},
				2: {3: struct{}{}},
			},
			expected: true,
		},
		{
			name: "Empty subnetwork",
			interactions: []types.InteractionID{
				types.FromToToID(1, 2),
			},
			subnetwork: types.GeneGeneMap[struct{}]{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interactions := types.NewInteractionIDSet()
			interactions.Add(tt.interactions)
			subnetwork := types.NewInteractionIDSet()
			for from, tos := range tt.subnetwork {
				for to := range tos {
					subnetwork.Set(types.FromToToID(from, to))
				}
			}
			result := pathInSubnetwork(interactions, subnetwork)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func pathsForTestFile(_ string) string {
	return "testdata/paths.txt"
}

func weightsForTestFile(_, _ string) string {
	return ""
}

func parseGeneData(geneData readers.FileData, gim *types.GeneIDMap) map[types.Condition][]types.GeneID {
	parsedGeneData := make(map[types.Condition][]types.GeneID)
	for _, entry := range geneData.Entries {
		condition := types.Condition(entry[geneData.Headers["condition"]])
		gene := gim.GetIDFromName(types.GeneName(entry[geneData.Headers["gene"]]))
		parsedGeneData[condition] = append(parsedGeneData[condition], gene)
	}
	return parsedGeneData
}

func readTestData() (
	*slog.Logger,
	*types.GeneIDMap,
	types.InteractionIDSet,
	map[string]GenesOfInterest,
	map[types.CNFHeader]types.CompactPathList,
	types.Conditions,
	types.InteractionIDSet,
) {
	// create logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// read the gene id map
	gim := readers.ReadGeneMap("testdata/gene-ids")
	// read the weighted network
	headers := []string{"from", "to", "rank"}
	weightedInteractions := readers.ReadSeparatedFile(
		logger,
		"testdata/weighted.network",
		"final subnetwork",
		headers,
		headers,
	)
	// convert the network file data to a map of interactions
	rankedInteractions := types.NewInteractionIDSet()
	for _, entry := range weightedInteractions.Entries {
		from := gim.GetIDFromName(types.GeneName(entry[weightedInteractions.Headers["from"]]))
		to := gim.GetIDFromName(types.GeneName(entry[weightedInteractions.Headers["to"]]))
		rankedInteractions.Set(types.FromToToID(from, to))
	}
	// the data files
	genesOfInterest := NewGenesOfInterestMap(gim, []readers.FileData{
		readers.ReadInputDataHeadersMutationFile(
			logger,
			"testdata/mutations",
		),
		readers.ReadExpressionFile(
			logger,
			"testdata/expression",
			"none",
		),
	})
	// gather conditions
	conditions := make(types.Conditions, 0)
	conditionMap := make(map[types.Condition]struct{})
	for _, coi := range genesOfInterest["mutations"].Genes {
		for condition := range coi {
			if _, ok := conditionMap[condition]; ok {
				continue
			}
			conditions = append(conditions, condition)
			conditionMap[condition] = struct{}{}
		}
	}
	for _, coi := range genesOfInterest["expression"].Genes {
		for condition := range coi {
			if _, ok := conditionMap[condition]; ok {
				continue
			}
			conditions = append(conditions, condition)
			conditionMap[condition] = struct{}{}
		}
	}
	// read the paths per sample
	interactionMap := types.NewInteractionIDSet()
	paths := readers.ReadPathList(
		logger,
		gim,
		math.MaxInt,
		"test",
		0,
		pathsForTestFile(""),
	)
	interactions := types.NewInteractionIDSet()
	for ids := range interactionMap {
		for id := range ids {
			from, to := types.IDToFromTo(id)
			interactions.Set(types.FromToToID(from, to))
		}
	}
	// return the data
	return logger, gim, rankedInteractions, genesOfInterest, paths, conditions, interactions
}

func TestEdgesInCondition(t *testing.T) {
	t.Skip("Skipping this test for now, manually compute the expected output first")
	dir := filepath.Join("testresult", "edgesInCondition")
	fileio.CreateEmptyDir(dir)
	outputFile := filepath.Join(dir, "output")
	// read the test data
	logger, gim, rankedInteractions, _, _, conditions, _ := readTestData()
	fw := &fileio.FileWriter{Logger: logger}

	// Call EdgesInCondition
	edgesInCondition, _, _ := EdgesInCondition(
		logger,
		gim,
		conditions,
		rankedInteractions,
		[]string{"eqtl"},
		0,
		pathsForTestFile,
		math.MaxInt,
	)

	// Write edgesInCondition to a file
	edgeLines := make([]string, 0)
	edgeLines = append(edgeLines, "#from\tto\tcondition")
	for interaction, conditionBools := range edgesInCondition {
		for i, condition := range conditionBools {
			if condition > 0 {
				edgeLines = append(edgeLines, fmt.Sprintf("%s\t%s\t%s",
					gim.GetNameFromID(interaction.From()),
					gim.GetNameFromID(interaction.To()),
					conditions[i],
				))
			}
		}
	}
	sort.Strings(edgeLines)
	fw.WriteLinesToNewFile(outputFile, edgeLines)
	// check that the new file is the same as the expected file
	expectedFile := "testdata/edgesInCondition.txt"
	if !fileio.CompareFiles(outputFile, expectedFile) {
		t.Errorf("Files %s and %s are not the same", outputFile, expectedFile)
	}
}

func TestConditionSpecificRanking(t *testing.T) {
	t.Skip("Skipping this test for now, manually compute the expected output first")
	dir := filepath.Join("testresult", "conditionSpecificRanking")
	fileio.CreateEmptyDir(dir)
	// read the test data
	logger, gim, rankedInteractions, _, _, conditions, _ := readTestData()
	fw := &fileio.FileWriter{Logger: logger}

	// setup
	_, pathsInCondition, _ := EdgesInCondition(
		logger,
		gim,
		conditions,
		rankedInteractions,
		[]string{"eqtl"},
		0,
		pathsForTestFile,
		math.MaxInt,
	)

	// Call ConditionSpecificRanking
	conditionSpecificGeneRanking := ConditionSpecificRanking(conditions, pathsInCondition)
	common := &arguments.Common{
		GeneIDMap:  gim,
		FileWriter: fw,
	}
	interpreter := Interpreter{common}
	interpreter.WriteConditionSpecificRanking(
		interpreter.FileWriter,
		conditionSpecificGeneRanking,
		conditions,
		dir,
		make(types.GeneTranslationMap),
	)

	// check that the new file is the same as the expected file
	outputFile := filepath.Join(dir, "conditionSpecificExpressionRanking.txt")
	expectedFile := "testdata/conditionSpecificExpressionRanking.txt"
	if !fileio.CompareFiles(outputFile, expectedFile) {
		t.Errorf("Files %s and %s are not the same", outputFile, expectedFile)
	}
	outputFile = filepath.Join(dir, "conditionSpecificMutationRanking.txt")
	expectedFile = "testdata/conditionSpecificMutationRanking.txt"
	if !fileio.CompareFiles(outputFile, expectedFile) {
		t.Errorf("Files %s and %s are not the same", outputFile, expectedFile)
	}
}
