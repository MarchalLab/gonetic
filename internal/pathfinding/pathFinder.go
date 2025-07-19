package pathfinding

import (
	"log/slog"
	"math"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
)

// pathFinder is a struct that contains the methods to perform path finding
type pathFinder struct {
	*slog.Logger
	expander       graph.PathExpander
	pathDefinition graph.PathDefinition
	sldCutoff      float64
}

// newPathFinder is a constructor for pathFinder
func newPathFinder(
	logger *slog.Logger,
	expander graph.PathExpander,
	pathDefinition graph.PathDefinition,
	sldCutoff float64,
) pathFinder {
	return pathFinder{
		Logger:         logger,
		expander:       expander,
		pathDefinition: pathDefinition,
		sldCutoff:      sldCutoff,
	}
}

func isConditionAllowed(condition types.Condition, include, exclude types.ConditionSet) bool {
	if _, bad := exclude[condition]; bad {
		// this condition is in the blacklist and thus not allowed here
		return false
	}
	if _, ok := include[condition]; !ok {
		// this condition is not in the whitelist and thus not allowed here
		return false
	}
	return true
}

func maxMutationScore(
	mutated types.GeneConditionMap[float64],
	include, exclude types.ConditionSet,
) float64 {
	maxScore := 0.0
	for gene := range mutated {
		for cond := range mutated[gene] {
			if isConditionAllowed(cond, include, exclude) {
				maxScore = math.Max(maxScore, mutated[gene][cond])
			}
		}
	}
	return maxScore
}

// addConditionallyScoredPaths is a method that adds conditionally scored paths to the results
func (search pathFinder) addConditionallyScoredPaths(
	current *graph.Path,
	from types.GeneID,
	fromScore float64,
	n int,
	cutoff float64,
	mutated types.GeneConditionMap[float64],
	include, exclude types.ConditionSet,
	results *PriorityQueue[*graph.Path],
) float64 {
	endGene := current.EndGene
	for condition := range mutated[endGene] {
		if !isConditionAllowed(condition, include, exclude) {
			continue
		}
		toScore, ok := mutated.Get(endGene, condition)
		if !ok {
			continue
		}
		pathToAdd, err := search.expander.CreatePathFrom(current.Interactions(), from, fromScore, toScore)
		if err != nil {
			search.Error("Error creating path from interactions",
				"from", from,
				"fromScore", fromScore,
				"toScore", toScore,
				"interactions", current.Interactions(),
				"err", err)
			continue
		}
		pathToAdd.EndCondition = condition
		cutoff = pushIfBetter(results, pathToAdd, n, cutoff)
	}
	return cutoff
}

// TODO: refactor this method to make it more readable
// findNBestPathsQTL is a method that finds the N best paths in the QTL case
func (search pathFinder) findNBestPathsQTL(
	pathlength, n int,
	from types.GeneID,
	conditionFromGene types.Condition,
	mutatedGenesMap types.GeneConditionMap[float64],
	filteredConditions, excludedConditions types.ConditionSet,
	endGenes types.GeneSet,
	startIsMutated bool,
) ([]*graph.Path, int) {
	cutoff := search.sldCutoff
	// In this case the weighting needs to be partially done directly on the paths (only network topology is incorporated through the weights on the edges)
	toVisit := NewPriorityQueue[*graph.Path]()
	results := NewReversePriorityQueue[*graph.Path]()

	// Get the score of the mutation with the best score in the gene from which the path starts for the condition from which the path starts.
	fromScore := 1.0
	if startIsMutated {
		fromMutationWeight, ok := mutatedGenesMap.Get(from, conditionFromGene)
		if !ok {
			return results.PopToReverseSlice(), 0
		}
		fromScore = fromMutationWeight
	}

	// find the max to score
	// TODO: only compute this once for the entire map (for each condition) and pass it to this method
	maxToScore := maxMutationScore(mutatedGenesMap, filteredConditions, excludedConditions)

	// compute the bound probability for the current path
	boundProbability := func(path *graph.Path) float64 {
		return path.Probability() * fromScore * maxToScore
	}

	// BFS with BB
	root := graph.RootPath(from, fromScore)
	toVisit.Push(root)
	maxToVisitSize := toVisit.Len()
	for !toVisit.Empty() {
		maxToVisitSize = max(maxToVisitSize, toVisit.Len())
		current := toVisit.Pop()
		// branch and bound
		if results.Len() >= n && results.Top().Probability() > boundProbability(current) {
			break // no better path is possible
		}
		// If a valid path was found, the scores of the end genes need to be calculated
		_, isEndGene := endGenes[current.EndGene]
		_, isMutatedGene := mutatedGenesMap[current.EndGene]
		isNotFromGene := from != current.EndGene
		isValidPath := search.pathDefinition(*current)
		if isEndGene && isMutatedGene && isNotFromGene && isValidPath {
			cutoff = search.addConditionallyScoredPaths(
				current,
				from,
				fromScore,
				n,
				cutoff,
				mutatedGenesMap,
				filteredConditions,
				excludedConditions,
				results,
			)
		}
		// construct next paths from current path
		if current.Length < pathlength {
			children := search.expandPath(current, cutoff, boundProbability)
			for _, child := range children {
				if child.Probability() > cutoff {
					toVisit.Push(child)
				}
			}
		}
	}
	return results.PopToReverseSlice(), maxToVisitSize
}

