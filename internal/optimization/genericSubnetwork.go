package optimization

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/normalform"
)

// *genericSubnetwork partially implements subnetwork, and is a base class for proper implementations of the interface
type genericSubnetwork struct {
	interactionSet     types.InteractionIDSet            // all interactions in the network
	selectedPaths      map[PathID]types.InteractionIDSet // pathId -> all interactions in path
	nonDominationLevel int
	crowdingDistance   float64
	scores             []float64
	opt                *NSGAOptimization
}

func newGenericSubnetwork(opt *NSGAOptimization) *genericSubnetwork {
	return &genericSubnetwork{
		interactionSet:     types.NewInteractionIDSet(),
		selectedPaths:      make(map[PathID]types.InteractionIDSet),
		nonDominationLevel: -1,
		crowdingDistance:   -1,
		scores:             nil,
		opt:                opt,
	}
}

func (network *genericSubnetwork) Scores() []float64 {
	return network.scores
}

func (network *genericSubnetwork) SetScores(scores []float64) {
	network.scores = scores
}

func (network *genericSubnetwork) subnetworkSize() int {
	return network.interactionCount()
}

func (network *genericSubnetwork) NonDominationLevel() int {
	return network.nonDominationLevel
}

func (network *genericSubnetwork) setNonDominationLevel(level int) {
	network.nonDominationLevel = level
}

func (network *genericSubnetwork) CrowdingDistance() float64 {
	return network.crowdingDistance
}

func (network *genericSubnetwork) setCrowdingDistance(dist float64) {
	network.crowdingDistance = dist
}

func (network *genericSubnetwork) SelectedPaths() map[PathID]types.InteractionIDSet {
	return network.selectedPaths
}

func (network *genericSubnetwork) addSelectedPath(pathID PathID, interactions types.InteractionIDSet) {
	network.selectedPaths[pathID] = interactions
}

func (network *genericSubnetwork) Interactions() types.InteractionIDSet {
	return network.interactionSet
}

func (network *genericSubnetwork) interactionCount() int {
	return network.interactionSet.Size()
}

func (network *genericSubnetwork) geneCount() int {
	return network.interactionSet.NetworkNodesSize()
}

func (network *genericSubnetwork) addInteraction(interactionID types.InteractionID) {
	network.interactionSet.Set(interactionID)
}

func (network *genericSubnetwork) expandWithPath(pathID PathID) {
	// check if any overlap, and if so select at random some paths that overlap
	originalInteractions := network.opt.pathRepositories.pathInteractionSetFromId(pathID)
	// copy the interaction set
	result := types.NewInteractionIDSet()
	for interactionID := range originalInteractions {
		result.Set(interactionID)
	}
	// add result to map of paths to interactions
	network.selectedPaths[pathID] = result
	// add result to interactionSet
	for interactionID := range result {
		network.interactionSet.Set(interactionID)
	}
	// invalidate scores
	network.scores = nil
}

func (network *genericSubnetwork) intersect(nnf normalform.NNF) types.InteractionIDSet {
	intersection := types.NewInteractionIDSet()
	for interactionID := range nnf.Values() {
		if network.interactionSet.Has(interactionID) {
			intersection.Set(interactionID)
		}
	}
	return intersection
}

// String is a method to convert a subnetwork to a string
func (network *genericSubnetwork) String() string {
	// gather interactions
	interactions := make([]string, 0, network.Interactions().Size())
	for interactionID := range network.Interactions() {
		interactions = append(interactions, interactionID.StringMinimal())
	}
	sort.Strings(interactions)
	interactionStr := strings.Join(interactions, " ")
	// gather paths
	paths := make([]string, 0, len(network.SelectedPaths()))
	for pathID := range network.SelectedPaths() {
		paths = append(paths, fmt.Sprintf("%d", pathID))
	}
	sort.Strings(paths)
	pathStr := strings.Join(paths, ";")
	// return the stringified network
	return fmt.Sprintf(
		"%s\t%s\t%v",
		interactionStr,
		pathStr,
		network.Scores(),
	)
}

// ParseString is a method to parse a string into a subnetwork
func (network *genericSubnetwork) ParseString(line string) {
	split := strings.Split(line, "\t")
	// parse interactions
	interactionStr := strings.Split(split[0], " ")
	interactions := types.NewInteractionIDSet()
	for _, edge := range interactionStr {
		interactionID := types.ParseInteractionID(edge)
		interactions.Set(interactionID)
	}
	// parse paths
	pathStr := strings.Split(split[1], ";")
	paths := make(map[PathID]types.InteractionIDSet)
	for _, path := range pathStr {
		pathIDint, _ := strconv.Atoi(path)
		pathID := PathID(pathIDint)
		paths[pathID] = network.opt.pathRepositories.pathInteractionSetFromId(pathID)
	}
	// parse scores
	split[2] = strings.Replace(split[2], "[", "", -1)
	split[2] = strings.Replace(split[2], "]", "", -1)
	scoreStr := strings.Split(split[2], " ")
	scores := make([]float64, 0)
	for _, score := range scoreStr {
		s, _ := strconv.ParseFloat(score, 64)
		scores = append(scores, s)
	}
	// set the parsed values
	network.interactionSet = interactions
	network.selectedPaths = paths
	network.scores = scores
}
