package interpretation

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/arguments"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// d3jsHtmlTemplate copies the html template
func d3jsHtmlTemplate(etcPath string) []string {
	// return the template
	return fileio.ReadListFromFile(filepath.Join(etcPath, "highestScoringSubnetwork.html"), false)
}

// WriteD3JS writes the HTML and javascript files
func (interpreter Interpreter) WriteD3JS(
	sortedConditions types.Conditions,
	etcPath string,
	nodeNameConversionMap types.GeneTranslationMap,
	resultsDirectory string,
	genesOfInterest map[string]GenesOfInterest,
	subnetwork types.InteractionIDSet,
) {
	// create the D3JS directory
	d3jsDirectory := filepath.Join(resultsDirectory, "d3js_visualization")
	fileio.CreateEmptyDir(d3jsDirectory)

	// Write the html file
	err := interpreter.WriteLinesToNewFile(filepath.Join(d3jsDirectory, "highestScoringSubnetwork.html"), d3jsHtmlTemplate(etcPath))
	if err != nil {
		return
	}

	// Write the JS visualisation file
	err = interpreter.WriteLinesToNewFile(filepath.Join(d3jsDirectory, "gonetic.js"), fileio.ReadListFromFile(filepath.Join(etcPath, "gonetic.js"), false))
	if err != nil {
		return
	}

	// gather the data
	sortedGenesOfInterest := interpreter.sortGenesOfInterest(genesOfInterest)
	genesOfInterestLines := interpreter.getGenesOfInterestLines(sortedGenesOfInterest)
	nodeLines := interpreter.getNodeLines(
		sortedGenesOfInterest,
		genesOfInterest,
		sortedConditions,
		nodeNameConversionMap,
		subnetwork,
	)
	linkLines := interpreter.getLinkLines(
		nodeNameConversionMap,
		subnetwork,
	)
	conditionLines := interpreter.getConditionLines(sortedConditions)

	// write the JS data file
	err = interpreter.WriteLinesToNewFile(
		filepath.Join(d3jsDirectory, "subnetwork.js"),
		[]string{"graph = {"},
		genesOfInterestLines,
		nodeLines,
		linkLines,
		conditionLines,
		[]string{"}"},
	)
	if err != nil {
		interpreter.Error("error writing d3js data file",
			"err", err,
			"file", "subnetwork.js",
		)
		return
	}

	encoder := encodeNetworkPaths(interpreter.Common, subnetwork)
	pathDataLines := interpreter.getPathDataLines(encoder)
	// write the JS data file
	err = interpreter.WriteLinesToNewFile(
		filepath.Join(d3jsDirectory, "paths.js"),
		pathDataLines,
	)
	if err != nil {
		interpreter.Error("error writing d3js paths file",
			"err", err,
			"file", "subnetwork.js",
		)
		return
	}
}

// getPathDataLines gathers the path data
func (interpreter Interpreter) getPathDataLines(
	encoder *pathDataEncoder,
) []string {
	lines := make([]string, 0)
	lines = append(lines, "paths = {")
	for pathType := range encoder.Paths {
		lines = append(lines, fmt.Sprintf("%s: [", pathType))
		for cnfHeader, paths := range encoder.Paths[pathType] {
			for _, path := range paths {
				split := strings.Split(path.TxtString(interpreter.GeneIDMap, cnfHeader.Gene, cnfHeader.ConditionName), "\t")
				if len(split) < 8 {
					// duplicate first sample
					split = append([]string{split[0]}, split...)
				}
				// trim the score to 5 digits
				split[2] = fmt.Sprintf("%.5f", path.Probability)
				lines = append(lines, fmt.Sprintf("\"%s\",",
					// only keep the first 4 entries: from sample, to sample, score, edges
					strings.Join(split[:4], "\t"),
				))
			}
		}
		lines = append(lines, "],")
	}
	lines = append(lines, "}")
	return lines
}

