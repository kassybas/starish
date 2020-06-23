// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The starlark command interprets a Starlark file.
// With no arguments, it starts a read-eval-print loop (REPL).
package main // import "go.starlark.net/cmd/starlark"

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"go.starlark.net/internal/compile"
	"go.starlark.net/repl"
	"go.starlark.net/resolve"
	"go.starlark.net/starish"
	"go.starlark.net/starlark"
)

// flags
var (
	cpuprofile = flag.String("cpuprofile", "", "gather Go CPU profile in this file")
	memprofile = flag.String("memprofile", "", "gather Go memory profile in this file")
	profile    = flag.String("profile", "", "gather Starlark time profile in this file")
	showenv    = flag.Bool("showenv", false, "on success, print final global environment")
	execprog   = flag.String("c", "", "execute program `prog`")
)

func init() {
	flag.BoolVar(&compile.Disassemble, "disassemble", compile.Disassemble, "show disassembly during compilation of each function")
	flag.BoolVar(&resolve.AllowGlobalReassign, "globalreassign", resolve.AllowGlobalReassign, "allow reassignment of globals, and if/for/while statements at top level")
	// starish defaults
	resolve.AllowFloat = true
	resolve.AllowSet = true
	resolve.AllowLambda = true
	resolve.AllowNestedDef = true
	resolve.AllowRecursion = true
}

func main() {
	os.Exit(doMain())
}

func doMain() int {
	log.SetPrefix("starish: ")
	log.SetFlags(0)
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		check(err)
		err = pprof.StartCPUProfile(f)
		check(err)
		defer func() {
			pprof.StopCPUProfile()
			err := f.Close()
			check(err)
		}()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		check(err)
		defer func() {
			runtime.GC()
			err := pprof.Lookup("heap").WriteTo(f, 0)
			check(err)
			err = f.Close()
			check(err)
		}()
	}

	if *profile != "" {
		f, err := os.Create(*profile)
		check(err)
		err = starlark.StartProfile(f)
		check(err)
		defer func() {
			err := starlark.StopProfile()
			check(err)
		}()
	}

	thread := &starlark.Thread{Load: repl.MakeLoad()}
	globals := make(starlark.StringDict)

	switch {
	case flag.NArg() == 1 || *execprog != "":
		var (
			filename string
			src      interface{}
			err      error
		)
		if *execprog != "" {
			// Execute provided program.
			filename = "cmdline"
			src = *execprog
		} else {
			// Execute specified file.
			filename = flag.Arg(0)
		}
		thread.Name = "exec " + filename

		predeclared := starlark.StringDict{
			"sh": starlark.NewBuiltin("sh", starish.Sh), //TODO: move this from builtins
		}
		globals, err = starlark.ExecFile(thread, filename, src, predeclared)

		if err != nil {
			repl.PrintError(err)
			return 1
		}
	case flag.NArg() == 0:
		fmt.Println("Welcome to Starish (go.starlark.net)")
		thread.Name = "REPL"
		repl.REPL(thread, globals)
		return 0
	default:
		log.Print("want at most one Starish file name")
		return 1
	}

	// Print the global environment.
	if *showenv {
		for _, name := range globals.Keys() {
			if !strings.HasPrefix(name, "_") {
				fmt.Fprintf(os.Stderr, "%s = %s\n", name, globals[name])
			}
		}
	}

	return 0
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
