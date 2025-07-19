package readers

import (
	"log/slog"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/types"
)

// networkLineParser is a struct that implements the line parsing for network files
type networkLineParser struct {
	*slog.Logger
}

// networkDefinitionPattern is a regex pattern to match network definition lines
var networkDefinitionPattern = regexp.MustCompile(`% (.*) (.*)`)

// tokenizeLine tokenizes a line into its components
func (nlp networkLineParser) tokenizeLine(line string) []string {
	var vals []string
	vals = networkDefinitionPattern.FindStringSubmatch(line)
	if len(vals) == 3 {
		return vals
	}
	for _, sep := range []string{",", ";", "\t", " "} {
		vals = strings.Split(line, sep)
		if len(vals) >= 2 && len(vals) <= 6 {
			return vals
		}
	}
	return make([]string, 0)
}

// parseEmptyLine parses a line with 0 tokens
func (nlp networkLineParser) parseEmptyLine(line string) {
	// noop
	nlp.Error("failed to parse line", "line", line)
}

// parseScore parses a score string and returns the parsed values
func (nlp networkLineParser) parseScore(scores string) []float64 {
	scores = strings.Trim(scores, "[]")
	parsedScores := make([]float64, 0, len(scores))
	for _, score := range strings.Split(scores, " ") {
		if len(score) == 0 {
			continue
		}
		parsed, err := strconv.ParseFloat(score, 64)
		if err != nil {
			nlp.Error("failed to parse score", "error", err)
			parsed = math.NaN()
		}
		parsedScores = append(parsedScores, parsed)
	}
	return parsedScores
}

// parseProbability parses a probability string and returns the parsed value
func (nlp networkLineParser) parseProbability(probability string) float64 {
	parsed, err := strconv.ParseFloat(probability, 64)
	if err != nil {
		nlp.Error("failed to parse probability, using default 1.0", "error", err)
		return 1.0
	}
	return parsed
}

// parseInteractionID parses an interaction ID string and returns the parsed value
func (nlp networkLineParser) parseInteractionID(id string, counter int64) int64 {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		nlp.Error("failed to parse interaction id", "error", err)
		return counter
	}
	return parsed
}

// parseInteractionTypeID parses an interaction type ID string and returns the parsed value
func (nlp networkLineParser) parseInteractionTypeID(id string) types.InteractionTypeID {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		nlp.Error("failed to parse interaction type id", "error", err)
		return 0
	}
	return types.InteractionTypeID(parsed)
}

// parsedInteractionType is a struct that holds the parsed header information
type parsedInteractionType struct {
	name  string
	isReg bool
}

// parseHeaderLine parses a header line and returns the parsed header information
func (nlp networkLineParser) parseHeaderLine(line string) (parsedInteractionType, bool) {
	tokens := strings.Split(line, " ")
	if len(tokens) < 2 || tokens[0] != "%" || len(tokens[1]) == 0 {
		nlp.Warn("Malformed header, expected format: \"% <type> [(non-)regulatory]\"", "line", line)
		return parsedInteractionType{}, false
	}

	name := tokens[1]
	reg := "non-regulatory"
	if len(tokens) >= 3 {
		reg = tokens[2]
	}

	return parsedInteractionType{
		name:  name,
		isReg: reg == "regulatory",
	}, true
}

// parsedInteraction is a struct that holds the parsed interaction information
type parsedInteraction struct {
	from, to    types.GeneID
	typ         string
	direction   string
	probability float64
	rawTypeID   string
}

// parseDataLine parses a data line and returns the parsed interaction information
func (nlp networkLineParser) parseDataLine(
	line string,
	parseGene func(string) types.GeneID,
	counter int64,
) (parsedInteraction, int64, bool) {
	tokens := nlp.tokenizeLine(line)
	if len(tokens) < 2 {
		nlp.parseEmptyLine(line)
		return parsedInteraction{}, counter, false
	}

	from := parseGene(tokens[0])
	to := parseGene(tokens[1])
	if from == to {
		return parsedInteraction{}, counter, false
	}

	typ := "unknown"
	if len(tokens) > 2 {
		typ = tokens[2]
	}

	direction := "directed"
	if len(tokens) > 3 {
		direction = tokens[3]
	}

	prob := 1.0
	if len(tokens) > 4 {
		prob = nlp.parseProbability(tokens[4])
	}

	if len(tokens) > 5 {
		counter = nlp.parseInteractionID(tokens[5], counter)
	}

	return parsedInteraction{
		from:        from,
		to:          to,
		typ:         typ,
		direction:   direction,
		probability: prob,
		rawTypeID:   typ,
	}, counter + 1, true
}
