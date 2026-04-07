# Tasks: ADR Intelligence

**Input**: Design documents from `/specs/006-p2-adr-intelligence/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: New types and store infrastructure that all stories depend on.

- [x] T001 Add `RawRelationship` struct and `RelatedADRs []RawRelationship` field to `ADR` struct in `internal/parser/parser.go`
- [x] T002 Add `ADRRelationship` struct and relationship type constants to `internal/store/store.go`
- [x] T003 Add `StoreRelationships`, `GetRelationships`, `GetAllRelationships` to `Store` interface in `internal/store/store.go`
- [x] T004 Implement `adr_relationships` table creation in `Reset()` in `internal/store/sqlite.go` — include `DROP TABLE IF EXISTS adr_relationships`, `CREATE TABLE`, and indexes per data-model.md
- [x] T005 Implement `StoreRelationships`, `GetRelationships`, `GetAllRelationships` in `internal/store/sqlite.go`
- [x] T006 Update mock stores in test files (`internal/rag/rag_test.go`, `internal/reindex/reindex_test.go`, `internal/server/handlers_test.go`) to satisfy updated Store interface with new relationship method stubs

**Checkpoint**: Store schema ready. All tests pass with new interface methods stubbed.

---

## Phase 2: User Story 2 — Relationship Parsing and Storage (Priority: P1)

**Goal**: During reindex, parse Related ADRs sections and classify relationship types via LLM. Store in the relationships table.

**Independent Test**: After reindex, verify ADR-015 → ADR-006 stored as `supersedes` and ADR-006 → ADR-015 stored as `superseded_by`.

### Implementation for User Story 2

- [x] T007 [US2] Implement `ParseRelatedADRs` in `internal/parser/markdown.go` — walk the AST for the "Related ADRs" H2 section, extract bullets, regex-match `ADR-(\d+)` for target number, capture full bullet text as description. Return `[]RawRelationship`.
- [x] T008 [US2] Update `ParseDir` or `parseFile` in `internal/parser/markdown.go` to populate the new `RelatedADRs` field on each ADR by calling `ParseRelatedADRs`.
- [x] T009 [US2] Also extract supersession from the `Status` field — if status contains "Superseded by ADR-NNN", add a `RawRelationship` with that target and "superseded by" description.
- [x] T010 [US2] Add `ClassifyRelationship(ctx, sourceTitle, bulletText string) (string, error)` method to `AnthropicLLM` in `internal/llm/anthropic.go` — sends the classification prompt to Haiku, returns one of the five relationship type constants.
- [x] T011 [US2] Add `RelationshipClassifier` interface in `internal/reindex/reindex.go` with `ClassifyRelationship` method. Add optional `RelClassifier` field to `Reindexer` struct.
- [x] T012 [US2] Add relationship extraction step to `Reindexer.Run()` in `internal/reindex/reindex.go` — after storing chunks, iterate all ADRs, for each `RawRelationship`, classify via LLM, build `[]ADRRelationship`, call `Store.StoreRelationships`. Handle missing/malformed references gracefully (log warning, skip).
- [x] T013 [US2] Wire the `RelClassifier` in `cmdReindex` in `cmd/adr-insight/main.go` — when `ANTHROPIC_API_KEY` is set, pass the `AnthropicLLM` as classifier.
- [x] T014 [US2] Write a test in `internal/store/sqlite_test.go` (or `internal/reindex/reindex_test.go`) that reindexes a small set of test ADR fixtures with known relationships, then queries `GetRelationships` and verifies: ADR-015 supersedes ADR-006, ADR-006 superseded_by ADR-015, ADR-004 drives ADR-014.

**Checkpoint**: `reindex` populates relationships table. Querying relationships returns correct types.

---

## Phase 3: User Story 1 — Relationship-Aware Answers (Priority: P1)

**Goal**: Improve synthesis by expanding retrieval with related ADRs and providing relationship context to the LLM.

**Independent Test**: Ask "What happened with the SQLite driver decision?" and verify the answer traces ADR-004 → ADR-006 → ADR-015.

### Implementation for User Story 1

- [ ] T015 [US1] Implement relationship expansion in `Pipeline.Query()` in `internal/rag/rag.go` — after reranking and dedup, load relationships for retrieved ADRs via `Store.GetRelationships`, add 1-hop related ADRs not already present. For supersession chains, walk the full chain with visited-node tracking to prevent loops. Cap at `topK * 2`.
- [ ] T016 [US1] Build relationship summary block in `Pipeline.Query()` — prepend a "## Relationship Context" section to the LLM prompt listing all relationships among the retrieved+expanded ADRs (e.g., "ADR-015 supersedes ADR-006").
- [ ] T017 [US1] Add eval test case for "What happened with the SQLite driver decision?" in `testdata/eval/cases.json` — expected ADRs: [4, 6, 15], key facts about the supersession chain.
- [ ] T018 [US1] Add eval test case for "What decisions were affected by the SQLite choice?" in `testdata/eval/cases.json` — expected ADRs: [6, 14, 15, 17].
- [ ] T019 [US1] Reindex, run eval, verify new test cases pass and no regressions on existing cases.

**Checkpoint**: Relationship-spanning questions produce answers that trace decision chains.

---

## Phase 4: User Story 3 — Relationship Display in UI (Priority: P2)

**Goal**: Show related ADRs as clickable links in the ADR detail panel.

**Independent Test**: Open ADR-015, verify "Supersedes: ADR-006" and related links are clickable.

### Implementation for User Story 3

- [ ] T020 [US3] Add `relationships` field to `adrDetailResponse` in `internal/server/handlers.go` — include target ADR number, title, relationship type, and description.
- [ ] T021 [US3] Update `handleGetADR` in `internal/server/handlers.go` — call `Store.GetRelationships` for the requested ADR, resolve target titles via `Store.ListADRs`, populate the relationships field.
- [ ] T022 [US3] Update `web/static/app.js` — in the ADR detail rendering, if relationships exist, display them grouped by type (Supersedes, Superseded By, Depends On, Drives, Related) as clickable links that navigate to the target ADR.
- [ ] T023 [US3] Add minimal styling for relationship links in `web/static/style.css`.
- [ ] T024 [US3] Manually verify: ADR-015 shows links to ADR-006 (supersedes), ADR-004, ADR-014. ADR-001 shows no relationship section (it has none). Clicking ADR-006 link from ADR-015 shows reciprocal "Superseded By: ADR-015".

**Checkpoint**: ADR detail views show navigable relationship links.

---

## Phase 5: User Story 4 — Status-Aware Retrieval Boost (Priority: P3)

**Goal**: Replace content-heuristic supersession penalty in reranker with authoritative relationship data.

**Independent Test**: Search "SQLite driver" — ADR-015 ranks above ADR-006 based on relationship data.

### Implementation for User Story 4

- [ ] T025 [US4] Update `DefaultReranker` in `internal/rag/rerank.go` — add a `Relationships map[int][]store.ADRRelationship` field. Replace the content-based "superseded"/"deprecated" string scan with a lookup: if the ADR has a `superseded_by` relationship, apply the penalty.
- [ ] T026 [US4] Load relationships into the reranker at pipeline startup — in `Pipeline.Query()` or at `Pipeline` construction, call `Store.GetAllRelationships()` and pass to the reranker.
- [ ] T027 [US4] Verify: search "SQLite driver" returns ADR-015 above ADR-006.

**Checkpoint**: Reranker uses authoritative relationship data.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, format convention, final validation.

- [ ] T028 Run `make lint` and fix any lint errors across all modified files
- [ ] T029 Run full eval suite and verify all success criteria met (SC-001 through SC-005)
- [ ] T030 Save new baseline: `./adr-insight eval --save-baseline`
- [ ] T031 Update ADR template at `.specify/templates/adr-template.md` — add standardized relationship type tags to the Related ADRs section example: `- ADR-NNN: Title [type] — description`
- [ ] T032 Update `CLAUDE.md` project instructions to document the relationship tag convention for new ADRs
- [ ] T033 Update `docs/architecture.md` with relationship parsing and graph traversal descriptions
- [ ] T034 Update `README.md` ADR table with ADR-018, ADR-019, and any new ADRs from this milestone

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — adds types and store infrastructure
- **US2 Parsing (Phase 2)**: Depends on Setup — needs store methods to exist
- **US1 Answers (Phase 3)**: Depends on US2 — needs relationships in the store
- **US3 UI (Phase 4)**: Depends on US2 — needs relationships in the store and API
- **US4 Reranker (Phase 5)**: Depends on US2 — needs relationships in the store
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US2 (P1)**: Depends on Phase 1 only — foundational
- **US1 (P1)**: Depends on US2 — needs parsed relationships for expansion
- **US3 (P2)**: Depends on US2 only — can run in parallel with US1
- **US4 (P3)**: Depends on US2 only — can run in parallel with US1 and US3

### Parallel Opportunities

- T001, T002, T003 can run in sequence (same files, dependent)
- T004, T005 are sequential (same file)
- T007, T010 can run in parallel (different files: parser vs llm)
- US1 (Phase 3), US3 (Phase 4), US4 (Phase 5) can all start after US2 completes
- T020, T025 can run in parallel (different files: handlers vs rerank)

---

## Implementation Strategy

### MVP First (US2 + US1)

1. Complete Phase 1: Setup (T001-T006)
2. Complete Phase 2: US2 — Relationship parsing (T007-T014)
3. Complete Phase 3: US1 — Relationship-aware answers (T015-T019)
4. **STOP and VALIDATE**: Eval shows relationship chain questions answered correctly

### Incremental Delivery

1. Setup → store infrastructure ready
2. US2 → relationships parsed and stored during reindex
3. US1 → answers trace decision chains → **core value delivered**
4. US3 → UI shows relationship links → **visible to users**
5. US4 → reranker uses authoritative data → **retrieval polish**
6. Polish → baseline saved, docs updated

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

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- LLM classification during reindex requires `ANTHROPIC_API_KEY`
- Graph traversal tracks visited nodes to prevent loops (clarification from spec)
- Relationship expansion capped at `topK * 2` to bound LLM context size
