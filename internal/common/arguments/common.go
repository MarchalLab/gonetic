package arguments

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/MarchalLab/gonetic/internal/common/semaphore"

	"github.com/MarchalLab/gonetic/internal/common/profiler"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// TODO: temporary global interaction store, should be removed after the transition is complete
var GlobalInteractionStore = types.NewInteractionStore()

type Common struct {
	// logger and writer
	*fileio.FileWriter

	// profiler
	*profiler.Profiler

	// VCS commit
	Commit string

	// Files and directories
	OutputFolder       string
	NetworkFiles       []string
	BannedNetworkFiles []string
	EtcPathAsString    string
	MappingFile        string
	// General settings
	NumCPU                    int
	LogFormat                 string
	LogFile                   string
	Verbose                   bool
	TopologyWeightingAddition string
	MinEdgeScore              float64
	// Path-finding settings
	PathTypes     []string
	PathLength    int
	BestPathCount int
	SldCutoff     float64
	// Optimization settings
	MaxPaths                   int
	NumGens                    int
	EarlyTermination           bool
	TimeLimitHours             float64
	WindowCount                int
	MinWindowSize              int
	MaxWindowSize              int
	RequiredProgressPercentage float64
	MutChance                  float64
	PopSize                    int
	OptimizeNetworkSize        bool
	OptimizeSampleCount        bool
	FocusFraction              float64
	TargetNetworkSize          int
	SampleObjectiveType        string
	// Interpretation settings
	// Skips
	UseIndex            string
	SkipPathFinding     bool
	UsePaths            string
	SkipCompilation     bool
	UseNNFs             string
	SkipOptimization    bool
	SkipInterpreter     bool
	SkipNetworkPrinting bool

	// TODO: Resume should use precomputed files wherever possible
	Resume bool

	// common data TODO: move this to a separate struct, since these are not arguments
	Sem semaphore.Semaphore
	*types.GeneIDMap
	InteractionStore *types.InteractionStore
}

func NewCommon() *Common {
	return &Common{
		Profiler: &profiler.Profiler{},
	}
}

const pathsDirectoryName = "paths"
const pathsFileName = "paths.txt"
const weightsFileName = "weights.txt"

func (arguments *Common) PathsDirectory(dir string) string {
	basePath := arguments.UsePaths
	if basePath == "" {
		basePath = filepath.Join(arguments.OutputFolder, fmt.Sprintf("%s_%d", pathsDirectoryName, arguments.PathLength))
	}
	if len(arguments.PathTypes) == 0 {
		return basePath
	}
	return filepath.Join(basePath, dir)
}

func (arguments *Common) PathsFileWithName(dir, filename string) string {
	return filepath.Join(arguments.PathsDirectory(dir), filename)
}

func (arguments *Common) PathsFile(dir string) string {
	return filepath.Join(arguments.PathsDirectory(dir), pathsFileName)
}

func (arguments *Common) WeightsFile(dir, prefix string) string {
	return filepath.Join(arguments.PathsDirectory(dir), prefix+weightsFileName)
}

const optimizationDirectoryName = "optimization"

func (arguments *Common) OptimizationDirectory() string {
	return filepath.Join(
		arguments.OutputFolder,
		fmt.Sprintf("%s_%d", optimizationDirectoryName, arguments.PopSize),
	)
}

const frontsDirectoryName = "fronts"

func (arguments *Common) FrontsDirectory() string {
	return filepath.Join(arguments.OutputFolder, frontsDirectoryName)
}

const resultsDirectoryName = "resulting_networks"
const filledDirectoryName = "filled_networks"
const weightedNetworkFileName = "weighted.network"

func (arguments *Common) ResultsDirectory(pathType string) string {
	return filepath.Join(arguments.OutputFolder, resultsDirectoryName, pathType)
}

func (arguments *Common) FillDirectory() string {
	return filepath.Join(arguments.OutputFolder, filledDirectoryName)
}

func (arguments *Common) WeightedNetworkFile(directory string) string {
	return filepath.Join(directory, weightedNetworkFileName)
}

const nnfDirectoryName = "NF"

func (arguments *Common) NormalFormDirectory() string {
	if arguments.UseNNFs != "" {
		return arguments.UseNNFs
	}
	if arguments.MaxPaths == 0 {
		return filepath.Join(
			arguments.OutputFolder,
			fmt.Sprintf(
				"%s_%d",
				nnfDirectoryName,
				arguments.PathLength,
			),
		)
	}
	return filepath.Join(
		arguments.OutputFolder,
		fmt.Sprintf(
			"%s_%d_%d",
			nnfDirectoryName,
			arguments.PathLength,
			arguments.MaxPaths,
		),
	)
}

