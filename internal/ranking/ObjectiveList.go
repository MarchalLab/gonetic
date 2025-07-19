package ranking

type ObjectiveList[S Solution] []Objective[S]

func (objectives ObjectiveList[S]) SetScores(solution S) {
	if len(solution.Scores()) == len(objectives) {
		// scores are already set, avoid recomputing
		return
	}
	scores := make([]float64, len(objectives))
	for i, obj := range objectives {
		scores[i] = obj.Compute(solution)
	}
	solution.SetScores(scores)
}

func (objectives ObjectiveList[S]) Dominates(solution1, solution2 S) bool {
	scores1 := solution1.Scores()
	scores2 := solution2.Scores()
	if scores1 == nil || scores2 == nil {
		panic("Scores not set for solution")
	}
	dominates := false
	for idx, obj := range objectives {
		switch obj.Compare(scores1[idx], scores2[idx]) {
		case -1:
			dominates = true
		case 0:
			continue
		case 1:
			return false
		}
	}
	return dominates
}
