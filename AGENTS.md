# agents.md

## Project
- Name: `gun`
- Module: `github.com/loheagn/gun`
- Purpose: Run nearby Go tests by file and line.

## Common Commands
- Run all tests: `go test ./...`
- Run race tests: `go test -race ./...`
- Build CLI: `go build ./cmd/gun`
- Run CLI locally: `go run ./cmd/gun --help`

## Repository Notes
- CLI entrypoint: `cmd/gun/main.go`
- Core logic: `internal/locator`
- Test execution: `internal/runner`
- CI workflows: `.github/workflows/ci.yml`, `.github/workflows/release.yml`
