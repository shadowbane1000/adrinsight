# Tasks: Retrieval Improvements

**Input**: Design documents from `/specs/005-p2-retrieval-improvements/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create experiment tracking directory and establish M4a baseline as the comparison point.

- [x] T001 Create directory `testdata/eval/experiments/` for storing per-experiment eval results
- [ ] T002 Run `./adr-insight eval --output testdata/eval/experiments/baseline-h2-sections.json` to capture the current retrieval scores as the experiment baseline

**Checkpoint**: Baseline scores captured in experiments directory for comparison.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add FTS5 infrastructure to the store — needed by hybrid search (US3) and benefits all stories.

**⚠️ CRITICAL**: Hybrid search cannot work until FTS5 is in place.

- [x] T003 Add `--skip-judge` flag to the `eval` command in `cmd/adr-insight/main.go` — when set, skip LLM judge scoring (set accuracy/completeness to 0, reason to "skipped"). This allows fast, cheap eval runs that only measure retrieval metrics (precision/recall/F1). Update `RunEval` in `internal/eval/eval.go` to accept a nil judge and skip scoring when nil.
- [x] T004 Add `fts_chunks` virtual table creation to `Reset()` in `internal/store/sqlite.go` — `CREATE VIRTUAL TABLE fts_chunks USING fts5(content)` alongside existing chunks and vec_chunks tables
- [x] T005 Update `StoreChunks()` in `internal/store/sqlite.go` — insert content into `fts_chunks` with matching rowid in the same transaction as chunks and vec_chunks
- [x] T006 Implement `SearchFTS(ctx, query string, topK int) ([]SearchResult, error)` in `internal/store/sqlite.go` — query `fts_chunks` with MATCH, join with `chunks` for metadata, normalize BM25 scores to 0-1 range using min-max normalization within the result set
- [x] T007 Add `SearchFTS` to the `Store` interface in `internal/store/store.go`
- [x] T008 Update mock stores in test files (`internal/rag/rag_test.go`, `internal/reindex/reindex_test.go`, `internal/server/handlers_test.go`) to satisfy updated Store interface with `SearchFTS` and `HybridSearch` stubs

**Checkpoint**: `reindex` populates FTS5 table. `SearchFTS` returns keyword-ranked results. All tests pass.

---

## Phase 3: User Story 2 — Evaluate Chunking Alternatives (Priority: P1) 🎯

**Goal**: Evaluate alternative chunking strategies against the baseline, select the best, ADR the decision.

**Independent Test**: At least two alternatives evaluated with documented before/after scores. ADR created.

**Note**: This is ordered before US1 because the chunking strategy affects all subsequent retrieval work.

### Implementation for User Story 2

- [x] T009 [US2] Implement "H2-section + whole-document embedding" chunking variant in `internal/parser/markdown.go` — add a method or flag that produces section chunks PLUS one whole-ADR chunk per document. Keep existing `ChunkADR` method unchanged; add `ChunkADRWithWholeDoc` or similar.
- [x] T010 [US2] Implement "H2-section with overlap preamble" chunking variant in `internal/parser/markdown.go` — each section chunk is prepended with the ADR title + first sentence of Context. Add as a separate method or flag.
- [ ] T011 [US2] Run experiment for each chunking variant: reindex with variant (reindex calls `Reset()` which clears all tables including FTS5, then rebuilds from scratch), run `./adr-insight eval --skip-judge --output testdata/eval/experiments/<strategy-name>.json`, record retrieval scores. Compare precision, recall, F1 against baseline. Skip LLM judge to save cost — retrieval metrics are sufficient for chunking comparison.
- [ ] T012 [US2] Select the best chunking strategy based on experiment results. Create an ADR that supersedes ADR-008, documenting experiment results, the chosen strategy, and why alternatives were rejected. Update ADR-008 status to "Superseded by ADR-NNN".
- [ ] T013 [US2] If the winning strategy differs from current H2-section splitting, update `ChunkADR` in `internal/parser/markdown.go` to use the new strategy as the default. Ensure `reindex` rebuilds correctly.

**Checkpoint**: Experiments documented, winning strategy selected and ADR'd, parser updated.

---

## Phase 4: User Story 1 — Better Answers to Sample Queries (Priority: P1)

**Goal**: Aggregate recall improves by at least 0.15 over M4a baseline with no regressions.

**Independent Test**: `./adr-insight eval` shows improved scores across all 6 test cases.

### Implementation for User Story 1

- [x] T014 [US1] Implement `HybridSearch(ctx, queryVec []float32, queryText string, topK int, vecWeight, kwWeight float64) ([]SearchResult, error)` in `internal/store/sqlite.go` — run vector search and FTS search independently, normalize scores to 0-1, combine with weighted merge, deduplicate by ADR number (keep highest combined score), return top-K by combined score. If FTS returns no matches (query has no keyword hits), return vector-only results — keyword weight effectively becomes 0.
- [x] T015 [US1] Add `HybridSearch` to the `Store` interface in `internal/store/store.go` and update mock stores in test files.
- [x] T016 [US1] Update `Pipeline.Query()` in `internal/rag/rag.go` — replace the current vector-only search with `HybridSearch`. Pass the raw query text alongside the embedded vector. Use default weights (0.7 vector, 0.3 keyword). The rest of the pipeline (dedup, expand, synthesize) remains unchanged.
- [ ] T017 [US1] Reindex and run eval: `./adr-insight reindex && ./adr-insight eval`. Verify aggregate recall improves by at least 0.15 and no test case regresses. If scores don't meet target, tune weights and re-evaluate.

**Checkpoint**: Eval harness shows measurable improvement. `RESULT: PASS`.

---

## Phase 5: User Story 3 — Keyword-Aware Search (Priority: P2)

**Goal**: Queries with exact technical terms reliably find the corresponding ADRs.

**Independent Test**: `./adr-insight search "goldmark"` returns ADR-007 in top 3.

### Implementation for User Story 3

- [x] T018 [US3] Update `cmdSearch` in `cmd/adr-insight/main.go` — use `HybridSearch` instead of vector-only `Search` for the `search` CLI command. Embed the query for vector search, pass raw text for FTS search.
- [ ] T019 [US3] Verify keyword search cases: search for "goldmark" (expect ADR-007), "ncruces" (expect ADR-006/015), "mattn" (expect ADR-015), "sqlite-vec" (expect ADR-004). Document results.

**Checkpoint**: Technical term queries reliably find the right ADRs.

---

## Phase 6: User Story 4 — Result Reranking (Priority: P3)

**Goal**: Reranking heuristics improve result ordering — active ADRs above superseded, title matches boosted.

**Independent Test**: Query "SQLite driver" returns ADR-015 above ADR-006.

### Implementation for User Story 4

- [x] T020 [P] [US4] Define `Reranker` interface and `RerankConfig` type in `internal/rag/rerank.go` per data-model.md. Implement `DefaultReranker` with three heuristics: title match boost (+0.2), status deprioritization for superseded/deprecated (-0.1), and section relevance boost for queries containing "why"/"rationale"/"alternative" (+0.1 for matching sections).
- [x] T021 [US4] Wire reranker into `Pipeline.Query()` in `internal/rag/rag.go` — after hybrid search returns deduplicated results, apply reranking before expansion to full ADR content. Use default config values.
- [ ] T022 [US4] Run eval and verify no regressions. Check that "SQLite driver" query returns ADR-015 above ADR-006.

**Checkpoint**: Reranking improves ordering. Eval passes.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, cleanup.

- [x] T023 Run `make lint` and fix any lint errors across all modified files
- [ ] T024 Run full eval suite: `./adr-insight eval --output testdata/eval/experiments/final.json` and verify all success criteria met (SC-001 through SC-005)
- [ ] T025 Save new baseline: `./adr-insight eval --save-baseline` to capture improved scores as the new baseline for future milestones
- [x] T026 Update `docs/architecture.md` with hybrid search and reranking descriptions
- [x] T027 Update `README.md` ADR table with new ADRs (chunking experiment ADR, ADR-017)

---

## ADR Handling During Implementation

- **NEVER edit the body** of an existing ADR (Context, Decision, Rationale,
  Consequences, Alternatives Considered sections) during task implementation
- If a task requires changing a previously recorded decision, **create a new
  superseding ADR** instead (use the ADR template with the `Supersedes` field)
- When a new ADR supersedes an old one, you MAY update the old ADR's
  **Status** (e.g., `Superseded by ADR-NNN`) and **Related ADRs** fields only
- This preserves decision history — see Constitution Principle II-A

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — capture baseline
- **Foundational (Phase 2)**: No dependency on Setup — adds FTS5 infrastructure
- **US2 Chunking (Phase 3)**: Depends on Setup (baseline) — should run before hybrid search to establish best chunks
- **US1 Hybrid Search (Phase 4)**: Depends on Foundational (FTS5) + US2 (best chunks in place)
- **US3 CLI Search (Phase 5)**: Depends on US1 (HybridSearch exists)
- **US4 Reranking (Phase 6)**: Depends on US1 (pipeline uses hybrid search)
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

- **US2 (P1)**: Depends on Phase 1 only — chunking experiments use existing vector search
- **US1 (P1)**: Depends on Phase 2 + US2 — hybrid search over the best chunking
- **US3 (P2)**: Depends on US1 — CLI uses HybridSearch
- **US4 (P3)**: Depends on US1 — reranking applied to hybrid results

### Parallel Opportunities

- T004 and T005 can run sequentially (same file, same transaction logic)
- T006 and T007 are sequential (implementation then interface)
- T009 and T010 can run in parallel (different chunking methods, same file but separate methods)
- T020 can run in parallel with T014-T017 (different files: rerank.go vs sqlite.go/rag.go)

---

## Implementation Strategy

### MVP First (US2 Chunking + US1 Hybrid Search)

1. Complete Phase 1: Capture baseline (T001-T002)
2. Complete Phase 2: FTS5 infrastructure + skip-judge flag (T003-T008)
3. Complete Phase 3: Chunking experiments (T009-T013)
4. Complete Phase 4: Hybrid search (T014-T017)
5. **STOP and VALIDATE**: Eval shows recall improvement ≥ 0.15

### Incremental Delivery

1. Baseline captured → measurement framework ready
2. FTS5 in store → keyword search available
3. Chunking experiments → best strategy selected, ADR'd
4. Hybrid search → measurable recall improvement
5. CLI search updated → keyword queries work from CLI
6. Reranking → ordering polish (active over superseded)
7. Polish → new baseline saved, docs updated

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Chunking experiments (US2) run BEFORE hybrid search (US1) so that hybrid search operates on the best chunks
- Each change is measured against the M4a baseline via the eval harness
- The chunking ADR is created during T012 — it supersedes ADR-008
- All retrieval changes go through the existing Store interface
