package cmd

import (
	"runtime"
	"runtime/debug"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
)

func init() {
	commonArguments.Commit = getCommit()

	// flags with shorthand
	// a - used in expression
	rootCmd.PersistentFlags().StringSliceVarP(&commonArguments.BannedNetworkFiles, "banned-network-file", "b", []string{}, "Path to the blacklisted network file. Each interaction in this file will be removed from the network, if it occurs in the --network-file networks. This parameter can be repeated. See example files for the format.")
	// c - used in QTL
	// d - used in pair / QTL
	// e - used in QTL
	// f - used in pair / QTL
	// g - used in expression
	// h "help"
	rootCmd.PersistentFlags().StringVarP(&commonArguments.TopologyWeightingAddition, "topology-weighting-addition", "i", "bayes", "Weighting addition method to reweight the network based on network topology. Valid values are: none, bayes, mean, and mult.")
	// j - used in expression
	// k - used in expression
	rootCmd.PersistentFlags().IntVarP(&commonArguments.PathLength, "path-length", "l", 4, "The maximum PathLength to be explored. Only values of 3 and 4 are realistic as lower values are biologically not relevant and higher values are impossible to calculate. 4 is seen as ideal.")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.MappingFile, "mapping-file", "m", "", "If mapping from e.g. systematic to trivial names is required, the path to the mapping file must be given here. A header in the format \"# from,to\" must be present at the top of the file (comma separated).")
	rootCmd.PersistentFlags().StringSliceVarP(&commonArguments.NetworkFiles, "network-file", "n", []string{}, "Path to the network file. This parameter can be repeated. See example files for the format.")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.OutputFolder, "output-folder", "o", "", "Path to desired output folder.")
	// p - used in QTL
	rootCmd.PersistentFlags().StringVarP(&commonArguments.EtcPathAsString, "etc-path-location", "q", "etc", "Path to the etc directory. Will assume etc directory is in the current directory when left blank.")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.Resume, "resume", "r", false, "Resume the optimization.")
	rootCmd.PersistentFlags().Float64VarP(&commonArguments.SldCutoff, "search-tree-cutoff", "s", -1, "The minimum probability a path must have to be retained. Paths with a lower probability are regarded as not biologically relevant. By default the sldCutoff will be estimated based on the weighted network.")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.MaxPaths, "max-paths", "t", 0, "Maximal number of paths used in optimization phase.")
	// u - used in QTL
	// v
	// w
	// x
	// y
	// z

	// flags without shorthand
	rootCmd.PersistentFlags().Float64VarP(&commonArguments.MinEdgeScore, "min-edge-score", "", 0.0, "The minimal edge score, lower scoring edges are rejected")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.BestPathCount, "best-path-count", "", 25, "Number of paths per possible pair. Increasing this might yield better results but is at the expense of longer computational times")

	// Resource flags
	rootCmd.PersistentFlags().IntVarP(&commonArguments.NumCPU, "numCPU", "", runtime.NumCPU()-2, "Limits the number of logical CPU cores that can be used.")

	// Logging flags
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.Verbose, "verbose", "", false, "Verbose logging")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.LogFormat, "log-format", "", "json", "Log output format, possible values are text or json.")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.LogFile, "log-file", "", "", "Log output file, stdout if not provided, possible values are auto, for default log file location, or a valid file location.")

	// Profiling flags
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipAllocsProfiling, "skip-allocs-profiling", "", true, "Skip the allocs profiling")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipGoroutineProfiling, "skip-goroutine-profiling", "", true, "Skip the goroutine profiling")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipHeapProfiling, "skip-heap-profiling", "", true, "Skip the heap profiling")

	// Precomputed files flags
	rootCmd.PersistentFlags().StringVarP(&commonArguments.UseIndex, "use-index", "", "", "Use gene index in the provided file as a base.")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.UsePaths, "use-paths", "", "", "Use paths in the given directory instead of finding new paths. Requires --use-index. Enables --skip-path-finding.")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.UseNNFs, "use-nnfs", "", "", "Use NNFs in the given directory instead of compiling from paths. Requires --use-paths. Enables --skip-compilation.")

	// Skip flags
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipPathFinding, "skip-path-finding", "", false, "Skip the path finding step to reuse existing paths")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipNetworkPrinting, "skip-network-printing", "", true, "Skip the network printing during path finding")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipCompilation, "skip-compilation", "", false, "Skip the compilation step to reuse existing NNFs")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipOptimization, "skip-optimization", "", false, "Skip the optimization step to reuse existing subnetworks")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.SkipInterpreter, "skip-interpreter", "", false, "Skip the interpreter step")

	// Multi objective core parameters
	rootCmd.PersistentFlags().IntVarP(&commonArguments.NumGens, "generations-count", "", -1, "The amount of generations used in the multi-objective optimization algorithm")
	rootCmd.PersistentFlags().Float64VarP(&commonArguments.MutChance, "mutation-chance", "", 0.5, "The mutation chance used by the multi-objective optimization algorithm")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.PopSize, "population-size", "", 500, "The population size used by the multi-objective optimization algorithm")

	// Scoring parameters
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.OptimizeNetworkSize, "optimize-network-size", "", true, "Force parsimony pressure on the network size")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.OptimizeSampleCount, "optimize-sample-count", "", true, "Maximize the explained sample count during the optimization")
	rootCmd.PersistentFlags().StringVarP(&commonArguments.SampleObjectiveType, "sample-objective-type", "", "entropy", "The type of sample objective to use. Possible values are \"entropy\" or \"effective\".")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.TargetNetworkSize, "target-network-size", "x", 100, "The target network size used in the optimization")

	// Multi objective early termination
	rootCmd.PersistentFlags().Float64VarP(&commonArguments.TimeLimitHours, "max-hours", "", 0, "The maximum number of hours to run the optimization before forced termination. By default no time limit is imposed")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.WindowCount, "window-count", "", 20, "The number of generation windows used in the optimization")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.MinWindowSize, "min-window-size", "", 2, "The minimal windows size for the optimization")
	rootCmd.PersistentFlags().IntVarP(&commonArguments.MaxWindowSize, "max-window-size", "", 100, "The maximal windows size for the optimization")
	rootCmd.PersistentFlags().BoolVarP(&commonArguments.EarlyTermination, "early-termination", "", true, "Stop early when the required progress is reached")
	rootCmd.PersistentFlags().Float64VarP(&commonArguments.RequiredProgressPercentage, "required-progress", "", 0.25, "The minimal required progress over an optimization window before forced termination")

	// Deprecated flags
	// defaultDeprecationMessage := "This flag is deprecated and will be removed in a future version."
}

var commonArguments = arguments.NewCommon()

func getCommit() string {
	commit := "unknown"
	modified := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value[:8]
			}
			if setting.Key == "vcs.modified" {
				modified = true
			}
		}
	}
	if modified {
		commit += "-dirty"
	}
	return commit
}
