package pathfinding

import (
	"sync"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
)

// runner is a struct containing all the necessary information to perform a pathfinding run
type runner struct {
	*arguments.Common
	gim                *types.GeneIDMap
	startGenes         types.GeneSet
	endGenes           types.GeneSet
	strains            types.ConditionSet
	excludedStrains    types.ConditionSet
	MutatedGeneWeights types.GeneConditionMap[float64]
	search             pathFinder
	outputFile         string
	pathLength         int
	nBest              int
	condition          types.Condition
}

func newRunner(
	args *arguments.Common,
	startGenes, endGenes types.GeneSet,
	conditions, excludedConditions types.ConditionSet,
	mutatedGeneWeights types.GeneConditionMap[float64],
	search pathFinder,
	outputFile string,
	pathLength, nBest int,
	condition types.Condition,
) runner {
	return runner{
		Common:             args,
		gim:                args.GeneIDMap,
		startGenes:         startGenes,
		endGenes:           endGenes,
		strains:            conditions,
		excludedStrains:    excludedConditions,
		MutatedGeneWeights: mutatedGeneWeights,
		search:             search,
		outputFile:         outputFile,
		pathLength:         pathLength,
		nBest:              nBest,
		condition:          condition,
	}
}

// findPaths performs the actual path finding for
// 1) the QTL case (qtl = true)
// A big difference with the "general" case is that here the found paths are not weighted solely by multiplying the
// weights of the edges which make up the path but that they are weighted based on the relevance scores of the start and
// end genes
// Paths are searched for between a specific start gene and ALL POSSIBLE end genes, this means that it is NOT true that
// the N-best paths between every pair of mutations is searched for, but that the N-best paths between all possible start
// genes and ALL OTHER genes are searched for. This is needed to reduce the total number of paths which would be
// computationally too difficult for the subsequent optimization step.
// 2) the "general" case. (qtl = false)
// Here, the probability of a path is calculated based on the multiplication of the weights of the edges which make up the path.
func (gpfr runner) findPaths(qtl bool, startIsMutated bool) {
	startTime := time.Now()
	// set up output file
	var outputFileMutex sync.Mutex
	// wait group to sync go routines
	wg := &sync.WaitGroup{}
	wg.Add(len(gpfr.startGenes))

	// channel to collect maxToVisitSize values
	tmpChannel := make(chan int, len(gpfr.startGenes))

	// log and close file at the end
	go func() {
		wg.Wait() // wait for timing, file closing
		close(tmpChannel)
		maxToVisitSize := 0
		for tmp := range tmpChannel {
			if tmp > maxToVisitSize {
				maxToVisitSize = tmp
			}
		}
		endTime := time.Now()
		gpfr.Info(
			"Finished pathfinding",
			"condition", gpfr.condition,
			"startGenes", len(gpfr.startGenes),
			"runTime", endTime.Sub(startTime).Milliseconds(),
			"maxToVisitSize", maxToVisitSize,
		)
	}()

	// loop over all start genes
	for fromGene := range gpfr.startGenes {
		gpfr.Sem.Acquire()
		go func(from types.GeneID) {
			defer func() {
				gpfr.Sem.Release()
				wg.Done()
			}()
			var result []*graph.Path
			var tmp int
			if qtl {
				result, tmp = gpfr.search.findNBestPathsQTL(
					gpfr.pathLength,
					gpfr.nBest,
					from,
					gpfr.condition,
					gpfr.MutatedGeneWeights,
					gpfr.strains,
					gpfr.excludedStrains,
					gpfr.endGenes,
					startIsMutated,
				)
			} else {
				result, tmp = gpfr.search.findNBestPathsFor(
					gpfr.pathLength,
					gpfr.nBest,
					from,
					gpfr.endGenes,
				)
			}
			tmpChannel <- tmp
			gpfr.writePathsToFile(result, from, &outputFileMutex)
		}(fromGene)
	}
}

func (gpfr runner) writePathsToFile(
	result []*graph.Path,
	from types.GeneID,
	outputFileMutex *sync.Mutex,
) {
	if len(result) > 0 {
		content := make([]string, 0, len(result))
		for _, path := range result {
			content = append(content, path.TxtString(gpfr.search.expander.Probabilities(), gpfr.gim, from, gpfr.condition))
		}
		content = append(content, "")
		outputFileMutex.Lock()
		if err := gpfr.AppendLinesToFile(gpfr.outputFile, content); err != nil {
			gpfr.Error("Error writing to output file", "error", err)
			panic("Cannot write paths to file")
		}
		outputFileMutex.Unlock()
	}
}
