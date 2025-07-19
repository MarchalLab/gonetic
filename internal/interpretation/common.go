package interpretation

import (
	"fmt"
	"path/filepath"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

type Interpreter struct {
	*arguments.Common
}

func (interpreter Interpreter) WriteWeightedSubnetwork(
	weightedNetworkfileName string,
	weightedInteractions map[types.InteractionID][]float64,
	geneMapping types.GeneTranslationMap,
) error {
	// write interactions (entries)
	interactionLines := make([]string, 0, len(weightedInteractions)+1)
	interactionLines = append(interactionLines, "#from\tto\tinteractionType\trank")
	for interaction, scores := range weightedInteractions {
		score := 0.0
		for _, s := range scores {
			score += s
		}
		from := interpreter.GetMappedName(interaction.From(), geneMapping)
		to := interpreter.GetMappedName(interaction.To(), geneMapping)
		interactionLines = append(interactionLines, fmt.Sprintf(
			"%s\t%s\t%d\t%f",
			from,
			to,
			interaction.Type(),
			score,
		))
	}
	return interpreter.WriteLinesToNewFile(weightedNetworkfileName, interactionLines)
}

func checkOfInterest(
	genesOfInterest GenesOfInterest,
	geneMap map[types.GeneID]int,
	gene types.GeneID,
	networkSize int,
) {
	if conditionMap, ok := genesOfInterest.Genes[gene]; ok && len(conditionMap) > 0 {
		if val, ok := geneMap[gene]; !ok || networkSize < val {
			geneMap[gene] = networkSize
		}
	}
}

type weightedGene struct {
	gene   types.GeneID
	weight int
}

type WeightedGenes []weightedGene

func (p WeightedGenes) Len() int           { return len(p) }
func (p WeightedGenes) Less(i, j int) bool { return p[i].weight > p[j].weight }
func (p WeightedGenes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (interpreter Interpreter) WriteMutationRanking(
	fileWriter *fileio.FileWriter,
	rankedGenes map[string]WeightedGenes,
	directory string,
	geneMapping types.GeneTranslationMap,
) error {
	for identifier := range rankedGenes {
		err := interpreter.writeMutationRanking(
			fileWriter,
			identifier,
			rankedGenes[identifier],
			directory,
			geneMapping,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (interpreter Interpreter) writeMutationRanking(
	fileWriter *fileio.FileWriter,
	identifier string,
	rankedGenes WeightedGenes,
	directory string,
	geneMapping types.GeneTranslationMap,
) error {
	index := 0
	previousWeight := -1
	linesToWrite := make([]string, 0, len(rankedGenes))
	for _, gene := range rankedGenes {
		name := interpreter.GetMappedName(gene.gene, geneMapping)
		weight := gene.weight
		if previousWeight != weight {
			previousWeight = weight
			index++
		}
		linesToWrite = append(linesToWrite, fmt.Sprintf("%s\t%d", name, index))
	}
	return fileWriter.WriteLinesToNewFile(filepath.Join(
		directory,
		fmt.Sprintf("%sRanking.txt", ToCamelCase(identifier)),
	), linesToWrite)
}
