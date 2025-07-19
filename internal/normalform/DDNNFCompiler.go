package normalform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/MarchalLab/gonetic/internal/common/arguments"
)

// DDNNFCompiler compiles the CNF file into d-DNNF using external software
type DDNNFCompiler struct {
	*arguments.Common
	compiler string
}

func NewDDNNFCompiler(args *arguments.Common, etcFolderLocation string) DDNNFCompiler {
	// Determine which compiler to use. The compiler is external software and OS dependent.
	var compiler string
	switch runtime.GOOS {
	case "windows":
		compiler = filepath.Join(etcFolderLocation, "c2d_windows.exe")
	case "linux":
		compiler = filepath.Join(etcFolderLocation, "c2d_linux")
	default:
		args.Error("There is no cnf to d-DDNNF compiler available for the OS %s", "OS", runtime.GOOS)
		log.Panic("unrecoverable error")
	}
	return DDNNFCompiler{
		Common:   args,
		compiler: compiler,
	}
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

// LoadDDNNFs loads d-DNNFs from disk
func (compiler DDNNFCompiler) LoadDDNNFs(nfDir string) ([]*NNF, error) {
	compiler.Info("Reading d-DNNFs.")

	ddnnfReader := NewDDNNFReader(compiler.Logger)
	compiler.Info("Loading d-DNNFs into memory.")

	ddnnfs := make([]*NNF, 0)
	err := filepath.Walk(nfDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if path == nfDir {
			return nil
		}
		interactionWeights := ReadInteractions(path)
		ddnnf := ddnnfReader.ReadNNF(
			path,
			"compiled.cnf.nnf",
			interactionWeights,
		)
		ddnnfs = append(ddnnfs, &ddnnf)
		return nil
	})
	if err != nil {
		return ddnnfs, err
	}
	return ddnnfs, nil
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

// compile compiles the CNF file to a d-DNNF file
func (compiler DDNNFCompiler) compile(location string) error {
	compiler.Debug("Compiling", "location", location)

	// Define the compile process
	compiledCNF := filepath.Join(location, "compiled.cnf")
	compilationArgs := []string{"-cache_size", "2048", "-dt_method", "4", "-smooth_all", "-in", compiledCNF}

	// create compilation command
	cmd := exec.Command(compiler.compiler, compilationArgs...)
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
