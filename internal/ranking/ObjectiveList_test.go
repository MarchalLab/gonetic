package ranking

import (
	"math"
	"testing"
)

// MockSolution is a mock implementation of the Solution interface
type MockSolution struct {
	values []float64
	scores []float64
}

func (m *MockSolution) Scores() []float64 {
	return m.scores
}

func (m *MockSolution) SetScores(scores []float64) {
	m.scores = scores
}

// MockObjective is a mock implementation of the Objective interface
type MockObjective struct {
	MaxObjective
}

func (o *MockObjective) Compute(solution *MockSolution) float64 {
	// Mock computation, just return the sum of absolute values of scores
	sum := 0.0
	for _, value := range solution.values {
		sum += math.Abs(value)
	}
	return sum
}

// Tests for te ObjectiveList type

func TestObjectiveList_SetScores(t *testing.T) {
	objective := &MockObjective{}
	objectiveList := ObjectiveList[*MockSolution]{objective}
	solution := &MockSolution{
		values: []float64{1.0, 2.0},
	}

	objectiveList.SetScores(solution)

	expectedScores := []float64{3.0}
	if len(solution.Scores()) != len(expectedScores) {
		t.Fatalf("expected %d scores (%+v), got %d scores (%+v)",
			len(expectedScores),
			expectedScores,
			len(solution.Scores()),
			solution.Scores(),
		)
	}
	for i, score := range solution.Scores() {
		if score != expectedScores[i] {
			t.Errorf("expected score %f, got %f", expectedScores[i], score)
		}
	}
}

func TestObjectiveList_Dominates(t *testing.T) {
	objectiveList := ObjectiveList[*MockSolution]{&MockObjective{}, &MockObjective{}}
	solutions := []*MockSolution{
		{scores: []float64{2.0, 2.0}},
		{scores: []float64{1.0, 2.0}},
		{scores: []float64{2.0, 1.0}},
		{scores: []float64{1.0, 1.0}},
	}
	for _, solution := range solutions {
		objectiveList.SetScores(solution)
	}
	expectedDomination := [][]bool{
		{false, true, true, true},
		{false, false, false, true},
		{false, false, false, true},
		{false, false, false, false},
	}

	for i, row := range expectedDomination {
		for j, expected := range row {
			if objectiveList.Dominates(solutions[i], solutions[j]) != expected {
				t.Errorf("expected solution %d to dominate solution %d: %t",
					i, j, expected)
			}
		}
	}
}
