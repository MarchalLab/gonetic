package optimization

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/fileio"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/compare"
	"github.com/MarchalLab/gonetic/internal/normalform"
	"github.com/MarchalLab/gonetic/internal/ranking"
	"github.com/MarchalLab/gonetic/internal/wfg"
)

// NSGAOptimization implements NSGA-II optimization according to https://ieeexplore.ieee.org/document/996017
type NSGAOptimization struct {
	// initialised fields
	*arguments.Common
	ranking.ObjectiveList[subnetwork]
	ObjectiveTypes     []objectiveType
	hvCalculatorOpt    int
	pathRepositories   *PathRepositories
	populationN        int
	numGenerations     int
	mutationChance     float64
	availableCores     int
	duplicateDetection bool
	initialPopulations int
	// computed fields
	Pt             []subnetwork
	Qt             []subnetwork
	gensPerWindow  int
	hyperVolumeArr []float64
	startTime      time.Time
	// track best front
	bestHypervolume      float64
	bestPopulation       int
	NetworkSizeOptimizer *NetworkSizeOptimizer
}

func creatObjectivesList(
	args *arguments.Common,
	dDNNFList [][]*normalform.NNF,
) (
	ranking.ObjectiveList[subnetwork],
	[]objectiveType,
) {
	objectives := make([]ranking.Objective[subnetwork], 0, 1+len(dDNNFList))
	objectiveTypes := make([]objectiveType, 0, 1+len(dDNNFList))
	if args.OptimizeNetworkSize {
		networkSizeObj := newNetworkSizeObjective()
		objectives = append(objectives, &networkSizeObj)
		objectiveTypes = append(objectiveTypes, networkSizeObjectiveType)
	}
	if args.OptimizeSampleCount {
		sampleObj := newSampleObjective(dDNNFList, args.SampleObjectiveType)
		objectives = append(objectives, &sampleObj)
		objectiveTypes = append(objectiveTypes, sampleObjectiveType)
	}
	for _, dDNNFs := range dDNNFList {
		dDNNFObj := newDDNNFObjective(dDNNFs)
		objectives = append(objectives, &dDNNFObj)
		objectiveTypes = append(objectiveTypes, dDNNFObjectiveType)
	}
	return objectives, objectiveTypes
}

func newNSGAOptimization(
	args *arguments.Common,
	pathRepositories *PathRepositories,
	dDNNFList [][]*normalform.NNF,
	popSize int,
	numGenerations int,
	mutChance float64,
	availableCores int,
) NSGAOptimization {
	// create objectives
	objectivesList, objectiveTypes := creatObjectivesList(args, dDNNFList)
	// create hv calculator options
	hvCalculatorOpt := 0
	if len(objectivesList) == 2 {
		hvCalculatorOpt = 2
	} else {
	}
	return NSGAOptimization{
		Common:             args,
		ObjectiveList:      objectivesList,
		ObjectiveTypes:     objectiveTypes,
		hvCalculatorOpt:    hvCalculatorOpt,
		pathRepositories:   pathRepositories,
		populationN:        popSize - popSize%2,
		numGenerations:     numGenerations,
		mutationChance:     mutChance,
		availableCores:     availableCores,
		duplicateDetection: true,
		initialPopulations: 10,
		bestHypervolume:    0,
		NetworkSizeOptimizer: NewNetworkSizeOptimizer(
			args.Logger,
			objectiveTypes,
			1,
			args.TargetNetworkSize,
			1000,
			args.FocusFraction,
		),
	}
}

// buildInitialPopulation generates the initial population of subnetworks for the NSGA optimization
func (opt *NSGAOptimization) buildInitialPopulation() {
	opt.Qt = make([]subnetwork, 0, opt.populationN)
	for i := 0; i < opt.populationN; i++ {
		networkSize := opt.NetworkSizeOptimizer.SampleNetworkSize()
		network := newFastSubnetwork(opt)
		opt.fillNetworkWithInitialPopulation(networkSize, network)
		opt.Qt = append(opt.Qt, network)
	}
}

