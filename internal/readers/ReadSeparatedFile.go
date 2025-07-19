package readers

import (
	"log"
	"log/slog"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
)

// ReadSeparatedFile reads a file with a header line starting with # and returns the data in a map
// headers is used to find the correct columns in the data
// required is used to check if all required columns are present
func ReadSeparatedFile(logger *slog.Logger, fileName, name string, headers, required []string) FileData {
	if fileName == "" {
		return FileData{}
	}
	// Read in data
	fileData := fileio.ReadListFromFile(fileName, true)

	// find separator
	seps := []string{",", ";", "\t"}
	separator := ","
	for _, sep := range seps {
		if strings.Count(fileData[0], sep) > strings.Count(fileData[0], separator) {
			separator = sep
		}
	}
	dataHeader := strings.Split(fileData[0], separator)

	// Check if header is present
	if len(dataHeader[0]) == 0 || dataHeader[0][0] != '#' {
		logger.Error("Header line starting with # is not present.", "fileName", name)
		log.Panic("unrecoverable error")
	}

	// Initialize map
	headersMap := make(map[string]int)

	// Fill map
	for _, header := range headers {
		for index, dataHeader := range dataHeader {
			if strings.Contains(strings.ToLower(dataHeader), header) {
				headersMap[header] = index
			}
		}
	}

	// Check for missing values which should always be present
	for _, requiredHeader := range required {
		if _, ok := headersMap[requiredHeader]; !ok {
			logger.Error("Missing required header", "header", requiredHeader, "fileName", name)
			log.Panic("unrecoverable error")
		}
	}

	// Check that the file is "comma" separated
	dataNoHeaders := make([][]string, 0, len(fileData)-1)
	for _, line := range fileData[1:] {
		split := strings.Split(line, separator)
		if len(split) != len(dataHeader) {
			logger.Error("Incorrect number of columns", "fileName", name, "line", line)
			continue
		}
		if line[0] != '#' {
			dataNoHeaders = append(dataNoHeaders, split)
		}
	}
	logger.Info("Finished reading data file", "fileName", name, "entries", len(dataNoHeaders))
	return FileData{name, headersMap, dataNoHeaders}
}
