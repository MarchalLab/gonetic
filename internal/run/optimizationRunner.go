package run

import (
	"path/filepath"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/normalform"
	"github.com/MarchalLab/gonetic/internal/optimization"
	"github.com/MarchalLab/gonetic/internal/readers"
)

// OptimizationRunner controls the Optimization step
type OptimizationRunner struct {
	*arguments.Common
	startIsMutated bool
}

func NewOptimizationRunner(args *arguments.Common, startIsMutated bool) OptimizationRunner {
	readers.ReadIndexes(args)
	return OptimizationRunner{
		args,
		startIsMutated,
	}
}

func (runner OptimizationRunner) Run() {
	runner.Info("Running optimization") //TODO: print arguments
	runner.DumpProfiles("opt-start")
	defer func() {
		runner.DumpProfiles("opt-end")
	}()

	// timing
	startTime := time.Now()
	defer func() {
		passedTime := time.Now().Sub(startTime).Seconds()
		runner.Info("Optimization finished", "seconds", passedTime)
	}()

	// clean and initialize the directories
	runner.prepareDirectories()

	// determine the relevant path types
	pathTypes := runner.PathTypes

	// read the path repositories
	cnfPathsList, pathRepositories := optimization.ReadPathRepositories(runner.Common)

	// compile ddnnfs
	if !runner.SkipCompilation {
		for pathType := range pathTypes {
			runner.compilePathsToDDNNF(pathTypes[pathType], cnfPathsList[pathType])
		}
	}

	// load ddnnfs
	dDNNFList := make([][]*normalform.NNF, 0, len(pathTypes))
	for pathType := range pathTypes {
		runner.Info("Processing path type", "pathType", pathTypes[pathType])
		dDNNFs := runner.loadDDNNFs(pathTypes[pathType])
		dDNNFList = append(dDNNFList, dDNNFs)
	}

	// close the semaphore channel
	runner.Sem.Wait()

	// run the optimization loop
	moRunner := optimization.NewMORunner(runner.Common)
	moRunner.Run(pathRepositories, dDNNFList, runner.NumCPU)
}

func (runner OptimizationRunner) loadDDNNFs(pathType string) []*normalform.NNF {
	startTime := time.Now()
	defer func() {
		passedTime := time.Now().Sub(startTime).Milliseconds()
		runner.Info("d-DNNFs loaded", "ms", passedTime, "path type", pathType)
	}()
	// read ddnnfs
	nfDir := filepath.Join(runner.NormalFormDirectory(), pathType)
	DDNNFCompiler := normalform.NewDDNNFCompiler(runner.Common, runner.EtcPathAsString)
	dDNNFs, err := DDNNFCompiler.LoadDDNNFs(nfDir)
	if err != nil {
		runner.Error("error in DDNNFCompiler.LoadDDNNFs", "err", err)
	}
	runner.Info("received d-DNNFs", "ddnnfs", len(dDNNFs))
	runner.computeIntersectionMap(dDNNFs)
	return dDNNFs
}

func (runner OptimizationRunner) computeIntersectionMap(ddnnfs []*normalform.NNF) {
	for _, nnf := range ddnnfs {
		runner.Sem.Acquire()
		go func(nnf *normalform.NNF) {
			defer runner.Sem.Release()
			nnf.InteractionIndex = make(map[types.InteractionID]int)
			i := 0
			for id := range nnf.Values() {
				i += 1
				nnf.InteractionIndex[id] = i
			}
		}(nnf)
	}
}

func (runner OptimizationRunner) prepareDirectories() {
	moDir := filepath.Join(runner.OutputFolder, "MO", "population")
	if runner.Resume {
		// Keep any previous results
		fileio.CreateDirKeepContent(runner.FrontsDirectory())
		fileio.CreateDirKeepContent(moDir)
		runner.Info("Creation of optimization directory finished.")
	} else {
		// Remove any previous results first so there is no chance of utilizing the wrong files when the program is run twice.
		fileio.CreateEmptyDir(runner.FrontsDirectory())
		fileio.CreateEmptyDir(moDir)
		runner.Info("Clearing and creation of optimization directory finished.")
	}
	// always wipe the optimization directory
	fileio.CreateEmptyDir(runner.OptimizationDirectory())
}

func (runner OptimizationRunner) compilePathsToDDNNF(pathType string, cnfPaths map[types.CNFHeader]types.CompactPathList) {
	startTime := time.Now()
	defer func() {
		passedTime := time.Now().Sub(startTime).Milliseconds()
		runner.Info("d-DNNFs compiled", "ms", passedTime, "path type", pathType)
	}()
	nfDir := filepath.Join(runner.NormalFormDirectory(), pathType)
	// Convert paths to a paths CNF
	cnf := normalform.NewCNF(runner.FileWriter)
	err := cnf.Conversion(cnfPaths, nfDir)
	if err != nil {
		runner.Error("error in cnf.Conversion", "err", err)
	}
	// Convert the paths CNF to a compiled CNF
	err = cnf.Compile(cnfPaths, nfDir)
	if err != nil {
		runner.Error("error in cnf.Compile", "err", err)
	}
	// compile d-DNNF's
	err = normalform.NewDDNNFCompiler(runner.Common, runner.EtcPathAsString).CompileDDNNFs(nfDir)
	if err != nil {
		runner.Error("error in DDNNFCompiler.CompileDDNNFs", "err", err)
	}
}
