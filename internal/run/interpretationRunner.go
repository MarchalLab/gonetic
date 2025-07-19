package run

import (
	"sort"

	"github.com/MarchalLab/gonetic/internal/graph"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/interpretation"
	"github.com/MarchalLab/gonetic/internal/readers"
)

type interpretationRunner struct {
	*arguments.Common
	*interpretation.Interpreter
	nwr *readers.NetworkReader
}

func NewInterpretation(args *arguments.Common) interpretationRunner {
	args.Info("Running interpreter")
	args.GeneIDMap = readers.ReadGeneMap(args.GeneMapFileToRead())
	nwr := readers.NewIntermediateNetworkReader(args.Logger)
	return interpretationRunner{
		Common:      args,
		Interpreter: &interpretation.Interpreter{Common: args},
		nwr:         nwr,
	}
}

func (runner interpretationRunner) Run(genesOfInterestData ...readers.FileData) {
	runner.DumpProfiles("int-start")
	defer func() {
		runner.DumpProfiles("int-end")
	}()

	// create genes of interest map
	genesOfInterest := interpretation.NewGenesOfInterestMap(runner.GeneIDMap, genesOfInterestData)
	// create gene name map
	geneNameMap := readers.ConvertMappingFile(runner.Logger, runner.MappingFile)
	// create an ordered list of condition names; the ordering is arbitrary, but fixed for indexing.
	orderedConditions := runner.orderConditions(genesOfInterest)

	// get the score offset and the score indices for the different score types
	scoreOffset, pathTypeScoreIdxs := runner.offsetAndScoreIdxs()

	// read all networks from the optimization directory
	networks := runner.nwr.ReadAllNetworks(runner.OptimizationDirectory())
	// Calculate the rank of each network for each objective
	rankMaps, topScores := createAllRankMaps(networks)

	// run the interpretation for each score type
	if runner.OptimizeSampleCount {
		runner.runPerScoreType(
			networks,
			rankMaps,
			topScores,
			"sample-rank",
			genesOfInterest,
			geneNameMap,
			orderedConditions,
			scoreSelector(scoreOffset-1),
			invRankSelector(pathTypeScoreIdxs),
		)
		runner.runPerScoreType(
			networks,
			rankMaps,
			topScores,
			"sample-norm",
			genesOfInterest,
			geneNameMap,
			orderedConditions,
			scoreSelector(scoreOffset-1),
			normalizedScoreSelector(pathTypeScoreIdxs),
		)
	}
	runner.runPerScoreType(
		networks,
		rankMaps,
		topScores,
		"ranksum",
		genesOfInterest,
		geneNameMap,
		orderedConditions,
		invRankSelector(pathTypeScoreIdxs),
	)
	runner.runPerScoreType(
		networks,
		rankMaps,
		topScores,
		"normsum",
		genesOfInterest,
		geneNameMap,
		orderedConditions,
		normalizedScoreSelector(pathTypeScoreIdxs),
	)
	for pathTypeIdx, pathType := range runner.PathTypes {
		scoreIdx := scoreOffset + pathTypeIdx
		runner.runPerScoreType(
			networks,
			rankMaps,
			topScores,
			pathType,
			genesOfInterest,
			geneNameMap,
			orderedConditions,
			scoreSelector(scoreIdx),
		)
	}
}

func (runner interpretationRunner) offsetAndScoreIdxs() (int, []int) {
	scoreOffset := 0
	if runner.OptimizeNetworkSize {
		scoreOffset += 1
	}
	if runner.OptimizeSampleCount {
		scoreOffset += 1
	}
	pathTypeScoreIdxs := make([]int, len(runner.PathTypes))
	for i := range runner.PathTypes {
		pathTypeScoreIdxs[i] = scoreOffset + i
	}
	return scoreOffset, pathTypeScoreIdxs
}

type scoreSummarizer func(*graph.Network, []map[float64]int, []float64) float64

func scoreSelector(scoreIdx int) scoreSummarizer {
	return func(network *graph.Network, _ []map[float64]int, _ []float64) float64 {
		return network.Scores()[scoreIdx]
	}
}

