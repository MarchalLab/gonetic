package run_test

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/run"
)

func TestExpression(t *testing.T) {
	outputFolder := path.Join("testresult", "expression")
	commonArgs := createCommonTestArgs(outputFolder)
	commonArgs.OutputFolder = outputFolder
	args := arguments.Expression{
		Common: commonArgs,

		DownUpstream:               true,
		ExpressionFile:             path.Join("testdata", "input-expression.csv"),
		DifferentialExpressionList: path.Join("testdata", "input-differential-expression.csv"),

		ExpressionWeightingMethod:   "none",
		ExpressionWeightingAddition: "none",
		ExpressionWeightingDefault:  0.0,
	}

	// run and log path finding separately
	commonArgs.SkipPathFinding = false
	commonArgs.SkipOptimization = true
	commonArgs.SkipInterpreter = true
	args.LogFile = filepath.Join(outputFolder, "output-path.log")
	args.Init()
	run.Expression(&args)

	// run and log optimization and interpretation separately
	commonArgs.SkipPathFinding = true
	commonArgs.SkipOptimization = false
	commonArgs.SkipInterpreter = false
	commonArgs.SldCutoff = 0.111
	args.LogFile = filepath.Join(outputFolder, "output-opt.log")
	args.Init()
	run.Expression(&args)
}
