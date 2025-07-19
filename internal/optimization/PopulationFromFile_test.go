package optimization

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/types"

	"github.com/MarchalLab/gonetic/internal/readers"

	"github.com/MarchalLab/gonetic/internal/common/fileio"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
)

func mockNSGA() *NSGAOptimization {
	args := arguments.NewCommon()
	args.OutputFolder = "testdata"
	args.Resume = true
	args.FileWriter = &fileio.FileWriter{
		Logger: slog.Default(),
	}
	args.PathLength = 5
	args.PathTypes = []string{"eqtl", "mutation", "expression"}
	arguments.GlobalInteractionStore = types.NewInteractionStore()
	args.Init()

	readers.ReadIndexes(args)

	_, pathRepositories := ReadPathRepositories(args)

	// Create a mock NSGAOptimization object
	return &NSGAOptimization{
		Common:           args,
		pathRepositories: pathRepositories,
	}
}

func TestPopulationFromFile(t *testing.T) {
	expectedGeneration := 7
	expectedSize := 100
	expectedHV := 3.1489916123

	opt := mockNSGA()

	// read the t=0 file
	opt.populationFromFile("", 0)

	// read the final file and validate the generation
	generation := opt.populationFromFile("", -1)
	if generation != expectedGeneration {
		t.Errorf("Expected generation %d, got %d", expectedGeneration, generation)
	}

	// Validate the results
	if len(opt.Pt) != expectedSize {
		t.Errorf("Expected %d subnetworks, got %d", expectedSize, len(opt.Pt))
	}
	// read and validate the hypervolume
	opt.hyperVolumeFromFile()
	if len(opt.hyperVolumeArr) < expectedGeneration+1 {
		t.Errorf("Insufficient (%d) hypervolumes, expected %d+", len(opt.hyperVolumeArr), expectedGeneration)
	}
	if expectedHV != opt.hyperVolumeArr[generation] {
		t.Errorf("Expected %f, got %f", expectedHV, opt.hyperVolumeArr[generation])
	}

	// write the population to file for comparison
	opt.OutputFolder = "testresult"
	fileio.CreateEmptyDir(filepath.Join(opt.OutputFolder, "MO", "population"))
	opt.populationToFile("result_", generation)
	// TODO: compare the files
	identical, err := filesAreIdentical(filepath.Join("testdata", "MO", "population", "7"), filepath.Join("testresult", "MO", "population", "result_7"))
	if !identical || err != nil {
		t.Errorf("Files are not identical")
	}
}

func filesAreIdentical(path1, path2 string) (bool, error) {
	file1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer file1.Close()

	file2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer file2.Close()

	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)

	for {
		n1, err1 := file1.Read(buf1)
		n2, err2 := file2.Read(buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			break
		}

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}
	}
	return true, nil
}

func TestParseString(t *testing.T) {
	opt := mockNSGA()

	// Call the methods
	populationFile := filepath.Join(opt.OutputFolder, "MO", "population", "7")
	lines := fileio.ReadListFromFile(populationFile, false)
	for _, line := range lines {
		// Parse the line to create a subnetwork
		subnet := newFastSubnetwork(opt)
		subnet.ParseString(line)
		if subnet.String() != line {
			t.Errorf("Expected %s, got %s", line, subnet.String())
		}
	}
}
