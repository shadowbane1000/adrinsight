# Tasks: Abuse Hardening for AI Cost Control

**Input**: Design documents from `/specs/009-abuse-hardening/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Configuration)

**Purpose**: Add configuration fields for rate limiting and query length

- [x] T001 Add RateLimitRequests (default 10), RateLimitWindow (default 60s), and MaxQueryLength (default 500) fields to Config struct and env var loading in `internal/config/config.go`

---

## Phase 2: User Story 1 - Rate Limiting (Priority: P1) 🎯 MVP

**Goal**: Per-IP rate limiting on the query endpoint to prevent API cost abuse

**Independent Test**: Send rapid queries from a single IP; verify excess requests return 429 with Retry-After header while other IPs and non-query endpoints remain unaffected

### Implementation for User Story 1

- [x] T002 [US1] Implement rate limiter with fixed-window counter per IP in `internal/server/middleware.go`: struct with `sync.Mutex`-protected map, `Allow(ip string) (bool, time.Duration)` method, and lazy cleanup of stale entries
- [x] T003 [US1] Implement `clientIP(r *http.Request) string` helper in `internal/server/middleware.go` that extracts real IP from `X-Forwarded-For` header, falling back to `RemoteAddr`
- [x] T004 [US1] Add `rateLimitMiddleware` in `internal/server/middleware.go` that wraps only the query handler: calls `Allow()`, returns 429 with `Retry-After` header and JSON error body if denied, logs rate-limited requests
- [x] T005 [US1] Wire rate limiter into the query route in `internal/server/server.go`: create rate limiter from config, apply middleware to `POST /query` only (not other routes)
- [x] T006 [US1] Handle 429 responses in the query store in `web/static/app.js`: parse `Retry-After` header, display user-friendly message with retry countdown

**Checkpoint**: Rate limiting functional — rapid queries from one IP get 429, other IPs unaffected, non-query endpoints unaffected

---

## Phase 3: User Story 2 - Query Length Limits (Priority: P2)

**Goal**: Reject oversized queries before calling any AI services

**Independent Test**: Submit a query exceeding the configured max length; verify 400 response and confirm no embedding/synthesis log entries

### Implementation for User Story 2

- [x] T007 [US2] Add query length validation at the top of `handleQuery` in `internal/server/handlers.go`: check `len(req.Query)` against configured max, return 400 with descriptive error before any AI service calls
- [x] T008 [US2] Handle 400 query-too-long responses in `web/static/app.js`: display the error message from the response body

**Checkpoint**: Oversized queries rejected with 400, no AI costs incurred

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T009 Update environment variable documentation in `README.md` with RATE_LIMIT_REQUESTS, RATE_LIMIT_WINDOW_S, and MAX_QUERY_LENGTH
- [x] T010 Run quickstart.md verification scenarios (all 5)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **US1 (Phase 2)**: Depends on T001 (config fields)
- **US2 (Phase 3)**: Depends on T001 (config fields); independent of US1
- **Polish (Phase 4)**: Depends on US1 and US2 completion

### User Story Dependencies

- **User Story 1 (P1)**: T002 → T003 → T004 → T005 → T006 (sequential — all touch middleware.go or depend on prior tasks)
- **User Story 2 (P2)**: T007 → T008 (sequential — handler then UI)
- **US1 and US2 can run in parallel** after Phase 1 (different files: middleware.go vs handlers.go, different UI changes)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Config setup
2. Complete Phase 2: Rate limiting
3. **STOP and VALIDATE**: Test with rapid queries
4. Deploy — primary cost protection is in place

### Incremental Delivery

1. Config setup → Rate limiting → Validate → Deploy (MVP)
2. Add query length limits → Validate → Deploy
3. Polish: README, full quickstart validation

---

## ADR Handling During Implementation

No existing ADRs are affected. No new ADRs planned.

## Notes

- All changes are in existing files — no new packages or files
- T002/T003/T004 are sequential because they all modify middleware.go
- US1 and US2 touch different files and can be parallelized
