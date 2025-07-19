package types

import (
	"fmt"
	"strconv"
	"strings"
)

func IsInteractionStringFormat(line string) bool {
	if strings.Count(line, ";") != 2 {
		return false
	}
	for _, char := range line {
		if char == ';' {
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func IsPathStringFormat(line string) bool {
	return strings.Count(line, ";") == 1
}

func ParseInteractionID(str string) InteractionID {
	parts := strings.Split(str, ";")
	from := GeneID(parseAbsoluteValue(parts[0]))
	to := GeneID(parseAbsoluteValue(parts[1]))
	typ := InteractionTypeID(parseAbsoluteValue(parts[2]))
	return FromToTypeToID(from, to, typ)
}

func parseAbsoluteValue(str string) int {
	if str[0] == '-' {
		str = str[1:]
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		panic(fmt.Sprintf("could not parse %s as int", str))
	}
	return val
}
