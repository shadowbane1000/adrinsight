# Phase 2 — Make It Good: Milestones

**Goal**: Polish the walking skeleton into something that demonstrates quality
engineering. Better answers, smarter ADR understanding, polished UI, and
production-grade code.

Each milestone is independently demonstrable and builds on the previous one.

---

## Milestone 4a — Evaluation Harness

**Demonstrates**: A repeatable, automated way to measure retrieval and answer
quality — the foundation for all subsequent improvements.

The current system has no way to measure whether a change to chunking, search,
or synthesis makes answers better or worse. Before changing anything, we need
a baseline and a way to measure against it.

**Deliverables**:

- **Test case corpus** — the six sample queries from the About page, plus
  additional edge cases, each annotated with:
  - Expected ADR citations (ground truth — which ADRs should be cited?)
  - Key facts the answer must contain
  - Store as JSON in `testdata/eval/cases.json`
- **Mechanical scoring** — retrieval metrics computed without an LLM:
  - **Precision**: what fraction of returned citations are in the expected set?
  - **Recall**: what fraction of expected citations were returned?
  - **F1**: harmonic mean of precision and recall
  - Run against the live system (requires Ollama + indexed ADRs)
- **LLM-as-judge scoring** — an Anthropic Claude call that receives the
  question, the full text of the expected ADRs, and the system's answer,
  then scores on a rubric:
  - **Accuracy** (0-5): does the answer correctly represent what the ADRs say?
  - **Completeness** (0-5): does it address all aspects of the question?
  - Returns structured JSON with scores and brief justification per dimension
- **Baseline capture** — run the harness against the current system and
  record scores as `testdata/eval/baseline.json`. This is the "before"
  snapshot.
- **CLI command** — `./adr-insight eval` runs the full harness and outputs
  a summary table (per-question scores + aggregate). Exit code 1 if any
  score drops below baseline thresholds.
- **CI integration** — the eval command runs in the pipeline. PRs that
  regress retrieval or answer quality below baseline thresholds are flagged.

**Done when**: `./adr-insight eval` runs all test cases, produces mechanical
and LLM-judge scores, compares against baseline, and the current system's
scores are documented. The About page sample queries all have ground truth
annotations.

**Key decisions to ADR**: Evaluation methodology (LLM-as-judge rubric,
threshold policy, mechanical vs subjective split).

---

## Milestone 4b — Retrieval Improvements

**Demonstrates**: Measurable improvement in search relevance and answer
quality, validated by the evaluation harness from M4a.

The current retrieval splits ADRs by H2 sections and searches by vector
similarity alone. This misses keyword-exact matches (e.g., searching for
"sqlite-vec" may not rank the SQLite ADR highest) and the chunking
granularity may not be optimal for all ADR structures.

**Deliverables**:

- **Chunking experiments** — ADR-008 deferred three alternatives to Phase 2:
  fixed-size token windows (handles varying section lengths), paragraph-level
  splitting (finer granularity), and whole-document embedding (broader context).
  Evaluate each against the harness. Also consider overlapping windows to
  address ADR-008's noted gap where queries spanning two sections miss both.
  ADR the winning strategy (supersedes or amends ADR-008).
- **Hybrid search** — combine keyword search (SQLite FTS5) with vector
  similarity. Score and rank results using a weighted combination. This
  catches cases where exact terminology matters (e.g., library names,
  acronyms).
- **Reranking** — after retrieving candidates, re-score them with a
  heuristic (e.g., boost exact title matches, penalize deprecated ADRs).
  Keep it simple — a tunable scoring function, not a separate ML model.
- **Search query preprocessing** — expand abbreviations, handle common
  question patterns ("why did we..." → focus on Rationale sections).
- **Iterative measurement** — each change is measured against the M4a
  baseline. Only ship changes that improve scores. Document before/after
  results for each experiment.

**Done when**: The evaluation harness shows measurable improvement over the
M4a baseline for the test question set, and at least one alternative chunking
strategy has been evaluated and documented in an ADR.

**Key decisions to ADR**: Chunking strategy change (supersedes/amends ADR-008),
hybrid search weighting.

---

## Milestone 5 — ADR Intelligence

**Demonstrates**: The system understands ADRs as a connected graph, not
isolated documents.