func invRankSelector(scoreIdxs []int) scoreSummarizer {
	return func(network *graph.Network, rankMaps []map[float64]int, _ []float64) float64 {
		totalRank := 0
		for _, scoreIdx := range scoreIdxs {
			rank := rankMaps[scoreIdx][network.Scores()[scoreIdx]]
			totalRank += rank
		}
		return 1 / float64(totalRank)
	}
}

func normalizedScoreSelector(scoreIdxs []int) scoreSummarizer {
	return func(network *graph.Network, _ []map[float64]int, topScores []float64) float64 {
		normalizedScoreSum := 0.0
		for _, scoreIdx := range scoreIdxs {
			normalizedScoreSum += network.Scores()[scoreIdx] / topScores[scoreIdx]
		}
		return normalizedScoreSum
	}
}

func getRanks(network *graph.Network, rankMaps []map[float64]int) []int {
	ranks := make([]int, len(rankMaps))
	for idx := range rankMaps {
		ranks[idx] = rankMaps[idx][network.Scores()[idx]]
	}
	return ranks
}

func getFractionOfTopScores(network *graph.Network, topScores []float64) []float64 {
	fractions := make([]float64, len(topScores))
	scores := network.Scores()
	for idx := range topScores {
		fractions[idx] = scores[idx] / topScores[idx]
	}
	return fractions
}

func compareNetworksByScore(
	networks []*graph.Network,
	scoreSummarizers []scoreSummarizer,
	rankMaps []map[float64]int,
	topScores []float64,
) func(i, j int) bool {
	return func(i, j int) bool {
		for _, score := range scoreSummarizers {
			scoreI := score(networks[i], rankMaps, topScores)
			scoreJ := score(networks[j], rankMaps, topScores)
			if scoreI != scoreJ {
				return scoreI > scoreJ
			}
		}
		return false
	}
}

func (runner interpretationRunner) runPerScoreType(
	networks []*graph.Network,
	rankMaps []map[float64]int,
	topScores []float64,
	scoreType string,
	genesOfInterest map[string]interpretation.GenesOfInterest,
	geneNameMap types.GeneTranslationMap,
	orderedConditions types.Conditions,
	scoreSummarizers ...scoreSummarizer,
) {
	// Use all remaining highest scoring subnetworks to produce a ranking of edges
	// based on the highest edge cost for which the edge was found.
	network := runner.selectNetwork(networks, rankMaps, topScores, scoreSummarizers...)

	runner.Info("Selected network",
		"scoreType", scoreType,
		"network node count", len(network.Genes()),
		"network edge count", network.InteractionCount(),
		"network scores", network.Scores(),
		"network ranks", getRanks(network, rankMaps),
		"network top score fractions", getFractionOfTopScores(network, topScores),
	)

	// create map[types.Interaction][]bool
	// representing whether an interaction is present in a condition-specific path in the subnetwork
	edgesInCondition, pathsInCondition, _ := runner.edgesInCondition(
		orderedConditions,
		network.Interactions(),
	)

	// create condition specific ranking of genes of interest
	conditionSpecificGeneRanking := interpretation.ConditionSpecificRanking(
		orderedConditions,
		pathsInCondition,
	)

	runner.writeResults(
		scoreType,
		geneNameMap,
		conditionSpecificGeneRanking,
		orderedConditions,
		genesOfInterest,
		edgesInCondition,
		network.Interactions(),
	)
}

