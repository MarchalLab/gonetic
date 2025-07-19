package readers

import (
	"fmt"
	"os"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

func ReadIndexes(args *arguments.Common) {
	ReadGeneIDMap(args)
	ReadInteractionTypeMap(args)
}

func ReadGeneIDMap(args *arguments.Common) {
	args.GeneIDMap = ReadIDMap[types.GeneID, types.GeneName](args.GeneMapFileToRead())
}

func ReadInteractionTypeMap(args *arguments.Common) {
	interactionTypes := ReadIDMap[types.InteractionTypeID, string](args.InteractionTypeMapFileToRead())
	args.InteractionStore.SetInteractionTypes(interactionTypes)
}

func ReadIDMap[ID ~uint64, Name ~string](filename string) *types.IDMap[ID, Name] {
	gim := types.NewIDMap[ID, Name]()
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
		gim.SetNameWithID(split[1], types.ParseID[ID](split[0]))
	}
	return gim
}
