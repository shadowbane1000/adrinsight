# Implementation Plan: RAG Pipeline and HTTP API

**Branch**: `002-p1-rag-pipeline` | **Date**: 2026-04-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-p1-rag-pipeline/spec.md`

## Summary

Wire the Milestone 1 indexing pipeline into a full RAG system: accept a
natural-language question via HTTP API, embed the question, retrieve relevant
ADR chunks, expand to full ADR content from disk, send to Anthropic Claude for
synthesis, and return a structured JSON response with answer text and citations.
Also expose ADR browsing endpoints.

## Technical Context

**Language/Version**: Go 1.24+ (same as M1)
**Primary Dependencies**:
- anthropics/anthropic-sdk-go (official Anthropic Go SDK)
- net/http stdlib ServeMux (Go 1.22+ method routing, no framework)
- Existing M1 packages: parser, embedder, store
**Storage**: SQLite with sqlite-vec (from M1, unchanged)
**Testing**: `go test` with table-driven tests, httptest for API tests
**Target Platform**: Linux (primary), macOS (dev)
**Project Type**: Web service (HTTP API serving JSON)
**Performance Goals**: Query response within 10 seconds (dominated by LLM latency)
**Constraints**: Single-user, no auth, no caching (this milestone)
**Scale/Scope**: Single instance, 3 HTTP endpoints

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | ✅ PASS | LLM behind interface, server uses interfaces for all deps |
| II. ADRs Are the Product | ✅ PASS | ADR checkpoint after planning |
| III. Skateboard First | ✅ PASS | Full ADR expansion (simple), no caching, no auth |
| IV. Idiomatic Go | ✅ PASS | stdlib net/http, no framework. Official Anthropic SDK. |
| V. Developer Experience | ✅ PASS | Single env var (ANTHROPIC_API_KEY) + existing Ollama |

## Project Structure

### Documentation (this feature)

```text
specs/002-p1-rag-pipeline/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (HTTP API contracts)
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (new packages added to repository root)

```text
internal/
├── llm/
│   ├── llm.go               # LLM interface
│   ├── anthropic.go          # Anthropic Claude implementation
│   └── anthropic_test.go
├── rag/
│   ├── rag.go                # RAG pipeline: query → embed → retrieve → expand → synthesize
│   └── rag_test.go
└── server/
    ├── server.go             # HTTP server setup + routes
    ├── handlers.go           # Handler functions (query, list ADRs, get ADR)
    └── handlers_test.go
```

Existing packages from M1 (unchanged):
- `internal/parser/` — used by server to read full ADR files for expansion
- `internal/embedder/` — used by RAG pipeline for query embedding
- `internal/store/` — used by RAG pipeline for similarity search

`cmd/adr-insight/main.go` — add `serve` command alongside existing `reindex`/`search`

**Structure Decision**: Three new packages under `internal/` following the
established pattern. LLM interface mirrors the Embedder/Store/Parser pattern
from M1 (Constitution Principle I).

## Complexity Tracking

No constitution violations — no entries needed.
