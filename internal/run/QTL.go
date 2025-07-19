package run

import (
	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/pathfinding"
	"github.com/MarchalLab/gonetic/internal/readers"
)

// QTL is the entry point for the QTL setting
func QTL(args *arguments.QTL) {
	args.PathTypes = []string{"mutation"}

	// Load the relevant data
	readers.ReadIndexes(args.Common)
	mutationFileData := readers.ReadInputDataHeadersMutationFile(args.Logger, args.MutationDataFile)

	// Run different parts of the program
	if !args.SkipPathFinding {
		pathfinding.Run(
			mutationFileData,
			readers.FileData{},
			readers.FileData{},
			&arguments.Expression{},
			args.QTLSpecific,
			args.Common,
			&arguments.EQTL{},
		)
	}
	if !args.SkipOptimization {
		optimizer := NewOptimizationRunner(args.Common, true)
		optimizer.Run()
	}
	if !args.SkipInterpreter {
		// run the interpreter
		interpreter := NewInterpretation(args.Common)
		interpreter.Run(
			mutationFileData,
		)
	}
}
