package normalform

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// CNF reads the paths out of the CNF file and converts it to an actual CNF
// it writes the CNF's, together with their probabilities in the normal form directory
type CNF struct {
	*fileio.FileWriter
}

func NewCNF(fw *fileio.FileWriter) CNF {
	return CNF{fw}
}

func (cnf CNF) Conversion(cnfPathMap map[types.CNFHeader]types.CompactPathList, nfDir string) error {
	cnf.Info("Converting paths to CNFs.")
	// make normal form directory
	fileio.CreateEmptyDir(nfDir)
	// convert paths to cnfs
	for header, paths := range cnfPathMap {
		// convert paths to cnf
		lines, _, _ := convertPathsToCNF(header.Name(), paths)
		// write cnf
		rootFolder := filepath.Join(nfDir, header.Name())
		fileio.CreateEmptyDir(rootFolder)
		err := cnf.WriteLinesToNewFile(
			filepath.Join(rootFolder, "cnf"),
			lines,
		)
		if err != nil {
			return err
		}
		// write interactions to condition file
		probabilities := interactionsFromCompactPaths(paths)
		interactionLines := make([]string, 0, len(*probabilities))
		for id, p := range *probabilities {
			interactionLines = append(interactionLines, id.StringWithProbability(p))
		}
		cnf.WriteLinesToNewFile(
			filepath.Join(rootFolder, "interactions"),
			interactionLines,
		)
	}
	return nil
}

// turn a single path into a cnf with an auxiliary variable
// each interaction on the path forms a disjunction with the negated auxiliary
// -aux interaction 0
// an additional disjunction is made with the negation of all interactions on the path, and the auxiliary
// aux -i1 -i2 ... 0
// a final disjuction is made with the cnfheader and the negation of the auxiliary
// cnfName -aux 0
func convertPathToCNF(cnfHeaderName string, path *types.CompactPath, auxiliaryVariable string) string {
	disjunctions := make([]string, 0, len(path.InteractionOrder)+2)
	allNegatedInteractions := make([]string, 0, len(path.InteractionOrder))
	// iterate over all interactions, order is irrelevant
	for _, interactionID := range path.InteractionOrder {
		// disjunction of the interaction and the negation of the auxiliary
		interactionDisjunction := fmt.Sprintf("-%s %s 0", auxiliaryVariable, interactionID.StringMinimal())
		disjunctions = append(disjunctions, interactionDisjunction)
		allNegatedInteractions = append(allNegatedInteractions, fmt.Sprintf("-%s", interactionID.StringMinimal()))
	}
	// disjunction of all interactions with negated aux
	allInteractionsDisjunction := fmt.Sprintf("%s %s 0", auxiliaryVariable, strings.Join(allNegatedInteractions, " "))
	disjunctions = append(disjunctions, allInteractionsDisjunction)
	// disjunction of header and negated aux
	headerDisjunction := fmt.Sprintf("%s -%s 0", cnfHeaderName, auxiliaryVariable)
	disjunctions = append(disjunctions, headerDisjunction)
	// separate all disjunctions with a new line for the sake of readability
	return strings.Join(disjunctions, "\n")
}

// convert every path to a CNF with `convertPathToCNF`
// add an additional disjunction with the negated cnfheader and all the auxiliaries
func convertPathsToCNF(cnfHeaderName string, paths types.CompactPathList) ([]string, map[string]float64, map[string]float64) {
	cnfs := make([]string, 0, len(paths)+1)
	auxiliaries := make([]string, 0, len(paths))
	toScores := make(map[string]float64, len(paths))
	fromScores := make(map[string]float64, len(paths))
	for idx, path := range paths {
		auxiliaryVariable := fmt.Sprintf("aux_%d", idx)
		toScores[auxiliaryVariable] = path.ToScore
		fromScores[auxiliaryVariable] = path.FromScore
		cnfs = append(cnfs, convertPathToCNF(cnfHeaderName, path, auxiliaryVariable))
		auxiliaries = append(auxiliaries, auxiliaryVariable)
	}
	// disjunction of negated header and all aux
	negatedHeaderDisjunction := fmt.Sprintf("-%s %s 0", cnfHeaderName, strings.Join(auxiliaries, " "))
	cnfs = append(cnfs, negatedHeaderDisjunction)
	return cnfs, toScores, fromScores
}

