package profiler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/MarchalLab/gonetic/internal/common/fileio"
)

type Profiler struct {
	pprofDumps             int
	SkipAllocsProfiling    bool
	SkipGoroutineProfiling bool
	SkipHeapProfiling      bool
	outputFolder           string
}

// DumpProfiles writes the current profiles to files
func (profiler *Profiler) DumpProfiles(state string) {
	// compute state name
	profiler.pprofDumps += 1
	state = fmt.Sprintf("%d_%s_%d", profiler.pprofDumps, state, time.Now().Unix())
	// dump profiles
	if !profiler.SkipAllocsProfiling {
		profiler.DumpProfile("allocs", state)
	}
	if !profiler.SkipGoroutineProfiling {
		profiler.DumpProfile("goroutine", state)
	}
	if !profiler.SkipHeapProfiling {
		profiler.DumpProfile("heap", state)
	}
}

// DumpProfile writes the current profile to a file
func (profiler *Profiler) DumpProfile(profile, state string) {
	filename := filepath.Join(
		profiler.outputFolder,
		"pprof",
		fmt.Sprintf("%s_%s.pprof", profile, state),
	)
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("could not create %s profile file %v: %v", profile, filename, err)
		return
	}
	defer file.Close()

	// Handle heap profile separately
	if profile == "heap" {
		if err := pprof.WriteHeapProfile(file); err != nil {
			log.Printf("could not write heap profile: %v", err)
		}
		return
	}

	// lookup and write the profile
	if p := pprof.Lookup(profile); p != nil {
		if err := p.WriteTo(file, 0); err != nil {
			log.Printf("could not write %s profile: %v", profile, err)
		}
	}
}

func (profiler *Profiler) Init(outputFolder string) {
	profiler.outputFolder = outputFolder
	// create the profiling folder
	if !profiler.SkipAllocsProfiling || !profiler.SkipGoroutineProfiling || !profiler.SkipHeapProfiling {
		fileio.CreateDirKeepContent(filepath.Join(profiler.outputFolder, "pprof"))
	}
}
