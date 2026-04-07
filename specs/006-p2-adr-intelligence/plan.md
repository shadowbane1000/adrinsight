# Implementation Plan: ADR Intelligence

**Branch**: `006-p2-adr-intelligence` | **Date**: 2026-04-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-p2-adr-intelligence/spec.md`

## Summary

Add relationship awareness to ADR Insight so the system understands ADRs as a connected graph. During reindex, parse "Related ADRs" sections and classify relationship types via LLM. Store relationships in a new table. Use relationship data to improve retrieval (co-retrieve related ADRs, replace content-heuristic supersession penalty), enrich LLM synthesis context (include relationship metadata so answers can trace decision chains), and display relationship links in the web UI.

## Technical Context

**Language/Version**: Go 1.24+ (same as M1-M4b)
**Primary Dependencies**: mattn/go-sqlite3 (CGO), sqlite-vec, Anthropic Go SDK, Ollama
**Storage**: SQLite with sqlite-vec + FTS5 + new `adr_relationships` table
**Testing**: Go table-driven tests, eval harness from M4a/M4b
**Target Platform**: Linux (local + Docker)
**Project Type**: CLI + web service
**Performance Goals**: Reindex with relationship parsing < 30s for 17 ADRs
**Constraints**: Relationship graph fits in memory (dozens of nodes/edges)
**Scale/Scope**: 17 ADRs, ~20 relationship edges currently

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | PASS | Relationship store methods added to existing Store interface. LLM-based classifier uses existing KeywordExtractor-style pattern. |
| II. ADRs Are the Product | PASS | New ADR for relationship storage schema and parsing approach. |
| II-A. ADR Immutability | PASS | No existing ADR bodies will be edited. New ADRs for new decisions. |
| III. Skateboard First | PASS | Parsing + synthesis context is the minimum viable slice. UI and reranker integration are P2/P3. |
| IV. Idiomatic Go | PASS | Uses cmd/ + internal/ layout, table-driven tests, standard error handling. |
| V. Developer Experience | PASS | Relationships parsed automatically during reindex. No new configuration needed. LLM classification uses existing API key. |

## Project Structure

### Documentation (this feature)

```text
specs/006-p2-adr-intelligence/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit-tasks)
```

### Source Code (repository root)

```text
internal/
├── parser/
│   └── markdown.go          # Extended: parse Related ADRs sections
├── store/
│   ├── store.go             # Extended: relationship methods on Store interface
│   └── sqlite.go            # Extended: adr_relationships table, CRUD methods
├── rag/
│   ├── rag.go               # Extended: relationship context in synthesis
│   └── rerank.go            # Extended: authoritative supersession from relationships
├── reindex/
│   └── reindex.go           # Extended: relationship extraction + LLM classification step
├── llm/
│   └── anthropic.go         # Extended: ClassifyRelationship method
└── server/
    └── handlers.go          # Extended: relationship data in ADR detail response

web/static/
├── app.js                   # Extended: display relationship links in detail panel
└── style.css                # Extended: relationship link styling

docs/adr/
└── ADR-018-*.md             # New: relationship storage and parsing approach
```

**Structure Decision**: All changes extend existing files in the established cmd/ + internal/ layout. No new packages needed — relationships are a cross-cutting concern that touches parser, store, rag, and server.

## ADR Impact

| Existing ADR | Impact | Action |
|-------------|--------|--------|
| None | No existing ADRs need to be superseded | ADR-020 (relationship model/storage), ADR-021 (LLM classification) |

## Complexity Tracking

No constitution violations. No complexity justification needed.
