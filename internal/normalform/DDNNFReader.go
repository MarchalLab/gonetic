package normalform

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type DDNNFReader interface {
	ReadNNF(fileLocation, fileName string, interactionWeights map[string]float64) NNF
}

func ReadDDNF(logger *slog.Logger, location, file string, weights map[string]float64) NNF {

	f, _ := os.Open(filepath.Join(location, file))
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	first := strings.TrimSpace(scanner.Text())

	if strings.HasPrefix(first, "nnf") {
		return NewC2DReader(logger).ReadNNF(location, file, weights)
	}

	return NewD4Reader(logger).ReadNNF(location, file, weights)
}
