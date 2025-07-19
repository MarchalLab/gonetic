package pathfinding

import (
	"log/slog"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func commonArgs(output string) *arguments.Common {
	commonArgs := arguments.NewCommon()
	commonArgs.FileWriter = &fileio.FileWriter{
		Logger: slog.Default(),
	}
	commonArgs.OutputFolder = output
	commonArgs.NetworkFiles = []string{"testdata/network"}
	commonArgs.TopologyWeightingAddition = "none"
	commonArgs.BestPathCount = 25
	commonArgs.PathLength = 5
	return commonArgs
}

func readTestData() (*arguments.EQTL, readers.FileData, readers.FileData) {
	// create the args
	args := &arguments.EQTL{
		Expression: &arguments.Expression{
			Common:                    commonArgs("testresult/eqtl"),
			ExpressionWeightingMethod: "none",
		},
		QTLSpecific: &arguments.QTLSpecific{},
	}
	args.Init()
	// create output folder
	fileio.CreateEmptyDir(args.PathsDirectory("eqtl"))
	// the data files
	mutationData := readers.ReadInputDataHeadersMutationFile(
		args.Logger,
		"testdata/mutations",
	)
	expressionData := readers.ReadExpressionFile(
		args.Logger,
		"testdata/expression",
		"none",
	)
	// return the data
	return args, mutationData, expressionData
}

func TestEQTL(t *testing.T) {
	args, mutationData, expressionData := readTestData()
	eqtl("eqtl", mutationData, expressionData, expressionData, args)
}

func TestExpression(t *testing.T) {
	args := &arguments.Expression{
		Common:                    commonArgs("testresult/expression"),
		ExpressionWeightingMethod: "none",
	}
	args.Init()
	// create output folder
	fileio.CreateEmptyDir(args.PathsDirectory("expression"))
	// the data files
	expressionData := readers.ReadExpressionFile(
		args.Logger,
		"testdata/expression",
		"none",
	)
	expression("expression", expressionData, expressionData, args)
}

func TestQTL(t *testing.T) {
	args := &arguments.QTL{
		Common:      commonArgs("testresult/qtl"),
		QTLSpecific: &arguments.QTLSpecific{},
	}
	args.Init()
	// create output folder
	fileio.CreateEmptyDir(args.PathsDirectory("qtl"))
	// the data files
	mutationData := readers.ReadInputDataHeadersMutationFile(
		args.Logger,
		"testdata/mutations",
	)
	qtl("qtl", mutationData, args.QTLSpecific, args.Common)
}
