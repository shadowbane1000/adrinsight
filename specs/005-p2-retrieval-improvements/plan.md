# Implementation Plan: Retrieval Improvements

**Branch**: `005-p2-retrieval-improvements` | **Date**: 2026-04-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-p2-retrieval-improvements/spec.md`

## Summary

Improve retrieval quality through three complementary approaches: chunking
experiments (evaluate alternatives to H2-section splitting), hybrid search
(FTS5 keyword search combined with vector similarity via weighted merge),
and result reranking (title boost, status-aware deprioritization). All
changes measured against the M4a evaluation harness baseline.

## Technical Context

**Language/Version**: Go 1.24+ (same as M1-M4a)
**Primary Dependencies**:
- Existing packages (parser, embedder, store, rag, eval)
- SQLite FTS5 (built into mattn/go-sqlite3, no additional dependency)
- sqlite-vec (unchanged)
**Storage**: SQLite with sqlite-vec + FTS5 virtual tables
**Testing**: M4a evaluation harness for retrieval quality, `go test` for units
**Target Platform**: CLI + web UI (unchanged)
**Project Type**: Backend retrieval improvements (no UI changes)
**Performance Goals**: No increase in query latency beyond 1 second
**Constraints**: All changes must pass M4a eval harness without regressions
**Scale/Scope**: 16 ADRs, ~80 chunks, 6 eval test cases

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | ✅ PASS | Store interface extended with FTS search method; Reranker behind interface |
| II. ADRs Are the Product | ✅ PASS | Chunking decision documented as ADR superseding/amending ADR-008 |
| II-A. ADR Immutability | ✅ PASS | ADR-008 body unchanged; new ADR supersedes it |
| III. Skateboard First | ✅ PASS | Simple BM25 + vector weighted merge; heuristic reranking, no ML |
| IV. Idiomatic Go | ✅ PASS | FTS5 via existing mattn/go-sqlite3; standard patterns |
| V. Developer Experience | ✅ PASS | `reindex` rebuilds all indices; eval harness validates changes |

## Project Structure

### Documentation (this feature)

```text
specs/005-p2-retrieval-improvements/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (new/modified files)

```text
internal/
├── store/
│   └── sqlite.go            # Add FTS5 table creation, FTS search method, hybrid search
├── rag/
│   ├── rag.go               # Update pipeline to use hybrid search + reranking
│   └── rerank.go            # Reranking heuristics (title boost, status deprioritization)
└── parser/
    └── markdown.go          # Chunking strategy changes (if needed after experiments)

testdata/
└── eval/
    └── experiments/         # Before/after scores for each chunking experiment
```

**Structure Decision**: FTS5 lives in the store package (it's a storage
concern). Reranking is a new file in the rag package (it's a retrieval
pipeline concern). Chunking changes stay in parser (existing location).

## ADR Impact

| Existing ADR | Impact | Action |
|-------------|--------|--------|
| ADR-008 | Chunking strategy may change based on experiments | Create new ADR superseding ADR-008 with experiment results |

## Complexity Tracking

No constitution violations — no entries needed.
