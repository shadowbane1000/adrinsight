# HTTP API Contract: adr-insight

Base URL: `http://localhost:{port}` (port configurable, default 8081)

## POST /query

Submit a natural-language question and receive a synthesized answer.

**Request**:
```json
{
  "query": "Why did we choose Go?"
}
```

**Response (200)**:
```json
{
  "answer": "Go was chosen primarily for its alignment with the platform engineering job market and its simplicity as a language for portfolio projects. The decision is documented in ADR-001, which notes that Go's static binary compilation, strong concurrency model, and opinionated simplicity make code easy to review — important for a portfolio piece.",
  "citations": [
    {
      "adr_number": 1,
      "title": "Why Go",
      "section": "Rationale"
    },
    {
      "adr_number": 1,
      "title": "Why Go",
      "section": "Context"
    }
  ]
}
```

**Response (400)** — empty or missing query:
```json
{
  "error": "query is required"
}
```

**Response (500)** — synthesis or embedding service unavailable:
```json
{
  "error": "synthesis service unavailable: connection refused"
}
```

## GET /adrs

List all indexed ADRs with metadata.

**Response (200)**:
```json
{
  "adrs": [
    {
      "number": 1,
      "title": "Why Go",
      "status": "Accepted",
      "path": "docs/adr/ADR-001-why-go.md"
    },
    {
      "number": 2,
      "title": "Anthropic Claude for LLM Synthesis",
      "status": "Accepted",
      "path": "docs/adr/ADR-002-anthropic-for-synthesis.md"
    }
  ]
}
```

**Response (200)** — empty index:
```json
{
  "adrs": []
}
```

## GET /adrs/{number}

Retrieve full content of a single ADR by number.

**Response (200)**:
```json
{
  "number": 1,
  "title": "Why Go",
  "status": "Accepted",
  "date": "2026-04-06",
  "content": "# ADR-001: Why Go\n\n**Status:** Accepted\n..."
}
```

**Response (404)** — ADR not found:
```json
{
  "error": "ADR 99 not found"
}
```

## CLI: serve command

```
adr-insight serve [--port <port>] [--db <path>] [--ollama-url <url>] [--model <model>]
```

| Flag           | Default                    | Description                  |
|----------------|----------------------------|------------------------------|
| `--port`       | `8081`                     | HTTP server port             |
| `--db`         | `./adr-insight.db`         | Path to SQLite database      |
| `--ollama-url` | `http://localhost:11434`   | Ollama API base URL          |
| `--adr-dir`    | `./docs/adr`               | ADR directory (for full reads)|
| `--model`      | `claude-sonnet-4-5`        | Anthropic model to use       |

Requires `ANTHROPIC_API_KEY` environment variable.
