# Tasks: Web UI and Docker Deployment

**Input**: Design documents from `/specs/003-p1-web-ui-docker/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ui-layout.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the static file directory structure and go:embed plumbing needed by all UI stories.

- [x] T001 Create directory structure: `web/static/` for HTML, CSS, JS files
- [x] T002 Create `web/embed.go` with `//go:embed` directive embedding `web/static/` filesystem (adjusted: embed.go in `web/` package, not `internal/server/`, since go:embed requires files in same directory tree)
- [x] T003 Add static file serving route to `internal/server/server.go` — serve embedded FS at `/` in production mode, serve from disk (`web/static/`) when `DevMode` bool is set on Server struct
- [x] T004 Add `--dev` flag to serve command in `cmd/adr-insight/main.go` that sets `Server.DevMode = true` for disk-based static file serving

**Checkpoint**: `go build` succeeds, serving an empty `web/static/` directory at `http://localhost:8081/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Auto-reindex on startup — MUST be complete before Docker story, and improves all other stories by ensuring the database is always populated.

**⚠️ CRITICAL**: No Docker story work can begin until this phase is complete.

- [x] T005 Add a method to the Store interface and SQLiteStore in `internal/store/store.go` and `internal/store/sqlite.go` that checks whether the chunks table has any rows (e.g., `IsEmpty(ctx) (bool, error)`)
- [x] T006 Add auto-reindex logic to `cmdServe` in `cmd/adr-insight/main.go` — after opening the store, call `IsEmpty`; if true, run `reindex.Reindexer.Run` before starting the HTTP server. Log progress ("Auto-indexing ADRs..." / "Index already populated, skipping reindex"). Handle Ollama not yet ready by retrying with exponential backoff (initial 1s, max 30s total wait). If Ollama is still unavailable after retries, log a warning and start the server anyway (ADR browsing works without embeddings; queries will fail until Ollama is ready).

**Checkpoint**: `./adr-insight serve` with no database auto-indexes ADRs; with existing database, skips reindex.

---

## Phase 3: User Story 1 — Ask a Question in the Browser (Priority: P1) 🎯 MVP

**Goal**: A user opens the web UI, types a question, sees a synthesized answer with clickable citation links.

**Independent Test**: Open `http://localhost:8081`, type "Why did we choose Go?", see answer with [ADR-001] citation link. Click citation to see full ADR.

### Implementation for User Story 1

