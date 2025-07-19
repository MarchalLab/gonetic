package ranking

import "math"

type MinObjective struct {
}

func (obj *MinObjective) Compare(score1, score2 float64) int {
	if score1 < score2 {
		return -1
	}
	if score1 == score2 {
		return 0
	}
	return 1
}

func (obj *MinObjective) BottomScore() float64 {
	return math.Inf(+1)
}

func (obj *MinObjective) TopScore() float64 {
	return math.Inf(-1)
}
