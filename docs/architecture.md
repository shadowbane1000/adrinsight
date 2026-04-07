# ADR Insight — Architecture

## System Overview

```
User Question
     │
     ▼
 [Web UI] ──► [HTTP API]
                  │
                  ▼
            [RAG Pipeline]
             /          \
   [Embed Question]   [Hybrid Search]
   (Ollama local)     (sqlite-vec + FTS5)
                          │
                          ▼
                     [Rerank]
                   (domain heuristics)
                          │
                          ▼
                   [Synthesize Answer]
                   (Anthropic Claude)
                          │
                          ▼
                   [Answer + Citations]
```

## Data Flows

### Index Time (reindex command)

1. Parse markdown ADRs from a directory — extract title, date, status, body
2. Chunk and embed via Ollama (`nomic-embed-text`)
3. Store vectors + metadata in SQLite with `sqlite-vec`

The SQLite database is a derived artifact. It can be deleted and fully
regenerated from the source ADR files at any time.

### Query Time (HTTP API)

1. User sends a question via `POST /query`
2. Question is embedded via Ollama with `search_query:` prefix
3. Hybrid search: vector similarity (sqlite-vec) + keyword search (FTS5 BM25), weighted 0.6/0.4, merged and deduplicated
4. Reranking: title match boost, superseded penalty (authoritative relationship data), section relevance boost
5. Results deduplicated by ADR number
6. Relationship expansion: related ADRs added via 1-hop graph traversal + full supersession chain walking
7. Relationship context summary prepended to LLM prompt
8. Full ADR content + question sent to Anthropic Claude for synthesis
9. Claude returns structured JSON (answer + citations array) via OutputConfig
10. HTTP API returns the structured response to the client

## Project Structure

```
adr-insight/
├── cmd/
│   └── adr-insight/
│       └── main.go              # entrypoint
├── internal/
│   ├── parser/                  # ADR markdown parsing
│   ├── embedder/                # embedding interface + Ollama impl
│   ├── store/                   # SQLite + sqlite-vec storage
│   ├── rag/                     # retrieval + synthesis orchestration
│   ├── llm/                     # LLM interface + Anthropic impl
│   ├── eval/                    # evaluation harness (scoring, judge, reporting)
│   └── server/                  # HTTP API + serves web UI
├── web/
│   ├── embed.go                 # go:embed directive for static files
│   └── static/                  # HTML/CSS/JS for the web UI
├── docs/
│   ├── adr/                     # project's own ADRs (also demo dataset)
│   ├── architecture.md          # this file
│   └── roadmap.md               # phased roadmap
├── testdata/
│   ├── eval/                    # evaluation test cases and baseline
│   └── *.md                     # sample ADRs for testing
├── .gitea/workflows/ci.yaml     # primary CI (lint + test + build)
├── .github/workflows/ci.yaml    # GitHub mirror CI
├── .specify/                    # spec-kit scaffolding
├── docker/
│   └── ollama-entrypoint.sh     # Ollama model pre-pull entrypoint
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── Makefile
└── .golangci.yml
```

## Key Interfaces

Every external dependency sits behind a Go interface so components are
swappable without changing callers (see ADR-001, Constitution Principle I).

| Interface | Purpose | Implementation |
|-----------|---------|----------------|
| Parser    | Parse ADR markdown files | goldmark + goldmark-meta |
| Embedder  | Convert text to vector embeddings | Ollama (`nomic-embed-text`) |
| Store     | Persist and search vectors + metadata | SQLite + `sqlite-vec` + FTS5 (mattn/go-sqlite3 + CGO) |
| LLM       | Synthesize answers from context | Anthropic Claude API (official Go SDK) |
| Reranker  | Reorder search results with domain heuristics | DefaultReranker (title boost, status penalty, section relevance) |

## Technical Notes

- **Anthropic has no embeddings API** — embeddings are handled separately
  via Ollama. See ADR-003.
- **sqlite-vec** is a young extension but appropriate at this scale. The
  interface abstraction allows swapping to pgvector later. See ADR-004.
- **Ollama Docker image** is official and well-maintained. First run
  downloads `nomic-embed-text` (~275MB).
- **CI runs on both Gitea Actions (primary) and GitHub Actions (mirror).**
  See ADR-005.
- **HTTP API uses Go 1.22+ stdlib ServeMux** — no framework needed for
  3 endpoints. See ADR-011.
- **Structured LLM output** via Anthropic OutputConfig with JSON schema —
  guarantees valid response format. See ADR-009.
- **Full ADR expansion** — search retrieves chunks but synthesis gets full
  ADR files from disk. See ADR-010.
- **Web UI** uses vanilla HTML/CSS/JS with marked.js (CDN) for markdown
  rendering. Static files embedded in binary via `go:embed`. See ADR-012, ADR-013.
- **Auto-reindex on startup** — the `serve` command checks if the database
  is empty and runs reindex automatically, with retry backoff for Ollama.
- **Docker multi-stage build** — `golang:bookworm` (build) to
  `debian:bookworm-slim` (runtime) for CGO/libc ABI compatibility. See ADR-014.
- **Evaluation harness** — `./adr-insight eval` runs test questions against
  the live system, scores answers with mechanical metrics (precision/recall/F1)
  and LLM-as-judge (accuracy/completeness), and detects regressions against
  a saved baseline. See ADR-016.
- **Hybrid search** — combines vector similarity (sqlite-vec, cosine distance)
  with keyword matching (FTS5 BM25). Both scores are min-max normalized to
  0-1 and merged with configurable weights (default 0.7 vector, 0.3 keyword).
  See ADR-017.
- **Reranking** — after hybrid search, domain-specific heuristics adjust result
  ordering: title match boost (+0.2), superseded/deprecated penalty (-0.1
  using authoritative relationship data), and section relevance boost for
  rationale/alternative queries (+0.1).
- **LLM keyword extraction** — during reindex, Haiku extracts domain-specific
  search terms from each ADR. FTS queries are filtered to this vocabulary,
  eliminating noise from generic English words. See ADR-018.
- **ADR relationship graph** — during reindex, "Related ADRs" sections are
  parsed and relationship types classified via Haiku. Five types: supersedes,
  superseded_by, depends_on, drives, related_to. Stored in
  `adr_relationships` table. Used for retrieval expansion (co-retrieve
  related ADRs), synthesis context (relationship summary for the LLM),
  reranker penalties, and UI navigation links. See ADR-020, ADR-021.
