package run

import (
	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/pathfinding"
	"github.com/MarchalLab/gonetic/internal/readers"
)

// Expression is the entry point for the expression setting
func Expression(args *arguments.Expression) {
	args.PathTypes = []string{"expression"}
	// Load the relevant data
	readers.ReadIndexes(args.Common)
	expressionFileData := readers.ReadExpressionFile(args.Logger, args.ExpressionFile, args.ExpressionWeightingMethod)
	differentialExpressionFileData := readers.ReadDifferentialExpressionFile(args.Logger, args.DifferentialExpressionList)

	// Run different parts of the program
	if !args.SkipPathFinding {
		pathfinding.Run(
			readers.FileData{},
			expressionFileData,
			differentialExpressionFileData,
			args,
			&arguments.QTLSpecific{},
			args.Common,
			&arguments.EQTL{},
		)
	}
	if !args.SkipOptimization {
		optimizer := NewOptimizationRunner(args.Common, false)
		optimizer.Run()
	}
	if !args.SkipInterpreter {
		interpreter := NewInterpretation(args.Common)
		interpreter.Run(
			differentialExpressionFileData,
		)
	}
}
