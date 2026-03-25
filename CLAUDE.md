# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific module
go test ./log/...
go test ./config/...
go test ./di/...
go test ./graceful/...
go test ./utils/...

# Run a single test
go test ./log/... -run TestFunctionName

# Run tests with verbose output
go test -v ./...

# Tidy dependencies
go mod tidy

# Build (verify compilation)
go build ./...
```

## Architecture

`fg-lib` is a Go library monorepo — a collection of lightweight utility packages with no `main.go`. Each subdirectory is an independent package under the module `github.com/fuguiw/fg-lib`.

### Modules

- **`log/`** — Structured logging wrapper. Exposes a global `slog.Logger` backed by Zap (`zap_handler.go` implements `slog.Handler`). Supports stdout/stderr/file output with log rotation (lumberjack), dynamic level control via `atomic.Value`, and context-injected fields (trace_id, request_id, user_id via `context.go`). Initialized once via `sync.Once`.

- **`config/`** — Reflection-based config loader. Populates structs using priority: env vars > file (YAML/JSON) > struct `default` tags. Key struct tags: `default`, `env`, `yaml`, `json`. Supports auto-refresh with file watching and callbacks. Type coercion handles primitives, slices, and `time.Duration`.

- **`di/`** — Thin singleton wrapper around `uber-go/dig`. Provides `MustProvide` and `MustInvoke` convenience methods that panic on error. Single global container managed via `sync.Once`.

- **`graceful/`** — Application lifecycle manager. Components implement `Name() string`, `Start(ctx) error`, `Stop(ctx) error`. The manager registers components, listens for SIGINT/SIGTERM (`signal.go`), and calls `stopAll` with a configurable timeout when shutdown is triggered.

- **`utils/`** — Terminal utilities: `ShowJsonDiff()` (colorized JSON diff via jsondiff), `PrintTable()` (ASCII tables via go-pretty), `EditInTempFile()` (opens `$EDITOR` on a temp file and returns edited content).

### Design Patterns

- **Functional options**: All modules with configuration use `WithXxx(...)` option functions (e.g., `graceful.WithTimeout`, `log.WithLevel`).
- **Singletons**: `log` and `di` expose package-level functions backed by a single instance; call their `Init`/`New` functions once at startup.
- **Wrapper pattern**: `log` wraps Zap behind the standard `slog` interface; `di` wraps `uber-go/dig` behind a simpler API.
