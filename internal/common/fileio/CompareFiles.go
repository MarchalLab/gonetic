package fileio

import (
	"bufio"
	"os"
	"reflect"
	"sort"
	"strings"
)

func CompareFiles(file1, file2 string) bool {
	lines1, err := ReadAndSortFile(file1)
	if err != nil {
		return false
	}

	lines2, err := ReadAndSortFile(file2)
	if err != nil {
		return false
	}

	if !reflect.DeepEqual(lines1, lines2) {
		return false
	}

	return true
}

func ReadAndSortFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Strings(lines)
	return lines, nil
}
