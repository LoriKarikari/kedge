# Kedge Development Guide

Long-running controller that syncs Docker Compose deployments to Git. Polls repos for changes, deploys on new commits, detects drift (running state ≠ Git state), and auto-remediates. NOT Kubernetes.

## Project Structure

- `cmd/kedge/` - CLI entrypoint
- `internal/` - All application code (controller, docker, git, state, telemetry, etc.)
- `docs/` - MkDocs documentation site

## Commands

```bash
make build          # Build binary to bin/kedge
make test           # Run tests with race detection
make lint           # Run golangci-lint
make check          # fmt, vet, lint, gosec, test (pre-commit)
make dev            # Hot reload with air
go test ./internal/docker/...  # Test single package
```

## Code Style

- No `fmt.Print*` or `log.*` - use `slog` logger instead
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Use `context.Context` as first param, propagate cancellation
- Prefer `lo` (samber/lo) for slice/map operations
- Table-driven tests with `t.Run()` subtests
- File order: types/constants first, then functions
- Imports: stdlib, blank, external, blank, internal

```go
import (
    "context"

    "github.com/samber/lo"

    "github.com/LoriKarikari/kedge/internal/docker"
)
```

### Naming

- Main logic in `<package>.go`, tests in `<package>_test.go`
- Test helpers in `testing.go` (unexported)
- Sentinel errors: `ErrNotFound`, `ErrInvalidStatus`
- Constants for enums: `StatusPending`, `ActionCreate`

## Architecture

| Package | Purpose |
|---------|---------|
| `controller` | Main reconciliation loop - watches git, detects drift, deploys |
| `docker` | Compose ops via Docker API (client, compose, differ, deploy) |
| `git` | Repository watcher with polling |
| `state` | SQLite persistence + golang-migrate migrations |
| `manager` | Multi-repo orchestration |
| `telemetry` | OpenTelemetry metrics, Prometheus `/metrics` endpoint |

## Key Libraries

- `slog` + `tint` - Structured logging
- `cobra` + `viper` - CLI and config
- `charmbracelet/huh` - Interactive CLI prompts
- `huma/v2` - HTTP API framework
- `go-git/go-git/v5` - Git operations
- `compose-spec/compose-go` - Parse compose files
- `docker/docker` - Docker API client
- `samber/lo` - Functional helpers
- `zog` - Schema validation
- `modernc.org/sqlite` - Pure Go SQLite

## Testing

- Helpers in `internal/docker/testing.go` for Docker mocks
- Use `t.TempDir()` for file-based tests

## References

- [Docs](https://lorikarikari.github.io/kedge/) | [Config](docs/docs/configuration.md) | [CLI](docs/docs/cli/index.md)
