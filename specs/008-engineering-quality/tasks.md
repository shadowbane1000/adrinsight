# Tasks: Engineering Quality

**Input**: Design documents from `/specs/008-engineering-quality/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/health.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create the config package and slog infrastructure that all stories depend on.

- [x] T001 Create `internal/config/config.go` — define `Config` struct with all fields from data-model.md (Port, DBPath, ADRDir, OllamaURL, AnthropicKey, EmbedModel, LogLevel, LogFormat, SlowRequestThreshold, ShutdownTimeout). Add `Load()` function that reads from environment variables with defaults. Validate required values and return error for invalid input.
- [x] T002 Create slog setup function in `internal/config/config.go` (or separate `logging.go`) — `SetupLogger(cfg Config) *slog.Logger` that creates a `slog.NewJSONHandler` or `slog.NewTextHandler` based on `cfg.LogFormat`, sets the level from `cfg.LogLevel`, and calls `slog.SetDefault()`. Return the logger for explicit passing.
- [x] T003 Update `cmd/adr-insight/main.go` — load config via `config.Load()` at startup, call `config.SetupLogger()`, replace all flag definitions with config struct fields. Pass config to subcommands (serve, reindex, eval, search). Keep `log.Fatal` only for unrecoverable startup failures.

**Checkpoint**: `go build` succeeds. Config loads from env vars. `LOG_FORMAT=text LOG_LEVEL=debug ./adr-insight serve --dev` produces text logs.

---

## Phase 2: Foundational — Middleware and Error Handling

**Purpose**: Request ID, timing, and logging middleware that all HTTP stories depend on. Error handling audit.

**CRITICAL**: Must be complete before story-specific work begins.

- [x] T004 Create `internal/server/middleware.go` — implement `requestIDMiddleware` that generates UUID v4 (or uses client-provided `X-Request-ID`), stores in `context.Context`, and sets `X-Request-ID` response header. Use `google/uuid` (already a transitive dependency).
- [x] T005 Implement `loggingMiddleware` in `internal/server/middleware.go` — logs method, path, status code, duration_ms, and request_id for every request. Log at warn level if duration exceeds `SlowRequestThreshold`. Log at info level otherwise.
- [x] T006 Wire middleware into `Server.NewServeMux()` in `internal/server/server.go` — apply requestID and logging middleware to all routes. Pass `*slog.Logger` and config to the server struct.
- [x] T007 Audit and fix error wrapping in `internal/rag/rag.go` — replace all bare `fmt.Errorf` (without `%w`) with wrapped errors. Ensure every error return includes context about what operation failed.
- [x] T008 [P] Audit and fix error wrapping in `internal/store/sqlite.go` — replace bare `fmt.Errorf` with `%w`, add context to error messages (which table, which operation).
- [x] T009 [P] Audit and fix error wrapping in `internal/llm/anthropic.go` — wrap errors with context (which API call, what input size).
- [x] T010 [P] Audit and fix error wrapping in `internal/reindex/reindex.go` — wrap errors with context (which ADR, which step).
- [x] T011 [P] Audit and fix error wrapping in `internal/embedder/ollama.go` — wrap errors with context (which model, what input).

**Checkpoint**: All HTTP requests produce structured log entries with request_id and duration. `grep -rn 'fmt.Errorf' internal/ | grep -v '%w'` returns zero results.

---

## Phase 3: User Story 1 — Structured Logging (Priority: P1)

**Goal**: Replace all unstructured log calls with slog throughout the codebase.

**Independent Test**: Run the server, submit a query, verify all log output is structured JSON with level, message, and contextual fields. `grep -rn 'log.Printf\|log.Println' internal/ cmd/` returns zero results (excluding `log.Fatal` at startup).

### Implementation for User Story 1

- [x] T012 [US1] Replace `log.Printf`/`log.Println` with `slog.Info`/`slog.Debug` in `internal/reindex/reindex.go` — add structured fields: adr_count, adr_dir, step name, keyword_count, relationship_count.
- [x] T013 [P] [US1] Replace `log.Printf`/`log.Println` with slog calls in `internal/rag/rag.go` — add structured fields: query, result_count, expansion_count.
- [x] T014 [P] [US1] Replace `log.Printf`/`log.Println` with slog calls in `internal/store/sqlite.go` — add structured fields: table, operation, row_count.
- [x] T015 [P] [US1] Replace `log.Printf`/`log.Println` with slog calls in `internal/llm/anthropic.go` — add structured fields: model, prompt_length, response_length.
- [x] T016 [P] [US1] Replace `log.Printf`/`log.Println` with slog calls in `internal/server/server.go` — add structured fields: port, dev_mode.
- [x] T017 [P] [US1] Replace `log.Printf`/`log.Println` with slog calls in `internal/embedder/ollama.go` — add structured fields: model, input_length.
- [x] T018 [US1] Replace `log.Fatalf` with `slog.Error` + `os.Exit(1)` in `cmd/adr-insight/main.go` for startup failures. Replace `log.Printf` with slog calls for non-fatal messages. Keep fatal exits only for truly unrecoverable errors (missing DB, config parse failure).
- [x] T019 [US1] Verify: `grep -rn 'log\.Printf\|log\.Println\|log\.Fatal' internal/` returns zero results. Only `cmd/` may retain `log.Fatal` for startup-only failures.

**Checkpoint**: All logging is structured. `LOG_FORMAT=json ./adr-insight serve` produces JSON logs parseable by `jq`.

---

## Phase 4: User Story 2 — Centralized Configuration (Priority: P1)

**Goal**: All settings configurable via env vars, documented in README.

**Independent Test**: `PORT=9090 ./adr-insight serve` starts on port 9090. `LOG_LEVEL=warn` suppresses info messages. README has a complete config table.

### Implementation for User Story 2

- [x] T020 [US2] Update `cmd/adr-insight/main.go` cmdServe — use `cfg.Port`, `cfg.DBPath`, `cfg.ADRDir`, `cfg.OllamaURL`, `cfg.AnthropicKey`, `cfg.EmbedModel` from the config struct instead of local flag parsing and `os.Getenv` calls.
- [x] T021 [US2] Update `cmd/adr-insight/main.go` cmdReindex — use config struct fields instead of local flags for db-path, adr-dir, ollama-url, embed-model.
- [x] T022 [US2] Update `cmd/adr-insight/main.go` cmdEval and cmdSearch — use config struct fields instead of local flags.
- [x] T023 [US2] Add configuration table to `README.md` — document all env vars, defaults, and descriptions per data-model.md.
- [x] T024 [US2] Verify: `grep -rn 'os.Getenv' internal/ cmd/ | grep -v config/` returns zero results (all env var access centralized in config package).

**Checkpoint**: All settings flow through config struct. README documents all options.

---

## Phase 5: User Story 5 — Error Handling Audit (Priority: P1)

**Goal**: Consistent wrapped errors with context throughout.

**Independent Test**: Run with invalid DB path — error message includes the path and what failed. Trace any HTTP error response back through the wrapped chain.

### Implementation for User Story 5

- [x] T025 [US5] Audit `internal/rag/rag.go` — verify all errors use `%w` wrapping with operation context. Fix any remaining bare strings.
- [x] T026 [P] [US5] Audit `internal/server/handlers.go` — verify user-facing error responses are friendly (no internal details). Ensure errors from Store/Pipeline are logged but not exposed to clients.
- [x] T027 [P] [US5] Audit `internal/eval/eval.go` — wrap errors with test case context (case ID, question).
- [x] T028 [US5] Audit `cmd/adr-insight/main.go` — verify startup errors are actionable (tell the user what to check, not just "failed").

**Checkpoint**: All `fmt.Errorf` calls use `%w`. User-facing errors are friendly. Error chain is traceable.

---

## Phase 6: User Story 3 — Health Check (Priority: P2)

**Goal**: `/health` endpoint with per-component status checks.

**Independent Test**: `curl http://localhost:8081/health` returns structured JSON. Stop Ollama — health shows degraded.

