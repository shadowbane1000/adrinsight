# ADR Insight — Phased Roadmap

## Phase 1 — Walking Skeleton

**Goal**: `docker-compose up`, open browser, ask a question, get a synthesized answer with citations.

### Week 1 — Foundation

1. Initialize Go module, set up directory structure (`cmd/` + `internal/`)
2. ADR markdown parser — read a directory, extract frontmatter/metadata (title, date, status) and body text
3. Ollama integration — call the embedding API, get vectors back
4. SQLite + sqlite-vec setup — store ADR content and vectors, run similarity queries
5. `reindex` command: parse → embed → store
6. Gitea Actions + GitHub Actions CI (lint + test + build)

### Week 2 — The Brain

7. Anthropic integration — send retrieved context + user question, get synthesized answer
8. RAG pipeline end-to-end: question → embed → retrieve top-K → synthesize → return with citations
9. HTTP API (at minimum: `POST /query` and `GET /adrs`)

### Week 3 — The Face

10. Minimal web UI — search bar, results with citations linking to source ADRs, ADR browse/list view
11. Dockerfiles (Go app + Ollama with model pre-pulled)
12. `docker-compose.yml` that brings it all up
13. README with setup instructions

## Phase 2 — Make It Good

**Goal**: Polish the walking skeleton into something that demonstrates quality engineering.

- Improve chunking strategy (experiment, document results in an ADR)
- Hybrid search (keyword + semantic), reranking
- Web UI polish: ADR browsing/filtering, citation links, status badges
- ADR metadata awareness (understands supersedes/deprecated relationships)
- Solid error handling, logging, configuration
- Revisit early Go code for idiom compliance
- Write ADRs for every significant decision along the way

## Phase 3 — Make It Impressive

**Goal**: Portfolio-ready presentation and documentation.

- Architecture diagram in the README
- Comprehensive ADR collection for the project itself
- Meaningful test strategy (integration tests on RAG pipeline, not coverage for its own sake)
- Expand CI/CD as needed
- Clean, documented API for extensibility
- Blog post or detailed README walkthrough of design philosophy

## Phase 4 — Optional Stretch

**Goal**: Features that demonstrate advanced thinking but aren't required for the portfolio.

- Decision dependency graph visualization in the UI
- Staleness detection ("this ADR is 2 years old and references a deprecated service")
- "Onboarding mode" — summarize the N most important decisions for new engineers
- Multi-LLM backend support (OpenAI, Ollama for synthesis, etc.)
- Multi-repo / org-wide federation
