
# Starish

Starish: python-like shell scripting, based on the Starlark langugage.

[![Travis CI](https://travis-ci.org/kassybas/starish.svg?branch=master)](https://travis-ci.org/kassybas/starish)


## Quickstart

**Create** a `Starfile`

```py 
name = "World"
# Variables in scope are injected as variables to shell
sh('echo "Hello ${name}"')
# output: "Hello World"
```

See more on the `sh` function below.

**Run** in the same working directory

``` shell
starish
```

**Call** a function directly from CLI

```py 
# ./Starfile
def foo():
  print("Hello foo")
```

``` sh
$ starish foo
# output: "Hello foo"
```

**Start** in interactive mode with flag `-i`

``` shell
$ starish -i
```

## Install

### MacOS

```
curl -Lo /usr/local/bin/starish https://github.com/kassybas/starish/releases/download/0.2.2/starish_amd64_darwin_0.2.2
chmod +x /usr/local/bin/starish
```

### Linux

```
curl -Lo /usr/local/bin/starish https://github.com/kassybas/starish/releases/download/0.2.2/starish_amd64_linux_0.2.2
chmod +x /usr/local/bin/starish
```

### Docker

Pull the latest release:

``` shell
docker pull kassybas/starish
```

Or copy starish to your Docker image:

``` Dockerfile
COPY --from=kassybas/starish /starish /usr/local/bin/starish
```

## Why

The purpose of starish is to provide a readable and extenadble way to run scripts locally, . 
It is not intended to be a general purpose programming language but rather glue code for the small bits of automations which are usually placed in shell scripts or Makefiles.

## Goals

### Portable interpreter

Written in Go, distributed as a single binary, it requires no packages or libraries to be installed on the system.


### Readable

The python-like syntax makes the scripts familiar, accessible and easily readable, by minimizing boilerplate.

### Reusable

Using the `load` function, you can import other starish (or starlark) files.

``` python
load("./docker.star", "docker")
docker.build(img, path)
```

### Interactive shell

With interactive shell (REPL) starish has the ability to execute each command one by one for testing and development.

### Shell integration

The main extension of starish is the special `sh()` function, which makes it possbile to interact with starish variables in the invoked shell scripts.

The captured the stdout, stderr and status code is returned.

``` python
foo = "bar"
out, err, rc = sh("""
  echo "hello world" >&2
  echo "${foo}"
  exit 42
""")
```

### Command line integration
TODO

### Complex variables

## Starlark vs Starish




## Documentation of the Starlark language

* Starlark README: [starlark-readme.md](./starlark-readme.md)

* Language definition: [doc/spec.md](doc/spec.md)

* About the Go implementation: [doc/impl.md](doc/impl.md)

* API documentation: [godoc.org/go.starlark.net/starlark](https://godoc.org/go.starlark.net/starlark)

* Mailing list: [starlark-go](https://groups.google.com/forum/#!forum/starlark-go)

* Issue tracker: [https://github.com/google/starlark-go/issues](https://github.com/google/starlark-go/issues)
