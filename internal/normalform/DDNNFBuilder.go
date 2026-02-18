package normalform

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

type DDNNFBuilder struct {
	*slog.Logger
	reverseTranslationTable map[NodeValue]string
	interactionWeights      map[string]float64
}

func NewDDNNFBuilder(
	logger *slog.Logger,
	location string,
	interactionWeights map[string]float64,
) DDNNFBuilder {

	return DDNNFBuilder{
		Logger:                  logger,
		reverseTranslationTable: loadTranslationTable(logger, location),
		interactionWeights:      interactionWeights,
	}
}

func loadTranslationTable(logger *slog.Logger, location string) map[NodeValue]string {
	table := make(map[NodeValue]string)

	file, err := os.Open(filepath.Join(location, "translation_table"))
	if err != nil {
		logger.Error("error constructing translation table", "err", err)
		return table
	}
	defer file.Close()

	counter := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		table[NodeValue(fmt.Sprintf("%d", counter))] = line
		counter++
	}

	return table
}
func (b DDNNFBuilder) BuildLiteralNode(id int, raw string) nnfNode {

	isNegated := false
	if strings.HasPrefix(raw, "-") {
		isNegated = true
		raw = raw[1:]
	}

	// resolve translation table
	name, ok := b.reverseTranslationTable[NodeValue(raw)]
	if !ok {
		name = raw
	}

	// interaction or auxiliary?
	if !types.IsInteractionStringFormat(name) {
		return b.buildAuxNode(id, name, isNegated)
	}

	// interaction
	probability, ok := b.interactionWeights[name]
	if !ok {
		b.Error("Interaction not found in interactionWeights", "interaction", name)
	}

	if isNegated {
		return newNegativeLeafNNFNode(id, name, probability)
	}

	return newPositiveLeafNNFNode(id, name, probability)
}

func (b DDNNFBuilder) buildAuxNode(id int, name string, negated bool) nnfNode {

	node := newAuxLeafNNFNode(id, name, negated)

	// classify special aux types
	if types.IsPathStringFormat(name) && !negated {
		node.leafType = core
	} else if !strings.Contains(name, "aux_") {
		node.leafType = erroneous
	}

	return node
}

func (b DDNNFBuilder) ParseChildList(arr []string) []int {
	children := make([]int, 0, len(arr))

	for _, child := range arr {
		idx, err := strconv.ParseInt(child, 10, 64)
		if err != nil {
			b.Error("error parsing child id", "err", err)
			continue
		}
		children = append(children, int(idx))
	}

	return children
}
