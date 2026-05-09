package main

import (
	"os"
	"path/filepath"
	"runtime/pprof"
)

type EnvVars struct {
	debug bool
	sep   string
}

var env_vars = EnvVars{
	debug: false,
	sep:   "  ",
}


func main() {
	// optional profiling controlled via ARE_PROFILE=1
	if os.Getenv("ARE_PROFILE") == "1" {
		cpuF, err := os.Create("cpu.prof")
		if err == nil {
			pprof.StartCPUProfile(cpuF)
			defer func() {
				pprof.StopCPUProfile()
				cpuF.Close()
			}()
		}
		defer func() {
			memF, err := os.Create("mem.prof")
			if err == nil {
				pprof.WriteHeapProfile(memF)
				memF.Close()
			}
		}()
	}

	// allow overriding the script to run via ARE_SCRIPT env var
	envPath := os.Getenv("ARE_LIB")
	var lib_path string
	if envPath != "" {
		lib_path = fs.Abs(envPath)
	} else {
		exec, err := os.Executable()
		if err != nil {
			throw(err)
		}
		exec, err = filepath.EvalSymlinks(exec)
		if err != nil {
			throw(err)
		}
		lib_path = fs.Abs(filepath.Dir(exec) + "/stdlib/main.as")
	}
	parser := newParser(lib_path)
	// parse-only mode for debugging (set ARE_PARSE_ONLY=1)
	if os.Getenv("ARE_PARSE_ONLY") == "1" {
		parser.Parse(true)
		return
	}
	interp := newInterpreter(parser)
	interp.Interpret()
}
