package readers

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/compare"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// setPathScoreCutoff implements quickselect to determine the cutoff score, which is the k-th highest score
func pathScoreCutoff(maxPaths int, cnfPaths map[types.CNFHeader]types.CompactPathList) float64 {
	if maxPaths == 0 {
		return 0.0
	}
	// first flatten the nested lists of paths to path scores
	flattenedList := make([]float64, 0)
	for _, paths := range cnfPaths {
		for _, path := range paths {
			// flip the sign so we can look for minimal scores in quickselect
			flattenedList = append(flattenedList, -path.Probability)
		}
	}
	if maxPaths >= len(flattenedList) {
		return 0.0
	}
	// quickselect, flip the sign to get the correct cutoff
	return -Quickselect(flattenedList, maxPaths)
}

// filterPaths removes paths with a score below the cutoff
func filterPaths(
	logger *slog.Logger,
	maxPaths int,
	pathType string,
	cnfPathsMap map[types.CNFHeader]types.CompactPathList,
) map[types.CNFHeader]types.CompactPathList {
	pathScoreCutoff := pathScoreCutoff(maxPaths, cnfPathsMap)
	if pathScoreCutoff > 0.0 {
		for cnf, cnfPaths := range cnfPathsMap {
			paths := make(types.CompactPathList, 0)
			for _, path := range cnfPaths {
				if path.Probability > pathScoreCutoff {
					paths[path.ID] = path
				}
			}
			cnfPathsMap[cnf] = paths
			if len(cnfPathsMap[cnf]) == 0 {
				delete(cnfPathsMap, cnf)
			}
		}
	}
	pathCount := 0
	for _, cnfPaths := range cnfPathsMap {
		pathCount += len(cnfPaths)
	}
	logger.Info("filtered path counts", "pathType", pathType, "pathCount", pathCount, "pathScoreCutoff", pathScoreCutoff)
	return cnfPathsMap
}

func ReadPathList(
	logger *slog.Logger,
	gim *types.GeneIDMap,
	maxPaths int,
	pathType string,
	cutoff float64,
	fileName string,
) map[types.CNFHeader]types.CompactPathList {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		logger.Error("failed to open file", "fileName", fileName, "err", err)
		return nil
	}
	defer file.Close()

	// Initialize the result map
	cnfs := make(map[types.CNFHeader]types.CompactPathList)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	pathIDs := make(map[int]struct{})
	for pathID := 0; scanner.Scan(); pathID++ {
		line := scanner.Text()
		processLine(line, logger, gim, pathID, pathIDs, cnfs, cutoff)
		// Dispatch each line to a helper function
	}
	if err := scanner.Err(); err != nil {
		logger.Error("error reading file", "fileName", fileName, "err", err)
	}

	return filterPaths(logger, maxPaths, pathType, cnfs)
}

func processLine(
	line string,
	logger *slog.Logger,
	gim *types.GeneIDMap,
	pathID int,
	pathIDs map[int]struct{},
	cnfs map[types.CNFHeader]types.CompactPathList,
	cutoff float64,
) *types.CompactPath {
	// Split the line based on \t characters
	parts := strings.Split(line, "\t")

	// if start and end sample are the same, then the end sample is omitted in the path file
	if len(parts) == 7 {
		// duplicate the first entry
		parts = append([]string{parts[0]}, parts...)
	}
	// Check the number of parts
	if len(parts) != 8 {
		logger.Error("invalid path format", "line", line)
		return types.EmptyCompactPath()
	}
	// assign each part to a meaningful variable
	startStr, endStr, pathProbabilityStr, interactionStr, fromScoreStr, edgeScoresStr, toScoreStr, typeStr := parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6], parts[7]

	// dispatch each part to a parser
	startSample := types.Condition(startStr)
	endSample := types.Condition(endStr)
	pathProbability, _ := strconv.ParseFloat(pathProbabilityStr, 64)
	edgeScores := parseEdgeScores(edgeScoresStr)
	interactionTypes := parseTypes(typeStr)
	interactions, _, interactionOrder, startGene, direction := parseInteractions(gim, interactionStr, edgeScores, interactionTypes)
	fromScore, _ := strconv.ParseFloat(fromScoreStr, 64)
	toScore, _ := strconv.ParseFloat(toScoreStr, 64)

	recomputedScore := fromScore * toScore
	for _, edgeScore := range edgeScores {
		recomputedScore *= edgeScore
	}
	if !compare.Tolerance(.00001).FloatEqualWithinTolerance(pathProbability, recomputedScore) {
		logger.Error("unexpected path probability",
			"line", line,
			"expected", recomputedScore,
			"actual", pathProbability,
		)
	}
	if pathProbability < cutoff {
		// skip paths with probability below the cutoff
		return types.EmptyCompactPath()
	}

	// reconstruct the CompactPath based on the parsed parts
	compactPath := types.NewCompactPath(
		pathID,
		pathProbability,
		interactionOrder,
		edgeScores,
		extractGenes(interactions),
		direction,
		extractStartGene(gim, interactionStr),
		startSample,
		extractEndGene(gim, interactionStr),
		endSample,
		fromScore,
		toScore,
	)
	// create CNFHeader
	cnfHeader := types.NewCNFHeader(startSample, startGene)

	// check for duplicate path IDs
	if _, ok := pathIDs[pathID]; ok {
		logger.Error("duplicate path ID", "ID", pathID)
		panic("Failed to read paths")
	}
	pathIDs[pathID] = struct{}{}

	// append the path to the cnfs map
	if _, ok := cnfs[cnfHeader]; !ok {
		cnfs[cnfHeader] = make(types.CompactPathList, 0)
	}
	cnfs[cnfHeader] = append(cnfs[cnfHeader], compactPath)
	return compactPath
}

