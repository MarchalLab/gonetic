package pathfinding

import (
	"log/slog"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/editor"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/readers"
)

// makeNetwork creates the network for path finding
func makeNetwork(args *arguments.Common, minEdgeScore float64) *graph.Network {
	weightingAddition, skipWeighting := editor.WeightingAddition(
		args.Logger,
		editor.FromOnly,
		args.TopologyWeightingAddition,
		"Invalid argument --topology-weighting-addition %s",
	)
	// initialize network with initial probability of 0 or 1 for each edge
	nwr := readers.NewInitialNetworkReader(args.Logger, args.GeneIDMap)
	network := nwr.NewNetworkFromFiles(args.NetworkFiles, args.BannedNetworkFiles, false, true)
	// There are no undirected edges in the network, undirected edges are represented as two directed edges.
	// There are no duplicates in the network. Undirected interactions which are each other's reverse can not exist, since `from <lex to` by design.
	// There are no self edges in the network.
	// Weight on network topology (sigmoidal)
	if !skipWeighting {
		network = editor.NetworkWeighting(
			editor.NewSigmoidalDegreeWeight(nwr.NewNetworkFromFiles(args.NetworkFiles, args.BannedNetworkFiles, false, true), 0.01).Score,
			weightingAddition,
			network,
		)
	}
	// remove low scoring edges
	if minEdgeScore > 0.0 {
		network = editor.RemoveLowScoringEdges(network, minEdgeScore)
	}
	return network
}

func expressionNetwork(args *arguments.Expression, data map[types.GeneID]float64, isEQTL bool) *graph.Network {
	network := makeNetwork(args.Common, args.MinEdgeScore)
	if args.ExpressionWeightingMethod != "none" {
		network = expressionWeighting(network, args, data, isEQTL)
	}
	// remove low scoring edges after weighting
	network = editor.RemoveLowScoringEdges(network, args.MinEdgeScore)
	args.Info(
		"network size",
		"genes", len(network.Genes()),
		"interactions", network.InteractionCount(),
		"interaction types", len(network.InteractionTypes()),
	)
	return network
}

func expressionWeighting(
	network *graph.Network,
	args *arguments.Expression,
	data map[types.GeneID]float64,
	isEQTL bool,
) *graph.Network {
	// determine addition method based on settings
	weightTarget := editor.BothBayes
	if isEQTL {
		weightTarget = editor.ToOnly
	}
	var weightingAddition, skipWeighting = editor.WeightingAddition(
		args.Logger,
		weightTarget,
		args.ExpressionWeightingAddition,
		"Invalid argument --expression-weighting-addition %s",
	)
	if skipWeighting {
		return network
	}
	// determine weighting method based on settings
	var weightingMethod func(*slog.Logger, map[types.GeneID]float64, float64) editor.ExpressionWeight
	switch args.ExpressionWeightingMethod {
	case "lfc":
		weightingMethod = editor.NewLfcExpressionWeight
	case "zscore":
		weightingMethod = editor.NewZscoreExpressionWeight
	default:
		args.Error("Invalid argument --expression-weighting-method", "ExpressionWeightingMethod", args.ExpressionWeightingMethod)
		return network
	}
	// compute gene weights
	weights := weightingMethod(args.Logger, data, args.ExpressionWeightingDefault)
	// print weights
	if args.PrintGeneScoreMap {
		weights.PrintScoreMap()
	}
	// weight network
	network = editor.NetworkWeighting(
		weights.Score,
		weightingAddition,
		network,
	)
	return network
}

func qtlNetwork(args *arguments.Common) *graph.Network {
	network := makeNetwork(args, args.MinEdgeScore)
	args.Info(
		"network size",
		"genes", len(network.Genes()),
		"interactions", network.InteractionCount(),
		"interaction types", len(network.InteractionTypes()),
	)
	return network
}
