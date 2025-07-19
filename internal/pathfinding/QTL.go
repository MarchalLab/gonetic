package pathfinding

import (
	"fmt"
	"strings"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/pathweight"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func qtlPrep(pathType string, mutFileData readers.FileData, commonArgs *arguments.Common, qtlArgs *arguments.QTLSpecific) (*graph.Network, float64, types.ConditionSet, map[string]struct{}, map[types.Condition]types.GeneSet, types.GeneConditionMap[float64]) {
	// Load in the freq increase data, if this is wanted for use
	freqIncreaseDataPerLine := make(map[string]struct{})
	if qtlArgs.FreqIncrease {
		if _, ok := mutFileData.Headers["freq increase"]; !ok {
			commonArgs.Error("Mutation data file does not contain frequency increase header while this feature is marked for being used.")
		} else {
			freqIncreaseDataPerLine = readers.LoadData(mutFileData, "gene name", "freq increase", "condition")
		}
	}
	commonArgs.Info("frequency increase data per line entries", "entries", len(freqIncreaseDataPerLine))

	// Load in the functional score data, if this is wanted for use
	functionalScoreDataPerLine := make(map[string]struct{})
	if qtlArgs.FuncScore {
		if _, ok := mutFileData.Headers["functional score"]; !ok {
			commonArgs.Info("Mutation data file does not contain functional score header while this feature is marked for being used.")
		} else {
			if _, ok := mutFileData.Headers["synonymous"]; ok {
				functionalScoreDataPerLine = readers.LoadData(mutFileData, "gene name", "functional score", "condition", "synonymous")
			} else {
				functionalScoreDataPerLine = readers.LoadData(mutFileData, "gene name", "functional score", "condition")
			}
		}
	}
	commonArgs.Info("functional score data per line entries", "entries", len(functionalScoreDataPerLine))

	// Load in the different strains
	conditions := readers.LoadConditionData(mutFileData, "condition")
	// Load in the mutation data
	mutatedGenes := readers.LoadData(mutFileData, "gene name", "condition")
	// Load in the mutation data per condition
	mutationPerCondition := readers.MakeGeneMap(commonArgs.GeneIDMap, mutFileData)

	// network
	network := qtlNetwork(commonArgs)
	commonArgs.Info(
		"network size",
		"genes", len(network.Genes()),
		"interactions", network.InteractionCount(),
		"interaction types", len(network.InteractionTypes()),
	)
	if !commonArgs.SkipNetworkPrinting {
		err := commonArgs.WriteStringLinerToFile(
			"qtl",
			commonArgs.PathsFileWithName(pathType, "qtl.network"),
			network,
		)
		if err != nil {
			commonArgs.Error(
				"Failed to write QTL network",
				"error", err,
			)
		}
	}

	// Setup path weighting values
	mutatorOutlierValues := map[types.Condition]float64{}
	if qtlArgs.Correction {
		pathweight.CalculateMutatorOutlierWeighting(commonArgs.Logger, mutatedGenes, qtlArgs.OutlierPopulations, conditions, qtlArgs.MutRateParam, qtlArgs.Correction)
	}
	frequencyDataWeighting := pathweight.FrequencyWeighting{}
	if qtlArgs.FreqIncrease {
		frequencyDataWeighting = pathweight.NewFrequencyWeighting(commonArgs.Logger, freqIncreaseDataPerLine, conditions, qtlArgs.FreqCutoff, qtlArgs.FreqIncrease)
	}
	functionalDataWeighting := pathweight.FunctionalWeighting{}
	if qtlArgs.FuncScore {
		pathweight.NewFunctionalWeighting(commonArgs.Logger, functionalScoreDataPerLine, qtlArgs.FuncScore)
	}
	// Calculate the relevance score for each gene, missing data in frequency or functional data gets assigned the mean value for frequency or functional data respectively.
	weightsPerGene := pathweight.RelevanceScorePerGene(
		commonArgs.GeneIDMap,
		functionalScoreDataPerLine,
		freqIncreaseDataPerLine,
		mutatedGenes,
		functionalDataWeighting,
		frequencyDataWeighting,
		mutatorOutlierValues,
	)

	// The sldCutoff is the minimal probability a path must have in order to be retained.
	// Setting this reduces the path finding time because paths through hubs do not need to be evaluated.
	sldCutoff := commonArgs.SldCutoff
	if sldCutoff < 0 {
		sldCutoff = pathweight.SldCutoffPrediction(network)
	}
	commonArgs.Info("sldCutoff", "sldCutoff", sldCutoff)

	return network, sldCutoff, conditions, mutatedGenes, mutationPerCondition, weightsPerGene
}

