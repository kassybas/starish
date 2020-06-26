package starish

import (
	"fmt"

	"github.com/kassybas/shell-exec/exec"
	"go.starlark.net/starlark"
)

func Sh(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	shellPath := "/bin/sh"
	sep := "_"
	silent := false
	shieldEnv := false
	if err := starlark.UnpackArgs("sh", nil, kwargs, "shell?", &shellPath, "shield_env?", &shieldEnv, "silent?", &silent, "sep?", &sep); err != nil {
		return nil, err
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("sh: got %d arguments, want exactly 1", len(args))
	}
	script, ok := starlark.AsString(args.Index(0))
	if !ok {
		return nil, fmt.Errorf("sh: non-string argument %+v, want string", args.Index(0))
	}

	opts := exec.Options{
		Silent:    silent,
		ShieldEnv: shieldEnv,
		ShellPath: shellPath,
	}
	starishEnv := thread.Local("starishEnv").(*starlark.Dict)
	envVars, err := getEnvVars(starishEnv, sep)
	if err != nil {
		return nil, fmt.Errorf("sh: could not parse variables")
	}
	output, errOutput, statusCode, err := exec.ShellExec(script, envVars, opts)
	if err != nil {
		return nil, err
	}
	result := []starlark.Value{starlark.String(output), starlark.String(errOutput), starlark.MakeInt(statusCode)}
	return starlark.NewList(result), nil
}

func getEnvVars(arg starlark.Value, sep string) ([]string, error) {
	if arg.Type() != "dict" {
		return nil, fmt.Errorf("sh internal error: expected dict, got: %s", arg.Type())
	}
	d := arg.(*starlark.Dict)
	res := []string{}
	for _, k := range d.Keys() {
		name, _ := starlark.AsString(k)
		v, _, _ := d.Get(k)
		vals := getDeepEnvVars(v, name, sep)
		res = append(res, vals...)
	}
	return res, nil
}

func getDeepEnvVars(v starlark.Value, prefix, sep string) []string {
	res := []string{}
	switch v.Type() {
	case "string":
		val, _ := starlark.AsString(v)
		res = []string{fmt.Sprintf("%s=%s", prefix, val)}
	case "int", "float", "bool", "NoneType":
		res = []string{fmt.Sprintf("%s=%s", prefix, v.String())}
	case "dict":
		d := v.(*starlark.Dict)
		for _, k := range d.Keys() {
			name, _ := starlark.AsString(k)
			newPrefix := fmt.Sprintf("%s%s%s", prefix, sep, name)
			v, _, _ := d.Get(k)
			res = append(res, getDeepEnvVars(v, newPrefix, sep)...)
		}
	case "tuple", "list", "set":
		iter := starlark.Iterate(v)
		var z starlark.Value
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