ADRs reference each other through "Related ADRs" sections, supersession
chains (ADR-015 supersedes ADR-006's driver choice), and topical clusters.
The current system ignores these relationships entirely.

**Deliverables**:

- **Relationship parsing** — extract "Related ADRs" sections and parse
  relationship types (supersedes, superseded by, amends, depends on,
  related to). Store relationships in a new `adr_relationships` table.
- **Status-aware retrieval** — when an ADR has been superseded or deprecated,
  the system should prefer the newer ADR and note the relationship in the
  answer. "ADR-006 was originally about ncruces/WASM but was amended to use
  mattn/CGO."
- **Relationship context in synthesis** — when the LLM receives ADRs for
  synthesis, include relationship metadata so answers can reference the
  decision chain: "This was decided in ADR-004 and later refined in ADR-006."
- **Relationship display in UI** — when viewing an ADR in the detail panel,
  show linked ADRs as clickable navigation (e.g., "Superseded by: ADR-006",
  "Related: ADR-001, ADR-004").
- **Graph query support** — answer questions like "what decisions were
  affected by the SQLite choice?" by traversing the relationship graph.

**Done when**: Asking "what happened with the SQLite driver decision?" returns
an answer that traces the chain from ADR-004 (SQLite chosen) → ADR-006
(ncruces/WASM chosen) → ADR-015 (superseded by mattn/CGO), and the ADR
detail view shows clickable relationship links.

**Key decisions to ADR**: Relationship storage schema, how relationships
influence retrieval ranking, graph traversal depth limits.

---

## Milestone 6 — UI & UX Polish

**Demonstrates**: A production-quality interface that a hiring manager would
be comfortable using without guidance.

The current UI is functional but minimal. This milestone makes it look and
feel like a real product.

**Deliverables**:

- **Status badges** — visual indicators for ADR status (Accepted, Proposed,
  Deprecated, Superseded) with color coding in the list and detail views.
- **Filtering and sorting** — filter ADR list by status, sort by number or
  date. Persist filter state across navigation.
- **Improved citation UX** — inline citation markers in the answer text
  (not just a list below), click to scroll to the relevant section in the
  ADR detail panel.
- **Relationship navigation** — clickable relationship links in the ADR
  detail panel (from M5 data), breadcrumb-style decision chains.
- **Search improvements** — search-as-you-type suggestions, recent query
  history, clear/reset search.
- **Loading and error states** — skeleton loading for ADR list, better
  error messages with retry actions, connection status indicator.
- **Responsive layout** — work reasonably on tablet-sized screens (not
  full mobile, but not broken either).
- **Visual refinement** — consistent spacing, typography hierarchy, subtle
  animations, dark/light consideration.

**Done when**: A non-technical user can navigate the UI without confusion.
ADR status is immediately visible, relationships are navigable, and the
overall feel is polished.

**Key decisions to ADR**: Any significant UX patterns chosen (if non-obvious).

---

## Milestone 7 — Engineering Quality

**Demonstrates**: Production-grade code quality, observability, and test
coverage that reflects real-world engineering standards.

This is the "would I ship this?" milestone. A reviewer should be able to
read the code and see patterns they'd expect in a well-maintained production
service.

**Deliverables**:

- **Structured logging** — replace `log.Printf` with structured logging
  (slog from Go stdlib). Log levels, request IDs, timing information.
  ADR the logging approach.
- **Configuration management** — support environment variables with sensible
  defaults for all settings. Document all configuration options. Consider
  a config file for complex deployments.
- **Error handling audit** — review all error paths for consistent, actionable
  error messages. Ensure errors propagate context (wrapped errors with %w).
  No silent failures.
- **Go idiom review** — audit the full codebase for non-idiomatic patterns
  introduced during rapid development. Fix variable naming, error handling
  style, package organization. Focus on code a Go reviewer would flag.
- **Integration test suite** — tests that exercise the RAG pipeline
  end-to-end: does retrieval return the right ADRs for known questions?
  Does synthesis produce answers that cite the correct sources? Use the
  evaluation harness from M4 as the foundation.
- **API documentation** — document the HTTP API endpoints with request/response
  examples. Consider generating OpenAPI spec from the handler definitions.
- **Health check endpoint** — `GET /health` that verifies database, Ollama,
  and (optionally) Anthropic connectivity. Used by Docker health checks.
- **Graceful shutdown** — handle SIGTERM properly, drain in-flight requests,
  close database connections cleanly.

**Done when**: `make check` passes with structured logging visible, integration
tests exercise the full pipeline, and a code reviewer would find no obvious
Go anti-patterns.

**Key decisions to ADR**: Logging framework choice (slog vs third-party),
configuration approach, test strategy.