func qtl(pathType string, mutFileData readers.FileData, qtlArgs *arguments.QTLSpecific, args *arguments.Common) {
	args.Info("Running QTL pathfinding")

	// timing
	startTime := time.Now()
	defer func() {
		passedTime := time.Now().Sub(startTime).Seconds()
		args.Info("Finished path finding", "seconds", passedTime)
	}()

	// Data preparation
	network, sldCutoff, conditions, mutatedGenes, mutationPerCondition, weightsPerGene := qtlPrep(pathType, mutFileData, args, qtlArgs)
	args.WriteGeneMapFile()
	args.WriteInteractionTypeMapFile()

	// Path Finding
	// TODO: fix comments
	// Build this specific run using the general run scheme
	// Expand downstream. A simple path definition would be problematic as
	expander := graph.NewDownstreamExpander(network)
	// The path definition is simple as regulatory path is of no interest.
	// The sldCutoff is the minimal probability a path must have in order to be retained. Setting this reduces the path
	// finding time because paths through hubs do not need to be evaluated.
	search := newPathFinder(args.Logger, expander, graph.SimplePathDefinition, sldCutoff)

	// Search for paths every time starting from a specific mutated gene from a specific line to the N-best mutated genes from OTHER lines. (So all lines (experiments) are used in one run here)
	// Note that the same gene can be an end point twice if it is mutated in two other lines. Doing so frequently mutated genes get selected more often as more overlapping paths will be found.
	// A cutoff can be defined in order to avoid assessing mutated genes with very low weights as this takes up a lot of time while the found paths will not be relevant.
	args.Info("Start processing", "samples", len(conditions), "parallelism", args.NumCPU)
	for condition := range conditions {
		// paths go from current condition to any other condition
		endGenes := make(types.GeneSet)
		for mutationEntry := range mutatedGenes {
			split := strings.Split(mutationEntry, "\t")
			if len(split) == 0 {
				continue
			}
			if types.Condition(split[len(split)-1]) != condition {
				endGene := args.SetName(split[0])
				endGenes[endGene] = struct{}{}
			}
		}
		// path finding object
		conditionPath := args.PathsFileWithName(pathType, fmt.Sprintf("%s.paths", condition))
		run := newRunner(
			args,
			mutationPerCondition[condition],
			endGenes,
			conditions,
			types.ConditionSet{condition: {}},
			weightsPerGene, // Add the weights for mutated genes in order to weigh the paths.
			search,
			conditionPath,      // The name of the output files
			args.PathLength,    // The maximum PathLength to be explored
			args.BestPathCount, // The number of paths from a start node. In theory more is better but the optimization step gets harder in that case. 25 is a realistic value.
			condition,          // Name of the strain from which a path is found. Needed as otherwise paths starting from the same gene would be overwritten in the optimization step.
		)
		run.findPaths(true, true)
		if qtlArgs.WithinCondition {
			// path finding object
			withinConditionPath := args.PathsFileWithName(pathType, fmt.Sprintf("%s.within.paths", condition))
			withinRun := newRunner(
				args,
				mutationPerCondition[condition],
				mutationPerCondition[condition],
				types.ConditionSet{condition: {}},
				types.ConditionSet{}, // don't have to exclude any, since only the current condition is allowed anyway
				weightsPerGene,       // Add the weights for mutated genes in order to weigh the paths.
				search,
				withinConditionPath, // The name of the output files
				args.PathLength,     // The maximum PathLength to be explored
				args.BestPathCount,  // The number of paths from a start node. In theory more is better but the optimization step gets harder in that case. 25 is a realistic value.
				condition,           // Name of the strain from which a path is found. Needed as otherwise paths starting from the same gene would be overwritten in the optimization step.
			)
			withinRun.findPaths(true, true)
		}
	}
	// sync go routines
	args.Sem.Wait()
	// Combine the found .paths files into one large .paths file to optimize.
	combinePathFiles(args.FileWriter, args.PathsDirectory(pathType), args.PathsFile(pathType))
	// Write the relevance scores
	writeWeights(args.FileWriter, args.WeightsFile(pathType, ""), weightsPerGene)
}
