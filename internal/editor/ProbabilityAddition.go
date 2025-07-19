package editor

import (
	"log/slog"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

// WeightTarget is an enum for which node contributes to the weighting
type WeightTarget int

const (
	FromOnly WeightTarget = iota
	ToOnly
	BothBayes
)

func (t WeightTarget) String() string {
	switch t {
	case FromOnly:
		return "FromOnly"
	case ToOnly:
		return "ToOnly"
	case BothBayes:
		return "BothBayes"
	default:
		return "Unknown"
	}
}

// targetSelector returns a probability for the interaction based on the selected node(s)
type targetSelector func(types.InteractionID) float64

func fromOnly(prob geneProbability) targetSelector {
	return func(id types.InteractionID) float64 { return prob(id.From()) }
}

func toOnly(prob geneProbability) targetSelector {
	return func(id types.InteractionID) float64 { return prob(id.To()) }
}

func bothBayes(prob geneProbability) targetSelector {
	return func(id types.InteractionID) float64 {
		return 1 - (1-prob(id.From()))*(1-prob(id.To()))
	}
}

// weightingMethod combines original interaction probability with gene-based probability
type weightingMethod func(types.InteractionID, float64, targetSelector) float64

func multWeight(i types.InteractionID, original float64, getProb targetSelector) float64 {
	return original * getProb(i)
}

func bayesWeight(i types.InteractionID, original float64, getProb targetSelector) float64 {
	return 1 - (1-original)*(1-getProb(i))
}

func meanWeight(i types.InteractionID, original float64, getProb targetSelector) float64 {
	return (original + getProb(i)) / 2
}

func identityWeight(_ types.InteractionID, original float64, _ targetSelector) float64 {
	return original
}

func WeightingAddition(
	logger *slog.Logger,
	weightTarget WeightTarget,
	tag, errMsg string,
) (additionType, bool) {
	method := ComposeWeighting(tag, weightTarget, logger, errMsg)
	switch tag {
	case "bayes":
		return method, false // 1 - (1 - original) * (1 - gene)
	case "mean":
		return method, false // original * (1 + gene) / 2
	case "mult":
		return method, false // original * gene
	case "none":
		return method, true // original
	default:
		return method, true
	}
}

func ComposeWeighting(
	tag string,
	target WeightTarget,
	logger *slog.Logger,
	errMsg string,
) func(i types.InteractionID, original float64, geneProbabilityMap geneProbability) float64 {
	var base weightingMethod
	switch tag {
	case "bayes":
		base = bayesWeight
	case "mean":
		base = meanWeight
	case "mult":
		base = multWeight
	case "none":
		base = identityWeight
	default:
		logger.Error("unknown weighting tag", "tag", tag, "err", errMsg)
		base = identityWeight
	}

	return func(i types.InteractionID, original float64, geneProbabilityMap geneProbability) float64 {
		switch target {
		case FromOnly:
			return base(i, original, fromOnly(geneProbabilityMap))
		case ToOnly:
			return base(i, original, toOnly(geneProbabilityMap))
		case BothBayes:
			return base(i, original, bothBayes(geneProbabilityMap))
		default:
			logger.Error("invalid WeightTarget", "target", target, "err", errMsg)
			return original
		}
	}
}
