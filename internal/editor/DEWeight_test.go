package editor

import (
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

type ZscoreArgs struct {
	expressionData     map[types.GeneID]float64
	defaultProbability float64
}

type ZscoreTestCase struct {
	name string
	args ZscoreArgs
	want ExpressionWeight
}

func generateZscoreTestCase(logger *slog.Logger, name string, zscore float64, weight float64) ZscoreTestCase {
	gene1 := types.GeneID(1)
	return ZscoreTestCase{
		name: name,
		args: ZscoreArgs{
			expressionData:     map[types.GeneID]float64{gene1: zscore},
			defaultProbability: 0.0,
		},
		want: ExpressionWeight{
			Logger:             logger,
			scoreMap:           map[types.GeneID]float64{gene1: weight},
			defaultProbability: 0.0,
		},
	}
}
func TestNewZscoreExpressionWeight(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	tests := []struct {
		name string
		args ZscoreArgs
		want ExpressionWeight
	}{
		generateZscoreTestCase(logger, "test zscore -2.5", -2.5, 0.9875806693484477),
		generateZscoreTestCase(logger, "test zscore -1.5", -1.5, 0.8663855974622838),
		generateZscoreTestCase(logger, "test zscore -0.5", -0.5, 0.38292492254802624),
		generateZscoreTestCase(logger, "test zscore 0.0", 0.0, 0.5),
		generateZscoreTestCase(logger, "test zscore 0.5", 0.5, 0.38292492254802624),
		generateZscoreTestCase(logger, "test zscore 1.5", 1.5, 0.8663855974622838),
		generateZscoreTestCase(logger, "test zscore 2.5", 2.5, 0.9875806693484477),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewZscoreExpressionWeight(logger, tt.args.expressionData, tt.args.defaultProbability); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewZscoreExpressionWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
