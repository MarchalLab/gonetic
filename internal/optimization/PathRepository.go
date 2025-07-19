package optimization

import (
	"log/slog"
	"math/rand"
	"sort"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

// PathRepositories is a slice of PathRepository
type PathRepositories struct {
	repos      []PathRepository
	reshuffles int
}

func NewPathRepositories() *PathRepositories {
	return &PathRepositories{
		repos:      []PathRepository{},
		reshuffles: 0,
	}
}

func (repositories *PathRepositories) Add(repo PathRepository) {
	repositories.repos = append(repositories.repos, repo)
}

// pathInteractionSetFromId dispatches the interaction set request to the correct PathRepository
func (repositories *PathRepositories) pathInteractionSetFromId(id PathID) types.InteractionIDSet {
	typeIndex := id.PathTypeIndex()
	return repositories.repos[typeIndex].pathInteractionSetFromId(id)
}

func (repositories *PathRepositories) NumberOfPaths() int {
	number := 0
	for _, repo := range repositories.repos {
		number += repo.NumberOfPaths
	}
	return number
}

func (repositories *PathRepositories) RandomRepo() (*PathRepository, int) {
	idx := rand.Intn(len(repositories.repos))
	return &repositories.repos[idx], idx
}

// PathRepository contains a map of paths
type PathRepository struct {
	pathType           string
	NumberOfPaths      int
	pathInteractionSet map[int]types.CompactPathList
}

func NewPathRepository(logger *slog.Logger, cnfPaths map[types.CNFHeader]types.CompactPathList, pathType string) PathRepository {
	pathCount := 0
	conditionMap := make(map[types.Condition]struct{})
	pathsPerSample := make(map[types.Condition]types.CompactPathList)
	for _, paths := range cnfPaths {
		for _, path := range paths {
			pathsPerSample[path.StartCondition] = append(pathsPerSample[path.StartCondition], path)
			// track the sample
			conditionMap[path.StartCondition] = struct{}{}
			// increment path count
			pathCount++
		}
	}
	// sort the samples
	sampleList := make([]types.Condition, 0, len(conditionMap))
	for sample := range conditionMap {
		sampleList = append(sampleList, sample)
	}
	sort.Slice(sampleList, func(i, j int) bool {
		return sampleList[i] < sampleList[j]
	})
	// create pathInteractionSet with sorted path lists
	pathInteractionSet := make(map[int]types.CompactPathList)
	totalDuration := time.Duration(0)
	peakDuration := time.Duration(0)
	for idx, sample := range sampleList {
		paths := pathsPerSample[sample]
		start := time.Now()
		sort.Sort(paths)
		duration := time.Since(start)
		totalDuration += duration
		peakDuration = max(duration, peakDuration)
		pathInteractionSet[idx] = paths
	}
	logger.Info("sorted all paths",
		"pathType", pathType,
		"samples", len(sampleList),
		"totalDuration", totalDuration.String(),
		"peakDuration", peakDuration.String(),
	)
	// return a new path repository
	return PathRepository{
		pathType,
		pathCount,
		pathInteractionSet,
	}
}

func (repository *PathRepository) pathInteractionSetFromId(id PathID) types.InteractionIDSet {
	sampleIndex := id.SampleIndex()
	pathIndex := id.PathIndex()
	return repository.pathInteractionSet[sampleIndex][pathIndex].InteractionSet()
}

func (repository *PathRepository) RandomSample() (types.CompactPathList, int) {
	idx := rand.Intn(len(repository.pathInteractionSet))
	return repository.pathInteractionSet[idx], idx
}

func (repository *PathRepository) RandomPath(sampleIdx int) (*types.CompactPath, int) {
	paths := repository.pathInteractionSet[sampleIdx]
	idx := rand.Intn(len(paths))
	return paths[idx], idx
}

// ReadPathRepositories reads the paths from the paths files
func ReadPathRepositories(args *arguments.Common) ([]map[types.CNFHeader]types.CompactPathList, *PathRepositories) {
	cnfPathsList := make([]map[types.CNFHeader]types.CompactPathList, 0, len(args.PathTypes))
	for _, pathType := range args.PathTypes {
		// Read the paths from the resulting .paths file from the pathfinding step
		var cnfPaths map[types.CNFHeader]types.CompactPathList
		cnfPaths = readers.ReadPathList(args.Logger, args.GeneIDMap, args.MaxPaths, pathType, args.SldCutoff, args.PathsFile(pathType))
		cnfPathsList = append(cnfPathsList, cnfPaths)
	}
	// load path repositories
	pathRepositories := loadPathRepositories(args.Logger, args.PathTypes, cnfPathsList)
	return cnfPathsList, pathRepositories
}

// loadPathRepositories converts the paths from the paths files to a PathRepositories object
func loadPathRepositories(logger *slog.Logger, pathTypes []string, cnfPathsList []map[types.CNFHeader]types.CompactPathList) *PathRepositories {
	pathRepositories := NewPathRepositories()
	// read path repository
	for idx, pathType := range pathTypes {
		pathRepository := NewPathRepository(logger, cnfPathsList[idx], pathType)
		pathRepositories.Add(pathRepository)
	}
	return pathRepositories
}
