# Quickstart: ADR Intelligence

## Prerequisites

- Go 1.24+ with CGO enabled
- `libsqlite3-dev` installed
- Ollama running with `nomic-embed-text`
- `ANTHROPIC_API_KEY` set (required for relationship classification during reindex)

## Build and Test

```bash
make build
ANTHROPIC_API_KEY=... ./adr-insight reindex --adr-dir ./docs/adr
ANTHROPIC_API_KEY=... ./adr-insight serve
```

## Verify Relationships

After reindexing, open `http://localhost:8081` and:

1. Click ADR-015 in the list — verify relationship links show "Supersedes: ADR-006", "Related: ADR-004, ADR-014"
2. Click the ADR-006 link — verify it shows "Superseded by: ADR-015"
3. Ask "What happened with the SQLite driver decision?" — verify the answer traces ADR-004 → ADR-006 → ADR-015

## Run Eval

```bash
ANTHROPIC_API_KEY=... ./adr-insight eval
```

Verify no regressions from M4b baseline.

## Key Files Changed

| File | Change |
|------|--------|
| `internal/parser/markdown.go` | Parse "Related ADRs" sections |
| `internal/store/sqlite.go` | New `adr_relationships` table, CRUD methods |
| `internal/store/store.go` | New methods on Store interface |
| `internal/reindex/reindex.go` | Relationship extraction + LLM classification step |
| `internal/llm/anthropic.go` | `ClassifyRelationship` method |
| `internal/rag/rag.go` | Relationship expansion + context enrichment |
| `internal/rag/rerank.go` | Authoritative supersession penalty |
| `internal/server/handlers.go` | Relationships in ADR detail response |
| `web/static/app.js` | Display relationship links |