func (opt *NSGAOptimization) fillNetworkWithInitialPopulation(sizeGoal int, network subnetwork) {
	// Initial build of a subnetwork, with size close to sizeGoal
	// Only attempt 10 * sizeGoal expansions, even if the required size is still not met at that point
	i := 0
	maxExpansions := sizeGoal * 10
	for ; network.subnetworkSize() < sizeGoal && i < maxExpansions; i++ {
		network.expansion()
	}
	if i >= maxExpansions {
		opt.Warn("max expansion limit reached", "size", network.subnetworkSize(), "target", sizeGoal)
	}
}

// fastNonDominatedSort performs fast non dominated sort with complexity O(M*N^2) based on paper NSGA-II
// Small optimization to terminate when size N is reached as suggested in paper is also implemented
// Non-domination level is written to subnetwork field, for binary selection tournament later
func (opt *NSGAOptimization) fastNonDominatedSort(targetSize int) map[int][]subnetwork {
	// concat Pt and Qt
	Rt := append(opt.Pt, opt.Qt...)

	// create return map
	fronts := make(map[int][]subnetwork)

	// get for each subnetwork the amount of subnetworks it is dominated by, and the subnetworks it dominates
	dominatedCounter := make(map[subnetwork]int)
	dominatingNetworks := make(map[subnetwork][]subnetwork)
	for i, p := range Rt {
		for j, q := range Rt {
			if i == j {
				// network never dominates itself
				continue
			}
			// TODO: check both p>q and p<q at once
			if opt.Dominates(p, q) {
				// if p dominates q, add q to dominated networks of p
				dominatingNetworks[p] = append(dominatingNetworks[p], q)
			} else if opt.Dominates(q, p) {
				// if q dominates p, increment domination counter of p
				dominatedCounter[p]++
			}
		}
		if dominatedCounter[p] == 0 {
			// if not dominated, append to first front and set nonDominationLevel
			fronts[1] = append(fronts[1], p)
			p.setNonDominationLevel(1)
		}
	}
	// final loop through networks
	k := 1
	totalNNetworks := len(fronts[1])
	// extra: early termination
	for len(fronts[k]) != 0 && totalNNetworks < targetSize {
		for _, p := range fronts[k] {
			for _, q := range dominatingNetworks[p] {
				// for each network in previous front, decrement the dominated counter of all networks it dominates
				dominatedCounter[q]--
				if dominatedCounter[q] == 0 {
					// if any of the counters reach 0, the network belongs to the next front
					fronts[k+1] = append(fronts[k+1], q)
					q.setNonDominationLevel(k + 1)
				}
			}
		}
		totalNNetworks += len(fronts[k+1])
		k++
	}
	return fronts
}

// crowdingDistanceAssignment assigns the crowding distance to each subnetwork in the given front
func (opt *NSGAOptimization) crowdingDistanceAssignment(front []subnetwork) {
	numSubnetworks := len(front)

	// Initialize crowding distance for each subnetwork
	for i := range front {
		front[i].setCrowdingDistance(0)
	}

	// Calculate crowding distance for each objective
	for i := 0; i < len(opt.ObjectiveList); i++ {
		// Sort the subnetworks based on the score of the i-th objective
		sort.Slice(front, func(j, k int) bool {
			return front[j].Scores()[i] < front[k].Scores()[i]
		})

		// Assign infinite crowding distance to boundary points
		front[0].setCrowdingDistance(math.Inf(1))
		front[numSubnetworks-1].setCrowdingDistance(math.Inf(1))

		normalization := front[numSubnetworks-1].Scores()[i] - front[0].Scores()[i]

		if normalization == 0 {
			// Avoid division by zero
			continue
		}
		// Calculate the normalized distance for intermediate points
		for j := 1; j < numSubnetworks-1; j++ {
			nextScore := front[j+1].Scores()[i]
			prevScore := front[j-1].Scores()[i]
			front[j].setCrowdingDistance(front[j].CrowdingDistance() + (nextScore-prevScore)/normalization)
		}
	}
}

