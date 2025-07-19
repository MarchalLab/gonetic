package editor_test

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"

	"github.com/MarchalLab/gonetic/internal/common/compare"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/editor"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/readers"
)

var update = flag.Bool("update", false, "update .golden files")

func createTestNetwork(t *testing.T, nwr *readers.NetworkReader, networkFile string, startingWeightNetwork float64) *graph.Network {
	// get actual network
	t.Logf("Reading network from %s", networkFile)
	network := nwr.NewNetworkFromFile(networkFile, true, true)
	t.Logf("Network size: %d", network.InteractionCount())
	return network
}

func updateGoldenNetworkFile(
	t *testing.T,
	goldenFileName string,
	network *graph.Network,
	gim *types.GeneIDMap,
) {
	if !*update {
		return
	}
	goldenFile, err := os.Create(goldenFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer goldenFile.Close()
	for id := range *network.Probabilities() {
		_, err := goldenFile.WriteString(fmt.Sprintf(
			"%s\t%s\t%f\n",
			gim.GetNameFromID(id.From()),
			gim.GetNameFromID(id.To()),
			network.Probabilities().GetProbability(id),
		))
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

func readGoldenNetworkFile(t *testing.T, goldenFileName string) map[types.GeneName]map[types.GeneName]float64 {
	goldenFile, err := os.Open(goldenFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer goldenFile.Close()
	scanner := bufio.NewScanner(goldenFile)
	var goldenInteractionWeights = make(map[types.GeneName]map[types.GeneName]float64)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, "\t")
		g1, g2 := types.GeneName(split[0]), types.GeneName(split[1])
		p, err := strconv.ParseFloat(split[2], 64)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if _, ok := goldenInteractionWeights[g1]; !ok {
			goldenInteractionWeights[g1] = make(map[types.GeneName]float64)
		}
		goldenInteractionWeights[g1][g2] = p
	}
	return goldenInteractionWeights
}

func compareNetworkToGolden(
	t *testing.T,
	goldenFileName string,
	network *graph.Network,
	gim *types.GeneIDMap,
) {
	errors := make([]string, 0)
	// get golden network
	goldenInteractionWeights := readGoldenNetworkFile(t, goldenFileName)
	goldenSize := 0
	for _, m := range goldenInteractionWeights {
		goldenSize += len(m)
	}
	if goldenSize != network.InteractionCount() {
		errors = append(errors, fmt.Sprintf(
			"Golden network size: %d, actual network size: %d",
			goldenSize,
			network.InteractionCount(),
		))
	}
	// create tolerance for edge weight comparison
	tolerance := compare.Tolerance(.00001)
	// check weights of all interactions
	count := 0
	checkedInteractionWeights := make(map[types.GeneName]map[types.GeneName]bool)
	for interactionID := range *network.Probabilities() {
		from := gim.GetNameFromID(interactionID.From())
		to := gim.GetNameFromID(interactionID.To())
		count++
		if _, ok := goldenInteractionWeights[from]; !ok {
			errors = append(errors, fmt.Sprintf(
				"Interaction %d->%d unexpected start gene",
				interactionID.From(),
				interactionID.To(),
			))
			continue
		}
		golden, ok := goldenInteractionWeights[from][to]
		if !ok {
			errors = append(errors, fmt.Sprintf(
				"Interaction %d->%d unexpected end gene",
				interactionID.From(),
				interactionID.To(),
			))
			continue
		}
		probability := network.Probabilities().GetProbability(interactionID)
		if !tolerance.FloatEqualWithinTolerance(golden, probability) {
			errors = append(errors, fmt.Sprintf(
				"Interaction %s->%s (%d) expected weight: %f, actual weight: %f",
				from,
				to,
				interactionID,
				golden,
				probability,
			))
			continue
		}
		if _, ok := checkedInteractionWeights[from]; !ok {
			checkedInteractionWeights[from] = make(map[types.GeneName]bool)
		}
		checkedInteractionWeights[from][to] = true
	}
	// check if we missed any golden interactions
	for g1, m := range goldenInteractionWeights {
		for g2 := range m {
			if _, ok := checkedInteractionWeights[g1][g2]; !ok {
				errors = append(errors, fmt.Sprintf("Interaction %s->%s not found in actual network", g1, g2))
			}
		}
	}
	if len(errors) > 1 {
		t.Errorf(
			"%d / %d Errors in %s:\n%s",
			len(errors),
			count,
			goldenFileName,
			strings.Join(errors, "\n"),
		)
	}
}

func TestSigmoidalNetworkWeighting(t *testing.T) {
	arguments.GlobalInteractionStore = types.NewInteractionStore()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	geneIdMap := types.NewGeneIDMap()
	nwr := readers.NewInitialNetworkReader(logger, geneIdMap)
	// create network
	networkFile := filepath.Join("testdata", "network.txt")
	network := createTestNetwork(t, nwr, networkFile, 0.0)
	t.Logf("Network size: %d", network.InteractionCount())
	// perform Bayesian addition hub weighting
	wayOfAddingProbabilitiesHubs, skipWeighting := editor.WeightingAddition(
		logger,
		editor.FromOnly,
		"bayes",
		"Invalid tag %s",
	)
	if !skipWeighting {
		network = editor.NetworkWeighting(
			editor.NewSigmoidalDegreeWeight(nwr.NewNetworkFromFile(networkFile, true, true), 0.01).Score,
			wayOfAddingProbabilitiesHubs,
			network,
		)
	}
	// golden network file contains the correct weights
	goldenFileName := filepath.Join("testdata", "interactionWeights.golden")
	// update golden
	updateGoldenNetworkFile(t, goldenFileName, network, geneIdMap)
	// compare actual to golden
	compareNetworkToGolden(t, goldenFileName, network, geneIdMap)
}

func TestLfcNetworkWeighting(t *testing.T) {
	methods := []string{
		"bayes",
		"mean",
		"mult",
		"none",
	}
	for _, method := range methods {
		lfcWeightingAux(t, method, 0.0)
	}
}

func lfcWeightingAux(t *testing.T, method string, defaultProbability float64) {
	arguments.GlobalInteractionStore = types.NewInteractionStore()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	geneIDMap := types.NewGeneIDMap()
	nwr := readers.NewInitialNetworkReader(logger, geneIDMap)
	// get actual network
	networkFile := filepath.Join("testdata", "expression_network.txt")
	network := createTestNetwork(t, nwr, networkFile, 1.0)
	// read expression data
	const weightingTag = "lfc"
	expressionFile := filepath.Join("testdata", "expression.csv")
	expressionFileData := readers.ReadExpressionFile(logger, expressionFile, weightingTag)
	expressionPerCondition := readers.MakeExpressionMap(logger, geneIDMap, expressionFileData, weightingTag)
	const condition = "Sample1"
	// way of adding probabilities
	wayOfAddingProbabilities, skipWeighting := editor.WeightingAddition(
		logger,
		editor.FromOnly,
		method,
		"Invalid weighting method %s",
	)
	// expression weighting
	weightingMethod := editor.NewLfcExpressionWeight
	weighting := weightingMethod(logger, expressionPerCondition[condition], defaultProbability)
	if !skipWeighting {
		network = editor.NetworkWeighting(
			weighting.Score,
			wayOfAddingProbabilities,
			network,
		)
	}
	// golden network file contains the correct weights
	goldenFileName := filepath.Join("testdata", fmt.Sprintf("expressionWeights-%s-%.2f.golden", method, defaultProbability))
	// update golden
	updateGoldenNetworkFile(t, goldenFileName, network, geneIDMap)
	// compare actual to golden
	compareNetworkToGolden(t, goldenFileName, network, geneIDMap)
}

// TestLfcSymmetry tests the symmetry of the LFC expression weighting up to a given tolerance of 1%.
func TestLfcSymmetry(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floatCompare := compare.Tolerance(1e-2)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := make([]struct {
		lfc   float64
		label string
	}, 100)
	for i := 0; i < 100; i++ {
		val := rng.Float64() * 10
		tests = append(tests, struct {
			lfc   float64
			label string
		}{
			lfc:   val,
			label: fmt.Sprintf("±%.4f", val),
		})
	}

	for _, tt := range tests {
		expressionData := make(map[types.GeneID]float64)

		// Add the test pair: +x and -x
		expressionData[0] = tt.lfc
		expressionData[1] = -tt.lfc

		// Add realistic background data centered around 0
		for i := 2; i < 1002; i++ {
			// Simulate Gaussian-like LFCs between -0.25 and 0.25
			expressionData[types.GeneID(i)] = (rng.Float64() - 0.5) * 0.5
		}

		weighting := editor.NewLfcExpressionWeight(logger, expressionData, 0.5)

		scorePos := weighting.Score(0)
		scoreNeg := weighting.Score(1)

		t.Run(tt.label, func(t *testing.T) {
			if !floatCompare.FloatEqualWithinTolerance(scorePos, scoreNeg) {
				t.Errorf("Expected symmetry for %s: score(+%.2f)=%.8f != score(-%.2f)=%.8f",
					tt.label, tt.lfc, scorePos, tt.lfc, scoreNeg)
			}
		})
	}
}
