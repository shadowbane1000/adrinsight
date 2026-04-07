# Data Model: Web UI and Docker Deployment

## No New Backend Entities

This milestone adds no new data entities. The web UI consumes the existing
HTTP API endpoints from Milestone 2:

- `POST /query` → QueryResponse (answer + citations)
- `GET /adrs` → ADR list (number, title, status, path)
- `GET /adrs/{number}` → ADR detail (number, title, status, date, content)

## UI State (Client-Side)

The single-page UI manages these client-side states:

| State | Description | Trigger |
|-------|-------------|---------|
| idle | Search bar visible, answer hidden, about panel shown | Page load |
| loading | Loading indicator visible in answer area | Query submitted |
| answered | Answer + citations rendered, about/ADR panel unchanged | Response received |
| adr-view | Right panel shows full ADR content | ADR clicked (list or citation) |
| about-view | Right panel shows project description | Page load or "about" click |
| error | Error message in answer area or panel | API error |

## Docker Composition

| Service | Image | Ports | Depends On |
|---------|-------|-------|------------|
| app | Built from Dockerfile | 8081:8081 | ollama |
| ollama | ollama/ollama | 11434:11434 | — |

Environment variables:
- `ANTHROPIC_API_KEY` — required, passed to app service
- `ADR_DIR` — optional, defaults to `/data/adr` (mounted volume)

## Auto-Reindex on Startup

The `serve` command checks if the database is empty or missing on startup.
If so, it runs the reindex pipeline (parse → embed → store) before starting
the HTTP server. This eliminates a manual `reindex` step from the Docker
workflow. If the database already contains data (from a previous run with
a persistent volume), reindex is skipped.
