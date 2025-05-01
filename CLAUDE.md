# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands
- Build: `make build` or `go build -o globus`
- Run: `go run main.go`
- Test: `make test` or `go test ./...`
- Test single package: `go test ./pkg/output`
- Format: `go fmt ./...`
- Lint: `make lint` or `golangci-lint run`
- Vet: `go vet ./...`
- Coverage: `make test-coverage`

## Code Style Guidelines
- **Imports**: Group stdlib first, then external deps, then internal packages
- **Formatting**: Follow Go standard (`go fmt`) with 2-space indentation
- **Error Handling**: Always check errors and propagate with context using `fmt.Errorf("context: %w", err)`
- **Naming**: Use camelCase for variables, PascalCase for exported names
- **Types**: Prefer strong typing with explicit struct definitions
- **Comments**: All exported functions require comments (godoc style)
- **File Headers**: Include SPDX license headers on all files
- **Command Structure**: Follow Cobra command pattern for all CLI commands
- **Git Workflow**: Use conventional commits format (`feat`, `fix`, `docs`, etc.)
- **Project Structure**: Follow Go standard project layout with `cmd`, `pkg`, etc.