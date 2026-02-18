package normalform

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type C2DReader struct {
	*slog.Logger
}

func NewC2DReader(logger *slog.Logger) C2DReader {
	return C2DReader{Logger: logger}
}

func (reader C2DReader) ReadNNF(
	fileLocation,
	fileName string,
	interactionWeights map[string]float64,
) NNF {

	builder := NewDDNNFBuilder(reader.Logger, fileLocation, interactionWeights)

	nodes := make([]nnfNode, 0)

	file, err := os.Open(filepath.Join(fileLocation, fileName))
	if err != nil {
		reader.Error("error in C2DReader.ReadNNF", "err", err)
		return newNNF(reader.Logger, nodes)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// skip header line: "nnf ..."
	if !scanner.Scan() {
		reader.Error("empty NNF file", "file", fileName)
		return newNNF(reader.Logger, nodes)
	}

	id := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch line[0] {

		case 'L':
			// format: L <literal>
			raw := strings.TrimSpace(line[2:])
			node := builder.BuildLiteralNode(id, raw)
			nodes = append(nodes, node)

		case 'O':
			// format: O <id> <childCount> <children...>
			parts := strings.Split(line, " ")
			children := builder.ParseChildList(parts[3:])
			nodes = append(nodes, newORNNFNode(id, children))

		case 'A':
			// format: A <childCount> <children...>
			parts := strings.Split(line, " ")
			children := builder.ParseChildList(parts[2:])
			nodes = append(nodes, newANDNNFNode(id, children))

		default:
			reader.Error("unknown C2D node type", "line", line)
		}

		id++
	}

	return newNNF(reader.Logger, nodes)
}
