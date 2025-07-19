package optimization

import (
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func arrayToConditionMap(array []int) map[types.Condition]int {
	conditionMap := make(map[types.Condition]int)
	for i, count := range array {
		conditionMap[types.Condition(rune(i))] = count
	}
	return conditionMap
}

func generateComparisonStrings() (string, string, string) {
	e := "equal"
	g := "greater than"
	l := "lesser than"
	return e, g, l
}

func generateTestCases() []struct {
	name          string
	first, second map[types.Condition]int
	expected      string
} {
	e, g, l := generateComparisonStrings()
	return []struct {
		name          string
		first, second map[types.Condition]int
		expected      string
	}{
		{
			name:     "Empty map e",
			first:    arrayToConditionMap([]int{}),
			second:   arrayToConditionMap([]int{}),
			expected: e,
		},
		{
			name:     "Single element e",
			first:    arrayToConditionMap([]int{5}),
			second:   arrayToConditionMap([]int{5}),
			expected: e,
		},
		{
			name:     "Two elements g",
			first:    arrayToConditionMap([]int{10, 10}),
			second:   arrayToConditionMap([]int{5, 10}),
			expected: g,
		},
		{
			name:     "Two elements l",
			first:    arrayToConditionMap([]int{0, 10}),
			second:   arrayToConditionMap([]int{5, 10}),
			expected: l,
		},
		{
			name:     "More zeros",
			first:    arrayToConditionMap([]int{1, 2, 3, 4, 5}),
			second:   arrayToConditionMap([]int{0, 2, 3, 4, 5}),
			expected: g,
		},
		{
			name:     "Even more zeros",
			first:    arrayToConditionMap([]int{0, 2, 3, 4, 5}),
			second:   arrayToConditionMap([]int{0, 0, 3, 4, 5}),
			expected: g,
		},
		{
			name:     "Uniform vs single peak",
			first:    arrayToConditionMap([]int{10, 10, 10, 10}),
			second:   arrayToConditionMap([]int{0, 0, 0, 40}),
			expected: g,
		},
		{
			name:     "Uniform vs slightly uneven",
			first:    arrayToConditionMap([]int{10, 10, 10}),
			second:   arrayToConditionMap([]int{9, 10, 11}),
			expected: g,
		},
		{
			name:     "One zero vs two zeros",
			first:    arrayToConditionMap([]int{0, 5, 5}),
			second:   arrayToConditionMap([]int{0, 0, 10}),
			expected: g,
		},
		{
			name:     "Zero vs all nonzero",
			first:    arrayToConditionMap([]int{1, 2, 3}),
			second:   arrayToConditionMap([]int{0, 2, 3}),
			expected: g,
		},
		{
			name:     "Single high vs two moderate",
			first:    arrayToConditionMap([]int{0, 20, 0}),
			second:   arrayToConditionMap([]int{10, 10, 0}),
			expected: l,
		},
		{
			name:     "Many zeros vs dense but skewed",
			first:    arrayToConditionMap([]int{0, 0, 0, 50}),
			second:   arrayToConditionMap([]int{5, 10, 15, 20}),
			expected: l,
		},
	}
}

func testComputeSampleBalanceScore(
	t *testing.T,
	balanceScore func(
		sampleObjective,
		map[types.Condition]int,
	) float64,
) {
	e, g, l := generateComparisonStrings()
	for _, tt := range generateTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			obj := sampleObjective{}
			firstResult := balanceScore(obj, tt.first)
			secondResult := balanceScore(obj, tt.second)
			var outcome string
			if firstResult == secondResult {
				outcome = e
			} else if firstResult > secondResult {
				outcome = g
			} else {
				outcome = l
			}
			if outcome != tt.expected {
				switch outcome {
				case e:
					t.Errorf("expected %v %v %v, got %f = %f", tt.first, tt.expected, tt.second, firstResult, secondResult)
				case g:
					t.Errorf("expected %v %v %v, got %f > %f", tt.first, tt.expected, tt.second, firstResult, secondResult)
				case l:
					t.Errorf("expected %v %v %v, got %f < %f", tt.first, tt.expected, tt.second, firstResult, secondResult)
				}
			}
		})
	}
}

func TestComputeEntropy(t *testing.T) {
	testComputeSampleBalanceScore(t, func(obj sampleObjective, samples map[types.Condition]int) float64 {
		return obj.computeEntropyScore(samples)
	})
}

func TestComputeEffectiveSampleScore(t *testing.T) {
	testComputeSampleBalanceScore(t, func(obj sampleObjective, samples map[types.Condition]int) float64 {
		return obj.computeEffectiveSampleScore(samples)
	})
}
