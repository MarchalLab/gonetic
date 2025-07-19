package graph

import (
	"errors"

	"github.com/MarchalLab/gonetic/internal/common/arguments"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

type PathExpander interface {
	Expand(path *Path) ([]*Path, error)
	CreatePathFrom(
		interactions []types.InteractionID,
		startGene types.GeneID,
		fromScore, toScore float64,
	) (*Path, error)
	Probabilities() *types.ProbabilityMap
}

type GenericExpander struct {
	network *Network
}

func NewGenericExpander(network *Network) *GenericExpander {
	return &GenericExpander{
		network: network,
	}
}

func (expander *GenericExpander) expandWIP(path *Path, interactionsToExpandList ...types.InteractionIDSet) ([]*Path, error) {
	newList := make([]*Path, 0)
	for _, interactionsToExpand := range interactionsToExpandList {
		for interactionID := range interactionsToExpand {
			newEnd, err := interactionID.OtherEndGene(path.EndGene)
			if err != nil {
				return newList, err
			}
			if path.hasGene(newEnd) {
				// paths should not contain loops
				continue
			}
			prob := path.probability * expander.network.probabilities.GetProbability(interactionID)
			direction := ExtendWithInteraction(interactionID, path)
			if direction == types.InvalidPathDirection {
				return newList, errors.New("invalid path direction while extending path")
			}
			newList = append(newList,
				ExtendPath(path, interactionID, prob, direction),
			)
		}
	}
	return newList, nil
}

// CreatePathFrom creates a path from the given interactions, starting from the given gene
func (expander *GenericExpander) CreatePathFrom(
	interactions []types.InteractionID,
	startGene types.GeneID,
	fromScore, toScore float64,
) (*Path, error) {
	currentPath := RootPath(startGene, fromScore)
	for _, interactionID := range interactions {
		tmpMap := types.NewInteractionIDSet()
		tmpMap.Set(interactionID)
		stack, err := expander.expandWIP(currentPath, tmpMap)
		if err != nil || len(stack) != 1 {
			return currentPath, err
		}
		currentPath = stack[0]
	}
	currentPath.toScore = toScore
	return currentPath, nil
}

// Probabilities returns the probability map of the expander
func (expander *GenericExpander) Probabilities() *types.ProbabilityMap {
	return expander.network.probabilities
}

type DownstreamExpander struct {
	*GenericExpander
}

func (expander *DownstreamExpander) Expand(path *Path) ([]*Path, error) {
	if !path.isExtended {
		path.Direction = types.DownstreamPath
	}
	return expander.expandWIP(path, expander.network.Outgoing()[path.EndGene])
}

func NewDownstreamExpander(network *Network) *DownstreamExpander {
	return &DownstreamExpander{
		NewGenericExpander(network),
	}
}

type UpstreamExpander struct {
	*GenericExpander
}

func (expander *UpstreamExpander) Expand(path *Path) ([]*Path, error) {
	if !path.isExtended {
		path.Direction = types.UpstreamPath
	}
	return expander.expandWIP(path, expander.network.Incoming()[path.EndGene])
}
func NewUpstreamExpander(network *Network) *UpstreamExpander {
	return &UpstreamExpander{
		NewGenericExpander(network),
	}
}

type SimpleExpander struct {
	*GenericExpander
}

func (expander *SimpleExpander) Expand(path *Path) ([]*Path, error) {
	return expander.expandWIP(path, expander.network.Incoming()[path.EndGene], expander.network.Outgoing()[path.EndGene])
}
func NewSimpleExpander(network *Network) *SimpleExpander {
	return &SimpleExpander{
		NewGenericExpander(network),
	}
}

type DownUpstreamExpander struct {
	*GenericExpander
}

func (expander *DownUpstreamExpander) Expand(path *Path) ([]*Path, error) {
	// first move is down stream
	if !path.isExtended {
		path.Direction = types.DownstreamPath
		return expander.expandWIP(path, expander.network.Outgoing()[path.EndGene])
	}
	// if last move was downstream, both directions are possible
	if path.Direction == types.DownstreamPath {
		return expander.expandWIP(path, expander.network.Incoming()[path.EndGene], expander.network.Outgoing()[path.EndGene])
	}
	// if last move was upstream, continue upstream
	return expander.expandWIP(path, expander.network.Incoming()[path.EndGene])
}
func NewDownUpstreamExpander(network *Network) *DownUpstreamExpander {
	return &DownUpstreamExpander{
		NewGenericExpander(network),
	}
}

type UpDownstreamExpander struct {
	*GenericExpander
}

func (expander *UpDownstreamExpander) Expand(path *Path) ([]*Path, error) {
	// first move is up stream
	if !path.isExtended {
		path.Direction = types.UpstreamPath
		return expander.expandWIP(path, expander.network.Incoming()[path.EndGene])
	}
	// if last move was upstream, both directions are possible
	if path.Direction == types.UpstreamPath {
		return expander.expandWIP(path, expander.network.Incoming()[path.EndGene], expander.network.Outgoing()[path.EndGene])
	}
	// if last move was downstream, continue downstream
	return expander.expandWIP(path, expander.network.Outgoing()[path.EndGene])
}
func NewUpDownstreamExpander(network *Network) *UpDownstreamExpander {
	return &UpDownstreamExpander{
		NewGenericExpander(network),
	}
}

// The RegulatoryUpDownstreamExpander expands the path first upstream and then downstream with the additional boundary condition that
// the first upstream expansion is regulatory and the last expansion is regulatory.

type RegulatoryUpDownstreamExpander struct {
	*GenericExpander
}

func (expander *RegulatoryUpDownstreamExpander) Expand(path *Path) ([]*Path, error) {
	// first move is up stream
	if !path.isExtended {
		path.Direction = types.UpstreamPath
		return expander.expandWIP(path, expander.network.Incoming()[path.EndGene])
	}
	// first expansion has to be regulatory (last expansion also, but we can't check this here)
	if !arguments.GlobalInteractionStore.IsRegulatoryInteraction(path.FirstInteraction()) {
		return nil, nil
	}
	// if last move was upstream, both directions are possible
	if path.Direction == types.UpstreamPath {
		return expander.expandWIP(path, expander.network.Incoming()[path.EndGene], expander.network.Outgoing()[path.EndGene])
	}
	// if last move was downstream, continue downstream
	return expander.expandWIP(path, expander.network.Outgoing()[path.EndGene])
}
func NewRegulatoryUpDownstreamExpander(network *Network) *RegulatoryUpDownstreamExpander {
	return &RegulatoryUpDownstreamExpander{
		NewGenericExpander(network)}
}

// ExtendWithInteraction determines the direction of the new path when extended with the given interaction
func ExtendWithInteraction(current types.InteractionID, previousPath *Path) types.PathDirection {
	previous := previousPath.Direction
	if previousPath.EndGene == current.To() {
		// moving up stream
		switch previous {
		case types.UndirectedPath:
			return types.UpstreamPath
		case types.UpstreamPath:
			return types.UpstreamPath
		case types.DownstreamPath:
			return types.DownUpstreamPath
		case types.UpDownstreamPath:
			return types.InvalidPathDirection
		case types.DownUpstreamPath:
			return types.DownUpstreamPath
		}
	}
	if previousPath.EndGene == current.From() {
		// moving down stream
		switch previous {
		case types.UndirectedPath:
			return types.DownstreamPath
		case types.UpstreamPath:
			return types.UpDownstreamPath
		case types.DownstreamPath:
			return types.DownstreamPath
		case types.UpDownstreamPath:
			return types.UpDownstreamPath
		case types.DownUpstreamPath:
			return types.InvalidPathDirection
		}
	}
	return types.InvalidPathDirection
}
