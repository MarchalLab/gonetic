package cmd

import (
	"fmt"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/run"

	"github.com/spf13/cobra"
)

func init() {
	// flags with shorthand

	// flags without shorthand
	eqtl.PersistentFlags().BoolVarP(&eqtlArguments.Regulatory, "regulatory", "", false, "Boolean whether or not the last interaction pointing to a differential expressed gene should be a regulatory interaction")
	eqtl.PersistentFlags().BoolVarP(&eqtlArguments.WithExpression, "with-expression", "", false, "Boolean whether or not to use within-sample expression optimisation")
	eqtl.PersistentFlags().BoolVarP(&eqtlArguments.WithMutation, "with-mutation", "", true, "Boolean whether or not to use across-sample mutation optimisation")

	// initialise expression args and add the command to the root
	initExpression(eqtl)
	initQTL(eqtl)

	// add the command to the root
	rootCmd.AddCommand(eqtl)
}

var eqtlArguments = arguments.EQTL{
	Expression:  &expressionArguments,
	QTLSpecific: &qtlSpecificArguments,
}

var eqtl = &cobra.Command{
	Use:   "EQTL [options...]",
	Short: "complete EQTL analysis: EQTL",
	Long:  `complete EQTL analysis: EQTL.`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := eqtlArguments.Init()
		if err != nil {
			panic(fmt.Sprintf("Error initializing EQTL arguments %s", err))
		}
		eqtlArguments.Info("Running EQTL",
			"args", fmt.Sprintf("%+v", args),
			"EQTL", fmt.Sprintf("%+v", eqtlArguments),
			"Common", fmt.Sprintf("%+v", eqtlArguments.Common),
			"Expression", fmt.Sprintf("%+v", eqtlArguments.Expression),
			"QTLSpecific", fmt.Sprintf("%+v", eqtlArguments.QTLSpecific),
		)
		run.EQTL(&eqtlArguments)
	},
}
