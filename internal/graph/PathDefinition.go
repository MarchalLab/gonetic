package graph

import (
	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

type PathDefinition func(path Path) bool

// Any path is a valid simple path. Used in the QTL case.
func SimplePathDefinition(_ Path) bool {
	return true
}

// BidirectionalRegulatoryUpDownstreamPathDefinition determines if a path first goes upstream and then goes downstream
// this is used to identify common regulatory mechanisms between differentially expressed genes
func BidirectionalRegulatoryUpDownstreamPathDefinition(path Path) bool {
	if path.Direction == types.UndirectedPath ||
		path.Direction == types.UpstreamPath ||
		path.Direction == types.UpDownstreamPath ||
		path.Direction == types.DownstreamPath {
		// some UpDownstreamPaths are classified as UpstreamPaths or DownstreamPaths or UndirectedPaths, because the Up/Down-part uses bidirectional edges
		return arguments.GlobalInteractionStore.IsRegulatoryInteraction(path.LastInteraction()) &&
			arguments.GlobalInteractionStore.IsRegulatoryInteraction(path.FirstInteraction())
	}
	return false
}

// Any path starting with a regulatory interaction is a valid path. Used in the EQTL case.
func InvertedRegulatoryPathDefinition(path Path) bool {
	return arguments.GlobalInteractionStore.IsRegulatoryInteraction(path.FirstInteraction())
}
