package arguments

type Expression struct {
	*Common

	DownUpstream               bool
	ExpressionFile             string
	DifferentialExpressionList string

	// Expression weighting
	ExpressionWeightingMethod   string
	ExpressionWeightingAddition string
	ExpressionWeightingDefault  float64
	PrintGeneScoreMap           bool
}
