package readers

import (
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

func TestParseInteractions(t *testing.T) {
	gim := types.NewGeneIDMap()
	tests := []struct {
		input            string
		startGene        types.GeneID
		expected         []types.InteractionID
		probabilities    []float64
		interactionTypes []types.InteractionTypeID
		direction        types.PathDirection
	}{
		{
			input:     "AAAAA->BBBBB<-CCCCC",
			startGene: gim.SetName("AAAAA"),
			expected: []types.InteractionID{
				types.FromToTypeToID(gim.SetName("AAAAA"), gim.SetName("BBBBB"), 1),
				types.FromToTypeToID(gim.SetName("CCCCC"), gim.SetName("BBBBB"), 1),
			},
			probabilities:    []float64{0.1, 0.2},
			interactionTypes: []types.InteractionTypeID{1, 1},
			direction:        types.DownUpstreamPath,
		},
		{
			input:     "A<-B->C<-D",
			startGene: gim.SetName("A"),
			expected: []types.InteractionID{
				types.FromToTypeToID(gim.SetName("B"), gim.SetName("A"), 1),
				types.FromToTypeToID(gim.SetName("B"), gim.SetName("C"), 2),
				types.FromToTypeToID(gim.SetName("D"), gim.SetName("C"), 3),
			},
			probabilities:    []float64{0.1, 0.2, 0.3},
			interactionTypes: []types.InteractionTypeID{1, 2, 3},
			direction:        types.InvalidPathDirection,
		},
		{
			input:     "A<-B->C",
			startGene: gim.SetName("A"),
			expected: []types.InteractionID{
				types.FromToTypeToID(gim.SetName("B"), gim.SetName("A"), 1),
				types.FromToTypeToID(gim.SetName("B"), gim.SetName("C"), 1),
			},
			probabilities:    []float64{0.1, 0.2},
			interactionTypes: []types.InteractionTypeID{1, 1},
			direction:        types.UpDownstreamPath,
		},
	}

	for _, test := range tests {
		result, _, _, startGene, direction := parseInteractions(gim, test.input, test.probabilities, test.interactionTypes)
		expected := types.NewInteractionIDSet()
		expected.Add(test.expected)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("parseInteractions(%q) = %v; want %v", test.input, result, expected)
		}
		if startGene != test.startGene {
			t.Errorf("parseInteractions(%q) = %v; want %v", test.input, startGene, test.startGene)
		}
		if direction != test.direction {
			t.Errorf("parseInteractions(%q) = %v; want %v", test.input, direction, test.direction)
		}
	}
}

func TestParseEdgeScores(t *testing.T) {
	tests := []struct {
		input    string
		expected []float64
	}{
		{
			input:    "[0.1 0.2 0.3]",
			expected: []float64{0.1, 0.2, 0.3},
		},
		{
			input:    "[1.0 2.0 3.0]",
			expected: []float64{1.0, 2.0, 3.0},
		},
		{
			input:    "[0.5 0.25 0.125]",
			expected: []float64{0.5, 0.25, 0.125},
		},
	}

	for _, test := range tests {
		result := parseEdgeScores(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("parseEdgeScores(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestReadPathList(t *testing.T) {
	logger := slog.Default()
	gim := ReadIDMap[types.GeneID, types.GeneName]("testdata/gene-ids")
	cutoff := 0.0
	inputDir := "testdata"
	outputDir := "testresult/ReadPathList"

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	// Read all files in the input directory
	files, err := os.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("failed to read input directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".paths" {
			continue
		}

		inputFile := filepath.Join(inputDir, file.Name())
		outputFile := filepath.Join(outputDir, file.Name())

		// Read the paths
		t.Logf("Reading paths from %s", inputFile)
		paths := ReadPathList(logger, gim, math.MaxInt, "test", cutoff, inputFile)

		// Write the paths back to a file
		writePaths(gim, outputFile, paths)
		// Compare the original and written files
		if !fileio.CompareFiles(inputFile, outputFile) {
			t.Errorf("files do not match: %v", err)
		}
	}
}

func writePaths(gim *types.GeneIDMap, fileName string, paths map[types.CNFHeader]types.CompactPathList) {
	fw := fileio.FileWriter{}
	lines := make([]string, 0, len(paths))
	for header, pathMap := range paths {
		from := header.Gene
		condition := header.ConditionName
		for _, path := range pathMap {
			lines = append(lines, path.TxtString(gim, from, condition))
		}
	}
	fw.WriteLinesToNewFile(fileName, lines)
}