- [x] T007 [P] [US1] Create `web/static/index.html` — single-page layout per contracts/ui-layout.md: header with title and search bar (text input + Go button), hidden answer area below, lower section with left ADR list panel and right detail panel. Include `<script>` tag for marked.js CDN and `<script src="app.js">`. Include `<link rel="stylesheet" href="style.css">`.
- [x] T008 [P] [US1] Create `web/static/style.css` — layout styles: header bar with search input and button, answer area (hidden by default, expands to fit content), two-column lower section (ADR list left ~250px, detail panel right flex), loading indicator animation, citation link styling, basic typography for rendered markdown.
- [x] T009 [US1] Create `web/static/app.js` — implement query functionality: on form submit (click Go or press Enter), show loading indicator in answer area, POST to `/query` with `{"query": text}`, on response render `answer` field as markdown via `marked.parse()`, render `citations` array as clickable links `[ADR-NNN]`. On citation click, fetch `GET /adrs/{number}` and display full ADR content (markdown rendered) in the right detail panel. Handle errors: show user-friendly message in answer area if API returns error or no relevant results. Handle empty query (don't submit). Disable submit button during loading.
- [x] T010 [US1] Verified same-origin — no CORS headers needed if needed in `internal/server/server.go` — since UI is served from the same origin, verify no CORS issues exist with fetch calls to `/query` and `/adrs`.

**Checkpoint**: Full query → answer → citation click flow works in browser.

---

## Phase 4: User Story 2 — Browse All ADRs (Priority: P2)

**Goal**: ADR list on the left is populated on page load; clicking any ADR shows its full content in the right panel.

**Independent Test**: Open `http://localhost:8081`, see ADR list populated on the left. Click "ADR-001" and see its full markdown-rendered content on the right.

### Implementation for User Story 2

- [x] T011 [US2] Add ADR list loading to `web/static/app.js` — on page load, fetch `GET /adrs`, populate left panel with ADR entries showing number, title, and status. On click, fetch `GET /adrs/{number}` and render full content as markdown in the right panel (replacing current right panel content). Handle empty list: show "No ADRs available" message. Handle API errors gracefully.

**Checkpoint**: Page loads with ADR list; clicking an ADR displays its content. Works independently of query functionality.

---

## Phase 5: User Story 4 — Discover the Project from the Web UI (Priority: P2)

**Goal**: Default right panel shows About content describing the project, with a GitHub link.

**Independent Test**: Open `http://localhost:8081`, right panel shows project description and clickable GitHub link without any user action.

### Implementation for User Story 4

- [x] T012 [US4] Add About panel content to `web/static/app.js` and `web/static/index.html` — the right detail panel defaults to showing: project name (ADR Insight), description (AI-powered search and reasoning over Architecture Decision Records), how it was built (spec-driven development with modified spec-kit including ADR generation), universality (any project with standard-format ADRs can use it), GitHub link (clickable, opens in new tab), and author (Tyler Colbert). Content can be hardcoded HTML or a markdown string rendered via marked.parse(). Clicking an ADR or citation replaces this with ADR content; provide a way to return to About (e.g., clicking the header title).

**Checkpoint**: Page loads with About content visible. GitHub link works. Clicking an ADR replaces About; clicking header returns to About.

---

## Phase 6: User Story 3 — Run Everything with Docker Compose (Priority: P3)

**Goal**: `ANTHROPIC_API_KEY=sk-... docker compose up` brings the full system up — Ollama with model, Go app with auto-reindex, web UI accessible.

**Independent Test**: On a machine with Docker, run `docker compose up` with API key set. Open browser to `http://localhost:8081`, ask a question, get an answer.

### Implementation for User Story 3

- [x] T013 [P] [US3] Create `Dockerfile` — multi-stage build per ADR-014: build stage uses `golang:bookworm`, installs `libsqlite3-dev`, copies source and runs `go build -o /app/adr-insight ./cmd/adr-insight`; runtime stage uses `debian:bookworm-slim`, installs `libsqlite3-0 ca-certificates`, copies binary from build stage. ENTRYPOINT runs `./adr-insight serve`. Expose port 8081.
- [x] T014 [P] [US3] Create `docker/ollama-entrypoint.sh` — entrypoint script that starts Ollama server in background, waits for it to be ready (poll `/api/tags`), runs `ollama pull nomic-embed-text` if model not already present, then kills background process and exec `ollama serve` as PID 1.
- [x] T015 [US3] Create `docker-compose.yml` — two services: `app` (built from Dockerfile, ports 8081:8081, depends_on ollama, environment ANTHROPIC_API_KEY and ADR_DIR=/data/adr, volumes: `./docs/adr:/data/adr:ro` and `adr-data:/data` for database persistence) and `ollama` (image ollama/ollama, ports 11434:11434, entrypoint from `docker/ollama-entrypoint.sh`, volume for model persistence). Define named volumes.
- [ ] T016 [US3] Verify end-to-end Docker flow: ensure `docker compose build` succeeds, `docker compose up` starts both services, Ollama pulls model, app auto-reindexes, web UI is accessible. Fix any networking issues (app must reach ollama by service name). Validate cold-start scenario where Ollama is slow to start — app should retry reindex and eventually succeed without crashing.

**Checkpoint**: `docker compose up` from clean state → web UI queryable. Second `docker compose up` skips model pull and reindex.

---

## Phase 7: User Story 5 — Setup from README (Priority: P3)

**Goal**: README has complete instructions for both local dev and Docker deployment.

**Independent Test**: Follow README from a fresh clone; system running in under 5 minutes.

### Implementation for User Story 5

- [x] T017 [US5] Update `README.md` — complete rewrite with: project description (what ADR Insight does), quick start with Docker (3 steps: clone, set API key, docker compose up), local development setup (prerequisites: Go 1.24+, Ollama, libsqlite3-dev; build, reindex, serve commands; --dev flag for UI iteration), architecture overview (link to docs/architecture.md), link to ADR collection in docs/adr/, project structure summary, author and license.

**Checkpoint**: A developer can follow README instructions to get the system running.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Error handling, edge cases, and validation across all stories.

- [x] T018 Handle edge case in `web/static/app.js`: multiple rapid query submissions — disable the submit button and input while a query is in progress, re-enable on response or error
- [x] T019 Handle edge case in `web/static/style.css`: very long answers — max-height 50vh with overflow scroll — ensure answer area scrolls within a max-height or the page scrolls naturally without breaking layout
- [x] T020 Add `<noscript>` tag to `web/static/index.html` indicating JavaScript is required
- [x] T021 Modify `cmdServe` in `cmd/adr-insight/main.go` to allow starting without ANTHROPIC_API_KEY — remove `log.Fatal`, log a warning instead ("ANTHROPIC_API_KEY not set — query endpoint disabled, ADR browsing available"). Set Pipeline to nil when key is missing. Update `handleQuery` in `internal/server/handlers.go` to return HTTP 503 with `{"error": "ANTHROPIC_API_KEY not configured"}` when Pipeline is nil. ADR list and detail endpoints must work regardless of API key.
- [x] T022 Run `make lint` and fix any lint errors across all modified files
- [ ] T023 Run quickstart.md scenarios to validate end-to-end functionality
- [x] T024 Update `docs/architecture.md` if project structure, interfaces, or dependencies changed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS Docker story (US3)
- **User Stories (Phase 3-7)**: US1/US2/US4 depend on Setup only; US3 depends on Foundational; US5 depends on all other stories being complete
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Phase 1 (Setup) — no other story dependencies
- **User Story 2 (P2)**: Depends on Phase 1 (Setup) — shares app.js with US1 but can be implemented independently
- **User Story 4 (P2)**: Depends on Phase 1 (Setup) — shares app.js with US1/US2 but is independent content
- **User Story 3 (P3)**: Depends on Phase 1 (Setup) + Phase 2 (Foundational) — auto-reindex must work before Docker
- **User Story 5 (P3)**: Depends on all other stories — README documents what's been built

### Within Each User Story

- HTML structure before JavaScript behavior
- CSS can be developed in parallel with HTML
- API integration after UI skeleton exists

### Parallel Opportunities

- T007 and T008 can run in parallel (HTML and CSS are independent files)
- T013 and T014 can run in parallel (Dockerfile and ollama entrypoint are independent)
- US1, US2, and US4 can be developed in parallel after Phase 1 (all work in app.js but on separate concerns)

---

## Parallel Example: User Story 1

```bash
# Launch HTML and CSS in parallel:
Task: "Create web/static/index.html" (T007)
Task: "Create web/static/style.css" (T008)

# Then sequentially:
Task: "Create web/static/app.js with query functionality" (T009)
Task: "Verify CORS/same-origin" (T010)
```

## Parallel Example: User Story 3

```bash
# Launch Dockerfile and entrypoint in parallel:
Task: "Create Dockerfile" (T013)
Task: "Create docker/ollama-entrypoint.sh" (T014)

# Then sequentially:
Task: "Create docker-compose.yml" (T015)
Task: "Verify end-to-end Docker flow" (T016)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 3: User Story 1 (T007-T010)
3. **STOP and VALIDATE**: Query → answer → citation click works in browser
4. This alone is a demoable milestone

### Incremental Delivery

1. Setup (T001-T004) → Static file serving ready
2. Foundational (T005-T006) → Auto-reindex on startup, unblocks Docker
3. Add US1 (query + answer) → Demo MVP
4. Add US2 (ADR browsing) + US4 (about panel) → Complete UI
5. Add US3 (Docker) → One-command deployment
6. Add US5 (README) → Fully documented
7. Polish → Production-ready

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- All UI work is in 3 files: index.html, style.css, app.js — coordinate sequential access
- The existing server.go, handlers.go, and main.go need modifications but no new API endpoints
- marked.js is loaded via CDN — no npm/node dependency