// crowdedSort sorts a list of subnetworks based on their non-domination level and crowding distance.
// The subnetworks are first sorted by their non-domination level in ascending order.
// If two subnetworks have the same non-domination level, they are then sorted by their crowding distance in descending order.
// Returns: sorted list of subnetworks
func (opt *NSGAOptimization) crowdedSort(subnetworks []subnetwork) []subnetwork {
	sort.Slice(subnetworks, func(i, j int) bool {
		if subnetworks[i].NonDominationLevel() != subnetworks[j].NonDominationLevel() {
			return subnetworks[i].NonDominationLevel() < subnetworks[j].NonDominationLevel()
		} else {
			return subnetworks[i].CrowdingDistance() > subnetworks[j].CrowdingDistance()
		}
	})
	// Returns: sorted list of subnetworks
	return subnetworks
}

// limitedChange changes old towards new, with a maximal change of 10% and a minimal change of 1
func limitedChange(old, new int) int {
	// no change required
	if old == new {
		return old
	}
	maxChange := max(1, int(0.1*float64(old)))
	// new is smaller
	if new < old {
		return old - min(old-new, maxChange)
	}
	// new is larger
	return old + min(new-old, maxChange)
}

// generateOffspring is a method that generates offspring by performing selection, crossover, and mutation operations.
// It builds a population of `populationN` children, saved in `Qt` for further processing.
// - Selection: It uses binary selection tournament to select two parent networks. The best one becomes the parent.
// - Crossover: It performs a crossover operation between the two parents to create a new child network.
// - Mutation: It applies mutation to the child network with a random chance.
//   - If `duplicateDetection` is enabled, the method checks if the child is already present in the population and generates a new child if it is.
//
// The target size for the child network is randomly generated between `minNetworkSize` and `maxNetworkSize`.
func (opt *NSGAOptimization) generateOffspring() {
	// build populationN children
	opt.Qt = make([]subnetwork, 0, opt.populationN)
	regenerations := 0
	regenerationLimit := 10
	for i := 0; i < opt.populationN; i++ {
		child := opt.generateChild()
		isDuplicateChild := opt.duplicateDetection && opt.detectDuplicates(child, i)
		if isDuplicateChild && regenerations < regenerationLimit {
			// retry with new child
			i -= 1
			regenerations += 1
			continue
		}
		// accept this child
		if isDuplicateChild {
			opt.Warn("Duplicate detected, could not find unique child",
				"index", i,
				"attempts", regenerations,
			)
		} else {
		}
		opt.Qt = append(opt.Qt, child)
		regenerations = 0
	}
}

func (opt *NSGAOptimization) binaryTournamentParent() subnetwork {
	// binary selection tournament: select two networks, best one gets to be the parent
	ptSize := len(opt.Pt)
	candidate1 := opt.Pt[rand.Intn(ptSize)]
	candidate2 := opt.Pt[rand.Intn(ptSize)]
	return opt.crowdedSort([]subnetwork{candidate1, candidate2})[0]
}

func (opt *NSGAOptimization) selectParentPaths() (parentPathIDs [2][]PathID) {
	parents := [2]subnetwork{
		opt.binaryTournamentParent(),
		opt.binaryTournamentParent(),
	}
	for i := 0; i < 2; i++ {
		parentPathIDs[i] = make([]PathID, len(parents[i].SelectedPaths()))
		j := 0
		for pathID := range parents[i].SelectedPaths() {
			parentPathIDs[i][j] = pathID
			j++
		}
	}
	return
}