### Implementation for User Story 3

- [x] T029 [US3] Implement `handleHealth` in `internal/server/handlers.go` — check database (SELECT 1), Ollama (HTTP HEAD with 3s timeout), Anthropic key (env var presence). Return structured JSON per contracts/health.md. Use 200 for healthy/degraded, 503 for unhealthy.
- [x] T030 [US3] Register `GET /health` route in `internal/server/server.go`.
- [x] T031 [US3] Add health check test in `internal/server/handlers_test.go` — test healthy, degraded (mock Ollama unreachable), and unhealthy (mock DB unreachable) scenarios.
- [x] T032 [US3] Update `docker-compose.yml` — add health check configuration using `/health` endpoint.

**Checkpoint**: Health check works. Docker health check configured.

---

## Phase 7: User Story 4 — Graceful Shutdown (Priority: P2)

**Goal**: Clean shutdown on SIGTERM with request draining.

**Independent Test**: Start server, send SIGTERM — server logs shutdown, drains requests, exits 0.

### Implementation for User Story 4

- [x] T033 [US4] Refactor `cmd/adr-insight/main.go` cmdServe — use `signal.NotifyContext` for SIGTERM/SIGINT. Create `http.Server` struct explicitly (not `http.ListenAndServe`). Call `server.Shutdown(ctx)` with `cfg.ShutdownTimeout` on signal.
- [x] T034 [US4] Add shutdown logging — log "shutting down" with reason (signal received), log "shutdown complete" with duration. Defer `store.Close()` for database cleanup.
- [x] T035 [US4] Remove `log.Fatal` calls in normal operation paths — replace with error returns or `slog.Error` + `os.Exit(1)` only at top level.

