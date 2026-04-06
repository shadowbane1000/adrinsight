# Phase 1 — Walking Skeleton: Milestones

**Goal**: `docker-compose up`, open browser, ask a question, get a synthesized
answer with citations.

Each milestone is independently demonstrable and builds on the previous one.

---

## Milestone 1 — Foundation

**Demonstrates**: Parse ADRs, generate embeddings, store and search vectors.

**Deliverables**:
- Go module initialized with `cmd/` + `internal/` layout
- ADR markdown parser (`internal/parser/`) — reads a directory, extracts
  title, date, status, and body text
- Ollama embedding integration (`internal/embedder/`) — calls the embedding
  API, returns vectors
- SQLite + sqlite-vec storage (`internal/store/`) — stores ADR content and
  vectors, runs similarity queries
- `reindex` CLI command — wires parse → embed → store
- CI on Gitea Actions + GitHub Actions (lint + test + build)

**Done when**: `go run ./cmd/adr-insight reindex` processes `docs/adr/`,
stores embeddings in SQLite, and a similarity query returns relevant results.

---

## Milestone 2 — The Brain

**Demonstrates**: Full RAG pipeline — question in, synthesized answer out.

**Deliverables**:
- Anthropic Claude integration (`internal/llm/`) — sends retrieved context +
  question, returns synthesized answer with citations
- RAG pipeline orchestration (`internal/rag/`) — question → embed → retrieve
  top-K → synthesize → return with citations
- HTTP API (`internal/server/`) — at minimum `POST /query` and `GET /adrs`

**Done when**: You can `curl POST /query` with a question about your ADRs and
get back a synthesized answer citing specific ADR numbers.

---

## Milestone 3 — The Face

**Demonstrates**: Complete end-to-end experience in a browser.

**Deliverables**:
- Minimal web UI (`web/static/`) — search bar, results with citations linking
  to source ADRs, ADR browse/list view
- Dockerfiles (Go app + Ollama with model pre-pulled)
- `docker-compose.yml` that brings it all up
- README with setup instructions

**Done when**: `docker-compose up`, open browser, ask "why did you choose Go?",
get a synthesized answer with a citation link to ADR-001.