// geneNameContinues checks if the gene name continues in the string or if the next pair of characters is an arrow
func geneNameContinues(i int, str string) bool {
	return i < len(str) &&
		!(i+1 < len(str) && str[i:i+2] == "->") &&
		!(i+1 < len(str) && str[i:i+2] == "<-")
}

func parseInteractions(
	gim *types.GeneIDMap,
	interactionsStr string,
	edgeScores []float64,
	interactionTypes []types.InteractionTypeID,
) (
	types.InteractionIDSet,
	*types.ProbabilityMap,
	[]types.InteractionID,
	types.GeneID,
	types.PathDirection,
) {
	interactions := types.NewInteractionIDSet()
	interactionOrder := make([]types.InteractionID, 0)
	probabilities := types.NewProbabilityMap()
	index := 0
	current := 0
	startGene := types.GeneID(0)
	i := 0
	direction := types.UndirectedPath
	for i < len(interactionsStr) {
		from := types.GeneID(0)
		to := types.GeneID(0)
		isDownStream := false
		switch {
		case i+2 <= len(interactionsStr) && interactionsStr[i:i+2] == "->":
			from = gim.GetIDFromName(types.GeneName(interactionsStr[current:i]))
			if startGene == 0 {
				startGene = from
			}
			i += 2
			current = i
			for geneNameContinues(i, interactionsStr) {
				i++
			}
			to = gim.GetIDFromName(types.GeneName(interactionsStr[current:i]))
			isDownStream = true
		case i+2 <= len(interactionsStr) && interactionsStr[i:i+2] == "<-":
			to = gim.GetIDFromName(types.GeneName(interactionsStr[current:i]))
			if startGene == 0 {
				startGene = to
			}
			i += 2
			current = i
			for geneNameContinues(i, interactionsStr) {
				i++
			}
			from = gim.GetIDFromName(types.GeneName(interactionsStr[current:i]))
			isDownStream = false
		default:
			i++
			continue
		}
		// create the interactionID with unknown type
		id := types.FromToTypeToID(from, to, interactionTypes[index])
		interactions.Set(id)
		probabilities.SetProbability(id, edgeScores[index])
		interactionOrder = append(interactionOrder, id)
		// update the direction
		direction = pathDirection(direction, isDownStream)
		// increment the index
		index++
	}
	return interactions, probabilities, interactionOrder, startGene, direction
}

func extractGenes(interactionIDs types.InteractionIDSet) types.GeneSet {
	genes := types.GeneSet{}
	for interactionID := range interactionIDs {
		genes[interactionID.From()] = struct{}{}
		genes[interactionID.To()] = struct{}{}
	}
	return genes
}

func extractStartGene(gim *types.GeneIDMap, interactionStr string) types.GeneID {
	parts := strings.Split(interactionStr, "->")
	parts = strings.Split(parts[0], "<-")
	return gim.GetIDFromName(types.GeneName(parts[0]))
}

func extractEndGene(gim *types.GeneIDMap, interactionStr string) types.GeneID {
	parts := strings.Split(interactionStr, "->")
	parts = strings.Split(parts[len(parts)-1], "<-")
	return gim.GetIDFromName(types.GeneName(parts[len(parts)-1]))
}

func parseTypes(typeStr string) []types.InteractionTypeID {
	// Remove the first and last characters
	typeStr = typeStr[1 : len(typeStr)-1]
	// Convert each score string to a float and store it in a slice
	parsed := make([]types.InteractionTypeID, 0)
	for _, str := range strings.Split(typeStr, " ") {
		typ, err := strconv.Atoi(str)
		if err != nil {
			panic(fmt.Sprintf("failed to parse type: %s", str))
		}
		parsed = append(parsed, types.InteractionTypeID(typ))
	}

	return parsed

}

func parseEdgeScores(scoresStr string) []float64 {
	// Remove the first and last characters
	scoresStr = scoresStr[1 : len(scoresStr)-1]

	// Convert each score string to a float and store it in a slice
	scores := make([]float64, 0)
	for _, scoreStr := range strings.Split(scoresStr, " ") {
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to parse score: %s", scoreStr))
		}
		scores = append(scores, score)
	}

	return scores
}

func pathDirection(soFar types.PathDirection, isDownStream bool) types.PathDirection {
	if isDownStream {
		// downstream
		switch soFar {
		case types.UpstreamPath:
			return types.UpDownstreamPath
		case types.DownstreamPath:
			return types.DownstreamPath
		case types.UndirectedPath:
			return types.DownstreamPath
		case types.UpDownstreamPath:
			return types.UpDownstreamPath
		case types.DownUpstreamPath:
			return types.InvalidPathDirection
		default:
			return types.InvalidPathDirection
		}
	}
	// upstream
	switch soFar {
	case types.UpstreamPath:
		return types.UpDownstreamPath
	case types.DownstreamPath:
		return types.DownUpstreamPath
	case types.UndirectedPath:
		return types.UpstreamPath
	case types.UpDownstreamPath:
		return types.InvalidPathDirection
	case types.DownUpstreamPath:
		return types.DownUpstreamPath
	default:
		return types.InvalidPathDirection
	}
}
