# Tasks: Foundation — ADR Indexing Pipeline

**Input**: Design documents from `/specs/001-p1-foundation/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize Go module, project structure, and shared configuration.

- [X] T001 Initialize Go module with `go mod init github.com/tylerc-atx/adr-insight` in go.mod
- [X] T002 Create directory structure: cmd/adr-insight/, internal/parser/, internal/embedder/, internal/store/, internal/reindex/, testdata/, web/static/
- [X] T003 [P] Create .gitignore with Go patterns (*.exe, *.test, *.out, vendor/, adr-insight.db)
- [X] T004 [P] Create .golangci.yml with standard Go linter configuration in .golangci.yml
- [X] T005 Add dependencies: goldmark, goldmark-meta, ncruces/go-sqlite3, sqlite-vec-go-bindings/ncruces, ollama/ollama in go.mod

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define interfaces and shared types that all user stories depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T006 Define ADR struct and Parser interface in internal/parser/parser.go — ADR struct with fields: FilePath, Number, Title, Date, Status, Deciders, Body; Chunk struct with fields: ADRNumber, SectionKey, Content; Parser interface with method: ParseDir(dir string) ([]ADR, error)
- [X] T007 Define Embedder interface in internal/embedder/embedder.go — Embedder interface with methods: Embed(ctx context.Context, texts []string) ([][]float32, error)
- [X] T008 Define Store interface in internal/store/store.go — Store interface with methods: Reset(ctx context.Context) error, StoreChunks(ctx context.Context, chunks []ChunkRecord) error, Search(ctx context.Context, query []float32, topK int) ([]SearchResult, error), Close() error; ChunkRecord struct with fields: ADRNumber, ADRTitle, ADRStatus, ADRPath, Section, Content, Embedding []float32; SearchResult struct with fields: ADRNumber, ADRTitle, ADRPath, Section, Content, Score float64
- [X] T009 [P] Create test fixture testdata/ADR-001-sample.md — valid ADR with title, date, status, deciders, and multiple H2 sections (Context, Decision, Rationale, Consequences)
- [X] T010 [P] Create test fixture testdata/ADR-002-sample.md — valid ADR with different topic for search differentiation
- [X] T011 [P] Create test fixture testdata/not-an-adr.md — a markdown file that does not match ADR-*.md pattern
- [X] T012 [P] Create test fixture testdata/ADR-bad-frontmatter.md — ADR file with missing or malformed metadata lines
- [X] T012b [P] Create test fixture testdata/ADR-empty-body.md — ADR file with valid frontmatter but no body content below metadata

**Checkpoint**: Interfaces defined, test fixtures ready. User story implementation can begin.

---

## Phase 3: User Story 1 — Reindex ADRs from a Directory (Priority: P1) 🎯 MVP

**Goal**: Run `adr-insight reindex` to parse ADR files, embed them, and store everything in SQLite.

**Independent Test**: Run reindex against testdata/, verify database contains expected records with valid embeddings.

### Implementation for User Story 1

- [X] T013 [US1] Implement goldmark-based markdown parser in internal/parser/markdown.go — ParseDir walks directory, filters ADR-*.md (case-insensitive), parses each file extracting H1 title, frontmatter metadata (Status/Date/Deciders lines), and splits body by H2 headings into Chunks; return raw content (no embedding prefixes — that's the pipeline's job)
- [X] T014 [US1] Write table-driven tests for parser in internal/parser/markdown_test.go — test cases: valid ADR with multiple sections, ADR with no H2 headings (single chunk), ADR with missing frontmatter (defaults), ADR with empty body (frontmatter only), non-ADR file skipped, empty directory, nonexistent directory returns error, malformed frontmatter; use testdata/ fixtures
- [X] T015 [US1] Implement Ollama embedder in internal/embedder/ollama.go — use official ollama/ollama/api Client.Embed() with model "nomic-embed-text"; accept base URL as constructor parameter; return [][]float32 vectors
- [X] T016 [US1] Write tests for Ollama embedder in internal/embedder/ollama_test.go — test construction and interface compliance; test that unreachable Ollama URL returns meaningful error; integration test (skipped without Ollama) that embeds sample text and verifies 768-dimension vector returned
- [X] T017 [US1] Implement SQLite store in internal/store/sqlite.go — constructor takes db path, auto-registers sqlite-vec via import, creates chunks table and vec_chunks virtual table (float[768]); Reset() drops and recreates tables; StoreChunks() inserts metadata rows and serialized vectors; Search() embeds query match with LIMIT, joins chunks table for metadata; Close() closes db connection
- [X] T018 [US1] Write tests for SQLite store in internal/store/sqlite_test.go — test Reset (tables created), StoreChunks (records inserted), Search (correct ranking), Close; use temporary database file
- [X] T019 [US1] Implement reindex pipeline orchestration in internal/reindex/reindex.go — Reindexer struct holds Parser, Embedder, Store; Run(ctx, adrDir) calls ParseDir, chunks all ADRs, prepends "search_document: " prefix to each chunk before embedding, batches chunks for embedding, stores results; logs progress (parsing N ADRs, embedding N chunks, storing)
- [X] T020 [US1] Write tests for reindex pipeline in internal/reindex/reindex_test.go — test with mock Parser, Embedder, Store implementations to verify orchestration flow
- [X] T021 [US1] Implement CLI entrypoint with reindex command in cmd/adr-insight/main.go — parse flags: --adr-dir (default ./docs/adr), --db (default ./adr-insight.db), --ollama-url (default http://localhost:11434); wire Parser, Embedder, Store into Reindexer; run and report results
- [X] T022 [US1] Add "search_query: " prefix handling to embedder for query-time embedding in internal/embedder/ollama.go — add EmbedQuery method or parameter to distinguish index-time vs query-time prefix

**Checkpoint**: `go run ./cmd/adr-insight reindex --adr-dir ./docs/adr` processes all ADRs and creates adr-insight.db

---

## Phase 4: User Story 2 — Verify Indexed ADRs via Similarity Search (Priority: P2)

**Goal**: Run `adr-insight search <query>` to find relevant ADR chunks by semantic similarity.

**Independent Test**: After reindex, search for "why did we choose Go?" and verify ADR-001 is the top result.

### Implementation for User Story 2

- [X] T023 [US2] Add search command to CLI in cmd/adr-insight/main.go — parse flags: --db, --ollama-url, --top-k (default 5); accept positional query argument; embed query with "search_query: " prefix; call Store.Search(); format and print results
- [X] T024 [US2] Implement search output formatting in cmd/adr-insight/main.go — display results as numbered list: "[ADR-NNN] Title — Section (score: X.XX)" followed by truncated content preview (first 150 chars)

**Checkpoint**: `go run ./cmd/adr-insight search "why did we choose Go?"` returns ADR-001 as top result

---

## Phase 5: User Story 3 — CI Validates Code Quality (Priority: P3)

**Goal**: CI pipeline runs lint, test, and build on every push to both Gitea and GitHub.

**Independent Test**: Push a commit, verify CI runs all three steps.

### Implementation for User Story 3

- [X] T025 [P] [US3] Create Gitea Actions workflow in .gitea/workflows/ci.yaml — trigger on push; steps: checkout, setup Go 1.22+, golangci-lint, go test ./..., go build ./cmd/adr-insight
- [X] T026 [P] [US3] Create GitHub Actions workflow in .github/workflows/ci.yaml — same steps as Gitea workflow, using GitHub Actions syntax

**Checkpoint**: Push triggers CI on both platforms; lint, test, build all pass.

---

## Phase 6: User Story 4 — Run Quality Checks Locally (Priority: P3)

**Goal**: Developer can run `make check` to execute the same lint/test/build steps locally.

**Independent Test**: Run `make check` locally, verify it runs lint, test, and build.

### Implementation for User Story 4

- [X] T027 [US4] Create Makefile with targets: lint (golangci-lint run), test (go test ./...), build (go build -o adr-insight ./cmd/adr-insight), check (lint + test + build in sequence), clean (remove binary and db file)

**Checkpoint**: `make check` runs lint, test, build and reports pass/fail.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup.

- [X] T028 Verify end-to-end quickstart scenarios from specs/001-p1-foundation/quickstart.md
- [X] T029 [P] Ensure all test files pass with `go test ./...`
- [X] T030 [P] Run `golangci-lint run` and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 Reindex (Phase 3)**: Depends on Foundational — this is the MVP
- **US2 Search (Phase 4)**: Depends on US1 (needs reindex to populate db)
- **US3 CI (Phase 5)**: Depends on Setup — can run in parallel with US1
- **US4 Makefile (Phase 6)**: Depends on Setup — can run in parallel with US1
- **Polish (Phase 7)**: Depends on all user stories complete

### Within User Story 1

- T013 (parser) and T015 (embedder) and T017 (store) can run in parallel after Foundational
- T014, T016, T018 (tests) run after their respective implementations
- T019 (reindex orchestration) depends on T013, T015, T017
- T020 (reindex tests) depends on T019
- T021 (CLI) depends on T019
- T022 (query prefix) depends on T015

### Parallel Opportunities

```
After Phase 2 completes:
  ├── US1: T013, T015, T017 can run in parallel (different packages)
  ├── US3: T025, T026 can run in parallel with US1 (different files)
  └── US4: T027 can run in parallel with US1 (different file)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (interfaces + fixtures)
3. Complete Phase 3: User Story 1 (reindex pipeline)
4. **STOP and VALIDATE**: `go run ./cmd/adr-insight reindex` works end-to-end
5. Continue to remaining stories

### Incremental Delivery

1. Setup + Foundational → Project compiles
2. Add US1 (Reindex) → Can index ADRs into database (MVP)
3. Add US2 (Search) → Can verify retrieval quality
4. Add US3 + US4 (CI + Makefile) → Quality infrastructure in place
5. Polish → All quickstart scenarios verified
