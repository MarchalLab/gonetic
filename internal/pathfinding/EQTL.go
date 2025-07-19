package pathfinding

import (
	"fmt"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func eqtl(pathType string, mutFileData, expressionFileData, differentialExpressionFileData readers.FileData, args *arguments.EQTL) {
	args.Info("Running EQTL pathfinding")

	// timing
	startTime := time.Now()
	defer func() {
		passedTime := time.Now().Sub(startTime).Seconds()
		args.Info("Finished path finding", "seconds", passedTime)
	}()

	// Data preparation
	network, sldCutoff, conditions, _, mutationPerCondition, weightsPerGene := qtlPrep(pathType, mutFileData, args.Common, args.QTLSpecific)
	args.WriteGeneMapFile()
	args.WriteInteractionTypeMapFile()
	// make map of expression data per condition per gene
	expressionPerCondition := readers.MakeExpressionMap(args.Logger, args.GeneIDMap, expressionFileData, args.ExpressionWeightingMethod)
	// make map of differentially expressed genes per condition
	dePerCondition := readers.MakeGeneMap(args.GeneIDMap, differentialExpressionFileData)

	// Path Finding
	args.Info("Start processing", "samples", len(conditions), "parallelism", args.NumCPU)
	for condition := range conditions {
		// create network for condition
		if args.ExpressionWeightingMethod != "none" {
			network = expressionNetwork(args.Expression, expressionPerCondition[condition], true)
			if !args.SkipNetworkPrinting {
				err := args.WriteStringLinerToFile(
					string(condition),
					args.PathsFileWithName(pathType, fmt.Sprintf("%s.network", condition)),
					network,
				)
				if err != nil {
					args.Error(
						"Failed to write EQTL network",
						"error", err,
						"condition", condition,
					)
				}
			}
		}
		args.Info("sample network",
			"sample", condition,
			"genes", len(network.Genes()),
			"interactions", network.InteractionCount(),
			"interaction types", len(network.InteractionTypes()),
		)

		// Depending on a user-defined parameter, accepted paths are whatever (simple) or start with a regulatory edge, enforcing the mutation to be connected to a differential expressed gene with a regulatory edge.
		var pathDefinition graph.PathDefinition
		if args.Regulatory {
			pathDefinition = graph.InvertedRegulatoryPathDefinition
		} else {
			pathDefinition = graph.SimplePathDefinition
		}

		// create search object
		expander := graph.NewUpstreamExpander(network)
		search := newPathFinder(args.Logger, expander, pathDefinition, sldCutoff)
		conditionPath := args.PathsFileWithName(pathType, fmt.Sprintf("%s.paths", condition))
		run := newRunner(
			args.Common,
			// if we go from mutation to DEG, we find very few paths because we have very few start genes
			// increasing the path limit per start point is a no-go, since that explodes the knowledge compilation phase
			dePerCondition[condition],       // Start from differentially expressed genes
			mutationPerCondition[condition], // End in mutated genes
			types.ConditionSet{condition: {}},
			types.ConditionSet{},
			weightsPerGene,
			search,
			conditionPath,      // The name of the output files
			args.PathLength,    // The maximum PathLength to be explored
			args.BestPathCount, // The number of paths from a start node. In theory more is better but the optimization step gets harder in that case. 25 is a realistic value.
			condition,          // Name of the strain from which a path is found. Needed as otherwise paths starting from the same gene would be overwritten in the optimization step.
		)
		run.findPaths(true, false)
	}
	// sync go routines
	args.Sem.Wait()
	// Combine the found .paths files into one large .paths file to optimize.
	combinePathFiles(args.FileWriter, args.PathsDirectory(pathType), args.PathsFile(pathType))
	// Write the relevance scores
	writeWeights(args.FileWriter, args.WeightsFile(pathType, ""), weightsPerGene)
}
