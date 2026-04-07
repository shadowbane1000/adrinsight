# ADR Insight — Phased Roadmap

## Phase 1 — Walking Skeleton

**Goal**: `docker-compose up`, open browser, ask a question, get a synthesized answer with citations.

### Milestone 1 — Foundation

1. Initialize Go module, set up directory structure (`cmd/` + `internal/`)
2. ADR markdown parser — read a directory, extract frontmatter/metadata (title, date, status) and body text
3. Ollama integration — call the embedding API, get vectors back
4. SQLite + sqlite-vec setup — store ADR content and vectors, run similarity queries
5. `reindex` command: parse → embed → store
6. Gitea Actions + GitHub Actions CI (lint + test + build)

### Milestone 2 — The Brain

7. Anthropic integration — send retrieved context + user question, get synthesized answer
8. RAG pipeline end-to-end: question → embed → retrieve top-K → synthesize → return with citations
9. HTTP API (at minimum: `POST /query` and `GET /adrs`)

### Milestone 3 — The Face

10. Minimal web UI — search bar, results with citations linking to source ADRs, ADR browse/list view
11. Dockerfiles (Go app + Ollama with model pre-pulled)
12. `docker-compose.yml` that brings it all up
13. README with setup instructions

## Phase 2 — Make It Good

**Goal**: Polish the walking skeleton into something that demonstrates quality engineering.

### Milestone 4a — Evaluation Harness

- Test cases from the About page sample queries with ground truth expected ADRs
- Mechanical scoring (retrieval precision/recall against expected citations)
- LLM-as-judge scoring (answer accuracy and completeness)
- CI integration — runs on every PR, fails if scores drop below baseline
- Establishes the current baseline so retrieval improvements are measurable

### Milestone 4b — Retrieval Improvements

- Improve chunking strategy (experiment with alternatives, measure against harness, ADR the decision)
- Hybrid search (keyword + semantic via FTS5) with reranking
- Each change measured against baseline, only shipped if scores improve

### Milestone 5 — ADR Intelligence

- ADR metadata awareness (understands supersedes/deprecated/amended relationships)
- Parse and store relationship data between ADRs
- Surface relationship context in synthesized answers and ADR detail views
- Treat ADRs as a connected graph, not isolated documents

### Milestone 6 — UI & UX Polish

- Web UI improvements: filtering by status, status badges, relationship links
- Better citation UX (inline highlighting, scroll-to-section)
- Responsive layout, improved loading states, error recovery
- Visual polish that makes the demo feel production-quality

### Milestone 7 — Engineering Quality

- Solid error handling and structured logging throughout
- Configuration management (environment variables, config file support)
- Revisit early Go code for idiom compliance
- Integration tests on the RAG pipeline (does retrieval return the right ADRs? does synthesis cite correctly?)
- Clean, documented API for extensibility

## Phase 3 — Make It Impressive

**Goal**: Portfolio-ready presentation and documentation.

- Architecture diagram in the README
- Comprehensive ADR collection for the project itself
- Expand CI/CD as needed
- Blog post or detailed README walkthrough of design philosophy

## Phase 4 — Optional Stretch

**Goal**: Features that demonstrate advanced thinking but aren't required for the portfolio.

- Decision dependency graph visualization in the UI
- Staleness detection ("this ADR is 2 years old and references a deprecated service")
- "Onboarding mode" — summarize the N most important decisions for new engineers
- Multi-LLM backend support (OpenAI, Ollama for synthesis, etc.)
- Multi-repo / org-wide federation
