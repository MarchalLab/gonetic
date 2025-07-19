package normalform

import (
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// ParsePathCNFs parses the CNF lines and reconstructs the interaction sets.
func ParsePathCNFs(cnfLines []string) (types.CNFHeader, types.CompactPathList) {
	paths := make(types.CompactPathList, 0)
	startGene := ""
	for _, line := range cnfLines {
		literals := strings.Fields(line)
		if len(literals) == 0 {
			continue
		}
		if !strings.HasPrefix(literals[0], "aux_") {
			if !strings.HasPrefix(literals[0], "-") {
				startGene = literals[0]
			}
			continue
		}
		// gather interactions of this path
		interactionOrder := make([]types.InteractionID, 0, len(literals)-2)
		for idx := 1; idx < len(literals)-1; idx++ {
			interactionID := types.ParseInteractionID(literals[idx])
			interactionOrder = append(interactionOrder, interactionID)
		}
		path := types.NewCompactPathWithInteractions(interactionOrder)
		paths = append(paths, path)
	}

	return types.NewCNFHeaderFromString(startGene), paths
}

func TestConversion(t *testing.T) {
	testCases := []string{
		"2",
		"2-2",
		"2-2-common",
		"2-2-same",
		//TODO "4",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			logger := slog.Default()
			fw := &fileio.FileWriter{Logger: logger}
			translationFilename := filepath.Join("testdata", tc, "translation_table")
			cnfFilename := filepath.Join("testdata", tc, tc+".cnf")
			translationTableFilename := filepath.Join("testdata", tc, "translation_table")
			resultDir := filepath.Join("testresult", tc)
			fileio.CreateEmptyDir(resultDir)
			pathsFilename := filepath.Join(resultDir, "paths")

			// Step 1: Read and parse CNF file
			clauses, err := ParseCNF(cnfFilename)
			if err != nil {
				t.Errorf("Error reading CNF file: %v", err)
			}

			// Step 2: Read and parse translation table
			translationMap := ParseTranslationTable(translationTableFilename)
			if err != nil {
				t.Errorf("Error reading translation table: %v", err)
			}

			// Step 3: Reconstruct paths
			paths := ReconstructPathCNF(clauses, translationMap)
			err = fw.WriteLinesToNewFile(pathsFilename, paths)
			if err != nil {
				t.Errorf("Error saving paths: %v", err)
			}

			// compare the paths file
			pathsCNFFilename := filepath.Join("testdata", tc, "cnf")
			newPathsCNFFilename := filepath.Join(resultDir, "paths")
			if !fileio.CompareFiles(pathsCNFFilename, newPathsCNFFilename) {
				t.Errorf("The paths files differ for directory %s", tc)
			}

			// Step 4: Read paths from file
			pathCNFs := ReadPathCNFs(pathsFilename)
			if err != nil {
				t.Errorf("Error reading paths: %v", err)
			}

			cnfPathMap := make(map[types.CNFHeader]types.CompactPathList)
			cnfHeader, cnfPaths := ParsePathCNFs(pathCNFs)
			cnfPathMap[cnfHeader] = cnfPaths

			// Step 6: Convert paths to CNF
			cnf := NewCNF(fw)
			conversionDir := filepath.Join(resultDir, "conversion")
			err = cnf.Conversion(cnfPathMap, conversionDir)
			if err != nil {
				t.Errorf("Error converting paths to CNFs: %v", err)
			}
			err = cnf.Compile(cnfPathMap, conversionDir)
			if err != nil {
				t.Errorf("Error compiling path CNFs to CNFs: %v", err)
			}

			newCnfFilename := filepath.Join(conversionDir, cnfHeader.Name(), "compiled.cnf")
			if !fileio.CompareFiles(cnfFilename, newCnfFilename) {
				x, err := fileio.ReadAndSortFile(cnfFilename)
				t.Errorf("%+v\t%+v", x, err)
				y, err := fileio.ReadAndSortFile(newCnfFilename)
				t.Errorf("%+v\t%+v", y, err)
				t.Errorf("The compiled CNFs differ for directory %s", tc)
			}
			newTranslationFileName := filepath.Join(conversionDir, cnfHeader.Name(), "translation_table")
			if !fileio.CompareFiles(translationFilename, newTranslationFileName) {
				t.Errorf("The translation tables differ for directory %s", tc)
			}
		})
	}
}