func (opt *NSGAOptimization) generateChild() subnetwork {
	// select parents and get their paths
	parentPathIDs := opt.selectParentPaths()

	// track duplicates
	addedPaths := make(map[PathID]struct{})

	// initialize child
	child := newFastSubnetwork(opt)

	// generate random target size
	targetSize := opt.NetworkSizeOptimizer.SampleNetworkSize()

	// perform crossover
	childDone := false
	currentParent := 0
	for !childDone {
		if len(parentPathIDs[currentParent]) == 0 {
			// only one parent with paths left, switch back
			currentParent = 1 - currentParent
		}
		if len(parentPathIDs[currentParent]) == 0 {
			// both parents have no paths left, break
			break
		}

		// get random path from current parent and add to child
		nPaths := len(parentPathIDs[currentParent])
		idx := rand.Intn(nPaths)
		pathID := parentPathIDs[currentParent][idx]

		// delete this element from paths slice, by copying last element and then removing last element from slice
		parentPathIDs[currentParent][idx] = parentPathIDs[currentParent][nPaths-1]
		parentPathIDs[currentParent] = parentPathIDs[currentParent][:nPaths-1]

		// check for duplicate paths
		if _, ok := addedPaths[pathID]; ok {
			continue
		}
		addedPaths[pathID] = struct{}{}

		// add selected path to child
		path := opt.pathRepositories.pathInteractionSetFromId(pathID)
		child.addSelectedPath(pathID, path)
		for interactionID := range path {
			child.addInteraction(interactionID)
		}

		// switch parents
		currentParent = 1 - currentParent

		// check if done, i.e. target size reached
		if child.subnetworkSize() >= targetSize {
			childDone = true
		}
	}

	// mutate with random chance
	if rand.Float64() < opt.mutationChance {
		// ONLY EXPANSION: (almost) all networks achieved by reduction can be achieved by crossover with lower target size
		if child.subnetworkSize() < opt.NetworkSizeOptimizer.CurrentMax {
			child.expansion()
		}
	}
	return child
}

// detectDuplicates checks whether child is already present in the population, by comparing path lists
func (opt *NSGAOptimization) detectDuplicates(child subnetwork, index int) bool {
	// detect if child already present in population
	if isDuplicate(child, opt.Pt) {
		return true
	}
	// detect if child already present in offspring
	if isDuplicate(child, opt.Qt[:index]) {
		return true
	}
	// child is unique
	return false
}

func isDuplicate(child subnetwork, population []subnetwork) bool {
	for _, network := range population {
		if haveSameKeys(
			network.Interactions(),
			child.Interactions(),
		) {
			return true
		}
	}
	return false
}

// Optimize is the main loop of the NSGA-II algorithm and returns the final approximation front Pt
func (opt *NSGAOptimization) Optimize() []subnetwork {
	// main function to perform NSGA-II optimization, returns P_t at end of runtime
	opt.Info("start optimisation",
		"cores", opt.availableCores,
		"population", opt.populationN,
		"generations", opt.numGenerations,
		"initialPopulations", opt.initialPopulations,
	)
	// load population from file
	t := -1
	if opt.Resume {
		opt.populationFromFile("", -1)
		t, _ = opt.findPrefixFileWithHighestNumber("")
	}
	if t < 0 {
		// initialize population if not loaded
		opt.initialPopulation()
		t = 0
		opt.hyperVolumeArr = make([]float64, 0)
	} else {
		opt.hyperVolumeFromFile()
		for len(opt.hyperVolumeArr) <= t {
			opt.hyperVolumeArr = append(opt.hyperVolumeArr, 0)
		}
		if len(opt.hyperVolumeArr) > t+1 {
			opt.hyperVolumeArr = opt.hyperVolumeArr[:t+1]
		}
		opt.bestHypervolume = opt.hyperVolumeArr[t]
		opt.bestPopulation = t
		opt.populationToFile("best_", t)
		opt.Info("loaded population from file",
			"hypervolume", opt.bestHypervolume,
		)
		t += 1
	}
	// compute optimization windows
	opt.computeWindows()
	done := t >= opt.numGenerations
	if done {
		opt.Info("nothing to optimize, increase number of iterations", "generation", t)
		return opt.Pt
	}
	//start main loop
	opt.startTime = time.Now()
	for ; !done; t++ {
		// compute target network size
		opt.NetworkSizeOptimizer.ComputeTargetNetworkSize(opt.Pt)
		// compute next generation
		opt.computeGeneration(t, opt.populationN)
		// write P_t to file
		opt.populationToFile("", t)
		// check stop conditions
		done = opt.isDone(t)
		// update bestFront, i.e. the population where fronts[1] has the highest hypervolume
		// if the current hypervolume is better
		if opt.hyperVolumeArr[t] > opt.bestHypervolume || opt.bestHypervolume == 0 {
			opt.bestHypervolume = opt.hyperVolumeArr[t]
			opt.bestPopulation = t
			opt.populationToFile("best_", t)
		}
	}
	endTime := time.Since(opt.startTime)
	opt.Info("finished optimization",
		"seed reshufles", opt.pathRepositories.reshuffles,
		"total generations", t,
		"total time", fmt.Sprintf("%.3f s", endTime.Seconds()),
		"time per generation", fmt.Sprintf("%.3f s", endTime.Seconds()/float64(t)),
	)
	// load best population
	opt.populationFromFile("best_", opt.bestPopulation)
	opt.Info("best population",
		"generation", opt.bestPopulation,
		"hv", opt.bestHypervolume,
	)
	// write hyperVolumes to file
	opt.hypervolumeToFile()
	// return final approximation front
	return opt.Pt
}

