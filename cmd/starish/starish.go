// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The starlark command interprets a Starlark file.
// With no arguments, it starts a read-eval-print loop (REPL).
package main // import "go.starlark.net/cmd/starlark"

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/urfave/cli"

	"go.starlark.net/internal/compile"
	"go.starlark.net/repl"
	"go.starlark.net/resolve"
	"go.starlark.net/starish"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlarkyaml"
)

type argConfig struct {
	cpuProfile string
	memProfile string
	profile    string
	showEnv    bool
	execProg   string
	fileName   string
	targetFunc string
	funcArgs   []string
	REPL       bool
}

func init() {
	// starish defaults
	resolve.AllowFloat = true
	resolve.AllowSet = true
	resolve.AllowLambda = true
	resolve.AllowNestedDef = true
	resolve.AllowRecursion = true
}

func main() {
	app := cli.NewApp()
	app.Name = "starish"
	app.Usage = "Starish: starlark integrated shell"
	app.Version = "0.1.0"

	var args argConfig

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       "Starfile",
			Usage:       "source file name",
			Destination: &args.fileName,
		},
		cli.StringFlag{
			Name:        "profile",
			Usage:       "gather Starlark time profile in this file",
			Destination: &args.profile,
		},
		cli.StringFlag{
			Name:        "cpuprofile",
			Usage:       "gather Go CPU profile in this file",
			Destination: &args.cpuProfile,
		},
		cli.StringFlag{
			Name:        "memprofile",
			Usage:       "gather Go memory profile in this file",
			Destination: &args.memProfile,
		},
		cli.BoolFlag{
			Name:        "showenv",
			Usage:       "on success, print final global environment",
			Destination: &args.showEnv,
		},
		cli.StringFlag{
			Name:        "c",
			Usage:       "execute program `prog`",
			Destination: &args.execProg,
		},
		cli.BoolFlag{
			Name:        "interactive, i",
			Usage:       "start interactive REPL shell",
			Destination: &args.REPL,
		},
		cli.BoolFlag{
			Name:        "disassemble",
			Usage:       "show disassembly during compilation of each function",
			Destination: &compile.Disassemble,
		},
		cli.BoolFlag{
			Name:        "globalreassign",
			Usage:       "allow reassignment of globals, and if/for/while statements at top level",
			Destination: &resolve.AllowGlobalReassign,
		},
	}
	app.Action = func(c *cli.Context) {
		if c.NArg() > 0 {
			args.targetFunc = c.Args()[0]
			args.funcArgs = c.Args()[1:]
		}
		os.Exit(doMain(args))
	}
	app.Run(os.Args)
}

func doMain(args argConfig) int {
	log.SetPrefix("starish: ")
	log.SetFlags(0)
	// flag.Parse()

	if args.cpuProfile != "" {
		f, err := os.Create(args.cpuProfile)
		check(err)
		err = pprof.StartCPUProfile(f)
		check(err)
		defer func() {
			pprof.StopCPUProfile()
			err := f.Close()
			check(err)
		}()
	}
	if args.memProfile != "" {
		f, err := os.Create(args.memProfile)
		check(err)
		defer func() {
			runtime.GC()
			err := pprof.Lookup("heap").WriteTo(f, 0)
			check(err)
			err = f.Close()
			check(err)
		}()
	}

	if args.profile != "" {
		f, err := os.Create(args.profile)
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

	// Ideally this statement would update the predeclared environment.
	// TODO(adonovan): plumb predeclared env through to the REPL.
	starlark.Universe["json"] = starlarkjson.Module

	// Starish extensions
	starlark.Universe["sh"] = starlark.NewBuiltin("sh", starish.Sh)
	starlark.Universe["module"] = starlark.NewBuiltin("module", starlarkstruct.MakeModule)
	starlark.Universe["yaml"] = starlarkyaml.Module
	starlark.Universe["file"] = starish.FileModule

	switch {
	case args.REPL:
		if args.fileName != "" && args.fileName != "Starfile" {
			log.Print("ambiguous command line arguments: interactive mode (-i) and file (-f)")
			return 1
		}
		fmt.Println("Welcome to Starish")
		thread.Name = "REPL"
		repl.REPL(thread, globals)
		return 0
	case args.fileName != "Starfile" || args.targetFunc != "" || args.execProg != "":
		var (
			src interface{}
			err error
		)
		if args.execProg != "" {
			// Execute provided program.
			args.fileName = "cmdline"
			src = args.execProg
		} else if args.targetFunc != "" {
			// Execute called function after loading the top level file
			// There is a nicer solution to this during eval, but this keeps starlark
			// pristine for uplift, decreases plumbing
			funcCall, err := makeFuncCall(args.targetFunc, args.funcArgs)
			check(err)
			src = fmt.Sprintf(`load("%s", "%s"); %s`, args.fileName, args.targetFunc, funcCall)
		}

		thread.Name = "exec " + args.fileName
		// if args.targetFunc != "" {
		// 	thread.SetLocal("targetFunc", args.targetFunc)
		// 	thread.SetLocal("funcArgs", args.funcArgs)
		// }
		globals, err = starlark.ExecFile(thread, args.fileName, src, nil)

		if err != nil {
			repl.PrintError(err)
			return 1
		}
	default:
		log.Print("want at most one Starish file name")
		return 1
	}

	// Print the global environment.
	if args.showEnv {
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

func makeFuncCall(targetFunc string, funcArgs []string) (string, error) {
	if len(funcArgs) == 2 && funcArgs[0] == "--" {
		// Complex function arguments after string --
		return fmt.Sprintf("%s%s", targetFunc, funcArgs[1]), nil
	}
	s := fmt.Sprintf("%s(", targetFunc)
	// Simple function arguments
	// TODO(kassybas): support '--arg=value' format
	for i := 0; i < len(funcArgs); i++ {
		if strings.HasPrefix(funcArgs[i], "--") {
			if i+1 < len(funcArgs) {
				if strings.HasPrefix(funcArgs[i+1], "--") {
					// Boolean arg
					s += fmt.Sprintf("%s=True, ", strings.TrimLeft(funcArgs[i], "-"))
				} else {
					// Key=value arg
					s += fmt.Sprintf(`%s="%s", `, strings.TrimLeft(funcArgs[i], "-"), funcArgs[i+1])
					i++
				}

			} else {
				// Last named boolean
				s += fmt.Sprintf("%s=True, ", strings.TrimLeft(funcArgs[i], "-"))
			}
		} else {
			// Positional argument
			s += fmt.Sprintf(`"%s", `, funcArgs[i])
		}
	}
	strings.TrimSuffix(s, ", ")
	s += ")"
	return s, nil
}
