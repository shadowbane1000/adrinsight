# Tasks: UI & UX Polish

**Input**: Design documents from `/specs/007-ui-ux-polish/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Add Alpine.js, restructure HTML for unified content area, establish Alpine stores.

- [x] T001 Add Alpine.js CDN script tag to `web/static/index.html` (before app.js, with `defer`)
- [x] T002 Restructure `web/static/index.html` — replace the separate `#answer-area` and `#detail-panel` with a single unified `#content-area` div. Keep `#about-content` and `#adr-detail` as child views. Add `#answer-view` div for query answers. Add `#breadcrumb` container at top of content area.
- [x] T003 Create Alpine stores skeleton in `web/static/app.js` — define `Alpine.store('nav')` (stack, push, pop, goto), `Alpine.store('adrs')` (all, filtered, filter, sort, loading, error), `Alpine.store('query')` (text, loading, error, lastQuery, history, showHistory), `Alpine.store('ui')` (sidebarOpen). Initialize stores before `Alpine.start()`.
- [x] T004 Migrate existing fetch/render logic in `web/static/app.js` to use Alpine stores — replace manual DOM manipulation for ADR list loading, ADR detail display, query submission, and about page toggling with Alpine store updates. Wire `x-data`, `x-show`, `x-for`, `x-on` directives in `web/static/index.html`.

**Checkpoint**: App loads with Alpine.js. ADR list renders via Alpine store. Query and ADR detail work through the unified content area. No visual changes yet — same functionality, new plumbing.

---

## Phase 2: Foundational — Navigation Stack and Breadcrumbs

**Purpose**: Implement the view stack that all user stories depend on.

**CRITICAL**: This must be complete before story-specific work begins.

- [x] T005 Implement navigation stack logic in `Alpine.store('nav')` in `web/static/app.js` — `push(viewEntry)`, `pop()`, `goto(index)`, `replace(viewEntry)`, `clear()`. ViewEntry has `{type, data, label}`. `currentView` is computed as top of stack. Track whether an answer is active in the stack for sidebar push/replace behavior.
- [x] T006 Implement breadcrumb rendering in `web/static/index.html` — use `x-for` over `nav.stack` to render clickable breadcrumb segments with `>` separators. Each segment calls `nav.goto(index)` on click. Hide breadcrumb when stack has 0 or 1 entries. Truncate with `...` when stack exceeds 5 entries.
- [x] T007 Wire content area view switching in `web/static/index.html` — use `x-show` with `x-transition` to display the correct view based on `nav.currentView.type` (`about`, `answer`, `adr`). Cache view data in stack entries so backward navigation is instant.
- [x] T008 Wire ADR list sidebar clicks to navigation stack in `web/static/app.js` — if an answer is active in the stack, push the ADR; otherwise replace. Wire citation links and relationship links to always push.
- [x] T009 Add breadcrumb and content area transition styles in `web/static/style.css` — breadcrumb bar styling, separator styling, truncation ellipsis, fade transitions for view switching.

**Checkpoint**: Unified content area works. Submit a query → answer shows. Click citation → ADR detail with breadcrumb "Answer > ADR-004". Click relationship → breadcrumb grows. Click breadcrumb segment → navigates back. Sidebar browse (no answer) replaces without stack growth.

---

## Phase 3: User Story 1 — Status Badges (Priority: P1)

**Goal**: Color-coded status badges visible in ADR list and detail view.

**Independent Test**: Open the app, verify each ADR shows a colored badge. Click ADR-008, verify "Accepted" badge in green. Click ADR-006, verify "Superseded" badge in gray.

### Implementation for User Story 1

- [x] T010 [US1] Add status badge CSS classes in `web/static/style.css` — `.badge`, `.badge-accepted` (green), `.badge-proposed` (blue), `.badge-deprecated` (amber), `.badge-superseded` (gray), `.badge-unknown` (neutral gray). Rounded corners, small inline display, white text.
- [x] T011 [US1] Add `normalizeStatus(status)` helper in `web/static/app.js` — extract first word from status string, lowercase. Map "accepted" → "accepted", "proposed" → "proposed", "deprecated" → "deprecated", "superseded" → "superseded", default → "unknown".
- [x] T012 [US1] Update ADR list rendering in `web/static/index.html` — replace plain text `.adr-list-status` with a `<span class="badge badge-{normalizedStatus}">` showing the normalized status word.
- [x] T013 [US1] Add status badge to ADR detail header in `web/static/index.html` — when an ADR is displayed in the content area, show a badge next to the title at the top of the detail view (above the rendered markdown).

