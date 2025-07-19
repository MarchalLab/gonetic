package cmd

import (
	"fmt"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/run"

	"github.com/spf13/cobra"
)

func init() {
	initExpression(expression)

	// add the command to the root
	rootCmd.AddCommand(expression)
}

func initExpression(expressionCmd *cobra.Command) {
	// flags with shorthand
	expressionCmd.PersistentFlags().StringVarP(&expressionArguments.ExpressionFile, "expression-file", "a", "", "File containing the differential expression data. This should be tab delimited with following informatiom <<gene name>> <<P value> <<LFC>> <<condition>> with the appropriate header.")
	expressionCmd.PersistentFlags().StringVarP(&expressionArguments.DifferentialExpressionList, "differential-expression-file", "g", "", "Path to file containing which genes to consider as differentially expressed. This should be tab delimited with following information <<gene name>> <<condition>> with the appropriate header.")
	expressionCmd.PersistentFlags().StringVarP(&expressionArguments.ExpressionWeightingMethod, "expression-weighting-method", "j", "lfc", "Weighting method to use for expression networks. Valid values are: none, lfc, and zscore.")
	expressionCmd.PersistentFlags().StringVarP(&expressionArguments.ExpressionWeightingAddition, "expression-weighting-addition", "k", "mean", "Weighting addition method to use for expression networks. Valid values are: none, bayes, mean, and mult.")

	// flags without shorthand
	expressionCmd.PersistentFlags().Float64VarP(&expressionArguments.ExpressionWeightingDefault, "expression-weighting-default", "", 0.0, "The default weighting probability when no expression data is available.")
	expressionCmd.PersistentFlags().BoolVarP(&expressionArguments.DownUpstream, "down-up-stream", "", true, "Boolean flag to determine GoNetic expression mode: true runs the 'downstream' configuration, false the 'upstream' configuration.")
	expressionCmd.PersistentFlags().BoolVarP(&expressionArguments.PrintGeneScoreMap, "print-gene-score-map", "", false, "Boolean flag to print the gene scores used for weighting the network.")
}

var expressionArguments = arguments.Expression{
	Common: commonArguments,
}

var expression = &cobra.Command{
	Use:   "expression [options...]",
	Short: "complete QTL analysis: expression",
	Long:  `complete QTL analysis: expression.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := expressionArguments.Init()
		if err != nil {
			panic(fmt.Sprintf("Error initializing expression arguments %s", err))
		}
		fmt.Println("Print: " + strings.Join(args, " "))
		expressionArguments.Info("Running expression",
			"args", fmt.Sprintf("%+v", args),
			"Expression", fmt.Sprintf("%+v", expressionArguments),
			"Common", fmt.Sprintf("%+v", expressionArguments.Common),
		)
		run.Expression(&expressionArguments)
	},
}
