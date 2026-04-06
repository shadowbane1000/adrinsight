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

### Query Time

1. User types a question in the web UI
2. Question is embedded via Ollama
3. Similarity search in SQLite returns top-K relevant ADR chunks
4. Retrieved chunks + question sent to Anthropic Claude for synthesis
5. Claude returns a natural-language answer with citations to specific ADRs
6. Web UI displays answer with clickable citation links

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
| Embedder  | Convert text to vector embeddings | Ollama (`nomic-embed-text`) |
| Store     | Persist and search vectors + metadata | SQLite + `sqlite-vec` |
| LLM       | Synthesize answers from context | Anthropic Claude API |

## Technical Notes

- **Anthropic has no embeddings API** — embeddings are handled separately
  via Ollama. See ADR-003.
- **sqlite-vec** is a young extension but appropriate at this scale. The
  interface abstraction allows swapping to pgvector later. See ADR-004.
- **Ollama Docker image** is official and well-maintained. First run
  downloads `nomic-embed-text` (~275MB).
- **CI runs on both Gitea Actions (primary) and GitHub Actions (mirror).**
  See ADR-005.
