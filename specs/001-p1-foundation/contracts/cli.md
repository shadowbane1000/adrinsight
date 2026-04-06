# CLI Contract: adr-insight

## Commands

### reindex

Parses ADR files, generates embeddings, and stores them in the local database.

```
adr-insight reindex [--adr-dir <path>] [--db <path>] [--ollama-url <url>]
```

| Flag          | Default              | Description                        |
|---------------|----------------------|------------------------------------|
| `--adr-dir`   | `./docs/adr`         | Directory containing ADR files     |
| `--db`        | `./adr-insight.db`   | Path to SQLite database file       |
| `--ollama-url`| `http://localhost:11434` | Ollama API base URL             |

**Exit codes**:
- 0: Success
- 1: Fatal error (missing directory, Ollama unreachable, etc.)

**Output** (stdout):
```
Parsing ADRs from ./docs/adr ...
  ADR-001: Why Go
  ADR-002: Anthropic for Synthesis
  Skipped: not-an-adr.md (does not match ADR-*.md pattern)
Embedding 8 chunks via Ollama ...
Stored 8 chunks in ./adr-insight.db
Reindex complete: 4 ADRs, 8 chunks
```

### search

Searches indexed ADRs by semantic similarity. (Used for verification in
Milestone 1 — becomes the foundation for the RAG pipeline in Milestone 2.)

```
adr-insight search <query> [--db <path>] [--ollama-url <url>] [--top-k <n>]
```

| Flag          | Default              | Description                        |
|---------------|----------------------|------------------------------------|
| `--db`        | `./adr-insight.db`   | Path to SQLite database file       |
| `--ollama-url`| `http://localhost:11434` | Ollama API base URL             |
| `--top-k`     | `5`                  | Number of results to return        |

**Exit codes**:
- 0: Success (even if zero results)
- 1: Fatal error (database missing, Ollama unreachable, etc.)

**Output** (stdout):
```
Query: "why did we choose Go?"

1. [ADR-001] Why Go — Context (score: 0.87)
   Go dominates the cloud-native and developer tooling ecosystem...

2. [ADR-001] Why Go — Rationale (score: 0.82)
   Market alignment: Go dominates the cloud-native...

3. [ADR-003] Local Embeddings — Context (score: 0.45)
   ADR Insight needs to convert text into vector embeddings...
```
