package compare

import (
	"testing"
)

func TestBetween(t *testing.T) {
	tests := []struct {
		name     string
		lower    int
		val      int
		upper    int
		expected int
	}{
		{"Value within bounds", 1, 5, 10, 5},
		{"Value equal to lower bound", 1, 1, 10, 1},
		{"Value equal to upper bound", 1, 10, 10, 10},
		{"Value below lower bound", 1, 0, 10, 1},
		{"Value above upper bound", 1, 15, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Between(tt.lower, tt.val, tt.upper)
			if result != tt.expected {
				t.Errorf("Between(%v, %v, %v) = %v; want %v", tt.lower, tt.val, tt.upper, result, tt.expected)
			}
		})
	}
}
