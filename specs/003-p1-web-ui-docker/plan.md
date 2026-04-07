# Implementation Plan: Web UI and Docker Deployment

**Branch**: `003-p1-web-ui-docker` | **Date**: 2026-04-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-p1-web-ui-docker/spec.md`

## Summary

Build a minimal web UI (vanilla HTML/CSS/JS) for querying ADRs and browsing
the ADR collection, served by the Go HTTP server. Containerize the application
and Ollama with docker-compose for one-command deployment. Update the README
with complete setup instructions.

## Technical Context

**Language/Version**: Go 1.24+ (same as M1/M2)
**Primary Dependencies**:
- marked.js via CDN (client-side markdown rendering, ~16KB)
- go:embed for static file serving in production
- Existing M1/M2 packages (parser, embedder, store, llm, rag, server)
**Storage**: SQLite with sqlite-vec (from M1, unchanged)
**Testing**: Browser-based manual testing, `go test` for server changes
**Target Platform**: Docker (linux/amd64), browser (Chrome, Firefox, Safari)
**Project Type**: Web application (Go backend + static frontend)
**Performance Goals**: Question-to-answer within 15 seconds
**Constraints**: No frontend framework, no build step, no node_modules
**Scale/Scope**: Single instance, single page UI
**Auto-Reindex**: The `serve` command checks for an empty/missing database on
startup and runs reindex automatically. This ensures `docker compose up`
produces a queryable system without a manual reindex step. If the database
already has data, reindex is skipped for fast restarts.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | ✅ PASS | No new interfaces needed — UI calls existing HTTP API |
| II. ADRs Are the Product | ✅ PASS | ADR checkpoint after planning |
| III. Skateboard First | ✅ PASS | Vanilla HTML/CSS/JS, marked.js via CDN, no framework |
| IV. Idiomatic Go | ✅ PASS | go:embed for static files, stdlib HTTP serving |
| V. Developer Experience | ✅ PASS | docker-compose up + one API key. README in 5 minutes. |

## Project Structure

### Documentation (this feature)

```text
specs/003-p1-web-ui-docker/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (UI layout contract)
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (new/modified files)

```text
web/
└── static/
    ├── index.html           # Single page: search, answer, ADR list, detail panel
    ├── style.css            # Layout and styling
    └── app.js               # API calls, markdown rendering, UI state

internal/
└── server/
    ├── server.go            # Add static file serving (go:embed + dev mode)
    └── embed.go             # go:embed directive for web/static/

cmd/
└── adr-insight/
    └── main.go              # Add --dev flag; auto-reindex on serve startup

Dockerfile                   # Multi-stage build (golang:bookworm → debian:bookworm-slim)
docker-compose.yml           # Go app + Ollama with model pre-pull
docker/
└── ollama-entrypoint.sh     # Entrypoint script for model pre-pull

README.md                    # Complete setup instructions
```

**Structure Decision**: Static files in `web/static/` per the existing
project structure. New `embed.go` file for the `go:embed` directive.
Docker files at project root following convention.

## Complexity Tracking

No constitution violations — no entries needed.
