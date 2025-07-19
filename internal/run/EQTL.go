package run

import (
	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/pathfinding"
	"github.com/MarchalLab/gonetic/internal/readers"
)

// EQTL is the entry point for the EQTL setting
func EQTL(args *arguments.EQTL) {
	args.PathTypes = make([]string, 0, 3)
	args.PathTypes = append(args.PathTypes, "eqtl")
	if args.WithMutation {
		args.PathTypes = append(args.PathTypes, "mutation")
	}
	if args.WithExpression {
		args.PathTypes = append(args.PathTypes, "expression")
	}

	// Load the relevant data
	readers.ReadIndexes(args.Common)
	mutationFileData := readers.ReadInputDataHeadersMutationFile(args.Logger, args.MutationDataFile)
	expressionFileData := readers.ReadExpressionFile(args.Logger, args.ExpressionFile, args.ExpressionWeightingMethod)
	differentialExpressionFileData := readers.ReadDifferentialExpressionFile(args.Logger, args.DifferentialExpressionList)

	// Run different parts of the program
	if !args.SkipPathFinding {
		pathfinding.Run(
			mutationFileData,
			expressionFileData,
			differentialExpressionFileData,
			args.Expression,
			args.QTLSpecific,
			args.Common,
			args,
		)
	}
	if !args.SkipOptimization {
		optimizer := NewOptimizationRunner(args.Common, false)
		optimizer.Run()
	}
	if !args.SkipInterpreter {
		interpreter := NewInterpretation(args.Common)
		interpreter.Run(
			mutationFileData,
			differentialExpressionFileData,
		)
	}
}
