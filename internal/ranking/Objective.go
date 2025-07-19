package ranking

type Objective[S Solution] interface {
	// Compute computes the score on this objective for this solution
	Compute(solution S) float64
	// Compare should return
	// -1 if score1 is better than score2,
	// 0 if they are equal, and
	// 1 if score1 is worse than score2
	Compare(score1, score2 float64) int
	// BottomScore returns the worst possible score for this objective
	BottomScore() float64
	// TopScore returns the best possible score for this objective
	TopScore() float64
}
