package main

import (
	"aspire/are/io"
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

const ARE_VERSION = "v0.2.12"

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
		lib_path = fs.Clean(filepath.Dir(exec) + "/stdlib/main.as")
	}

	if !fs.pathExists(lib_path) {
		if len(os.Args) > 1 {
			lib_path = os.Args[1]
		} else {
			io.Printf("ARE %s; could not find standard library.", ARE_VERSION)
		}
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