// initialPopulation creates the initial population of subnetworks for the NSGA-II optimization
func (opt *NSGAOptimization) initialPopulation() {
	// build initial population and calculate score (if only 1 core, score is calculated during sorting)
	P0 := make([]subnetwork, 0, opt.populationN)
	for i := 0; i < opt.initialPopulations; i++ {
		opt.buildInitialPopulation()
		opt.Pt = make([]subnetwork, 0, opt.populationN/10)
		opt.computeGeneration(-opt.initialPopulations+i, opt.populationN/10)
		P0 = append(P0, opt.Pt...)
		opt.Info("initial population",
			"id", i+1,
			"size", len(opt.Qt),
			"total", len(P0),
		)
	}
	opt.Pt = P0
}

// arrayToString converts an array into a string with a custom delimiter
func arrayToString[S any](a []S, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

// computeWindows calculates the number of generations based on the calculations done in the thesis.
// It also determines the window size based on the WindowCount, MinWindowSize, and MaxWindowSize parameters.
// If the calculated window size is smaller than the minimum window size, it adjusts the window size and the termination generations accordingly.
// If the calculated window size is larger than the maximum window size, it adjusts the window size.
// Finally, it sets the number of generations and logs the calculated termination generations, window size, and minimum improvement percentage.
func (opt *NSGAOptimization) computeWindows() {
	// by default, number of generations is based on calculations done in thesis
	termGens := opt.calculateGenerations()
	// These values control the window of generations and minimal improvement, could be parameters
	windowCount := opt.WindowCount
	opt.gensPerWindow = termGens / windowCount
	minWindowSize := opt.MinWindowSize
	maxWindowSize := opt.MaxWindowSize
	if opt.gensPerWindow < minWindowSize {
		opt.gensPerWindow = minWindowSize
		termGens = windowCount*minWindowSize + windowCount - 1
	}
	if opt.gensPerWindow > maxWindowSize {
		opt.gensPerWindow = maxWindowSize
	}
	opt.numGenerations = termGens
	opt.Info("calculated termination generations",
		"maximal generations", termGens,
		"window size", opt.gensPerWindow,
		"minimal progress", fmt.Sprintf("%.3f", opt.RequiredProgressPercentage),
	)
}

// isDone checks if the optimization process should be terminated
// Check maximum amount of generations
func (opt *NSGAOptimization) isDone(t int) bool {
	message := "computed generation"
	hyperVolume := 0.0
	progress := 0.0
	hoursPassed := time.Since(opt.startTime).Hours()
	defer (func() {
		opt.Info(message,
			"generation", t,
			"hyper volume", hyperVolume,
			"progress", fmt.Sprintf("%.3f", progress*100)+"%",
			"time", fmt.Sprintf("%.3fh", hoursPassed),
			"next minSize", opt.NetworkSizeOptimizer.CurrentMin,
			"next maxSize", opt.NetworkSizeOptimizer.CurrentMax,
			"window", opt.gensPerWindow,
		)
	})()
	hyperVolume = opt.hyperVolumeArr[t]
	// Early termination using minimal percentage improvement in window of generations
	if t > 0 && t > opt.gensPerWindow {
		previousDifference := opt.hyperVolumeArr[t-opt.gensPerWindow] - opt.hyperVolumeArr[0]
		if previousDifference > 0 {
			progress = (hyperVolume - opt.hyperVolumeArr[t-opt.gensPerWindow]) / previousDifference
		}
		if opt.EarlyTermination && progress*100 < opt.RequiredProgressPercentage {
			message = "early termination"
			return true
		}
	}
	// Check maximum amount of generations
	if t+1 >= opt.numGenerations {
		message = "termination"
		return true
	}
	// Early termination when out of time
	if opt.TimeLimitHours > 0 && hoursPassed >= opt.TimeLimitHours {
		message = "early termination"
		return true
	}
	// Continue
	return false
}

// computeGeneration performs a single iteration of the main loop of the NSGA-II algorithm
// the final approximation front is stored in Pt
// TODO: what is space efficiency for first iteration ?
// TODO: why is it slow with large number of paths ?
func (opt *NSGAOptimization) computeGeneration(t, targetSize int) {
	// log generation
	opt.Debug("start generation",
		"generation", t,
	)
	opt.DumpProfiles(fmt.Sprintf("opt-gen-%d", t))
	if t >= 0 {
		// generate offspring
		opt.generateOffspring()
		opt.Debug("Generated offspring")
	}

	// calculate scores
	opt.parallelScoreCalc()
	// do non-dominated sorting
	fronts := opt.fastNonDominatedSort(targetSize)
	opt.Debug("Performed non-dominated sort")
	//select new population
	selected := make([]subnetwork, 0, targetSize)
	// add fronts until population size is reached
	for i := 1; i <= len(fronts); i++ {
		front := fronts[i]
		// calculate crowdingDistanceAssignment
		opt.crowdingDistanceAssignment(front)
		if len(selected)+len(front) > targetSize {
			// sort split up front
			front := opt.crowdedSort(front)
			opt.Debug("Performed crowded sort")
			length := compare.Between(0, targetSize-len(selected), len(front))
			// update front for hypervolume calculation
			if i == 1 {
				fronts[i] = front[:length]
			}
			selected = append(selected, front[:length]...)
		} else {
			selected = append(selected, front...)
		}
		if len(selected) == targetSize {
			break
		}
	}

	// update population
	opt.Pt = selected
	// calculate hyperVolume and store for early termination calculation
	if t >= 0 {
		hyperVolume := opt.calculateHyperVolume(t, fronts[1])
		opt.hyperVolumeArr = append(opt.hyperVolumeArr, hyperVolume)
	}
}

// parallelScoreCalc calculates scores for subnetworks in parallel
// It checks if scores for each subnetwork in (opt.Pt + opt.Qt) are already computed
// And computes them if not
// It uses goroutines to parallelize the score calculation
// The scores are cached in the subnetworks themselves
// Waits for all goroutines to finish before returning
func (opt *NSGAOptimization) parallelScoreCalc() {
	nets := append(opt.Pt, opt.Qt...)
	// loop over networks using goroutines
	var wg sync.WaitGroup
	wg.Add(len(nets))
	for _, network := range nets {
		go func(network subnetwork, wg *sync.WaitGroup) {
			defer wg.Done()
			opt.ObjectiveList.SetScores(network)
		}(network, &wg)
	}
	wg.Wait()
}

func (opt *NSGAOptimization) populationToFile(prefix string, t int) {
	outDir := filepath.Join(
		opt.OutputFolder,
		"MO",
		"population",
	)
	outFile := filepath.Join(
		outDir,
		fmt.Sprintf("%s%d", prefix, t),
	)
	// Check if the file already exists
	if _, err := os.Stat(outFile); err == nil {
		opt.Info("File already exists", "file", outFile)
		return
	}
	// write P_t to file
	lines := make([]string, 0, len(opt.Pt))
	for _, network := range opt.Pt {
		lines = append(lines, network.String())
	}
	sort.Strings(lines)
	err := opt.WriteLinesToNewFile(
		outFile,
		lines,
	)
	if err != nil {
		return
	}
	// wipe previous population files
	opt.removeOldPopulationFiles(outDir, prefix, t, 3)
}

// removeOldPopulationFiles removes old population files, keeping the X most recent
func (opt *NSGAOptimization) removeOldPopulationFiles(outDir, prefix string, currentGen, keepRecent int) {
	files, err := os.ReadDir(outDir)
	if err != nil {
		opt.Error("Failed to read directory", "dir", outDir, "err", err)
		return
	}

	// Collect all matching files
	var generations []int
	fileMap := make(map[int]string)

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
			name := strings.TrimPrefix(file.Name(), prefix)
			if gen, err := strconv.Atoi(name); err == nil && gen < currentGen {
				filePath := filepath.Join(outDir, file.Name())
				generations = append(generations, gen)
				fileMap[gen] = filePath
			}
		}
	}

	// Sort generations in descending order (newest first)
	sort.Sort(sort.Reverse(sort.IntSlice(generations)))

	// Keep the X most recent, delete the rest
	for i := keepRecent; i < len(generations); i++ {
		gen := generations[i]
		filePath := fileMap[gen]
		if err := os.Remove(filePath); err != nil {
			opt.Warn("Failed to remove old file", "file", filePath, "err", err)
		} else {
			opt.Info("Removed old file", "file", filePath)
		}
	}
}

