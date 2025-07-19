package pathfinding

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func Run(
	mutationFileData readers.FileData,
	expressionFileData readers.FileData,
	differentialExpressionFileData readers.FileData,
	expressionArgs *arguments.Expression,
	qtlArgs *arguments.QTLSpecific,
	commonArgs *arguments.Common,
	eqtlArgs *arguments.EQTL,
) {
	commonArgs.DumpProfiles("path-start")
	defer func() {
		commonArgs.DumpProfiles("path-end")
	}()
	for _, pathType := range commonArgs.PathTypes {
		fileio.CreateEmptyDir(commonArgs.PathsDirectory(pathType))
		switch pathType {
		case "eqtl":
			eqtl(
				pathType,
				mutationFileData,
				expressionFileData,
				differentialExpressionFileData,
				eqtlArgs,
			)
			break
		case "expression":
			expression(
				pathType,
				expressionFileData,
				differentialExpressionFileData,
				expressionArgs,
			)
			break
		case "mutation":
			qtl(
				pathType,
				mutationFileData,
				qtlArgs,
				commonArgs,
			)
			break
		}
	}
	commonArgs.WriteGeneMapFile()
	commonArgs.WriteInteractionTypeMapFile()
}

func writeWeights(fw *fileio.FileWriter, outputFileName string, weightsPerGene types.GeneConditionMap[float64]) {
	var lines []string
	for gene, conditionMap := range weightsPerGene {
		for condition, weight := range conditionMap {
			line := fmt.Sprintf("%d;%s;%f", gene, condition, weight)
			lines = append(lines, line)
		}
	}
	err := fw.WriteLinesToNewFile(outputFileName, lines)
	if err != nil {
		fw.Error("error writing weights to file", "err", err)
	}
}

func combinePathFiles(fw *fileio.FileWriter, outputFolder, outputFileName string) {
	var lines []string
	err := filepath.Walk(outputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// skip directories
			return nil
		}
		if filepath.Ext(path) != ".paths" {
			// skip files that are not paths-files
			return nil
		}
		// copy paths
		inputFile, err := os.Open(path)
		defer inputFile.Close()
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(inputFile)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
		}
		return nil
	})
	if err != nil {
		fw.Error("error walking the path", "err", err)
		return
	}
	err = fw.WriteLinesToNewFile(outputFileName, lines)
	if err != nil {
		fw.Error("error writing combined paths to file", "err", err)
	}
}
