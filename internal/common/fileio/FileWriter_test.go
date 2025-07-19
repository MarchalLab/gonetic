package fileio

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func newFileWriter() *FileWriter {
	return &FileWriter{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func TestAppendLinesToFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		lineSets [][]string
		wantErr  bool
	}{
		{
			name:     "valid",
			filePath: "test_append.txt",
			lineSets: [][]string{{"line1", "line2"}, {"line3"}},
			wantErr:  false,
		},
		{
			name:     "invalid path",
			filePath: "\x00",
			lineSets: [][]string{{"line1"}},
			wantErr:  true,
		},
	}

	dir := filepath.Join("testresult", "AppendLinesToFile")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateEmptyDir(dir)
			filePath := filepath.Join(dir, tt.filePath)

			fw := newFileWriter()
			err := fw.AppendLinesToFile(filePath, tt.lineSets...)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				expectedContent := ""
				for _, lines := range tt.lineSets {
					for _, line := range lines {
						expectedContent += line + "\n"
					}
				}
				expectedContent = expectedContent[:len(expectedContent)-1]
				if string(content) != expectedContent {
					t.Errorf("file content = %s, want %s", string(content), expectedContent)
				}
			}
		})
	}
	os.RemoveAll("testresult")
}

func TestWriteLinesToNewFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		lineSets [][]string
		wantErr  bool
	}{
		{
			name:     "valid",
			filePath: "test_write.txt",
			lineSets: [][]string{{"line1", "line2"}, {"line3"}},
			wantErr:  false,
		},
		{
			name:     "invalid path",
			filePath: "\x00",
			lineSets: [][]string{{"line1"}},
			wantErr:  true,
		},
	}

	dir := filepath.Join("testresult", "WriteLinesToNewFile")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateEmptyDir(dir)
			filePath := filepath.Join(dir, tt.filePath)

			fw := newFileWriter()
			err := fw.WriteLinesToNewFile(filePath, tt.lineSets...)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				expectedContent := ""
				for _, lines := range tt.lineSets {
					for _, line := range lines {
						expectedContent += line + "\n"
					}
				}
				expectedContent = expectedContent[:len(expectedContent)-1]
				if string(content) != expectedContent {
					t.Errorf("file content = %s, want %s", string(content), expectedContent)
				}
			}
		})
	}
	os.RemoveAll("testresult")
}

type mockStringLiner struct {
	lines [][]string
}

func (m mockStringLiner) StringArray() [][]string {
	return m.lines
}
func TestWriteStringLinerToFile(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		mockData        mockStringLiner
		expectedContent string
		wantErr         bool
	}{
		{
			name:     "valid",
			filePath: "test_stringliner.txt",
			mockData: mockStringLiner{
				lines: [][]string{{"line1", "line2"}, {"line3"}},
			},
			expectedContent: "line1\nline2\nline3",
			wantErr:         false,
		},
		{
			name:     "invalid path",
			filePath: "\x00",
			mockData: mockStringLiner{
				lines: [][]string{{"line1"}},
			},
			expectedContent: "",
			wantErr:         true,
		},
	}

	dir := filepath.Join("testresult", "WriteStringLinerToFile")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateEmptyDir(dir)
			filePath := filepath.Join(dir, tt.filePath)

			fw := newFileWriter()
			err := fw.WriteStringLinerToFile("testTag", filePath, tt.mockData)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				if string(content) != tt.expectedContent {
					t.Errorf("file content = %s, want %s", string(content), tt.expectedContent)
				}
			}
		})
	}
	os.RemoveAll("testresult")
}
