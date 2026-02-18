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

// LoadDDNNFs loads d-DNNFs from disk
func LoadDDNNFs(logger *slog.Logger, nfDir string) ([]*NNF, error) {
	logger.Info("Reading d-DNNFs.")
	logger.Info("Loading d-DNNFs into memory.")

	ddnnfs := make([]*NNF, 0)
	err := filepath.Walk(nfDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if path == nfDir {
			return nil
		}
		interactionWeights := ReadInteractions(path)
		ddnnf := ReadDDNF(
			logger,
			path,
			"compiled.cnf.nnf",
			interactionWeights,
		)
		ddnnfs = append(ddnnfs, &ddnnf)
		return nil
	})
	if err != nil {
		return ddnnfs, err
	}
	return ddnnfs, nil
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
