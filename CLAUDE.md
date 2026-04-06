# CLAUDE.md — Project Instructions for Claude Code

## Project

ADR Insight — AI-powered search and reasoning over Architecture Decision Records. Written in Go.

## Key Files

- `CONTEXT.md` — Full project context, architecture, roadmap, and decisions. **Read this first.**
- `docs/adr/` — Architecture Decision Records for this project (also the demo dataset)

## Conventions

- Go project using `cmd/` + `internal/` layout
- Every external dependency (LLM, embedder, store) behind a Go interface
- Use idiomatic Go: explicit error handling, short variable names in small scopes, table-driven tests
- Lint with golangci-lint (config in `.golangci.yml`)
- CI runs on Gitea Actions and GitHub Actions: lint + test + build

## Stack

- **Language:** Go
- **LLM Synthesis:** Anthropic Claude API
- **Embeddings:** Ollama with `nomic-embed-text` (runs locally in Docker)
- **Storage:** SQLite with `sqlite-vec` extension
- **Web UI:** Simple HTML/CSS/JS served by the Go HTTP server (no frontend framework for v1)

## When Making Decisions

If a design choice is non-trivial, write an ADR in `docs/adr/` following the existing format (see ADR-001 through ADR-004 for examples). ADRs are a core part of this project's value — they're the demo dataset and interview talking points.

## Current Phase

Phase 1 — Walking Skeleton. See CONTEXT.md for the full breakdown. The immediate next steps are:

1. Initialize Go module and scaffold the directory structure
2. Build the ADR markdown parser (`internal/parser/`)
3. Set up Gitea Actions + GitHub Actions CI (lint + test + build)
4. Ollama embedding integration (`internal/embedder/`)
5. SQLite + sqlite-vec storage (`internal/store/`)
6. `reindex` command wiring parse → embed → store

## Testing

- Prefer table-driven tests
- Focus integration tests on the RAG pipeline (does retrieval return the right ADRs? does synthesis cite correctly?)
- Don't chase coverage numbers — test what matters
- Use `testdata/` for sample ADR fixtures

## Tyler's Go Level

Tyler is learning Go through this project. He's an expert in C#/C++/Java/Python. Help him write idiomatic Go — flag non-idiomatic patterns, suggest Go-specific approaches, but don't over-explain basic programming concepts.

## Active Technologies
- Go 1.22+ (001-p1-foundation)
- SQLite with sqlite-vec extension (embedded, single file) (001-p1-foundation)

## Recent Changes
- 001-p1-foundation: Added Go 1.22+
