package profiler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "net/http/pprof"
)

// TestProfiler_DumpProfiles tests the DumpProfiles method
func TestProfiler_DumpProfiles(t *testing.T) {
	profiler := &Profiler{}
	profiler.Init("testresult")
	profiler.DumpProfiles("test_state")

	// Check if the profiling files are created
	profiles := []string{"allocs", "goroutine", "heap"}
	for _, profile := range profiles {
		filename := filepath.Join(profiler.outputFolder, "pprof", fmt.Sprintf(
			"%s_1_test_state_%d.pprof",
			profile,
			time.Now().Unix(),
		))
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Expected profile file %s to be created", filename)
		}
	}
	// Clean up
	os.RemoveAll(profiler.outputFolder)
}

// TestProfiler_Init tests the Init method
func TestProfiler_Init(t *testing.T) {
	profiler := &Profiler{}
	profiler.Init("testresult")

	// Check if the profiling directory is created
	dir := filepath.Join(profiler.outputFolder, "pprof")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Expected profiling directory %s to be created", dir)
	}
	// Clean up
	os.RemoveAll(profiler.outputFolder)
}
