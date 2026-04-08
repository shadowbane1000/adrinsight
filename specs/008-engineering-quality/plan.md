# Implementation Plan: Engineering Quality

**Branch**: `008-engineering-quality` | **Date**: 2026-04-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-engineering-quality/spec.md`

## Summary

Bring the ADR Insight codebase to production-grade quality: replace all unstructured logging with `log/slog`, centralize configuration in a config struct, add request ID tracing and timing middleware, implement a `/health` endpoint, add graceful shutdown with signal handling, audit and fix all error handling, review Go idioms, add integration tests, and document the HTTP API.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `log/slog` (stdlib, Go 1.21+), `net/http` (existing), `os/signal` (stdlib)
**Storage**: SQLite with sqlite-vec + FTS5 (unchanged)
**Testing**: `go test` with real SQLite (in-memory), mock embedder, table-driven tests
**Target Platform**: Linux (Docker), macOS (development)
**Project Type**: Web service with embedded static frontend
**Performance Goals**: Health check responds within 5 seconds; no regression in query latency
**Constraints**: No new external dependencies — all stdlib. Constitution Principle IV (Idiomatic Go).
**Scale/Scope**: Internal refactoring — touches all Go packages but no new features

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | PASS | No new interfaces needed. Existing interfaces unchanged. |
| II. ADRs Are the Product | PASS | ADR for slog choice. No ADR body edits. |
| II-A. ADR Immutability | PASS | No existing ADRs modified. |
| III. Skateboard First | PASS | slog is the simplest structured logging (stdlib, zero deps). Config struct is the simplest centralization. No over-engineering. |
| IV. Idiomatic Go | PASS | This milestone IS the idiom review. slog is the idiomatic Go logging solution. Wrapped errors with %w. |
| V. Developer Experience | PASS | `docker-compose up` still works. Config defaults unchanged. Health check aids debugging. |

All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/008-engineering-quality/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (config struct, health response)
├── contracts/           # Phase 1 output (health endpoint contract)
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── config/
│   └── config.go        # NEW: centralized Config struct, env loading
├── server/
│   ├── server.go         # Modified: graceful shutdown, config injection
│   ├── handlers.go       # Modified: health endpoint, structured logging
│   ├── middleware.go      # NEW: request ID, timing, logging middleware
│   └── handlers_test.go  # Modified: health endpoint tests
├── rag/
│   └── rag.go            # Modified: slog, error wrapping
├── store/
│   └── sqlite.go         # Modified: slog, error wrapping
├── reindex/
│   └── reindex.go        # Modified: slog, error wrapping
├── llm/
│   └── anthropic.go      # Modified: slog, error wrapping
├── parser/
│   └── markdown.go       # Modified: slog (minimal)
├── embedder/
│   └── ollama.go         # Modified: slog, error wrapping
└── eval/
    └── eval.go           # Modified: slog

cmd/adr-insight/
└── main.go              # Modified: config loading, slog setup, signal handling, graceful shutdown

docs/
└── api.md               # NEW: HTTP API documentation
```

**Structure Decision**: One new package (`internal/config/`), one new file (`internal/server/middleware.go`), one new doc (`docs/api.md`). Everything else is modifications to existing files.

## ADR Impact

| Existing ADR | Impact | Action |
|-------------|--------|--------|
| None | No existing decisions overridden | No superseding ADRs needed |

New ADR needed: slog structured logging choice (ADR-024).

## Complexity Tracking

No constitution violations to justify.
