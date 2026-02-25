package normalform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
	"github.com/MarchalLab/gonetic/internal/common/fileio"
)

// DDNNFCompiler compiles the CNF file into d-DNNF using external software
type DDNNFCompiler struct {
	*arguments.Common
	compiler     string
	compilerType string
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func NewDDNNFCompiler(args *arguments.Common, etcFolderLocation string) DDNNFCompiler {
	// 1. Explicit compiler override
	if args.DDNNFCompilerPath != "" {
		if args.DDNNFCompilerType == "" {
			args.Error("compiler type required when custom compiler path is set")
			log.Panic("invalid compiler configuration")
		}
		return DDNNFCompiler{
			Common:       args,
			compiler:     args.DDNNFCompilerPath,
			compilerType: args.DDNNFCompilerType,
		}
	}

	// 2. Find first existing candidate
	for _, name := range args.DDNNFCompilerTypes {
		path := filepath.Join(etcFolderLocation, fmt.Sprintf("%s%s", name, args.ExeSuffix()))
		if fileExists(path) {
			return DDNNFCompiler{
				Common:       args,
				compiler:     path,
				compilerType: name,
			}
		}
	}

	args.Error("No d-DNNF compiler found in etc folder",
		"OS", runtime.GOOS,
		"location", etcFolderLocation,
		"known compilers", args.DDNNFCompilerTypes,
	)
	log.Panic("no compiler found")
	return DDNNFCompiler{}
}

// CompileDDNNFs compiles CNFs to d-DNNFs
func (compiler DDNNFCompiler) CompileDDNNFs(nfDir string) error {
	compiler.Info("Compiling CNFs to d-DNNFs.")

	// gather directories to compile
	var dirs []string
	err := filepath.Walk(nfDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if path == nfDir {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "compiled.cnf.nnf")); err == nil {
			// compiled file already exists
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	if err != nil {
		return err
	}

	// compile in parallel
	var wg sync.WaitGroup
	wg.Add(len(dirs))
	for _, dir := range dirs {
		dir := dir
		compiler.Sem <- struct{}{}
		go func() {
			compiler.compileDDNNF(dir)
			wg.Done()
			<-compiler.Sem
		}()
	}
	wg.Wait()
	compiler.LogActiveGoRoutines()
	return nil
}

func ReadInteractions(location string) map[string]float64 {
	filename := filepath.Join(location, "interactions")
	lines := fileio.ReadListFromFile(filename, false)
	interactions := make(map[string]float64)

	for _, line := range lines {
		split := strings.Split(line, ";")
		if len(split) != 4 {
			panic("Parsing error in interaction: " + line)
		}
		probability, _ := strconv.ParseFloat(split[3], 64)
		interactions[split[0]+";"+split[1]+";"+split[2]] = probability
	}

	return interactions
}

// compileDDNNF compiles a CNF to a d-DNNF
func (compiler DDNNFCompiler) compileDDNNF(location string) {
	fileName := filepath.Join(location, "compiled.cnf.nnf")
	if _, err := os.Stat(fileName); err == nil {
		compiler.Warn("file exists already", "fileName", fileName)
	} else {
		err := compiler.compile(location)
		if err != nil {
			compiler.Error("error in DDNNFCompiler.compileDDNNF", "err", err)
		}
	}
}

// key value pair struct
type kvPair struct {
	key   string
	value int
}

func (compiler DDNNFCompiler) getArgs(location string) []string {
	compiledCNF := filepath.Join(location, "compiled.cnf")
	switch compiler.compilerType {
	case "c2d":
		return []string{"-cache_size", "2048", "-dt_method", "4", "-smooth_all", "-in", compiledCNF}
	case "d4":
		return []string{"-dDNNF", compiledCNF, fmt.Sprintf("-out=%s.nnf", compiledCNF)}
	}
	return []string{}
}

// compile compiles the CNF file to a d-DNNF file
func (compiler DDNNFCompiler) compile(location string) error {
	compiler.Debug("Compiling", "location", location)

	// create compilation command
	cmd := exec.Command(compiler.compiler, compiler.getArgs(location)...)
	// create output file
	outputFile, err := os.Create(filepath.Join(location, "compilation_log.txt"))
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
	compiler.Debug("Successfully compiled", "compiled", location)
	return nil
}

func convertToIntIfAtom(atomMap map[string]int, counter int, atomName string) (string, int) {
	negative := atomName[0] == '-'
	if negative {
		atomName = atomName[1:] // shave the '-' off
	}
	if _, ok := atomMap[atomName]; !ok {
		atomMap[atomName] = counter
		counter++
	}
	if negative {
		return "-" + fmt.Sprintf("%d", atomMap[atomName]), counter
	}
	return fmt.Sprintf("%d", atomMap[atomName]), counter
}
