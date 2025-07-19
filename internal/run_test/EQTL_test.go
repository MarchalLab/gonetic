package run_test

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/run"
)

func TestEQTL(t *testing.T) {
	outputFolder := path.Join("testresult", "eqtl")
	commonArgs := createCommonTestArgs(outputFolder)
	commonArgs.OutputFolder = outputFolder
	commonArgs.SkipOptimization = true
	commonArgs.SkipInterpreter = true
	expressionArgs := arguments.Expression{
		Common: commonArgs,

		DownUpstream:               true,
		ExpressionFile:             path.Join("testdata", "input-expression.csv"),
		DifferentialExpressionList: path.Join("testdata", "input-differential-expression.csv"),

		ExpressionWeightingMethod:   "none",
		ExpressionWeightingAddition: "none",
		ExpressionWeightingDefault:  0.0,
	}
	qtlArgs := arguments.QTLSpecific{
		FreqIncrease:     true,
		Correction:       true,
		FuncScore:        true,
		MutationDataFile: path.Join("testdata", "input-qtl.csv"),
	}
	args := arguments.EQTL{
		Expression:     &expressionArgs,
		QTLSpecific:    &qtlArgs,
		WithMutation:   true,
		WithExpression: true,
	}

	// run and log path finding separately
	commonArgs.SkipPathFinding = false
	commonArgs.SkipOptimization = true
	commonArgs.SkipInterpreter = true
	args.LogFile = filepath.Join(outputFolder, "output-path.log")
	args.Init()
	run.EQTL(&args)

	// run and log optimization and interpretation separately
	commonArgs.SkipPathFinding = true
	commonArgs.SkipOptimization = false
	commonArgs.SkipInterpreter = false
	args.LogFile = filepath.Join(outputFolder, "output-opt.log")
	args.Init()
	run.EQTL(&args)

	// now resume the optimization
	args.Resume = true
	args.EarlyTermination = false
	args.LogFile = filepath.Join(outputFolder, "output-res.log")
	args.Init()
	run.EQTL(&args)
}
