package wfg_test

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/wfg"
)

type mockScoreContainer struct {
	scores []float64
}

func (m mockScoreContainer) Scores() []float64 {
	return m.scores
}

func TestConvertToFront(t *testing.T) {
	tests := []struct {
		name  string
		input []mockScoreContainer
		want  wfg.Front
	}{
		{
			"Empty input",
			[]mockScoreContainer{},
			wfg.Front{},
		},
		{
			"Unique points",
			[]mockScoreContainer{
				{[]float64{1.0}},
				{[]float64{2.0}},
			},
			wfg.Front{{1.0}, {2.0}},
		},
		{
			"Duplicate points",
			[]mockScoreContainer{
				{[]float64{1.0}},
				{[]float64{1.0}},
			},
			wfg.Front{{1.0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wfg.ConvertToFront(tt.input)
			if !equalFronts(got, tt.want) {
				t.Errorf("ConvertToFront() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToPoint(t *testing.T) {
	tests := []struct {
		name  string
		input mockScoreContainer
		want  wfg.Point
	}{
		{
			"Single score",
			mockScoreContainer{[]float64{
				1.0,
			}},
			wfg.Point{
				1.0,
			},
		},
		{
			"Multiple scores",
			mockScoreContainer{[]float64{
				1.0,
				2.0,
			}},
			wfg.Point{
				1.0,
				2.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wfg.ConvertToPoint(tt.input)
			if !equalPoints(got, tt.want) {
				t.Errorf("ConvertToPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateWfgInput(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	writer := &fileio.FileWriter{Logger: logger}
	dir := filepath.Join("testresult", "create_wfg_input")
	fileio.CreateEmptyDir(dir)

	tests := []struct {
		name  string
		input wfg.Front
	}{
		{
			"Empty front",
			wfg.Front{}},
		{
			"Multiple points",
			wfg.Front{
				{1.0, 2.0},
				{3.0, 4.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wfg.CreateWfgInput(
				writer,
				filepath.Join(dir, tt.name),
				tt.input,
			)
			content, err := os.ReadFile(filepath.Join(
				dir,
				fmt.Sprintf("%s.fronts", tt.name),
			))
			if err != nil {
				t.Errorf("Error reading file: %v", err)
			}
			if !contains(content, "#") {
				t.Errorf("File content does not contain expected substring: %s", content)
			}
		})
	}
}

func copyFile(src, dst string) error {
	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func TestRunWfg(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	writer := &fileio.FileWriter{Logger: logger}
	dir := filepath.Join("testresult", "run_wfg")
	fileio.CreateEmptyDir(dir)
	copyFile(
		filepath.Join("testdata", "valid.fronts"),
		filepath.Join(dir, "valid.fronts"),
	)
	validWFG := filepath.Join("..", "..", "etc", "wfg0")
	validFronts := filepath.Join(dir, "valid")
	tests := []struct {
		name      string
		wfgPath   string
		frontFile string
		wantErr   bool
	}{
		{
			name:      "Valid WFG executable and front file",
			wfgPath:   validWFG,
			frontFile: validFronts,
			wantErr:   false,
		},
		{
			name:      "Invalid WFG executable path",
			wfgPath:   filepath.Join("invalid", "path", "to", "wfg"),
			frontFile: validFronts,
			wantErr:   true,
		},
		{
			name:      "Missing front file",
			wfgPath:   validWFG,
			frontFile: filepath.Join(dir, "missing"),
			wantErr:   true,
		},
		{
			name:      "Invalid file name",
			wfgPath:   validWFG,
			frontFile: filepath.Join(dir, "\x00"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wfg.RunWfg(writer, tt.wfgPath, tt.frontFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunWfg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseWfgResult(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dir := "testdata"

	tests := []struct {
		name     string
		filePath string
		want     float64
	}{
		{
			name:     "Valid hypervolume result file",
			filePath: filepath.Join(dir, "valid"),
			want:     1.2991487529,
		},
		{
			name:     "Invalid data in hypervolume result file",
			filePath: filepath.Join(dir, "garbage"),
			want:     0.0,
		},
		{
			name:     "Empty hypervolume result file",
			filePath: filepath.Join(dir, "empty"),
			want:     0.0,
		},
		{
			name:     "Missing hypervolume result file",
			filePath: filepath.Join(dir, "missing"),
			want:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hv := wfg.ParseWfgResult(logger, tt.filePath)
			if hv != tt.want {
				t.Errorf("ParseWfgResult() = %v, want %v", hv, tt.want)
			}
		})
	}
}

func equalFronts(a, b wfg.Front) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalPoints(a[i], b[i]) {
			return false
		}
	}
	return true
}

func equalPoints(a, b wfg.Point) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(content []byte, substr string) bool {
	return strings.Contains(string(content), substr)
}