**Checkpoint**: Status badges visible in list and detail. Colors match the mapping.

---

## Phase 4: User Story 2 — Filter and Sort (Priority: P1)

**Goal**: Filter ADR list by status, sort by number or date.

**Independent Test**: Select "Accepted" filter → only accepted ADRs visible. Sort by "Date (newest)" → order changes. Navigate to ADR and back → filter and sort persist.

### Implementation for User Story 2

- [x] T014 [US2] Add filter/sort control HTML in `web/static/index.html` — add a filter bar above `#adr-list-items` with status filter buttons (All, Accepted, Proposed, Deprecated, Superseded) and a sort dropdown (Number ↑, Number ↓, Date ↑, Date ↓). Use Alpine `x-on:click` to set `adrs.filter` and `adrs.sort`.
- [x] T015 [US2] Add `date` field to `adrSummaryJSON` in `internal/server/handlers.go` and populate it in `handleListADRs` by calling `extractDate` on each ADR's file content. Update the `adrSummaryJSON` struct and list response. Update mock in `internal/server/handlers_test.go` if needed.
- [x] T016 [US2] Implement `filtered` computed property in `Alpine.store('adrs')` in `web/static/app.js` — filter `all` by `filter` value (null = show all), then sort by `sort` value using the date field from the list response.
- [x] T017 [US2] Persist filter and sort to localStorage in `web/static/app.js` — on change, write to `adr-insight-filter` and `adr-insight-sort`. On init, read and restore.
- [x] T018 [US2] Update ADR list rendering in `web/static/index.html` to use `adrs.filtered` instead of `adrs.all`. Show "No ADRs match the current filter" with a clear-filter link when filtered list is empty.
- [x] T019 [US2] Add filter/sort control styling in `web/static/style.css` — filter button group, active filter highlight (use accent color), sort dropdown styling, compact layout within sidebar.

**Checkpoint**: Filtering and sorting work. State persists across ADR navigation and page reloads.

---

## Phase 5: User Story 3 — Inline Citations and Decision Chain Exploration (Priority: P1)

**Goal**: ADR references in answer text become clickable links. Full navigation chain works.

**Independent Test**: Submit query "What are all the ways SQLite is used?" → answer contains clickable ADR-004, ADR-015, ADR-017 links. Click ADR-004 → breadcrumb "Answer > ADR-004". Click relationship to ADR-015 → breadcrumb "Answer > ADR-004 > ADR-015". Click "Answer" in breadcrumb → returns to answer instantly.

### Implementation for User Story 3

- [x] T020 [US3] Implement `linkifyCitations(html, citations)` function in `web/static/app.js` — scan rendered answer HTML for `ADR-\d+` patterns. For each match where the ADR number is in the citations array, wrap in `<a class="citation-inline" data-adr="N">ADR-NNN</a>`. Leave non-cited ADR references as plain text.
- [x] T021 [US3] Wire inline citation clicks in `web/static/app.js` — add delegated click handler on the answer view for `.citation-inline` elements. On click, fetch the ADR detail and push onto the navigation stack.
- [x] T022 [US3] Update answer rendering in `web/static/app.js` — after `marked.parse()`, pass result through `linkifyCitations()` before inserting into the DOM. Retain the citation summary list below the answer.
- [x] T023 [US3] Add inline citation styling in `web/static/style.css` — `.citation-inline` with accent color, underline on hover, cursor pointer. Distinguish from regular links.

**Checkpoint**: Inline citations work. Full decision chain navigation works with breadcrumbs. Backward navigation is instant.

---

## Phase 6: User Story 4 — Error States with Retry (Priority: P2)