**Checkpoint**: `kill -TERM <pid>` triggers clean shutdown. Exit code 0. Shutdown logged.

---

## Phase 8: User Story 6 — Integration Tests (Priority: P2)

**Goal**: Tests exercise the full pipeline with real SQLite, no external services.

**Independent Test**: `go test -tags fts5 ./internal/server/ -run TestIntegration` passes without Ollama or Anthropic.

### Implementation for User Story 6

- [x] T036 [US6] Create `internal/server/integration_test.go` — test helper that sets up real SQLite (in-memory), mock embedder with deterministic vectors (hash-based), real parser against `testdata/` fixtures. Wire up a full Server with real Store and mock Pipeline.
- [x] T037 [US6] Add integration test: parse test fixtures → store chunks → GET /adrs returns correct list with titles, statuses, dates.
- [x] T038 [US6] Add integration test: store chunks → GET /adrs/{number} returns correct ADR content with relationships.
- [x] T039 [US6] Add integration test: GET /health returns healthy when DB is accessible.

**Checkpoint**: Integration tests pass in CI. Real SQLite + real parser + mock embedder. No Ollama/Anthropic required.

---

## Phase 9: User Story 8 — Go Idiom Review (Priority: P1)

**Goal**: Codebase passes idiom review — no patterns a Go reviewer would flag.

**Independent Test**: Code reviewer finds no non-idiomatic patterns.

### Implementation for User Story 8

- [x] T040 [US8] Review and fix variable naming across all packages — short names in small scopes, descriptive names for exports. Fix any single-letter names used in large scopes or unexported globals with unclear names.
- [x] T041 [P] [US8] Review and fix comment style — add package comments to all packages missing them. Add doc comments to exported functions where the purpose is non-obvious. Remove stale/misleading comments.
- [x] T042 [P] [US8] Review package boundaries — verify no circular dependencies, clean imports, no package doing too much. Check that `internal/config/` doesn't import other internal packages.

**Checkpoint**: `golangci-lint` passes. Code is idiomatic.

---

## Phase 10: User Story 7 — API Documentation (Priority: P3)

**Goal**: All HTTP endpoints documented with examples.

**Independent Test**: Developer can use the API from docs alone, without reading source.

### Implementation for User Story 7