func (runner interpretationRunner) writeResults(
	scoreType string,
	geneNameMap types.GeneTranslationMap,
	conditionSpecificGeneRanking map[string]map[types.Condition]map[types.GeneID]float64,
	orderedConditions types.Conditions,
	genesOfInterest map[string]interpretation.GenesOfInterest,
	edgesInCondition map[types.InteractionID][]float64,
	subnetwork types.InteractionIDSet,
) {
	// write the results to the results directory

	// create results directory and remove old files if any exist
	resultsDirectory := runner.ResultsDirectory(scoreType)
	runner.Info("Initializing and creating results folder")
	fileio.CreateEmptyDir(resultsDirectory)

	// Write the resulting subnetwork. Each file has different properties for easy visualizing of the solution

	// Use the edge ranks to write the resulting subnetwork (which is the union of all edges present in all highest scoring subnetworks) and annotate the edges by their ranks.
	err := runner.WriteWeightedSubnetwork(
		runner.WeightedNetworkFile(resultsDirectory),
		edgesInCondition,
		geneNameMap,
	)
	if err != nil {
		runner.Error("error writing weighted subnetwork", "err", err)
	}

	// write the condition specific gene ranking to file
	err = runner.WriteConditionSpecificRanking(
		runner.FileWriter,
		conditionSpecificGeneRanking,
		orderedConditions,
		resultsDirectory,
		geneNameMap,
	)
	if err != nil {
		runner.Error("error writing condition specific ranking", "err", err)
	}
	// TODO: write a sif file for the resulting subnetwork
	// TODO: write XGMML file for resulting subnetwork
	// write HTML visualization for resulting subnetwork
	runner.WriteD3JS(
		orderedConditions,
		runner.EtcPathAsString,
		geneNameMap,
		resultsDirectory,
		genesOfInterest,
		subnetwork,
	)
	runner.Info("interpretation finished")
}

// edgesInCondition dispatches the calculation of edges in condition to the interpretation package
func (runner interpretationRunner) edgesInCondition(
	orderedConditions types.Conditions,
	interactions types.InteractionIDSet,
) (
	map[types.InteractionID][]float64,
	map[string]map[types.Condition]map[types.GeneID][2]float64,
	map[string]map[types.Condition]map[types.GeneID]struct{},
) {
	return interpretation.EdgesInCondition(
		runner.Logger,
		runner.GeneIDMap,
		orderedConditions,
		interactions,
		runner.PathTypes,
		runner.SldCutoff,
		runner.PathsFile,
		runner.MaxPaths,
	)
}

// selectNetwork reads all networks in the optimization directory and determines the optimal network based on the given
// score indices and score summarizers.
func (runner interpretationRunner) selectNetwork(
	networks []*graph.Network,
	rankMaps []map[float64]int,
	topScores []float64,
	scoreSummarizers ...scoreSummarizer,
) *graph.Network {
	// Sort the networks by their scores
	sort.Slice(networks, compareNetworksByScore(networks, scoreSummarizers, rankMaps, topScores))

	selectedNetwork := networks[0]

	return selectedNetwork
}

func createAllRankMaps(networks []*graph.Network) ([]map[float64]int, []float64) {
	objectiveCount := len(networks[0].Scores())
	rankMaps := make([]map[float64]int, objectiveCount)
	topScores := make([]float64, objectiveCount)
	for scoreIdx := range objectiveCount {
		scores := make([]float64, len(networks))
		for i, network := range networks {
			scores[i] = network.Scores()[scoreIdx]
		}
		rankMaps[scoreIdx], topScores[scoreIdx] = createRankMap(scores)
	}
	return rankMaps, topScores
}

func createRankMap(scores []float64) (map[float64]int, float64) {
	sort.Float64s(scores)
	rankMap := make(map[float64]int)
	topScore := scores[len(scores)-1]
	rank := 1
	for i := len(scores) - 1; i >= 0; i-- {
		if i < len(scores)-1 && scores[i] != scores[i+1] {
			rank = len(scores) - i
		}
		rankMap[scores[i]] = rank
	}
	return rankMap, topScore
}

func (runner interpretationRunner) orderConditions(
	genesOfInterest map[string]interpretation.GenesOfInterest,
) types.Conditions {
	conditions := interpretation.ConditionsOfInterest(genesOfInterest)
	orderedConditions := make(types.Conditions, 0, len(conditions))
	for condition := range conditions {
		orderedConditions = append(orderedConditions, condition)
	}
	sort.Sort(orderedConditions)
	return orderedConditions
}
