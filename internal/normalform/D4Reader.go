package normalform

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type D4Reader struct {
	*slog.Logger
}

func NewD4Reader(logger *slog.Logger) D4Reader {
	return D4Reader{Logger: logger}
}

func (reader D4Reader) ReadNNF(
	fileLocation,
	fileName string,
	interactionWeights map[string]float64,
) NNF {
	builder := NewDDNNFBuilder(reader.Logger, fileLocation, interactionWeights)
	file, err := os.Open(filepath.Join(fileLocation, fileName))
	if err != nil {
		reader.Error("error in D4Reader.ReadNNF", "err", err)
		return newNNF(reader.Logger, nil)
	}
	defer file.Close()
	declared, arcs, maxID := reader.collectD4Structure(file)
	nodes := buildDeclaredNodes(declared, maxID)
	nodes = expandD4Arcs(builder, nodes, arcs, maxID)
	return newNNF(reader.Logger, nodes)
}

type d4Arc struct {
	parent   int
	child    int
	literals []string
}

func (reader D4Reader) collectD4Structure(file *os.File) (
	map[int]string,
	[]d4Arc,
	int,
) {
	scanner := bufio.NewScanner(file)
	declared := make(map[int]string)
	arcs := make([]d4Arc, 0)
	maxID := -1
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		switch parts[0] {
		case "o", "a", "t", "f":
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				reader.Error("invalid node id", "value", parts[1], "line", line)
				continue
			}
			declared[id] = parts[0]
			maxID = max(maxID, id)
		default:
			parent, err := strconv.Atoi(parts[0])
			if err != nil {
				reader.Error("invalid parent id", "value", parts[0], "line", line)
				continue
			}
			child, err := strconv.Atoi(parts[1])
			if err != nil {
				reader.Error("invalid child id", "value", parts[1], "line", line)
				continue
			}
			arcs = append(arcs, d4Arc{
				parent:   parent,
				child:    child,
				literals: parts[2 : len(parts)-1],
			})
			maxID = max(maxID, parent)
			maxID = max(maxID, child)
		}
	}

	return declared, arcs, maxID
}

func buildDeclaredNodes(
	declared map[int]string,
	maxID int,
) []nnfNode {
	nodes := make([]nnfNode, maxID+1)
	for id, typ := range declared {
		switch typ {
		case "o":
			nodes[id] = newORNNFNode(id, nil)
		case "a":
			nodes[id] = newANDNNFNode(id, nil)
		case "t":
			nodes[id] = newTrueLeafNNFNode(id)
		case "f":
			nodes[id] = newFalseLeafNNFNode(id)
		}
	}
	return nodes
}

func expandD4Arcs(
	builder DDNNFBuilder,
	nodes []nnfNode,
	arcs []d4Arc,
	maxID int,
) []nnfNode {
	nextID := maxID + 1
	literalNodeMap := make(map[string]int)
	for _, a := range arcs {
		andChildren := []int{a.child}
		for _, lit := range a.literals {
			if existing, ok := literalNodeMap[lit]; ok {
				andChildren = append(andChildren, existing)
				continue
			}
			node := builder.BuildLiteralNode(nextID, lit)
			nodes = append(nodes, node)
			literalNodeMap[lit] = nextID
			andChildren = append(andChildren, nextID)
			nextID++
		}
		andNode := newANDNNFNode(nextID, andChildren)
		nodes = append(nodes, andNode)
		nodes[a.parent].children = append(nodes[a.parent].children, nextID)
		nextID++
	}
	return nodes
}