// findPrefixFileWithHighestNumber finds the file with the highest number in the given directory
func (opt *NSGAOptimization) findPrefixFileWithHighestNumber(prefix string) (int, string) {
	dir := filepath.Join(opt.OutputFolder, "MO", "population")
	files, err := os.ReadDir(dir)
	if err != nil {
		opt.Error("Failed to read directory", "dir", dir, "err", err)
		return -1, ""
	}
	fileName := ""
	generation := -1
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			name = strings.TrimPrefix(name, prefix)
			if t, err := strconv.Atoi(name); err == nil {
				if t > generation {
					generation = t
					fileName = file.Name()
				}
			}
		}
	}
	return generation, fileName
}

// findPopulationFile finds the latest population file
func (opt *NSGAOptimization) findPopulationFile() (int, string) {
	for _, prefix := range []string{"best_", ""} {
		lastGen, lastFile := opt.findPrefixFileWithHighestNumber(prefix)
		if lastFile != "" {
			return lastGen, lastFile
		}
	}
	return -1, ""
}

// populationFromFile reads the population from a file
func (opt *NSGAOptimization) populationFromFile(prefix string, t int) int {
	fileName := fmt.Sprintf("%s%d", prefix, t)
	if t < 0 {
		t, fileName = opt.findPopulationFile()
		opt.Info("found population file",
			"generation", t,
			"file", fileName,
		)
	}
	if t == -1 {
		return -1
	}
	// Read the population from the file with the highest number
	populationFile := filepath.Join(opt.OutputFolder, "MO", "population", fileName)
	opt.Info("reading population from file",
		"generation", t,
		"file", populationFile,
	)
	lines := fileio.ReadListFromFile(populationFile, false)
	opt.Pt = make([]subnetwork, 0, len(lines))
	opt.Qt = make([]subnetwork, 0, opt.populationN)
	minNetworkSize := math.MaxInt64
	maxNetworkSize := 0
	for _, line := range lines {
		// Parse the line to create a subnetwork
		subnet := newFastSubnetwork(opt)
		subnet.ParseString(line)
		opt.Pt = append(opt.Pt, subnet)
		// Update the min and max network sizes
		minNetworkSize = min(minNetworkSize, subnet.subnetworkSize())
		maxNetworkSize = max(maxNetworkSize, subnet.subnetworkSize())
	}
	opt.NetworkSizeOptimizer = NewNetworkSizeOptimizer(
		opt.Logger,
		opt.ObjectiveTypes,
		minNetworkSize,
		maxNetworkSize,
		1000,
		opt.FocusFraction,
	)
	return t
}

