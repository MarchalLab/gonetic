package fileio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareFiles(t *testing.T) {
	tests := []struct {
		name     string
		content1 []string
		content2 []string
		want     bool
	}{
		{
			name:     "identical files",
			content1: []string{"line1", "line2", "line3"},
			content2: []string{"line1", "line2", "line3"},
			want:     true,
		},
		{
			name:     "different files",
			content1: []string{"line1", "line2", "line3"},
			content2: []string{"line1", "line2", "line4"},
			want:     false,
		},
		{
			name:     "same lines different order",
			content1: []string{"line1", "line2", "line3"},
			content2: []string{"line3", "line2", "line1"},
			want:     true,
		},
		{
			name:     "one file empty",
			content1: []string{},
			content2: []string{"line1", "line2", "line3"},
			want:     false,
		},
		{
			name:     "both files empty",
			content1: []string{},
			content2: []string{},
			want:     true,
		},
	}

	dir := filepath.Join("testresult", "CompareFiles")
	os.MkdirAll(dir, os.ModePerm)
	defer os.RemoveAll(dir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file1 := filepath.Join(dir, "file1.txt")
			file2 := filepath.Join(dir, "file2.txt")

			writeLinesToFile(t, file1, tt.content1)
			writeLinesToFile(t, file2, tt.content2)

			got := CompareFiles(file1, file2)
			if got != tt.want {
				t.Errorf("CompareFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func writeLinesToFile(t *testing.T, filePath string, lines []string) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	for _, line := range lines {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			t.Fatalf("failed to write to file: %v", err)
		}
	}
}