**Goal**: Errors include actionable retry buttons.

**Independent Test**: Stop server, submit query → error with "Try Again" button. Restart server, click "Try Again" → query re-fires. ADR list load failure shows "Retry" in sidebar.

### Implementation for User Story 4

- [x] T024 [US4] Update query error handling in `web/static/app.js` — on query failure, set `query.error` with user-friendly message and store `query.lastQuery`. Do not show raw error codes.
- [x] T025 [US4] Add "Try Again" button to query error display in `web/static/index.html` — show when `query.error` is set. On click, re-submit `query.lastQuery`.
- [x] T026 [US4] Update ADR list error handling in `web/static/app.js` — on list load failure, set `adrs.error`. Add "Retry" button to sidebar in `web/static/index.html` that calls `loadADRList()`.
- [x] T027 [US4] Add error state styling in `web/static/style.css` — error container, retry button styling consistent with accent color.

**Checkpoint**: Query and ADR list errors show user-friendly messages with retry actions.

---

## Phase 7: User Story 6 — Recent Query History (Priority: P2)

**Goal**: Recent queries stored in localStorage, accessible via dropdown.

**Independent Test**: Submit queries → focus search input → dropdown shows recent queries. Click one → fires immediately. Reload page → history persists. Clear history → dropdown empty.

### Implementation for User Story 6

- [x] T028 [US6] Implement query history management in `Alpine.store('query')` in `web/static/app.js` — `addToHistory(query)` prepends to array, deduplicates, caps at 10, writes to `adr-insight-query-history` in localStorage. `clearHistory()` empties array and localStorage. On init, load from localStorage.
- [x] T029 [US6] Add history dropdown HTML in `web/static/index.html` — below the search form, show a dropdown list when `query.showHistory` is true and `query.history.length > 0`. Each item shows the query text. Include a "Clear history" link at the bottom.
- [x] T030 [US6] Wire history dropdown behavior in `web/static/app.js` — show dropdown on search input focus (if history exists), hide on blur (with delay for click), hide on Escape. Clicking a history item populates input and submits.
- [x] T031 [US6] Add history dropdown styling in `web/static/style.css` — positioned below search input, subtle shadow, hover highlight, max-height with scroll if needed.

**Checkpoint**: Query history works end-to-end. Persists across reloads. Clear works.

---

## Phase 8: User Story 5 — Responsive Layout (Priority: P3)

**Goal**: Usable layout at 768px viewport width.

**Independent Test**: Open dev tools at 768px → sidebar collapses, toggle button appears. Click toggle → sidebar slides out. Search, browse, and read all work.

### Implementation for User Story 5

- [x] T032 [US5] Add sidebar toggle button to header in `web/static/index.html` — hamburger icon visible only below 768px. Toggles `ui.sidebarOpen` in Alpine store.
- [x] T033 [US5] Add responsive CSS in `web/static/style.css` — `@media (max-width: 768px)` rules: sidebar becomes `position: fixed` overlay with slide-in transition, content area takes full width, search form adapts to smaller width, header stacks if needed.
- [x] T034 [US5] Wire sidebar auto-close on navigation in `web/static/app.js` — when user selects an ADR or submits a query on small screens, auto-close the sidebar so the content area is visible.

**Checkpoint**: Layout is usable at 768px. No overlapping content. Core functions work.

---

## Phase 9: User Story 7 — Search Clear (Priority: P3)

**Goal**: Clear button in search input, Escape key support.

**Independent Test**: Type in search → X button appears. Click X → input clears, content returns to about page. Press Escape while typing → same behavior.

### Implementation for User Story 7

- [x] T035 [US7] Add clear button to search form in `web/static/index.html` — an X button inside or adjacent to the search input, visible only when `query.text` is non-empty. Use `x-show` with transition.
- [x] T036 [US7] Wire clear behavior in `web/static/app.js` — on clear button click or Escape key, set `query.text = ''`, clear the navigation stack, and show the about page. Discard cached answer.
- [x] T037 [US7] Add clear button styling in `web/static/style.css` — positioned inside the search input (right side), subtle appearance, hover highlight.