func (arguments *Common) AutoLogFile() string {
	return filepath.Join(arguments.OutputFolder, "output.log")
}

// GeneMapFileToRead returns the path to the gene map file that will be read from
func (arguments *Common) GeneMapFileToRead() string {
	if arguments.UseIndex != "" {
		return arguments.UseIndex
	}
	return arguments.geneMapFileToWrite()
}

// InteractionTypeMapFileToRead returns the path to the gene map file that will be read from
func (arguments *Common) InteractionTypeMapFileToRead() string {
	if arguments.UseIndex != "" {
		return arguments.UseIndex
	}
	return arguments.interactionTypeFileToWrite()
}

// geneMapFileToWrite returns the path to the gene map file that will be written to
func (arguments *Common) geneMapFileToWrite() string {
	return filepath.Join(arguments.OutputFolder, "gene-ids")
}

// interactionTypeFileToWrite returns the path to the interaction type map file that will be written to
func (arguments *Common) interactionTypeFileToWrite() string {
	return filepath.Join(arguments.OutputFolder, "interaction-type-ids")
}

func (arguments *Common) WriteGeneMapFile() {
	arguments.GeneIDMap.WriteNameMap(arguments.FileWriter, arguments.geneMapFileToWrite())
}

func (arguments *Common) WriteInteractionTypeMapFile() {
	arguments.InteractionStore.InteractionTypes().WriteNameMap(arguments.FileWriter, arguments.interactionTypeFileToWrite())
}

// LogActiveGoRoutines logs the number of active goroutines
func (arguments *Common) LogActiveGoRoutines() {
	arguments.Info("Active goroutines", "num", runtime.NumGoroutine())
}

func (arguments *Common) Init() error {
	// set the log file
	var logWriter io.Writer
	var err error
	switch arguments.LogFile {
	case "":
		logWriter = os.Stdout
	case "auto":
		logWriter, err = createLogFile(arguments.AutoLogFile())
	default:
		logWriter, err = createLogFile(arguments.LogFile)
	}
	if err != nil {
		return err
	}
	var handler slog.Handler
	switch arguments.LogFormat {
	case "text":
		handler = slog.NewTextHandler(logWriter, nil)
	case "json":
		handler = slog.NewJSONHandler(logWriter, nil)
	default:
		handler = slog.NewJSONHandler(logWriter, nil)
	}
	logger := slog.New(handler)
	arguments.FileWriter = &fileio.FileWriter{Logger: logger}

	// set skips when alternative locations are provided
	if arguments.UsePaths != "" {
		logger.Info("Skipping path finding")
		arguments.SkipPathFinding = true
	}
	if arguments.UseNNFs != "" {
		logger.Info("Skipping path compilation")
		arguments.SkipCompilation = true
	}

	// check validity of skips and use of precomputed files
	if (arguments.UseNNFs != "" && arguments.UsePaths == "") ||
		(arguments.UseNNFs != "" && arguments.UseIndex == "") ||
		(arguments.UsePaths != "" && arguments.UseIndex == "") {
		arguments.Error("UseNNF requires UsePaths requires UseIndex, this is not the case here.",
			"UseNNFs", arguments.UseNNFs,
			"UsePaths", arguments.UsePaths,
			"UseIndex", arguments.UseIndex,
		)
		return errors.New("invalid use of precomputed files")
	}
	// MaxPaths should be positive
	arguments.MaxPaths = max(arguments.MaxPaths, 0)

	// if sample objective type is "", set it to "entropy"
	if arguments.SampleObjectiveType == "" {
		arguments.SampleObjectiveType = "entropy"
	}

	// set the number of logical CPU cores to use
	if arguments.NumCPU < 1 {
		arguments.NumCPU = 1
	}
	runtime.GOMAXPROCS(arguments.NumCPU)
	arguments.Sem = semaphore.NewSemaphore(arguments.NumCPU)
	logger.Info("Using CPUs", "maxProcs", runtime.GOMAXPROCS(0), "numCPU", runtime.NumCPU())

	// create the gene - id and interaction - id maps
	arguments.GeneIDMap = types.NewGeneIDMap()
	arguments.InteractionStore = GlobalInteractionStore

	// Init the profiler
	arguments.Profiler.Init(arguments.OutputFolder)

	// log the arguments
	logger.Info("Commmon arguments", SlowAttributeArray("", arguments)...)

	// finished initialization
	return nil
}

func createLogFile(filename string) (*os.File, error) {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.New(fmt.Sprintf("error creating log directory: %v", err))
	}

	// Open the file
	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error opening log file: %v", err))
	}
	return f, nil
}