- [x] T043 [US7] Create `docs/api.md` — document `POST /query` (request body, response body, error responses, status codes, example curl).
- [x] T044 [P] [US7] Document `GET /adrs` and `GET /adrs/{number}` in `docs/api.md` — response format, fields, status codes, examples.
- [x] T045 [P] [US7] Document `GET /health` in `docs/api.md` — response format, status values, component checks, example.

**Checkpoint**: All 4 endpoints documented with request/response examples.

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, docs, cleanup.

- [x] T046 Run `make lint` and fix any lint errors across all modified files
- [x] T047 Run full test suite `go test -tags fts5 ./...` and verify all pass
- [x] T048 Update `README.md` ADR table with ADR-024
- [x] T049 Update `docs/architecture.md` — add structured logging, config management, health check, graceful shutdown to Technical Notes
- [x] T050 Manual verification of all 8 quickstart.md scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — config and slog infrastructure
- **Foundational (Phase 2)**: Depends on Setup — middleware needs config and slog
- **US1 Logging (Phase 3)**: Depends on Setup — needs slog infrastructure
- **US2 Config (Phase 4)**: Depends on Setup — needs config struct
- **US5 Errors (Phase 5)**: Depends on Setup — needs slog for error logging
- **US3 Health (Phase 6)**: Depends on Foundational — needs middleware for request ID
- **US4 Shutdown (Phase 7)**: Depends on Setup — needs config for timeout
- **US6 Integration (Phase 8)**: Depends on US1 + US3 — needs health endpoint and slog in place
- **US8 Idioms (Phase 9)**: Depends on US1 + US5 — review after logging and error fixes are in
- **US7 API Docs (Phase 10)**: Depends on US3 — needs health endpoint to document
- **Polish (Phase 11)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 1 only
- **US2 (P1)**: Depends on Phase 1 only — can run in parallel with US1
- **US5 (P1)**: Depends on Phase 1 only — can run in parallel with US1 and US2
- **US3 (P2)**: Depends on Phase 2
- **US4 (P2)**: Depends on Phase 1 only
- **US6 (P2)**: Depends on US1 + US3
- **US8 (P1)**: Depends on US1 + US5
- **US7 (P3)**: Depends on US3

### Parallel Opportunities

- After Phase 1: US1 (T012-T019), US2 (T020-T024), and US5 (T025-T028) can run in parallel
- After Phase 2: US3 (T029-T032) and US4 (T033-T035) can run in parallel
- Within US1: T013-T017 can all run in parallel (different files)
- Within US5: T026-T027 can run in parallel
- Within US7: T043-T045 can run in parallel

---

## Implementation Strategy

### MVP First (Setup + Foundational + US1 + US2 + US5)

1. Complete Phase 1: Setup — config struct, slog setup (T001-T003)
2. Complete Phase 2: Middleware, error audit (T004-T011)
3. Complete Phase 3: US1 — Replace all unstructured logging (T012-T019)
4. Complete Phase 4: US2 — Centralize configuration (T020-T024)
5. Complete Phase 5: US5 — Error handling audit (T025-T028)
6. **STOP and VALIDATE**: Zero unstructured logs, config centralized, errors wrapped

### Incremental Delivery

1. Setup + Foundational → structured logging and middleware working
2. US1 → all logs structured → **observability foundation**
3. US2 → all config centralized → **operator-friendly**
4. US5 → all errors wrapped → **debuggable**
5. US3 → health check → **operational readiness**
6. US4 → graceful shutdown → **production-ready**
7. US6 → integration tests → **confidence**
8. US8 → idiom review → **code quality**
9. US7 → API docs → **documentation**
10. Polish → final sweep

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
- `google/uuid` is already a transitive dependency — no new deps needed for request IDs
- `log/slog` is stdlib since Go 1.21 — no new deps for logging
- Integration tests use in-memory SQLite — fast, no cleanup, CI-safe
- Error audit tasks (Phase 2 + US5) overlap by design — Phase 2 does the initial fix, US5 does the thorough audit
