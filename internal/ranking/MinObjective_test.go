package ranking

import (
	"testing"
)

func TestMinObjective_Compare(t *testing.T) {
	obj := &MinObjective{}

	tests := []struct {
		score1, score2 float64
		expected       int
	}{
		{5.0, 3.0, 1},
		{3.0, 5.0, -1},
		{4.0, 4.0, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := obj.Compare(tt.score1, tt.score2)
			if result != tt.expected {
				t.Errorf("Compare(%f, %f) = %d; want %d", tt.score1, tt.score2, result, tt.expected)
			}
		})
	}
}

func TestMinObjective_BottomScore(t *testing.T) {
	obj := &MinObjective{}
	result := obj.BottomScore()
	if obj.Compare(result, 0) != 1 {
		t.Errorf("BottomScore() = %f should be worse than 0", result)
	}
}

func TestMinObjective_TopScore(t *testing.T) {
	obj := &MinObjective{}
	result := obj.TopScore()
	if obj.Compare(result, 0) != -1 {
		t.Errorf("TopScore() = %f should be better than 0", result)
	}
}
