# Implementation Plan: Foundation — ADR Indexing Pipeline

**Branch**: `001-p1-foundation` | **Date**: 2026-04-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-p1-foundation/spec.md`

## Summary

Build the core data pipeline: parse ADR markdown files from a directory,
generate vector embeddings via Ollama, store everything in SQLite with
sqlite-vec, and expose a `reindex` CLI command plus basic similarity search.
Establish CI and local dev tooling (Makefile) for the project.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**:
- goldmark + goldmark-meta (markdown parsing + YAML frontmatter)
- ncruces/go-sqlite3 (SQLite via WASM, no CGO)
- asg017/sqlite-vec-go-bindings/ncruces (sqlite-vec extension)
- ollama/ollama/api (official Ollama Go client)
**Storage**: SQLite with sqlite-vec extension (embedded, single file)
**Testing**: `go test` with table-driven tests, testdata/ fixtures
**Target Platform**: Linux (primary), macOS (dev)
**Project Type**: CLI tool (will become web service in Milestone 2)
**Performance Goals**: Reindex ≤50 ADRs in under 30 seconds
**Constraints**: No CGO — pure Go + WASM for cross-platform builds
**Scale/Scope**: Dozens to low hundreds of ADR files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | ✅ PASS | Embedder, Store, and Parser behind Go interfaces |
| II. ADRs Are the Product | ✅ PASS | ADR generation checkpoint after planning |
| III. Skateboard First | ✅ PASS | Full rebuild on reindex, no incremental optimization |
| IV. Idiomatic Go | ✅ PASS | cmd/ + internal/ layout, table-driven tests, standard lib where possible |
| V. Developer Experience | ✅ PASS | Makefile for local dev, single `reindex` command |

## Project Structure

### Documentation (this feature)

```text
specs/001-p1-foundation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI contract)
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
cmd/
└── adr-insight/
    └── main.go              # entrypoint, CLI command routing

internal/
├── parser/
│   ├── parser.go            # Parser interface + ADR struct
│   ├── markdown.go          # goldmark-based implementation
│   └── markdown_test.go
├── embedder/
│   ├── embedder.go          # Embedder interface
│   ├── ollama.go            # Ollama implementation
│   └── ollama_test.go
├── store/
│   ├── store.go             # Store interface
│   ├── sqlite.go            # SQLite + sqlite-vec implementation
│   └── sqlite_test.go
└── reindex/
    ├── reindex.go           # Pipeline orchestration (parse → embed → store)
    └── reindex_test.go

testdata/
├── ADR-001-sample.md        # Valid ADR fixture
├── ADR-002-sample.md        # Valid ADR fixture
├── not-an-adr.md            # Non-matching file for skip tests
└── ADR-bad-frontmatter.md   # Malformed frontmatter fixture

Makefile                     # lint, test, build targets
.golangci.yml                # linter configuration
.gitea/workflows/ci.yaml     # Gitea Actions CI
.github/workflows/ci.yaml    # GitHub Actions CI
```

**Structure Decision**: Standard Go `cmd/` + `internal/` layout per
Constitution Principle IV. The `reindex` package orchestrates the pipeline
and depends on the three interface packages (parser, embedder, store).

## Complexity Tracking

No constitution violations — no entries needed.
