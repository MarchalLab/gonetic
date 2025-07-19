package run_test

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/run"
)

func createCommonTestArgs(outputFolder string) *arguments.Common {
	fileio.CreateEmptyDir(outputFolder)
	testArgs := arguments.NewCommon()

	// Files and directories
	testArgs.OutputFolder = outputFolder
	testArgs.NetworkFiles = []string{path.Join("testdata", "network.csv")}
	testArgs.EtcPathAsString = path.Join("..", "..", "etc")
	testArgs.MappingFile = ""

	// General Settings
	testArgs.NumCPU = 1
	testArgs.LogFormat = ""
	testArgs.LogFile = "auto"
	testArgs.Verbose = false
	testArgs.TopologyWeightingAddition = "bayes"
	testArgs.MinEdgeScore = 0.7

	// Path-finding settings
	testArgs.PathLength = 5
	testArgs.BestPathCount = 25
	testArgs.SldCutoff = 0

	// Optimization settings
	testArgs.NumGens = 10
	testArgs.EarlyTermination = true
	testArgs.TimeLimitHours = 0
	testArgs.WindowCount = 20
	testArgs.MinWindowSize = 2
	testArgs.MaxWindowSize = 100
	testArgs.RequiredProgressPercentage = 0.25
	testArgs.MutChance = 0.2
	testArgs.PopSize = 100
	testArgs.MaxPaths = 50
	testArgs.OptimizeNetworkSize = true
	testArgs.OptimizeSampleCount = true
	testArgs.TargetNetworkSize = 10

	// Skips
	testArgs.SkipPathFinding = false
	testArgs.SkipOptimization = false
	testArgs.SkipInterpreter = false
	testArgs.SkipNetworkPrinting = false

	// Initialize the common arguments
	testArgs.Init()
	return testArgs
}

func TestQTL(t *testing.T) {
	outputFolder := path.Join("testresult", "qtl")
	commonArgs := createCommonTestArgs(outputFolder)
	commonArgs.OutputFolder = outputFolder
	qtlArgs := arguments.QTLSpecific{
		FreqIncrease:     true,
		Correction:       true,
		FuncScore:        true,
		MutRateParam:     0,
		MutationDataFile: path.Join("testdata", "input-qtl.csv"),
		FreqCutoff:       0,
		WithinCondition:  false,
	}
	args := arguments.QTL{
		Common:      commonArgs,
		QTLSpecific: &qtlArgs,
	}

	// run and log path finding separately
	commonArgs.SkipPathFinding = false
	commonArgs.SkipOptimization = true
	commonArgs.SkipInterpreter = true
	args.LogFile = filepath.Join(outputFolder, "output-path.log")
	args.Init()
	run.QTL(&args)

	// run and log optimization and interpretation separately
	commonArgs.SkipPathFinding = true
	commonArgs.SkipOptimization = false
	commonArgs.SkipInterpreter = false
	commonArgs.SldCutoff = 0.111
	args.LogFile = filepath.Join(outputFolder, "output-opt.log")
	args.Init()
	run.QTL(&args)
}