**Checkpoint**: Clear button and Escape work. Returns to about page.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Loading states, animations, visual refinement, final validation.

- [x] T038 Add loading skeleton for ADR list in `web/static/index.html` and `web/static/style.css` — show placeholder items with animated shimmer while `adrs.loading` is true
- [x] T039 Add "Still thinking..." message to query loading in `web/static/app.js` — after 5 seconds of loading, update the loading indicator text. Use `x-show` with transition.
- [x] T040 Add answer appearance animation in `web/static/style.css` and `web/static/index.html` — fade-in on answer view using `x-transition`. Citations fade in with slight stagger delay.
- [x] T041 Visual refinement pass on `web/static/style.css` — consistent spacing, typography hierarchy, hover transitions on all interactive elements, clean visual separation between header/sidebar/content.
- [x] T042 Update `web/static/index.html` about page sample queries — ensure they still work as clickable items that submit via the new Alpine-based form handler.
- [x] T043 Run `make lint` and fix any lint errors across all modified files
- [x] T044 Update `README.md` ADR table with ADR-023
- [x] T045 Update `docs/architecture.md` with Alpine.js and unified content area description
- [x] T046 Manual verification of all 8 quickstart.md scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — Alpine.js integration and HTML restructuring
- **Foundational (Phase 2)**: Depends on Setup — navigation stack and breadcrumbs
- **US1 Badges (Phase 3)**: Depends on Setup — needs Alpine rendering in place
- **US2 Filter/Sort (Phase 4)**: Depends on Setup — needs Alpine store for ADR list
- **US3 Citations (Phase 5)**: Depends on Foundational — needs navigation stack for citation clicks
- **US4 Errors (Phase 6)**: Depends on Setup — needs Alpine store for error state
- **US6 History (Phase 7)**: Depends on Setup — needs Alpine store for query state
- **US5 Responsive (Phase 8)**: Depends on Foundational — needs sidebar and content area finalized
- **US7 Clear (Phase 9)**: Depends on Foundational — needs navigation stack for clear-to-about behavior
- **Polish (Phase 10)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 1 only — can start after setup
- **US2 (P1)**: Depends on Phase 1 only — can start after setup
- **US3 (P1)**: Depends on Phase 2 — needs navigation stack
- **US4 (P2)**: Depends on Phase 1 only — can start after setup
- **US5 (P3)**: Depends on Phase 2 — needs finalized layout
- **US6 (P2)**: Depends on Phase 1 only — can start after setup
- **US7 (P3)**: Depends on Phase 2 — needs navigation stack

### Parallel Opportunities

- After Phase 1: US1 (T010-T013) and US2 (T014-T019) can run in parallel (different concerns)
- After Phase 1: US4 (T024-T027) and US6 (T028-T031) can run in parallel
- After Phase 2: US3 (T020-T023), US5 (T032-T034), and US7 (T035-T037) can run in parallel

---

## Implementation Strategy

### MVP First (Setup + Foundational + US1 + US3)

1. Complete Phase 1: Setup — Alpine.js, unified content area, stores (T001-T004)
2. Complete Phase 2: Navigation stack and breadcrumbs (T005-T009)
3. Complete Phase 3: US1 — Status badges (T010-T013)
4. Complete Phase 5: US3 — Inline citations and decision chain (T020-T023)
5. **STOP and VALIDATE**: Unified content area works, breadcrumbs navigate, badges visible, citations clickable

### Incremental Delivery

1. Setup + Foundational → app works with Alpine.js, unified content area
2. US1 → status badges visible → **visual impact**
3. US3 → inline citations, full navigation chain → **core UX improvement**
4. US2 → filter and sort → **productivity feature**
5. US4 + US6 → error retry, query history → **polish**
6. US5 + US7 → responsive, search clear → **edge cases**
7. Polish → loading skeletons, animations, docs → **final sheen**

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
- All changes are frontend (HTML/CSS/JS) — minimal or no Go backend changes
- Alpine.js loaded via CDN — no build step, no node_modules
- Filter/sort date requires the date field from `/adrs` response — verify backend returns it; if not, add it as a minor backend task
