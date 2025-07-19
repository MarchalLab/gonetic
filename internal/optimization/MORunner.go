package optimization

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/normalform"
)

type MORunner struct {
	*arguments.Common
	directory    string
	nGenerations int
	mutChance    float64
	popSize      int
	verbose      bool
}

func NewMORunner(args *arguments.Common) MORunner {
	return MORunner{
		args,
		args.OptimizationDirectory(),
		args.NumGens,
		args.MutChance,
		args.PopSize,
		args.Verbose,
	}
}

func (runner MORunner) Run(pathRepositories *PathRepositories, dDNNFList [][]*normalform.NNF, numCPUs int) {
	start := time.Now()
	if pathRepositories.NumberOfPaths() == 0 {
		runner.Error("No paths available for optimization")
		return
	}

	// run the optimization
	runner.Info("Starting MO optimization")
	opt := newNSGAOptimization(
		runner.Common,
		pathRepositories,
		dDNNFList,
		runner.popSize,
		runner.nGenerations,
		runner.mutChance,
		numCPUs,
	)
	subnetworks := opt.Optimize()

	// make sure every network has a score assigned
	opt.parallelScoreCalc()
	networksPerSize := make(map[int]int)
	writtenSubnetworks := make([]subnetwork, 0, len(subnetworks))
	for _, network := range subnetworks {
		// final duplicate check
		if isDuplicate(network, writtenSubnetworks) {
			continue
		}
		writtenSubnetworks = append(writtenSubnetworks, network)
		// write the subnetwork to a file
		networksPerSize[network.subnetworkSize()]++
		runner.writeSubNetwork(network, network.Scores(), networksPerSize[network.subnetworkSize()])
	}
	//save time
	endTime := time.Since(start)
	err := runner.AppendLinesToFile(
		filepath.Join(runner.OutputFolder, "MO", "runTime"),
		[]string{fmt.Sprintf("%f", endTime.Seconds())},
	)
	if err != nil {
		runner.Error("Could not write running time to file", "err", err)
	}
	runner.Info("Finished MO optimization", "seconds", endTime.Seconds())
}

func (runner MORunner) writeSubNetwork(network subnetwork, scores []float64, idx int) {
	outDir := filepath.Join(runner.directory, fmt.Sprintf("size_%d", network.subnetworkSize()))
	fileio.CreateDirKeepContent(outDir)
	outFilePath := filepath.Join(outDir, fmt.Sprintf("result-%d.network", idx))
	// Write the optimal subnetwork to a file
	interactionTypeLines := runner.InteractionStore.InteractionTypeStringList()
	interactionLines := make([]string, 0, network.interactionCount())
	for interactionID := range network.Interactions() {
		interactionLines = append(interactionLines, interactionID.StringMinimal())
	}
	sort.Strings(interactionLines)
	err := runner.WriteLinesToNewFile(outFilePath, []string{fmt.Sprintf("%% score %v", scores)}, interactionTypeLines, interactionLines)
	if err != nil {
		runner.Error("Could not write subnetwork to file", "err", err)
	}
}
