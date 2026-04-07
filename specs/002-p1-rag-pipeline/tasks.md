# Tasks: RAG Pipeline and HTTP API

**Input**: Design documents from `/specs/002-p1-rag-pipeline/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/http-api.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add new dependencies and create directory structure for M2 packages.

- [X] T001 Add anthropic-sdk-go dependency with `go get github.com/anthropics/anthropic-sdk-go` in go.mod
- [X] T002 Create directory structure: internal/llm/, internal/rag/, internal/server/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define interfaces and shared types that all user stories depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 Define LLM interface and response types in internal/llm/llm.go — LLM interface with method: Synthesize(ctx context.Context, query string, adrContents []ADRContext) (QueryResponse, error); ADRContext struct with fields: Number int, Title string, Content string; QueryResponse struct with fields: Answer string, Citations []Citation; Citation struct with fields: ADRNumber int, Title string, Section string
- [X] T004 Add ListADRs method to Store interface in internal/store/store.go — ListADRs(ctx context.Context) ([]ADRSummary, error); ADRSummary struct with fields: Number int, Title string, Status string, Path string
- [X] T005 Implement ListADRs in SQLite store in internal/store/sqlite.go — query chunks table for distinct ADR metadata: SELECT DISTINCT adr_number, adr_title, adr_status, adr_path FROM chunks ORDER BY adr_number
- [X] T006 Write test for ListADRs in internal/store/sqlite_test.go — insert sample chunks, verify ListADRs returns correct unique ADR summaries

**Checkpoint**: Interfaces defined, Store extended. User story implementation can begin.

---

## Phase 3: User Story 1 — Ask a Question and Get a Synthesized Answer (Priority: P1) 🎯 MVP

**Goal**: POST /query accepts a question, retrieves relevant ADRs, synthesizes an answer with citations via Claude.

**Independent Test**: Index sample ADRs, POST a question, verify response contains answer + citations referencing correct ADRs.

### Implementation for User Story 1

- [X] T007 [US1] Implement Anthropic LLM client in internal/llm/anthropic.go — constructor takes API key and model name; Synthesize() builds system prompt instructing Claude to answer using provided ADR content and return structured JSON; use OutputConfig with JSONOutputFormatParam defining schema: {answer: string, citations: [{adr_number: int, title: string, section: string}]}; parse response into QueryResponse
- [X] T008 [US1] Write tests for Anthropic LLM client in internal/llm/anthropic_test.go — test construction and interface compliance; test with mock HTTP server returning canned JSON response; integration test (skipped without ANTHROPIC_API_KEY) that sends a real query
- [X] T009 [US1] Implement RAG pipeline in internal/rag/rag.go — Pipeline struct holds Embedder, Store, LLM, and ADRDir string; Query(ctx, question) method: (1) embed question with "search_query:" prefix, (2) search store for top-K chunks, (3) deduplicate results by ADR number, (4) read full ADR files from disk using ADRPath with os.ReadFile, (5) build []ADRContext, (6) call LLM.Synthesize, (7) return QueryResponse
- [X] T010 [US1] Write tests for RAG pipeline in internal/rag/rag_test.go — test with mock Embedder, Store, LLM; verify deduplication (two chunks from same ADR produce one ADRContext); verify full file read; verify error handling for missing ADR file
- [X] T011 [US1] Implement HTTP server setup and query handler in internal/server/server.go and internal/server/handlers.go — Server struct holds RAG pipeline, Store, ADRDir, and port; NewServeMux() registers routes: POST /query, GET /adrs, GET /adrs/{number}; query handler: decode QueryRequest JSON, validate non-empty query, call pipeline.Query(), encode QueryResponse as JSON; return 400 for empty query, 500 for service errors
- [X] T012 [US1] Write handler tests for POST /query in internal/server/handlers_test.go — use httptest; test valid query returns 200 with answer+citations; test empty query returns 400; test LLM error returns 500 with error message; test Embedder error returns 500 with error message
- [X] T013 [US1] Add `serve` command to CLI in cmd/adr-insight/main.go — parse flags: --port (default 8080), --db, --ollama-url, --adr-dir, --model (default claude-sonnet-4-5); read ANTHROPIC_API_KEY from env and exit with clear error message if missing; wire Embedder, Store, LLM, RAG pipeline, Server; start HTTP server with log output

**Checkpoint**: `curl -X POST localhost:8080/query -d '{"query":"Why Go?"}' | jq .` returns synthesized answer with citations.

---

## Phase 4: User Story 2 — Browse Indexed ADRs via API (Priority: P2)

**Goal**: GET /adrs returns all indexed ADRs with metadata.

**Independent Test**: Index sample ADRs, GET /adrs, verify all ADRs listed with correct metadata.

### Implementation for User Story 2

- [X] T014 [US2] Implement GET /adrs handler in internal/server/handlers.go — call Store.ListADRs(), encode as JSON with {"adrs": [...]} wrapper; return empty array for no results
- [X] T015 [US2] Write handler test for GET /adrs in internal/server/handlers_test.go — test with populated store returns all ADRs; test with empty store returns empty array

**Checkpoint**: `curl localhost:8080/adrs | jq .` returns list of all indexed ADRs.

---

## Phase 5: User Story 3 — Retrieve a Single ADR's Full Content (Priority: P3)

**Goal**: GET /adrs/{number} returns full content of a specific ADR.

**Independent Test**: Index sample ADRs, GET /adrs/1, verify full content returned.

### Implementation for User Story 3

- [X] T016 [US3] Implement GET /adrs/{number} handler in internal/server/handlers.go — parse ADR number from path, find ADR in store via ListADRs (filter by number), read full file from ADRPath with os.ReadFile, extract date from file content by scanning for "Date:" frontmatter line (reuse extractMetadata pattern from parser or add a ParseFile method to the Parser interface), encode as JSON with number/title/status/date/content; return 404 for unknown ADR number, 400 for invalid number format
- [X] T017 [US3] Write handler test for GET /adrs/{number} in internal/server/handlers_test.go — test valid ADR returns 200 with full content; test nonexistent ADR returns 404; test invalid number returns 400

**Checkpoint**: `curl localhost:8080/adrs/1 | jq .` returns full ADR-001 content.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup.

- [X] T018 Verify end-to-end quickstart scenarios from specs/002-p1-rag-pipeline/quickstart.md
- [X] T019 [P] Ensure all test files pass with `go test -short ./...`
- [X] T020 [P] Run linter and fix any issues
- [X] T021 Update docs/architecture.md with new packages (llm, rag, server) and HTTP API details

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 Query (Phase 3)**: Depends on Foundational — this is the MVP
- **US2 List ADRs (Phase 4)**: Depends on Foundational — can run in parallel with US1
- **US3 Get ADR (Phase 5)**: Depends on Foundational — can run in parallel with US1
- **Polish (Phase 6)**: Depends on all user stories complete

### Within User Story 1

- T007 (LLM client) and T009 (RAG pipeline) are sequential — RAG depends on LLM interface
- T008 (LLM tests) follows T007
- T010 (RAG tests) follows T009
- T011 (server + handlers) depends on T009 (RAG pipeline)
- T012 (handler tests) follows T011
- T013 (CLI serve command) depends on T011

### Parallel Opportunities

```
After Phase 2 completes:
  ├── US1: T007 → T008 → T009 → T010 → T011 → T012 → T013 (sequential)
  ├── US2: T014, T015 can run in parallel with US1 (different handler file sections)
  └── US3: T016, T017 can run in parallel with US1 (different handler file sections)
```

Note: US2 and US3 handlers are in the same file as US1's query handler, but they're
independent functions with no shared state beyond the Server struct.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (dependencies + directories)
2. Complete Phase 2: Foundational (LLM interface + Store extension)
3. Complete Phase 3: User Story 1 (LLM client → RAG pipeline → server → CLI)
4. **STOP and VALIDATE**: `curl POST /query` works end-to-end
5. Continue to remaining stories

### Incremental Delivery

1. Setup + Foundational → Interfaces ready
2. Add US1 (Query) → Can ask questions and get synthesized answers (MVP)
3. Add US2 (List ADRs) → Can browse indexed ADRs
4. Add US3 (Get ADR) → Can drill into individual ADRs from citations
5. Polish → All quickstart scenarios verified, lint clean, docs updated