// hypervolumeToFile writes the hypervolume values to a file
func (opt *NSGAOptimization) hypervolumeToFile() {
	err := opt.WriteLinesToNewFile(
		filepath.Join(opt.OutputFolder, "MO", "hyperVolumes"),
		[]string{arrayToString(opt.hyperVolumeArr, "\n")},
	)
	if err != nil {
		opt.Error("failed to write hyper volumes", "err", err)
	}
}

// hyperVolumeFromFile reads the hypervolume values from a file
func (opt *NSGAOptimization) hyperVolumeFromFile() {
	filename := filepath.Join(opt.OutputFolder, "MO", "hyperVolumes")
	lines := fileio.ReadListFromFile(filename, false)
	hvArr := make([]float64, 0, len(lines))
	for _, line := range lines {
		hv, _ := strconv.ParseFloat(line, 64)
		hvArr = append(hvArr, hv)
	}
	opt.hyperVolumeArr = hvArr
}

// calculateHyperVolume approximates the hypervolume enclosed by the front and the reference point 0*
// It assumes all objectives are to be maximized
func (opt *NSGAOptimization) calculateHyperVolume(t int, front []subnetwork) float64 {
	startTime := time.Now()
	points := wfg.ConvertToFront(front)
	wfgFile := filepath.Join(opt.EtcPathAsString, fmt.Sprintf("wfg%d", opt.hvCalculatorOpt))
	frontFile := filepath.Join(opt.FrontsDirectory(), fmt.Sprintf("%d", t))
	frontFile = wfg.CreateWfgInput(opt.FileWriter, frontFile, points)
	err := wfg.RunWfg(opt.FileWriter, wfgFile, frontFile)
	if err != nil {
		opt.Error("Failed to compute hypervolume", "err", err)
		return 0
	}
	hv := wfg.ParseWfgResult(opt.Logger, frontFile)
	wfgTime := time.Since(startTime)
	opt.Debug("front HV",
		"front", len(front),
		"points", len(points),
		"wfg.opt", opt.hvCalculatorOpt,
		"hv", hv,
		"time", wfgTime.String(),
	)
	return hv
}

