package starish

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var FileModule = &starlarkstruct.Module{
	Name: "file",
	Members: starlark.StringDict{
		"read":  starlark.NewBuiltin("file.read", readFile),
		"write": starlark.NewBuiltin("file.write", writeFile),
	},
}

func readFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	path, ok := starlark.AsString(args.Index(0))
	if !ok {
		return nil, fmt.Errorf("open: non-string argument %+v, want string", args.Index(0))
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return starlark.String(content), nil
}

func writeFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 2 {
		return nil, fmt.Errorf("write: want exactly 2 arguments: file name, source", args.Index(0))

	}
	path, ok := starlark.AsString(args.Index(0))
	if !ok {
		return nil, fmt.Errorf("write: non-string argument %+v, want string", args.Index(0))
	}
	content, ok := starlark.AsString(args.Index(1))
	if !ok {
		return nil, fmt.Errorf("write: non-string argument %+v, want string", args.Index(1))
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(content)
	f.Sync()

	return nil, nil
}
