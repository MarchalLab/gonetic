package wfg

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
)

// wfg is a package that provides functions to interact with the external WFG toolkit

type Point []float64

type Front []Point

type ScoreContainer interface {
	Scores() []float64
}

// ConvertToFront converts a ScoreContainer array to a Front
func ConvertToFront[S ScoreContainer](front []S) Front {
	result := make(Front, 0, len(front))
	points := make(map[string]struct{})
	for _, scores := range front {
		point := ConvertToPoint(scores)
		pointString := fmt.Sprintf("%v", point)
		if _, ok := points[pointString]; !ok {
			result = append(result, point)
			points[pointString] = struct{}{}
		}
	}
	return result
}

// ConvertToPoint converts a ScoreContainer to a Point
func ConvertToPoint(point ScoreContainer) Point {
	return point.Scores()
}

func CreateWfgInput(writer *fileio.FileWriter, name string, front Front) string {
	writer.Debug("create test", "name", name)
	points := make([]string, 0, len(front))
	for _, point := range front {
		parts := make([]string, 0, len(point))
		for _, f := range point {
			// Convert each float64 to a string
			str := strconv.FormatFloat(f, 'f', -1, 64)
			parts = append(parts, str)
		}
		points = append(points, strings.Join(parts, " "))
	}
	sort.Strings(points)
	outFile := fmt.Sprintf("%s.fronts", name)
	_, err := os.Stat(outFile)
	idx := 0
	for err == nil {
		writer.Info("File already exists", "file", outFile)
		idx++
		outFile = fmt.Sprintf("%s-%d.fronts", name, idx)
		_, err = os.Stat(outFile)
	}
	writer.AppendLinesToFile(
		outFile,
		[]string{"#"},
		points,
		[]string{"#"},
	)
	if idx == 0 {
		return name
	} else {
		return fmt.Sprintf("%s-%d", name, idx)
	}
}

func RunWfg(writer *fileio.FileWriter, wfg, frontFile string) error {
	writer.Debug("run test", "frontFile", frontFile)
	// Define the arguments
	arguments := []string{frontFile + ".fronts"}

	// create hypervolume command
	cmd := exec.Command(wfg, arguments...)
	// create output file
	outputFile, err := os.Create(frontFile + ".hv")
	if err != nil {
		return err
	}
	defer outputFile.Close()
	// set output file
	cmd.Stdout = outputFile
	// execute the actual compiling
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	writer.Debug("Successfully computed hypervolume",
		"in", frontFile+".fronts",
	)
	return nil
}

func ParseWfgResult(logger *slog.Logger, frontFile string) float64 {
	file, err := os.Open(frontFile + ".hv")
	if err != nil {
		logger.Error("error opening file", "err", err)
		return 0
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, " = ")
		if len(split) == 2 {
			hv, err := strconv.ParseFloat(split[1], 64)
			if err != nil {
				logger.Error("error parsing hypervolume file", "frontFile", frontFile, "error", err)
				return 0
			}
			return hv
		}
	}
	logger.Error("error parsing hypervolume file", "frontFile", frontFile)
	return 0
}
