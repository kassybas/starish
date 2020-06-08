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
	var env *Dict
	if err := UnpackArgs("sh", nil, kwargs, "shell?", &shellPath, "shield_env?", &shieldEnv, "silent?", &silent, "sep?", &sep, ":env?", &env); err != nil {
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
	envVars, err := getEnvVars(env)
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

func getEnvVars(arg Value) ([]string, error) {
	// TODO(kassybas): handle errors sanely
	if arg.Type() != "dict" {
		return nil, fmt.Errorf("sh env vars: expected dict, got: %s", arg.Type())
	}
	d := arg.(*Dict)
	res := make([]string, len(d.Keys()))
	for i, k := range d.Keys() {
		v, _, _ := d.Get(k)
		name, _ := AsString(k)
		val, _ := AsString(v)
		res[i] = fmt.Sprintf("%s=%s", name, val)
	}
	return res, nil
}
