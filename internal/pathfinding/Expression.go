package pathfinding

import (
	"fmt"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/pathweight"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func expression(pathType string, expressionFileData, differentialExpressionFileData readers.FileData, args *arguments.Expression) {
	args.Info("Running expression path finding")

	// timing
	startTime := time.Now()
	defer func() {
		passedTime := time.Now().Sub(startTime).Seconds()
		args.Info("Finished path finding", "seconds", passedTime)
	}()

	// Data preparation
	// Load in the different strains
	conditions := readers.LoadConditionData(differentialExpressionFileData, "condition")
	// make map of expression data per condition per gene
	expressionPerCondition := readers.MakeExpressionMap(args.Logger, args.GeneIDMap, expressionFileData, args.ExpressionWeightingMethod)
	// make map of differentially expressed genes per condition
	dePerCondition := readers.MakeGeneMap(args.GeneIDMap, differentialExpressionFileData)

	// Path Finding
	// Search for paths every time starting from a specific mutated gene from a specific line to the N-best mutated genes from OTHER lines. (So all lines (experiments) are used in one run here)
	// Note that the same gene can be an end point twice if it is mutated in two other lines. Doing so frequently mutated genes get selected more often as more overlapping paths will be found.
	// A cutoff can be defined in order to avoid assessing mutated genes with very low weights as this takes up a lot of time while the found paths will not be relevant.
	args.Info("Processing samples in parallel.", "samples", len(conditions), "parallelism", args.NumCPU)
	for condition := range conditions {
		// create network for condition
		network := expressionNetwork(args, expressionPerCondition[condition], false)
		if !args.SkipNetworkPrinting {
			args.Info(
				"network size",
				"genes", len(network.Genes()),
				"interactions", network.InteractionCount(),
				"interaction types", len(network.InteractionTypes()),
			)
			err := args.WriteStringLinerToFile(
				string(condition),
				args.PathsFileWithName(pathType, fmt.Sprintf("%s.network", condition)),
				network,
			)
			if err != nil {
				args.Error(
					"Failed to write expression network",
					"error", err,
					"condition", condition,
				)
			}
		}

		// downstream mode (default): find a shared node downstream
		// upstream mode: find a shared node upstream that has a regulatory impact on the nodes of interest
		var expander graph.PathExpander
		var pathDefinition graph.PathDefinition
		if args.DownUpstream {
			expander = graph.NewDownUpstreamExpander(network)
			pathDefinition = graph.SimplePathDefinition
		} else {
			expander = graph.NewRegulatoryUpDownstreamExpander(network)
			pathDefinition = graph.BidirectionalRegulatoryUpDownstreamPathDefinition
		}
		// The sldCutoff is the minimal probability a path must have in order to be retained. Setting this reduces the path
		// finding time because paths through hubs do not need to be evaluated.
		sldCutoff := args.SldCutoff
		if sldCutoff < 0 {
			sldCutoff = pathweight.SldCutoffPrediction(network)
		}
		args.Info("sldCutoff", "sldCutoff", sldCutoff)
		// create search object
		search := newPathFinder(args.Logger, expander, pathDefinition, sldCutoff)

		conditionPath := args.PathsFileWithName(pathType, fmt.Sprintf("%s.paths", condition))
		run := newRunner(
			args.Common,
			dePerCondition[condition],
			dePerCondition[condition],
			conditions,
			nil,
			nil,
			search,
			conditionPath,      // The name of the output files
			args.PathLength,    // The maximum PathLength to be explored
			args.BestPathCount, // The number of paths from a start node. In theory more is better but the optimization step gets harder in that case. 25 is a realistic value.
			condition,          // Name of the strain from which a path is found. Needed as otherwise paths starting from the same gene would be overwritten in the optimization step.
		)
		run.findPaths(false, false)
	}
	// sync go routines
	args.Sem.Wait()
	// Combine the found .paths files into one large .paths file to optimize.
	combinePathFiles(args.FileWriter, args.PathsDirectory(pathType), args.PathsFile(pathType))
}
