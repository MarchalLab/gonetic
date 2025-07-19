package fileio

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func ReadListFromFile(fileName string, trim bool) []string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error opening file %s: %v", fileName, err)
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	list := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if trim {
			line = strings.TrimSpace(line)
		}
		if len(line) > 0 {
			list = append(list, line)
		}
	}
	return list
}
