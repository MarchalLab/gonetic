package readers

import (
	"fmt"
	"os"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

func ReadGeneMap(filename string) *types.GeneIDMap {
	gim := types.NewGeneIDMap()
	/// check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// return empty gene map if file does not exist
		return gim
	}
	// read the index if it exists
	for _, entry := range fileio.ReadListFromFile(filename, true) {
		if len(entry) == 0 {
			continue
		}
		split := strings.Split(entry, "\t")
		if len(split) != 2 {
			panic(fmt.Sprintf("GeneID map file entries should have two columns: %s", entry))
		}
		gim.SetNameWithID(split[1], types.ParseID[types.GeneID](split[0]))
	}
	return gim
}
