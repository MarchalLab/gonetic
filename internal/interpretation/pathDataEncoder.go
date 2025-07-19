package interpretation

import (
	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

type pathDataEncoder struct {
	*arguments.Common
	GeneMap map[types.GeneID]struct{}
	Paths   map[string]map[types.CNFHeader][]*types.CompactPath
}

func encodeNetworkPaths(
	args *arguments.Common,
	interactions types.InteractionIDSet,
) *pathDataEncoder {
	subnetwork := types.NewInteractionIDSet()
	for interaction := range interactions {
		subnetwork.Set(interaction)
	}
	enc := &pathDataEncoder{
		Common:  args,
		GeneMap: make(map[types.GeneID]struct{}),
	}
	enc.gatherGenes(subnetwork)
	enc.gatherPathsInSubnetwork(subnetwork)
	return enc
}

func (enc *pathDataEncoder) gatherGenes(subnetwork types.InteractionIDSet) {
	enc.GeneMap = make(map[types.GeneID]struct{})
	for interactionID := range subnetwork {
		enc.GeneMap[interactionID.From()] = struct{}{}
		enc.GeneMap[interactionID.To()] = struct{}{}
	}
}

func (enc *pathDataEncoder) gatherPathsInSubnetwork(
	subnetwork types.InteractionIDSet,
) {
	enc.Paths = make(map[string]map[types.CNFHeader][]*types.CompactPath)
	for _, pathType := range enc.PathTypes {
		enc.Paths[pathType] = make(map[types.CNFHeader][]*types.CompactPath)
		enc.gatherTypedPathsInSubnetwork(subnetwork, pathType)
	}
}

func (enc *pathDataEncoder) gatherTypedPathsInSubnetwork(
	subnetwork types.InteractionIDSet,
	pathType string,
) {
	allPaths := readers.ReadPathList(
		enc.Logger,
		enc.GeneIDMap,
		enc.MaxPaths,
		pathType,
		enc.SldCutoff,
		enc.PathsFile(pathType),
	)
	for header, paths := range allPaths {
		if _, ok := enc.GeneMap[header.Gene]; !ok {
			continue
		}
		for _, path := range paths {
			if pathInSubnetwork(path.InteractionSet(), subnetwork) {
				enc.Paths[pathType][header] = append(enc.Paths[pathType][header], path)
			}
		}
	}
}
