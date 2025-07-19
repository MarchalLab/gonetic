package ranking

type Solution interface {
	// SetScores sets scores on the Solution object
	// call with argument nil to reset the scores
	SetScores([]float64)
	// Scores returns the scores of the Solution object
	// if scores have not been set, this should return nil
	Scores() []float64
}
