package optimization

import (
	"testing"
)

func TestLimitedChange(t *testing.T) {
	tests := []struct {
		old, new, expected int
	}{
		{100, 110, 110}, // Increase within 10%
		{100, 90, 90},   // Decrease within 10%
		{100, 150, 110}, // Increase more than 10%
		{100, 50, 90},   // Decrease more than 10%
		{100, 100, 100}, // No change
		{9, 5, 8},       // Decrease with minimal change of 1
		{9, 15, 10},     // Increase with minimal change of 1
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := limitedChange(tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("limitedChange(%d, %d) = %d; want %d", tt.old, tt.new, result, tt.expected)
			}
		})
	}
}
