package readers

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/arguments"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
)

// NetworkReader reads networks from the original data
// depending on the geneParser, it can read the input network (using gene names)
// or the intermediate networks (
type NetworkReader struct {
	networkLineParser
	geneParser
}

func NewInitialNetworkReader(logger *slog.Logger, gim *types.GeneIDMap) *NetworkReader {
	return newNetworkReader(logger, initialGeneParser{gim})
}

func NewIntermediateNetworkReader(logger *slog.Logger) *NetworkReader {
	return newNetworkReader(logger, intermediateGeneParser{})
}

func newNetworkReader(logger *slog.Logger, gp geneParser) *NetworkReader {
	return &NetworkReader{
		networkLineParser{logger},
		gp,
	}
}

func (nwr NetworkReader) ReadAllNetworks(optimizationDirectory string) []*graph.Network {
	// For every network size, get the networks
	resultFolders := make(map[string][]string)
	resultFileNamePattern := regexp.MustCompile(`^result-.*\.network$`)
	err := filepath.Walk(
		optimizationDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// skip directories
				return nil
			}
			dirName, fileName := filepath.Split(path)
			if !resultFileNamePattern.MatchString(fileName) {
				// only consider directories with valid result network
				return nil
			}
			resultFolders[dirName] = append(resultFolders[dirName], fileName)
			return nil
		},
	)
	if err != nil {
		nwr.Error("error traversing optimization directory", "err", err)
	}
	// For every network size, get the networks
	networksPerSize := nwr.readNetworks(resultFolders)
	// Combine all networks into a single list
	networks := make([]*graph.Network, 0)
	for _, networkList := range networksPerSize {
		for _, network := range networkList {
			networks = append(networks, network)
		}
	}
	return networks
}

func (nwr NetworkReader) readNetworks(resultFolders map[string][]string) [][]*graph.Network {
	networks := make([][]*graph.Network, 0)
	networkCounts := make(map[int]int)
	for folder, filenames := range resultFolders {
		for _, filename := range filenames {
			network := nwr.NewNetworkFromFile(filepath.Join(folder, filename), false, false)
			// add network to the list of networks
			for len(networks) < network.InteractionCount()+1 {
				networks = append(networks, make([]*graph.Network, 0))
			}
			networks[network.InteractionCount()] = append(networks[network.InteractionCount()], network)
			networkCounts[network.InteractionCount()]++
		}
	}
	nwr.Info("readNetworks", "counts per size", networkCounts)
	return networks
}

func (nwr NetworkReader) NewNetworkFromFile(
	fileName string,
	verbose bool,
	initial bool,
) *graph.Network {
	parsed := nwr.newWeightedNetworkFromFile(fileName, verbose, initial)
	return graph.NewNetwork(arguments.GlobalInteractionStore, parsed.probabilities, parsed.interactionTypes, parsed.scores)
}

func (nwr NetworkReader) NewNetworkFromFiles(
	networkFiles []string,
	bannedFiles []string,
	verbose bool,
	initial bool,
) *graph.Network {
	store := arguments.GlobalInteractionStore
	combined := newParsedNetwork()

	// Load and merge networks
	for _, file := range networkFiles {
		parsed := nwr.newWeightedNetworkFromFile(file, verbose, initial)
		nwr.mergeNetworks(combined, parsed)
	}

	// Load and subtract banned interactions
	for _, file := range bannedFiles {
		banned := nwr.newWeightedNetworkFromFile(file, false, initial)
		for id := range *banned.probabilities {
			combined.probabilities.Delete(id)
		}
	}

	return graph.NewNetwork(store, combined.probabilities, combined.interactionTypes, combined.scores)
}

func (nwr NetworkReader) mergeNetworks(combined, network *parsedNetwork) {
	// Merge interaction types
	for name, id := range network.interactionTypes {
		if previous, exists := combined.interactionTypes[name]; exists && previous != id {
			nwr.Error("interaction type name conflict", "name", name, "id1", previous, "id2", id)
			continue
		}
		combined.interactionTypes[name] = id
	}
	// Merge probabilities
	for id, probability := range *network.probabilities {
		if combined.probabilities.Has(id) {
			probability = max(combined.probabilities.GetProbability(id), probability)
		}
		combined.probabilities.SetProbability(id, probability)
	}
	// Append scores
	combined.scores = append(combined.scores, network.scores...)
}

func (nwr NetworkReader) newWeightedNetworkFromFile(
	fileName string,
	verbose, initial bool,
) *parsedNetwork {
	file, err := os.Open(fileName)
	if err != nil {
		nwr.Error("failed to open network file", "error", err)
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	return nwr.fromScanner(fileName, scanner, verbose, initial)
}

type parsedNetwork struct {
	interactionTypes types.InteractionTypeSet
	probabilities    *types.ProbabilityMap
	scores           []float64
}

func newParsedNetwork() *parsedNetwork {
	return &parsedNetwork{
		interactionTypes: make(types.InteractionTypeSet),
		probabilities:    types.NewProbabilityMap(),
		scores:           make([]float64, 0),
	}
}

func (nwr NetworkReader) useRawTypeIDs(initial bool) bool {
	return !initial
}

func (nwr NetworkReader) fromScanner(
	fileName string,
	scanner *bufio.Scanner,
	verbose, initial bool,
) *parsedNetwork {
	parser := networkLineParser{Logger: nwr.Logger}
	parsed := newParsedNetwork()
	parsedInt := parsedInteraction{}
	ok := false
	store := arguments.GlobalInteractionStore
	store.AddInteractionType("unknown", false)
	counter := int64(0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "% score"):
			parsed.scores = parser.parseScore(strings.TrimPrefix(line, "% score "))
		case strings.HasPrefix(line, "%"):
			if ph, ok := parser.parseHeaderLine(line); ok {
				nwr.processHeader(ph, parsed, store)
			}
		default:
			parsedInt, counter, ok = nwr.parseDataLine(line, nwr.parseGene, counter)
			if !ok {
				continue
			}
			nwr.processInteraction(parsedInt, store, parsed, nwr.useRawTypeIDs(initial))
		}
	}

	if verbose {
		nwr.Info("Finished loading network file", "interactionCount", counter, "fileName", fileName)
	}
	if err := scanner.Err(); err != nil {
		nwr.Error("error while loading network file", "error", err)
	}
	return parsed
}

func (nwr NetworkReader) processHeader(ph parsedInteractionType, parsed *parsedNetwork, store *types.InteractionStore) {
	parsed.interactionTypes[ph.name] = types.NewInteractionType(ph.name, ph.isReg)
	store.AddInteractionType(ph.name, ph.isReg)
}

func (nwr NetworkReader) processInteraction(
	intx parsedInteraction,
	store *types.InteractionStore,
	parsed *parsedNetwork,
	useRawIDs bool,
) {
	var typeID types.InteractionTypeID
	if useRawIDs {
		typeID = nwr.parseInteractionTypeID(intx.rawTypeID)
	} else {
		typeID = store.GetInteractionTypeID(intx.typ)
	}

	id := types.FromToTypeToID(intx.from, intx.to, typeID)
	store.AddInteraction(id)
	parsed.probabilities.SetProbability(id, intx.probability)

	if intx.direction != "directed" {
		store.AddInteraction(id.Reverse())
		parsed.probabilities.SetProbability(id.Reverse(), intx.probability)
	}

	parsed.interactionTypes[intx.typ] = types.NewInteractionType(intx.typ, false)
}