// calculateGenerations calculates the number of generations based on the input parameters.
// If opt.NumGens is greater than 0, the value of opt.NumGens is used as the number of generations.
// If opt.NumGens is less than 0, the number of generations is computed using the following formula:
//   - Calculate the expected amount of paths present in the initial population using the formula: ENInit = n * (1 - ((n-1)/n)^k),
//     where n is the number of paths in the path repository and k is a constant computed based on the input parameters.
//   - Approximate the harmonic number using the formula: H = ln(n-ENInit) + EULER + 1/(2*(n-ENInit)) - 1/(12*(n-ENInit)^2),
//     where EULER is a predefined constant.
//   - Compute the number of generations using the formula: generations = ceil((n - ENInit) * H / (populationN * MutChance)).
//
// The calculated number of generations is returned by the method.
func (opt *NSGAOptimization) calculateGenerations() int {
	generations := int(1e7)
	if opt.NumGens > 0 {
		generations = opt.NumGens
	} else if opt.NumGens < 0 {
		EULER := 0.577215665
		n := float64(opt.pathRepositories.NumberOfPaths())
		k := (opt.NetworkSizeOptimizer.CurrentMax + opt.NetworkSizeOptimizer.CurrentMin) * opt.populationN / (opt.PathLength + 2)

		// expected amount of paths present in initial population
		ENInit := n * (1 - math.Pow((n-1)/n, float64(k)))

		// approximation for harmonic number
		H := math.Log(n-ENInit) + EULER + 1/(2*(n-ENInit)) - 1/(12*math.Pow(n-ENInit, 2))

		generations = int(math.Ceil((n - ENInit) * H / (float64(opt.populationN) * opt.MutChance)))
	}
	return generations
}
