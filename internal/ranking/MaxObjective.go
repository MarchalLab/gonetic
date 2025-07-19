package ranking

import "math"

type MaxObjective struct {
}

func (obj *MaxObjective) Compare(score1, score2 float64) int {
	if score1 > score2 {
		return -1
	}
	if score1 == score2 {
		return 0
	}
	return 1
}

func (obj *MaxObjective) BottomScore() float64 {
	return math.Inf(-1)
}

func (obj *MaxObjective) TopScore() float64 {
	return math.Inf(+1)
}
