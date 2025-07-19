package types

import (
	"testing"
)

func TestNewPathDirection(t *testing.T) {
	tests := map[string]PathDirection{
		"Upstream":     UpstreamPath,
		"Downstream":   DownstreamPath,
		"Undirected":   UndirectedPath,
		"UpDownstream": UpDownstreamPath,
		"DownUpstream": DownUpstreamPath,
		"InvalidInput": InvalidPathDirection,
		"":             InvalidPathDirection,
	}

	for input, expected := range tests {
		result := NewPathDirection(input)
		if result != expected {
			t.Errorf("NewPathDirection(%q): expected %v, got %v", input, expected, result)
		}
	}
}

func TestPathDirection_String(t *testing.T) {
	tests := map[PathDirection]string{
		UpstreamPath:         "Upstream",
		DownstreamPath:       "Downstream",
		UndirectedPath:       "Undirected",
		UpDownstreamPath:     "UpDownstream",
		DownUpstreamPath:     "DownUpstream",
		InvalidPathDirection: "Invalid",
	}

	for direction, expected := range tests {
		result := direction.String()
		if result != expected {
			t.Errorf("PathDirection.String(): expected %q, got %q", expected, result)
		}
	}
}

func TestPathDirection_Matches(t *testing.T) {
	directions := []PathDirection{
		UpstreamPath,
		DownstreamPath,
		UndirectedPath,
		UpDownstreamPath,
		DownUpstreamPath,
		InvalidPathDirection,
	}

	for _, d1 := range directions {
		for _, d2 := range directions {
			expected := expectedMatch(d1, d2)
			result := d1.Matches(d2)
			if result != expected {
				t.Errorf("PathDirection.Matches(%v, %v): expected %v, got %v", d1, d2, expected, result)
			}
		}
	}
}

// Helper function to determine expected match results for PathDirection.Matches
func expectedMatch(d1, d2 PathDirection) bool {
	// Exact match
	if d1 == d2 {
		return true
	}
	// Undirected matches anything
	if d1 == UndirectedPath || d2 == UndirectedPath {
		return true
	}
	// Downstream or Upstream matches UpDownstream or DownUpstream
	if (d1 == DownstreamPath || d1 == UpstreamPath) && (d2 == UpDownstreamPath || d2 == DownUpstreamPath) {
		return true
	}
	// Otherwise, no match
	return false
}
