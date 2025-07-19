package editor

import (
	"log/slog"
	"math"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

type ExpressionWeight struct {
	*slog.Logger
	scoreMap           map[types.GeneID]float64
	defaultProbability float64
}

func (w ExpressionWeight) PrintScoreMap() {
	for gene, score := range w.scoreMap {
		w.Info("gene-score %s %f", "gene", gene, "score", score)
	}
}

func (w ExpressionWeight) Score(gene types.GeneID) float64 {
	if val, ok := w.scoreMap[gene]; ok {
		return val
	}
	return w.defaultProbability
}

func NewLfcExpressionWeight(logger *slog.Logger, expressionData map[types.GeneID]float64, defaultProbability float64) ExpressionWeight {
	genes := make([]float64, 0, len(expressionData))
	for _, lfc := range expressionData {
		genes = append(genes, lfc)
	}
	mean, std := MeanStdDev(genes)
	logger.Info("LFC expression weight distribution", "mean", mean, "std", std)
	normal := Normal{Mu: mean, Sigma: std}
	scoreMap := make(map[types.GeneID]float64)
	for gene, lfc := range expressionData {
		scoreMap[gene] = 2 * math.Abs(normal.CDF(lfc)-0.5)
	}
	return ExpressionWeight{
		logger,
		scoreMap,
		defaultProbability,
	}
}

func NewZscoreExpressionWeight(logger *slog.Logger, expressionData map[types.GeneID]float64, defaultProbability float64) ExpressionWeight {
	scoreMap := make(map[types.GeneID]float64)
	for gene, z := range expressionData {
		if z == 0 {
			scoreMap[gene] = 0.5
			continue
		}
		cumulProb := 0.5 * (1.0 + math.Erf(z/math.Sqrt(2.0)))
		if z < 0 {
			cumulProb = 1 - cumulProb
		}
		scoreMap[gene] = (cumulProb - 0.5) * 2
	}
	return ExpressionWeight{
		logger,
		scoreMap,
		defaultProbability,
	}
}
