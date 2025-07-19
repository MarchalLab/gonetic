package fileio

import (
	"log/slog"
	"os"
)

type FileWriter struct {
	*slog.Logger
}

func (fw *FileWriter) AppendLinesToFile(filePath string, lineSets ...[]string) error {
	return fw.writeLinesToFile(filePath, os.O_APPEND, lineSets...)
}

func (fw *FileWriter) WriteLinesToNewFile(filePath string, lineSets ...[]string) error {
	return fw.writeLinesToFile(filePath, os.O_TRUNC, lineSets...)
}

func (fw *FileWriter) writeLinesToFile(filePath string, flag int, lineSets ...[]string) error {
	// create file
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|flag, 0666)
	if err != nil {
		return err
	}
	// defer close file
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
		if err != nil {
			fw.Error("Failed to close file", "file", filePath, "error", err)
		}
	}()
	// write all lines from all line sets to the file
	for i, lines := range lineSets {
		for j, line := range lines {
			var err error
			if i == len(lineSets)-1 && j == len(lines)-1 {
				// no newline after the last line
				_, err = file.WriteString(line)
			} else {
				// append newline
				_, err = file.WriteString(line + "\n")
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type stringLiner interface {
	StringArray() [][]string
}

func (fw *FileWriter) WriteStringLinerToFile(tag, filename string, data stringLiner) error {
	fw.Info("converting StringLiner to file", "tag", tag, "filename", filename)
	lineSets := data.StringArray()
	err := fw.WriteLinesToNewFile(
		filename,
		lineSets...,
	)
	return err
}
