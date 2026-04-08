# ADR Insight — HTTP API

All endpoints return JSON. All responses include an `X-Request-ID` header for log correlation.

## `POST /query`

Submit a natural language question about architecture decisions.

**Request:**
```json
{"query": "Why was SQLite chosen over PostgreSQL?"}
```

**Response (200 OK):**
```json
{
  "answer": "SQLite was chosen for single-file simplicity...",
  "citations": [
    {"adr_number": 4, "title": "SQLite with Vector Extension", "section": "Rationale"}
  ]
}
```

**Response (400 Bad Request):**
```json
{"error": "query is required"}
```

**Response (503 Service Unavailable):**
```json
{"error": "ANTHROPIC_API_KEY not configured"}
```

**Notes:**
- Requires `ANTHROPIC_API_KEY` and a running Ollama instance.
- Query is embedded via Ollama, searched via hybrid vector+keyword search, and synthesized via Anthropic Claude.

---

## `GET /adrs`

List all indexed ADRs.

**Response (200 OK):**
```json
{
  "adrs": [
    {"number": 1, "title": "Why Go", "status": "Accepted", "date": "2026-04-06", "path": "docs/adr/ADR-001-why-go.md"},
    {"number": 4, "title": "SQLite with Vector Extension", "status": "Accepted", "date": "2026-04-06", "path": "docs/adr/ADR-004-sqlite-vec.md"}
  ]
}
```

**Notes:**
- Returns ADRs from the indexed database, not directly from disk.
- Run `reindex` to update the list after adding new ADRs.

---

## `GET /adrs/{number}`

Get full content and metadata for a specific ADR.

**Response (200 OK):**
```json
{
  "number": 4,
  "title": "SQLite with Vector Extension",
  "status": "Accepted",
  "date": "2026-04-06",
  "content": "# ADR-004: SQLite with Vector Extension\n\n**Status:** Accepted...",
  "relationships": [
    {"target_adr": 15, "target_title": "mattn/CGO SQLite Driver", "rel_type": "drives", "description": "..."}
  ]
}
```

**Response (404 Not Found):**
```json
{"error": "ADR 99 not found"}
```

**Response (400 Bad Request):**
```json
{"error": "invalid ADR number"}
```

**Notes:**
- `content` is raw markdown (rendered client-side by marked.js).
- `relationships` is omitted if the ADR has no related ADRs.

---

## `GET /health`

Check system readiness.

**Response (200 OK — healthy):**
```json
{
  "status": "healthy",
  "components": {
    "database": {"status": "healthy"},
    "ollama": {"status": "healthy"},
    "anthropic_key": {"status": "healthy"}
  }
}
```

**Response (200 OK — degraded):**
```json
{
  "status": "degraded",
  "components": {
    "database": {"status": "healthy"},
    "ollama": {"status": "unhealthy", "error": "connection refused"},
    "anthropic_key": {"status": "healthy"}
  }
}
```

**Response (503 Service Unavailable — unhealthy):**
```json
{
  "status": "unhealthy",
  "components": {
    "database": {"status": "unhealthy", "error": "database is locked"},
    "ollama": {"status": "healthy"},
    "anthropic_key": {"status": "healthy"}
  }
}
```

**Status values:**
- `healthy` — all components operational
- `degraded` — non-critical components unreachable (Ollama, API key)
- `unhealthy` — database unreachable (503 returned)

**Notes:**
- Per-component checks timeout after 3 seconds.
- Used by Docker health checks.
