package arguments

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestCommon_PathsDirectory tests the PathsDirectory method
func TestCommon_PathsDirectory(t *testing.T) {
	// Create an instance of Common with a test output folder
	common := Common{
		OutputFolder: "/test/output",
	}

	// Expected value
	expected := filepath.Join(
		"/test/output",
		fmt.Sprintf("%s_%d", pathsDirectoryName, common.PathLength),
	)

	// Actual value
	actual := common.PathsDirectory("")

	// Compare
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_PathsFile tests the PathsFile method
func TestCommon_PathsFile(t *testing.T) {
	common := Common{OutputFolder: "/test/output"}
	expected := filepath.Join(
		"/test/output",
		fmt.Sprintf("%s_%d", pathsDirectoryName, common.PathLength),
		pathsFileName,
	)
	actual := common.PathsFile("")
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_WeightsFile tests the WeightsFile method
func TestCommon_WeightsFile(t *testing.T) {
	common := Common{OutputFolder: "/test/output"}
	prefix := "test_"
	expected := filepath.Join(
		"/test/output",
		fmt.Sprintf("%s_%d", pathsDirectoryName, common.PathLength),
		prefix+weightsFileName,
	)
	actual := common.WeightsFile("", prefix)
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_OptimizationDirectory tests the OptimizationDirectory method
func TestCommon_OptimizationDirectory(t *testing.T) {
	common := Common{OutputFolder: "/test/output"}
	expected := filepath.Join("/test/output", optimizationDirectoryName+"_0")
	actual := common.OptimizationDirectory()
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_ResultsDirectory tests the ResultsDirectory method
func TestCommon_ResultsDirectory(t *testing.T) {
	common := Common{OutputFolder: "/test/output"}
	pathType := "test"
	expected := filepath.Join("/test/output", resultsDirectoryName, pathType)
	actual := common.ResultsDirectory(pathType)
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_FillDirectory tests the FillDirectory method
func TestCommon_FillDirectory(t *testing.T) {
	common := Common{OutputFolder: "/test/output"}
	expected := filepath.Join("/test/output", filledDirectoryName)
	actual := common.FillDirectory()
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_WeightedNetworkFile tests the WeightedNetworkFile method
func TestCommon_WeightedNetworkFile(t *testing.T) {
	directory := "/test/specific/directory"
	common := Common{}
	expected := filepath.Join(directory, weightedNetworkFileName)
	actual := common.WeightedNetworkFile(directory)
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_WorkingDirectory tests the normalFormDirectory method
func TestCommon_WorkingDirectory(t *testing.T) {
	common := Common{
		OutputFolder: filepath.Join("", "test", "output"),
		PathLength:   17,
		MaxPaths:     300,
	}
	expected := filepath.Join("", "test", "output", "NF_17_300")
	actual := common.NormalFormDirectory()
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_AutoLogFile tests the AutoLogFile method
func TestCommon_AutoLogFile(t *testing.T) {
	common := Common{OutputFolder: "/test/output"}
	expected := filepath.Join("/test/output", "output.log")
	actual := common.AutoLogFile()
	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TestCommon_Init tests the Init method of the Common struct
func TestCommon_Init(t *testing.T) {
	testCases := []struct {
		name        string
		numCPU      int
		logFile     string
		logFormat   string
		expectedCPU int
	}{
		{"DefaultSettings", 0, "", "", 1},
		{"SetNumCPU", 2, "", "text", 2},
		{"AutoLogFile", runtime.NumCPU(), "auto", "text", runtime.NumCPU()},
		{"CustomLogFile", 0, "pathto/logfile", "json", 1},
		// Add more test cases as needed
	}
	for _, testCase := range testCases {
		initProcs := runtime.GOMAXPROCS(0)
		args := NewCommon()

		// Set the test case values
		args.NumCPU = testCase.numCPU
		args.LogFile = testCase.logFile
		args.LogFormat = testCase.logFormat

		// Run the Init method
		args.Init()

		// Check that NumCPU is set correctly
		procs := runtime.GOMAXPROCS(0)
		if procs != testCase.expectedCPU {
			t.Errorf("Expected NumCPU to be set to %d, got %d", testCase.expectedCPU, procs)
		}

		// Check that Logger and FileWriter are initialized
		if args.Logger == nil {
			t.Errorf("Expected Logger to be initialized")
		}
		if args.FileWriter == nil {
			t.Errorf("Expected FileWriter to be initialized")
		}

		// restore max procs
		runtime.GOMAXPROCS(initProcs)

		// remove log file
		switch testCase.logFile {
		case "":
			break
		case "auto":
			removeLogFile(t, "output.log")
		default:
			removeLogFile(t, testCase.logFile)
		}
	}
}

// TestCommon_Init_UsePrecomputed tests the use of precomputed values in the Init method of the Common struct
func TestCommon_Init_UsePrecomputed(t *testing.T) {
	testCases := []struct {
		name        string
		useIndex    string
		usePaths    string
		useNNFs     string
		expectError bool
	}{
		{"ValidCombination1", "path/to/index", "path/to/paths", "path/to/nnfs", false},
		{"ValidCombination2", "path/to/index", "path/to/paths", "", false},
		{"ValidCombination3", "path/to/index", "", "", false},
		{"ValidCombination4", "", "", "", false},
		{"InvalidCombination1", "", "path/to/paths", "path/to/nnfs", true},
		{"InvalidCombination2", "", "", "path/to/nnfs", true},
		{"InvalidCombination3", "path/to/index", "", "path/to/nnfs", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			args := NewCommon()
			// Set the test case values
			args.UsePaths = testCase.usePaths
			args.UseNNFs = testCase.useNNFs
			args.UseIndex = testCase.useIndex

			// Capture log output
			logOutput := &bytes.Buffer{}
			log.SetOutput(logOutput)
			defer log.SetOutput(os.Stderr)

			err := args.Init()

			if testCase.expectError && err == nil {
				t.Errorf("Expected error, but none occurred")
			}
			if !testCase.expectError && err != nil {
				t.Errorf("Did not expect error, but one occurred")
			}
		})
	}
}

func removeLogFile(t *testing.T, filePath string) {
	// Remove the file
	err := os.Remove(filePath)
	if err != nil {
		t.Fatalf("error removing log file: %v", err)
	}
	// Remove the directory
	dir := filepath.Dir(filePath)
	t.Log(dir)
	if dir != "" && dir != "." && dir != ".." {
		if err := os.Remove(dir); err != nil {
			t.Fatalf("error removing log directory: %v", err)
		}
	}
}
