package readers

import (
	"log/slog"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func ReadInputDataHeadersMutationFile(logger *slog.Logger, mutationFileName string) FileData {
	headers := []string{"gene name", "condition", "freq increase", "functional score", "synonymous"}
	required := []string{"gene name", "condition"}
	return ReadSeparatedFile(logger, mutationFileName, "mutation", headers, required)
}

func readMappingFile(logger *slog.Logger, mappingFileName string) FileData {
	headers := []string{"from", "to"}
	required := []string{"from", "to"}
	return ReadSeparatedFile(logger, mappingFileName, "mapping", headers, required)
}

func ConvertMappingFile(logger *slog.Logger, mappingFileName string) types.GeneTranslationMap {
	data := readMappingFile(logger, mappingFileName)
	mappingFileDataNoHeaders := make(types.GeneTranslationMap)
	for _, entry := range data.Entries {
		from := types.GeneName(entry[data.Headers["from"]])
		to := types.GeneName(entry[data.Headers["to"]])
		mappingFileDataNoHeaders[from] = to
	}
	return mappingFileDataNoHeaders
}

func ReadExpressionFile(logger *slog.Logger, fileName, tag string) FileData {
	headers := []string{"gene name", "condition", "p value", tag}
	required := []string{"gene name", "condition"}
	if tag != "none" {
		required = append(required, tag)
	}
	return ReadSeparatedFile(logger, fileName, "expression", headers, required)
}

func ReadDifferentialExpressionFile(logger *slog.Logger, fileName string) FileData {
	headers := []string{"gene name", "condition"}
	required := []string{"gene name", "condition"}
	return ReadSeparatedFile(logger, fileName, "differential expression", headers, required)
}
