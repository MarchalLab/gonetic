package graph_test

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/readers"
)

var update = flag.Bool("update", false, "update .golden files")

func TestNetworkDegree(t *testing.T) {
	arguments.GlobalInteractionStore = types.NewInteractionStore()
	// get actual network
	networkFile := filepath.Join("testdata", "network.txt")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	gim := types.NewGeneIDMap()
	nwr := readers.NewInitialNetworkReader(logger, gim)
	network := nwr.NewNetworkFromFile(networkFile, true, true)

	// golden file
	goldenFileName := filepath.Join("testdata", "networkDegree.golden")

	// update golden
	if *update {
		goldenFile, err := os.Create(goldenFileName)
		if err != nil {
			t.Fatal(err)
		}
		defer goldenFile.Close()
		for gene := range network.Genes() {
			_, err := goldenFile.WriteString(fmt.Sprintf("%d\t%d\t%d\n", gene, network.InDegree(gene), network.OutDegree(gene)))
			if err != nil {
				t.Fatalf("err: %s", err)
			}
		}
	}

	// get golden network
	goldenFile, err := os.Open(goldenFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer goldenFile.Close()
	scanner := bufio.NewScanner(goldenFile)
	var goldenInDegree = make(map[types.GeneID]int)
	var goldenOutDegree = make(map[types.GeneID]int)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, "\t")
		in, err := strconv.Atoi(split[1])
		gene := gim.SetName(split[0])
		out, err := strconv.Atoi(split[2])
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		goldenInDegree[gene] = int(in)
		goldenOutDegree[gene] = int(out)
	}

	// compare actual to golden
	for gene := range network.Genes() {
		if network.InDegree(gene) != goldenInDegree[gene] {
			t.Errorf("GeneID %d found in-degree %d expected %d", gene, network.InDegree(gene), goldenInDegree[gene])
		}
		if network.OutDegree(gene) != goldenOutDegree[gene] {
			t.Errorf("GeneID %d found out-degree %d expected %d", gene, network.OutDegree(gene), goldenOutDegree[gene])
		}
	}
}