// sortGenesOfInterest sorts the genes of interest
func (interpreter Interpreter) sortGenesOfInterest(genesOfInterest map[string]GenesOfInterest) []string {
	sortedGenesOfInterest := make([]string, 0, len(genesOfInterest))
	for identifier := range genesOfInterest {
		sortedGenesOfInterest = append(sortedGenesOfInterest, identifier)
	}
	sort.Strings(sortedGenesOfInterest)
	return sortedGenesOfInterest
}

// getGenesOfInterestLines gathers the genes of interest
func (interpreter Interpreter) getGenesOfInterestLines(sortedGenesOfInterest []string) []string {
	// get types of genes of interest
	genesOfInterestLines := make([]string, 0, len(sortedGenesOfInterest)+2)
	genesOfInterestLines = append(genesOfInterestLines, "genesOfInterest: [")
	for _, identifier := range sortedGenesOfInterest {
		genesOfInterestLines = append(genesOfInterestLines, fmt.Sprintf("\"%s\",", identifier))
	}
	genesOfInterestLines = append(genesOfInterestLines, "],")
	return genesOfInterestLines
}

// getNodeLines gathers the nodes of the network
func (interpreter Interpreter) getNodeLines(
	sortedGenesOfInterest []string,
	genesOfInterest map[string]GenesOfInterest,
	sortedConditions types.Conditions,
	nodeNameConversionMap types.GeneTranslationMap,
	edgesInCondition types.InteractionIDSet,
) []string {
	// Get header of database file
	nodeLines := make([]string, 0)
	nodeLines = append(nodeLines, "nodes: [")
	genes := make(types.GeneSet)
	for interaction := range edgesInCondition {
		genes[interaction.From()] = struct{}{}
		genes[interaction.To()] = struct{}{}
	}
	// For every node in the subnetwork, get the appropriate information (whichever is available).
	for gene := range genes {
		// When mutation/DE data is available, fill the "samples" field with {id, []bool} representing this data.
		samplesString := "["
		for _, identifier := range sortedGenesOfInterest {
			genesOfInterestMap := genesOfInterest[identifier]
			samplesList := make([]bool, 0, len(sortedConditions))
			if _, ok := genesOfInterestMap.Genes[gene]; ok {
				for _, condition := range sortedConditions {
					_, mutated := genesOfInterestMap.Genes[gene][condition]
					samplesList = append(samplesList, mutated)
				}
			}
			samplesString += fmt.Sprintf("\"%s\",", BoolsToBase64(samplesList))
		}
		samplesString += "]"
		nodeLines = append(nodeLines, fmt.Sprintf(
			`{id:"%s",samples: %s},`,
			interpreter.GetMappedName(gene, nodeNameConversionMap),
			samplesString,
		))
	}
	nodeLines = append(nodeLines, "],")
	return nodeLines
}

// getLinkLines gathers the edges (links) of the network
func (interpreter Interpreter) getLinkLines(
	nodeNameConversionMap types.GeneTranslationMap,
	edgesInCondition types.InteractionIDSet,
) []string {
	linkLines := make([]string, 0)
	linkLines = append(linkLines, "links: [")
	// For every edge, get the data types needed
	for interaction := range edgesInCondition {
		edgeContent := fmt.Sprintf(`{source:"%s", target:"%s", type:"%s"},`,
			interpreter.GetMappedName(interaction.From(), nodeNameConversionMap),
			interpreter.GetMappedName(interaction.To(), nodeNameConversionMap),
			arguments.GlobalInteractionStore.InteractionType(interaction),
		)
		linkLines = append(linkLines, edgeContent)
	}
	linkLines = append(linkLines, "],")
	return linkLines
}

// getConditionLines gathers (sorted, having a fixed order is important) conditions
func (interpreter Interpreter) getConditionLines(sortedConditions types.Conditions) []string {
	conditionLines := make([]string, 0)
	conditionLines = append(conditionLines, "conditions: [")
	for _, condition := range sortedConditions {
		conditionLines = append(conditionLines, fmt.Sprintf(`"%s",`, string(condition)))
	}
	conditionLines = append(conditionLines, "],")
	return conditionLines
}
