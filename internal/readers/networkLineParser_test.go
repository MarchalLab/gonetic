package readers

import (
	"log/slog"
	"math"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func TestTokenizeLine(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}

	tests := []struct {
		line     string
		expected []string
	}{
		{"A,B,pp,directed,0.9", []string{"A", "B", "pp", "directed", "0.9"}},
		{"A;B;pp;directed;0.9", []string{"A", "B", "pp", "directed", "0.9"}},
		{"A\tB\tpp\tdirected\t0.9", []string{"A", "B", "pp", "directed", "0.9"}},
		{"% type regulatory", []string{"% type regulatory", "type", "regulatory"}},
		{"garbage", []string{}},
	}

	for _, tt := range tests {
		result := parser.tokenizeLine(tt.line)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenizeLine(%q): expected %d tokens, got %d (%v)", tt.line, len(tt.expected), len(result), result)
		}
	}
}

func TestParseScore(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}

	tests := []struct {
		name     string
		input    string
		expected []float64
	}{
		{"valid 3 scores", "[0.1 0.2 0.3]", []float64{0.1, 0.2, 0.3}},
		{"empty list", "[]", []float64{}},
		{"NaN", "[1.0 NaN 2.0]", []float64{1.0, math.NaN(), 2.0}},
		{"invalid float", "[1.0 three 2.0]", []float64{1.0, math.NaN(), 2.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.parseScore(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d scores, got %d", len(tt.expected), len(got))
			}
			for i := range got {
				if got[i] != tt.expected[i] && !(math.IsNaN(got[i]) && math.IsNaN(tt.expected[i])) {
					t.Errorf("score %d: got %.2f, want %.2f", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseProbability(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}
	tests := []struct {
		name   string
		in     string
		expect float64
	}{
		{"valid float", "0.7", 0.7},
		{"invalid fallback", "invalid", 1.0},
		{"negative", "-0.1", -0.1},
		{"empty", "", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.parseProbability(tt.in)
			if got != tt.expect {
				t.Errorf("parseProbability(%q): got %v, want %v", tt.in, got, tt.expect)
			}
		})
	}
}

func TestParseInteractionID(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}

	tests := []struct {
		name     string
		in       string
		counter  int64
		expected int64
	}{
		{"valid int", "123", 0, 123},
		{"invalid fallback", "abc", 42, 42},
		{"empty fallback", "", 7, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.parseInteractionID(tt.in, tt.counter)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestParseInteractionTypeID(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}

	tests := []struct {
		name     string
		in       string
		expected types.InteractionTypeID
	}{
		{"valid type ID", "3", 3},
		{"invalid fallback", "xyz", 0},
		{"empty fallback", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.parseInteractionTypeID(tt.in)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestParseHeaderLine(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}

	tests := []struct {
		name       string
		line       string
		expectOK   bool
		expectName string
		expectReg  bool
	}{
		{
			name:       "valid non-regulatory header",
			line:       "% pp non-regulatory",
			expectOK:   true,
			expectName: "pp",
			expectReg:  false,
		},
		{
			name:       "valid regulatory header",
			line:       "% inferred regulatory",
			expectOK:   true,
			expectName: "inferred",
			expectReg:  true,
		},
		{
			name:       "valid implicitly non-regulatory header",
			line:       "% inferred",
			expectOK:   true,
			expectName: "inferred",
			expectReg:  false,
		},
		{
			name:     "malformed line, name too short",
			line:     "% ",
			expectOK: false,
		},
		{
			name:     "not a header at all",
			line:     "just,a,data,line",
			expectOK: false,
		},
		{
			name:     "empty line",
			line:     "",
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph, ok := parser.parseHeaderLine(tt.line)
			if ok != tt.expectOK {
				t.Fatalf("expected ok=%v, got %v", tt.expectOK, ok)
			}
			if !ok {
				return
			}
			if ph.name != tt.expectName {
				t.Errorf("expected name=%q, got %q", tt.expectName, ph.name)
			}
			if ph.isReg != tt.expectReg {
				t.Errorf("expected isReg=%v, got %v", tt.expectReg, ph.isReg)
			}
		})
	}
}

func TestParseDataLine(t *testing.T) {
	parser := networkLineParser{Logger: slog.Default()}

	tests := []struct {
		name       string
		line       string
		expectOK   bool
		expectTyp  string
		expectDir  string
		expectProb float64
	}{
		{
			name:       "basic 2-token line",
			line:       "A,B",
			expectOK:   true,
			expectTyp:  "unknown",
			expectDir:  "directed",
			expectProb: 1.0,
		},
		{
			name:       "full valid line",
			line:       "A,B,pp,directed,0.9,42",
			expectOK:   true,
			expectTyp:  "pp",
			expectDir:  "directed",
			expectProb: 0.9,
		},
		{
			name:       "undirected line",
			line:       "X,Y,pp,undirected,0.7",
			expectOK:   true,
			expectTyp:  "pp",
			expectDir:  "undirected",
			expectProb: 0.7,
		},
		{
			name:       "bad probability fallback",
			line:       "X,Y,pp,directed,abc",
			expectOK:   true,
			expectTyp:  "pp",
			expectDir:  "directed",
			expectProb: 1.0, // default
		},
		{
			name:     "self-loop should fail",
			line:     "A,A",
			expectOK: false,
		},
		{
			name:     "empty line",
			line:     "",
			expectOK: false,
		},
		{
			name:     "malformed line",
			line:     "onlyonefield",
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gim := types.NewGeneIDMap()
			intx, _, ok := parser.parseDataLine(tt.line, initialGeneParser{gim}.parseGene, 0)
			if ok != tt.expectOK {
				t.Fatalf("expected ok=%v, got %v", tt.expectOK, ok)
			}
			if !ok {
				return
			}
			if intx.typ != tt.expectTyp {
				t.Errorf("expected reg %q, got %q", tt.expectTyp, intx.typ)
			}
			if intx.direction != tt.expectDir {
				t.Errorf("expected direction %q, got %q", tt.expectDir, intx.direction)
			}
			if intx.probability != tt.expectProb {
				t.Errorf("expected prob %.2f, got %.2f", tt.expectProb, intx.probability)
			}
		})
	}
}