// findNBestPathsFor is a method that finds the N best paths for a given path length
func (search pathFinder) findNBestPathsFor(
	pathlength, n int,
	from types.GeneID,
	endGenes types.GeneSet,
) ([]*graph.Path, int) {
	boundProbability := func(path *graph.Path) float64 {
		return path.Probability()
	}
	toVisit := NewPriorityQueue[*graph.Path]()
	results := NewReversePriorityQueue[*graph.Path]()
	root := graph.RootPath(from, 1)
	cutoff := search.sldCutoff
	toVisit.Push(root)
	maxToVisitSize := toVisit.Len()
	for !toVisit.Empty() {
		maxToVisitSize = max(maxToVisitSize, toVisit.Len())
		current := toVisit.Pop()
		if search.isAcceptablePath(current, from, endGenes) {
			cutoff = pushIfBetter(results, current, n, cutoff)
		}
		// construct next paths from current path
		if current.Length < pathlength {
			children := search.expandPath(current, cutoff, boundProbability)
			for _, child := range children {
				if child.Probability() > cutoff {
					toVisit.Push(child)
				}
			}
		}
	}
	return results.PopToReverseSlice(), maxToVisitSize
}

func (search pathFinder) isAcceptablePath(
	path *graph.Path,
	from types.GeneID,
	endGenes types.GeneSet,
) bool {
	// Check if the path has no genes or if it ends with the starting gene
	if path == nil || path.Length == 0 || path.EndGene == from {
		return false
	}
	// Check if the path ends with a gene in the endGenes set
	if _, ok := endGenes[path.EndGene]; !ok {
		return false
	}
	// Check if the path is acceptable according to the path definition
	return search.pathDefinition(*path)
}

// pushIfBetter Adds to results only if:
// - we still need more results
// - OR this path is better than the worst so far
// returns the updated cutoff
func pushIfBetter(
	results *PriorityQueue[*graph.Path],
	path *graph.Path,
	n int,
	cutoff float64,
) float64 {
	probability := path.Probability()
	// if the path is not better than the cutoff, we can skip it
	if probability < cutoff {
		return cutoff
	}
	// try to add the path to the results
	if results.Len() < n {
		results.Push(path)
	} else if probability > results.Top().Probability() {
		results.Pop()
		results.Push(path)
	}
	// update cutoff and return it
	if results.Len() >= n {
		cutoff = max(cutoff, results.Top().Probability())
	}
	return cutoff
}

// expandPath expands the current path and filters the results based on the cutoff
func (search pathFinder) expandPath(
	current *graph.Path,
	cutoff float64,
	boundProbability func(*graph.Path) float64,
) []*graph.Path {
	expansions, err := search.expander.Expand(current)
	if err != nil {
		search.Error("Error expanding path", "path", current, "err", err)
		return nil
	}
	filtered := make([]*graph.Path, 0, len(expansions))
	for _, newPath := range expansions {
		score := boundProbability(newPath)
		if score > cutoff {
			filtered = append(filtered, newPath)
		}
	}
	return filtered
}
