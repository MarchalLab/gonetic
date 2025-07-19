package cmd

import (
	"fmt"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/run"

	"github.com/spf13/cobra"
)

func init() {
	initQTL(qtl)

	// add the command to the root
	rootCmd.AddCommand(qtl)
}

func initQTL(qtlCmd *cobra.Command) {
	// flags with shorthand
	qtlCmd.PersistentFlags().BoolVarP(&qtlSpecificArguments.FreqIncrease, "freq-increase", "c", true, "Boolean whether or not to include frequency increases.")
	qtlCmd.PersistentFlags().StringVarP(&qtlSpecificArguments.MutationDataFile, "mutation-data-file", "d", "", "File containing all mutations. See example files for the format.")
	qtlCmd.PersistentFlags().BoolVarP(&qtlSpecificArguments.FuncScore, "func-score", "e", true, "Boolean whether or not to include functional scores.")
	qtlCmd.PersistentFlags().Float64VarP(&qtlSpecificArguments.FreqCutoff, "frequency-cutoff", "f", 1, "The minimal frequency increase which is not considered as a measurement error. Default is 1%. This can be ignored when frequency increase is not used.")
	qtlCmd.PersistentFlags().Float64VarP(&qtlSpecificArguments.MutRateParam, "mutation-rate-parameter", "p", 3, "The parameter which determines how populations with significantly higher mutation rates are corrected. Higher values of this parameter will punish population with a higher mutation rate less (weights will be closer to 1). Setting this parameter to 0 will cause GoNetic to ignore any populations with a significantly higher mutation rate with respect to the other populations. Mutation rate parameter should be between 0 and 3.5")
	qtlCmd.PersistentFlags().StringVarP(&qtlSpecificArguments.OutlierPopulations, "outliers-file", "u", "", "File containing a list of the populations. This should be a file with a population name in each line. If no file is provided, the algorithm calculates the outliers using the modified Z-score. If no weighting is to be performed, provide a blank file")

	// flags without shorthand
	qtlCmd.PersistentFlags().BoolVarP(&qtlSpecificArguments.Correction, "correction", "", false, "Boolean whether or not to include mutator outlier correction.")
	qtlCmd.PersistentFlags().BoolVarP(&qtlSpecificArguments.WithinCondition, "within-condition", "", false, "Enable if paths should also be detected within a condition.")
}

var qtlSpecificArguments = arguments.QTLSpecific{}

var qtlArguments = arguments.QTL{
	Common:      commonArguments,
	QTLSpecific: &qtlSpecificArguments,
}

var qtl = &cobra.Command{
	Use:   "QTL [options...]",
	Short: "complete QTL analysis",
	Long:  `complete QTL analysis.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := qtlArguments.Init()
		if err != nil {
			panic(fmt.Sprintf("Error initializing QTL arguments %s", err))
		}
		fmt.Println("Print: " + strings.Join(args, " "))
		qtlArguments.Info("Running EQTL",
			"args", fmt.Sprintf("%+v", args),
			"QTL", fmt.Sprintf("%+v", qtlArguments),
			"Common", fmt.Sprintf("%+v", qtlArguments.Common),
			"QTLSpecific", fmt.Sprintf("%+v", qtlArguments.QTLSpecific),
		)
		run.QTL(&qtlArguments)
	},
}
