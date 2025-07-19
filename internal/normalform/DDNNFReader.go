package normalform

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

// DDNNFReader contains the mechanism and functions to read NNF's from the compiled d-DNNF
type DDNNFReader struct {
	*slog.Logger
}

func NewDDNNFReader(logger *slog.Logger) DDNNFReader {
	reader := DDNNFReader{
		Logger: logger,
	}
	return reader
}

func (reader DDNNFReader) translationTable(location string) map[NodeValue]string {
	table := make(map[NodeValue]string)
	file, err := os.Open(filepath.Join(location, "translation_table"))
	if err != nil {
		reader.Error("error constructing translation table", "err", err)
		return table
	}
	defer file.Close()
	counter := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		value := NodeValue(fmt.Sprintf("%d", counter))
		table[value] = line
		counter++
	}
	return table
}

func trimLine(line string) (string, bool) {
	line = line[2:]
	isNegated := line[0] == '-'
	if isNegated {
		line = line[1:]
	}
	return line, isNegated
}

func (reader DDNNFReader) convertZippedLine(reverseTranslationTable map[NodeValue]string, interactionWeights map[string]float64, line string, id int) nnfNode {
	switch line[0] {
	case 'L':
		line, isNegated := trimLine(line)
		name, ok := reverseTranslationTable[NodeValue(line)]
		if !ok {
			name = line
		}
		isAux := !types.IsInteractionStringFormat(name)
		if isAux {
			return newAuxLeafNNFNode(id, name, isNegated)
		}
		probability, ok := interactionWeights[name]
		if !ok {
			reader.Error("Interaction not found in interactionWeights", "interaction", name)
		}
		if isNegated {
			return newNegativeLeafNNFNode(id, name, probability)
		}
		return newPositiveLeafNNFNode(id, name, probability)
	case 'O':
		return newORNNFNode(id, reader.childList(strings.Split(line, " ")[3:]))
	case 'A':
		return newANDNNFNode(id, reader.childList(strings.Split(line, " ")[2:]))
	}
	reader.Error("Line: " + line + " could not be converted to a NNFNode.")
	return nnfNode{}
}

func (reader DDNNFReader) childList(arr []string) []int {
	children := make([]int, 0)
	for _, child := range arr {
		idx, err := strconv.ParseInt(child, 10, 64)
		if err != nil {
			reader.Error("error in DDNNFReader.childList", "err", err)
			continue
		}
		children = append(children, int(idx))
	}
	return children
}

func ReadInteractions(location string) map[string]float64 {
	filename := filepath.Join(location, "interactions")
	lines := fileio.ReadListFromFile(filename, false)
	interactions := make(map[string]float64)
	for _, line := range lines {
		split := strings.Split(line, ";")
		if len(split) != 4 {
			panic("Parsing error in interaction: " + line)
		}
		probability, _ := strconv.ParseFloat(split[3], 64)
		interactions[split[0]+";"+split[1]+";"+split[2]] = probability
	}
	return interactions
}

func (reader DDNNFReader) ReadNNF(fileLocation, fileName string, interactionWeights map[string]float64) NNF {
	reverseTranslationTable := reader.translationTable(fileLocation)
	nodes := make([]nnfNode, 0)
	file, err := os.Open(filepath.Join(fileLocation, fileName))
	if err != nil {
		reader.Error("error in DDNNFReader.ReadNNF", "err", err)
		return newNNF(reader.Logger, nodes)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	counter := 0
	// skip first line
	scanner.Scan()
	// log.Infof("scanning NNF: %s", scanner.Text())
	for scanner.Scan() {
		line := scanner.Text()
		node := reader.convertZippedLine(reverseTranslationTable, interactionWeights, line, counter)
		// further classify auxiliary nodes
		if node.leafType == aux {
			if types.IsPathStringFormat(node.name) && !node.IsNegated() {
				node.leafType = core
			} else if !strings.Contains(node.name, "aux_") {
				node.leafType = erroneous
			}
		}
		nodes = append(nodes, node)
		counter += 1
	}
	return newNNF(reader.Logger, nodes)
}
