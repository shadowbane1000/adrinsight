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
   [Embed Question]   [Retrieve Top-K]
   (Ollama local)     (SQLite + sqlite-vec)
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
3. Similarity search in SQLite returns top-K relevant ADR chunks
4. Results deduplicated by ADR number; full ADR files read from disk
5. Full ADR content + question sent to Anthropic Claude for synthesis
6. Claude returns structured JSON (answer + citations array) via OutputConfig
7. HTTP API returns the structured response to the client

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
│   └── server/                  # HTTP API + serves web UI
├── web/
│   └── static/                  # HTML/CSS/JS for the web UI
├── docs/
│   ├── adr/                     # project's own ADRs (also demo dataset)
│   ├── architecture.md          # this file
│   └── roadmap.md               # phased roadmap
├── testdata/                    # sample ADRs for testing
├── .gitea/workflows/ci.yaml     # primary CI (lint + test + build)
├── .github/workflows/ci.yaml    # GitHub mirror CI
├── .specify/                    # spec-kit scaffolding
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
| Store     | Persist and search vectors + metadata | SQLite + `sqlite-vec` (mattn/go-sqlite3 + CGO) |
| LLM       | Synthesize answers from context | Anthropic Claude API (official Go SDK) |

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
