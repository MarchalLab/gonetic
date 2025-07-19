package graph_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/MarchalLab/gonetic/internal/common/arguments"

	"github.com/MarchalLab/gonetic/internal/common/types"
	"github.com/MarchalLab/gonetic/internal/graph"
	"github.com/MarchalLab/gonetic/internal/readers"
)

func TestUpstreamExpander_Expand(t *testing.T) {
	arguments.GlobalInteractionStore = types.NewInteractionStore()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	gim := types.NewGeneIDMap()
	nwr := readers.NewInitialNetworkReader(logger, gim)
	network := nwr.NewNetworkFromFile("testdata/expand-network.csv", false, true)

	expander := graph.NewUpstreamExpander(network)

	tests := []struct {
		name     string
		path     *graph.Path
		expected []*graph.Path
		wantErr  bool
	}{
		{
			name: "Expand from gene PIK3C3",
			path: &graph.Path{EndGene: gim.GetIDFromName("PIK3C3")},
			expected: []*graph.Path{
				{EndGene: gim.GetIDFromName("FRS2"), Direction: types.UpstreamPath},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for incoming := range network.Incoming()[tt.path.EndGene] {
				t.Logf("incoming = %s -> %s", gim.GetNameFromID(incoming.From()), gim.GetNameFromID(incoming.To()))
			}
			got, err := expander.Expand(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expand() error = %v, want error %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("Expand() = %v, expected %v", got, tt.expected)
				return
			}
			for i, path := range got {
				if path.EndGene != tt.expected[i].EndGene || path.Direction != tt.expected[i].Direction {
					t.Errorf("Expand() = %v, expected %v", path, tt.expected)
				}
			}
		})
	}
}
