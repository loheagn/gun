# gun

`gun` is a Go CLI for running tests near a specific file + line.

The name comes from Gungnir (Odin's spear).

## Features

- Input formats: `<file> <line>` and `<file>:<line>`
- Subcommands: `leaf`, `parent`, `test`, `file`, `pkg`, `project`
- Default mode without subcommand: auto choose `leaf` or `test`
- `--` passthrough to `go test` flags

## Install

```bash
go build ./cmd/gun
```

## Command Overview

```bash
gun leaf    <file> <line> [-- <go test args...>]
gun parent  <file> <line> [--up N] [-- <go test args...>]
gun test    <file> <line> [-- <go test args...>]
gun file    <file> <line> [-- <go test args...>]
gun pkg     <file> <line> [-- <go test args...>]
gun project <file> <line> [project-root] [-- <go test args...>]

# auto mode (no subcommand)
gun <file> <line> [-- <go test args...>]
```

## Behavior Details

- `leaf`: run the deepest matching `t.Run`.
- `leaf` fallback: if no `t.Run` exists in the containing `TestXxx`, fallback to `test`.
- `parent`: move up by `--up` levels (default `1`).
- `parent` overflow fallback: if `--up` exceeds depth, fallback to `test`.
- `test`: run containing top-level `TestXxx`.
- `file`: run all top-level `TestXxx` in the given file.
- `pkg`: run all tests in the package that contains the file.
- `project`: run `go test ./...` at `project-root` if provided, otherwise at the nearest module root (`go.mod`) of the file.

## Auto Mode (No Subcommand)

For `gun <file> <line>`:

- If line is in a resolvable `t.Run` chain: behaves like `leaf`.
- If there is no `t.Run` chain: fallback to `test`.
- If subtest name is dynamic and cannot be resolved: fallback to `test`.
- If line is outside any `Test/t.Run` block: error.

## Name Resolution for `t.Run`

`gun` statically resolves subtest names from:

- string literals
- constants
- string concatenation with `+`
- `fmt.Sprintf(...)` (when all args are statically resolvable)
- `strconv.Itoa(...)` (when arg is statically resolvable)

If explicit `leaf` or `parent` hits an unresolvable subtest name, `gun` returns an error and suggests broader scopes.

## Passthrough Flags

Use `--` to pass extra flags to `go test`:

```bash
gun leaf ./foo/bar_test.go 42 -- -v -count=1
```

For commands that already generate `-run` (`leaf`, `parent`, `test`, `file`, and auto mode), passing your own `-run` is rejected to avoid conflicts.

## Scope and Limitations

- Input file must end with `_test.go`.
- Supported test style: standard `testing` (`TestXxx` + `t.Run`).
- Not supported in v1: `Benchmark`, `Fuzz`, `Example`, `testify/suite`, Ginkgo.

## Exit Codes

- `0`: success
- `1`: `go test` failed
- `2`: input/locator/usage errors
