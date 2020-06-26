package starlark

import (
	"fmt"

	"github.com/kassybas/shell-exec/exec"
)

func sh(thread *Thread, b *Builtin, args Tuple, kwargs []Tuple) (Value, error) {
	shellPath := "/bin/sh"
	sep := "_"
	silent := false
	shieldEnv := false
	if err := UnpackArgs("sh", nil, kwargs, "shell?", &shellPath, "shield_env?", &shieldEnv, "silent?", &silent, "sep?", &sep); err != nil {
		return nil, err
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("sh: got %d arguments, want exactly 1", len(args))
	}
	script, ok := AsString(args.Index(0))
	if !ok {
		return nil, fmt.Errorf("sh: non-string argument %+v, want string", args.Index(0))
	}

	opts := exec.Options{
		Silent:    silent,
		ShieldEnv: shieldEnv,
		ShellPath: shellPath,
	}
	starishEnv := thread.Local("starishEnv").(*Dict)
	envVars, err := getEnvVars(starishEnv, sep)
	if err != nil {
		return nil, fmt.Errorf("sh: could not parse variables")
	}
	output, errOutput, statusCode, err := exec.ShellExec(script, envVars, opts)
	if err != nil {
		return nil, err
	}
	result := []Value{String(output), String(errOutput), MakeInt(statusCode)}
	return NewList(result), nil
}

func getEnvVars(arg Value, sep string) ([]string, error) {
	if arg.Type() != "dict" {
		return nil, fmt.Errorf("sh internal error: expected dict, got: %s", arg.Type())
	}
	d := arg.(*Dict)
	res := []string{}
	for _, k := range d.Keys() {
		name, _ := AsString(k)
		v, _, _ := d.Get(k)
		vals := getDeepEnvVars(v, name, sep)
		res = append(res, vals...)
	}
	return res, nil
}

func getDeepEnvVars(v Value, prefix, sep string) []string {
	res := []string{}
	switch v.Type() {
	case "string":
		val, _ := AsString(v)
		res = []string{fmt.Sprintf("%s=%s", prefix, val)}
	case "int", "float", "bool", "NoneType":
		res = []string{fmt.Sprintf("%s=%s", prefix, v.String())}
	case "dict":
		d := v.(*Dict)
		for _, k := range d.Keys() {
			name, _ := AsString(k)
			newPrefix := fmt.Sprintf("%s%s%s", prefix, sep, name)
			v, _, _ := d.Get(k)
			res = append(res, getDeepEnvVars(v, newPrefix, sep)...)
		}
	case "tuple", "list", "set":
		iter := Iterate(v)
		var z Value
		i := 0
		for iter.Next(&z) {
			i++
			newPrefix := fmt.Sprintf("%s%s%d", prefix, sep, i)
			res = append(res, getDeepEnvVars(z, newPrefix, sep)...)
		}
		iter.Done()
	}
	return res
}