func interactionsFromCompactPaths(paths types.CompactPathList) *types.ProbabilityMap {
	probabilities := types.NewProbabilityMap()
	for _, path := range paths {
		for idx, id := range path.InteractionOrder {
			p := path.ProbabilityOrder[idx]
			probabilities.SetProbability(id, p)
		}
	}
	return probabilities
}

// Compile converts the paths CNF files to compiled CNF files
func (cnf CNF) Compile(cnfPathMap map[types.CNFHeader]types.CompactPathList, nfDir string) error {
	cnf.Info("Compiling paths to compiled CNFs.")
	for header := range cnfPathMap {
		rootFolder := filepath.Join(nfDir, header.Name())
		err := cnf.compile(rootFolder)
		if err != nil {
			return err
		}
	}
	return nil
}

// compile converts the paths CNF file to a compiled CNF file
func (cnf CNF) compile(rootFolder string) error {
	// Load cnf content
	lines := fileio.ReadListFromFile(filepath.Join(rootFolder, "cnf"), false)
	// create the translation table
	atomMap := make(map[string]int)
	counter := 1
	// force the CNF header to have index 1 in the translation table
	_, counter = convertToIntIfAtom(atomMap, counter, strings.Split(lines[len(lines)-1], " ")[0])
	// force the auxiliary variables to have indices 2, 3, 4, ...
	for idx, entry := range strings.Split(lines[len(lines)-1], " ") {
		if idx == 0 || entry == "0" {
			continue
		}
		_, counter = convertToIntIfAtom(atomMap, counter, entry)
	}
	// parse the CNF lines to a suitable form for the compilation
	content := make([]string, 0)
	for _, line := range lines {
		split := strings.Split(line, " ")
		tmp := make([]string, 0)
		for _, entry := range split {
			if len(entry) == 0 {
				continue
			}
			_, err := strconv.ParseInt(entry, 10, 64)
			if err != nil {
				entry, counter = convertToIntIfAtom(atomMap, counter, entry)
			}
			tmp = append(tmp, entry)
		}
		content = append(content, strings.Join(tmp, " "))
	}

	// Write the content to a compiled cnf file.
	compiledCNF := filepath.Join(rootFolder, "compiled.cnf")
	err := cnf.WriteLinesToNewFile(
		compiledCNF,
		[]string{fmt.Sprintf("p cnf %d %d", len(atomMap), len(content))},
		content,
	)
	if err != nil {
		return err
	}

	// Write the translation table
	return cnf.writeTranslationTable(atomMap, rootFolder)
}

func (cnf CNF) writeTranslationTable(atomMap map[string]int, location string) error {
	// write the translation table of the atoms. This reflects the number of time a specific atom from the cnf file.
	translationTable := make([]kvPair, 0, len(atomMap))
	for k, v := range atomMap {
		translationTable = append(translationTable, kvPair{k, v})
	}
	sort.Slice(translationTable, func(i, j int) bool {
		return translationTable[i].value < translationTable[j].value
	})
	translationKeys := make([]string, 0, len(translationTable))
	for _, kv := range translationTable {
		translationKeys = append(translationKeys, kv.key)
	}
	return cnf.WriteLinesToNewFile(filepath.Join(location, "translation_table"), translationKeys)
}

// ParseCNF parses the CNF file content and returns the clauses.
func ParseCNF(filename string) ([][]int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var clauses [][]int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "p cnf") || len(line) == 0 {
			continue
		}
		tokens := strings.Fields(line)
		var clause []int
		for _, token := range tokens[:len(tokens)-1] { // Exclude the trailing 0
			literal, _ := strconv.Atoi(token)
			clause = append(clause, literal)
		}
		clauses = append(clauses, clause)
	}
	return clauses, scanner.Err()
}

// ParseTranslationTable parses the translation table content and returns a map of literals to conditions.
func ParseTranslationTable(filename string) []string {
	return fileio.ReadListFromFile(filename, false)
}

// ReconstructPaths reconstructs the path cnf using the clauses and translation map.
func ReconstructPathCNF(clauses [][]int, translationMap []string) []string {
	var paths []string
	for _, clause := range clauses {
		var path []string
		for _, literal := range clause {
			translation := translationMap[abs(literal)-1]
			if literal < 0 {
				translation = fmt.Sprintf("-%s", translation)
			}
			path = append(path, translation)
		}
		path = append(path, "0") // Add the trailing 0
		paths = append(paths, strings.Join(path, " "))
	}
	return paths
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ReadPathCNFs reads the paths from a file.
func ReadPathCNFs(filename string) []string {
	return fileio.ReadListFromFile(filename, false)
}
