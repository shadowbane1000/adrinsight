# ADR Insight

AI-powered search and reasoning over Architecture Decision Records.

ADR Insight lets you ask natural-language questions about your project's architectural decisions and get synthesized answers with citations to the source ADRs. It combines a RAG pipeline (retrieval-augmented generation) with a web UI for browsing and querying.

Built using **spec-driven development** with a [fork of spec-kit](https://github.com/shadowbane1000/spec-kit) that adds ADR generation and specification analysis passes to the original workflow. Any project that maintains ADRs in the standard markdown format can use ADR Insight.

## Quick Start (Docker)

The fastest way to run ADR Insight:

```bash
git clone https://github.com/shadowbane1000/adrinsight.git
cd adrinsight
ANTHROPIC_API_KEY=sk-... docker compose up
```

Open [http://localhost:8081](http://localhost:8081) in your browser.

On first start, Ollama downloads the embedding model (~275MB) and the app indexes the bundled ADRs. Subsequent starts skip indexing if the database is already populated.

## Local Development

### Prerequisites

- **Go 1.24+**
- **GCC / C toolchain** — required for CGO (SQLite + sqlite-vec)
- **libsqlite3-dev** — SQLite development headers
  ```bash
  # Debian/Ubuntu
  sudo apt-get install libsqlite3-dev
  ```
- **Ollama** with `nomic-embed-text` model
  ```bash
  ollama pull nomic-embed-text
  ```
- **Anthropic API key** — set as `ANTHROPIC_API_KEY` environment variable

### Build and Run

```bash
# Build
make build

# Index ADRs (required once, or after ADR changes)
./adr-insight reindex --adr-dir ./docs/adr

# Start the server
ANTHROPIC_API_KEY=sk-... ./adr-insight serve

# Or with live UI editing (serves static files from disk)
ANTHROPIC_API_KEY=sk-... ./adr-insight serve --dev
```

Open [http://localhost:8081](http://localhost:8081).

### CLI Commands

| Command | Description |
|---------|-------------|
| `reindex` | Parse, embed, and store ADRs |
| `search` | Search indexed ADRs by similarity |
| `serve` | Start the HTTP server with web UI |
| `eval` | Evaluate answer quality against test cases |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--adr-dir` | `./docs/adr` | Directory containing ADR files |
| `--db` | `./adr-insight.db` | Path to SQLite database |
| `--ollama-url` | `http://localhost:11434` | Ollama API base URL |
| `--port` | `8081` | HTTP server port (serve only) |
| `--dev` | `false` | Serve static files from disk (serve only) |
| `--model` | `claude-sonnet-4-5` | Anthropic model (serve only) |
| `--top-k` | `5` | Number of search results (search only) |
| `--cases` | `./testdata/eval/cases.json` | Test case corpus (eval only) |
| `--baseline` | `./testdata/eval/baseline.json` | Baseline file (eval only) |
| `--save-baseline` | `false` | Save results as new baseline (eval only) |
| `--delta` | `0.2` | Max per-question score drop (eval only) |

### Environment Variables

All settings can be configured via environment variables. Flags take precedence when both are set.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP server port |
| `DB_PATH` | `./adr-insight.db` | SQLite database file path |
| `ADR_DIR` | `./docs/adr` | ADR markdown files directory |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama API base URL |
| `ANTHROPIC_API_KEY` | (none) | Anthropic API key (required for queries) |
| `EMBED_MODEL` | `nomic-embed-text` | Ollama embedding model name |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `LOG_FORMAT` | `json` | Log output format: json, text |
| `SLOW_REQUEST_MS` | `2000` | Threshold (ms) for slow request warnings |
| `SHUTDOWN_TIMEOUT_S` | `10` | Graceful shutdown drain timeout (seconds) |

### Quality Checks

```bash
make check    # lint + test + build (all three)
make lint     # golangci-lint only
make test     # tests only
make eval     # run evaluation harness (requires Ollama + API key)
make clean    # remove binary and database
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for the full system overview, data flows, and project structure.

### Key Components

- **Parser** — Extracts metadata and sections from ADR markdown files (goldmark)
- **Embedder** — Generates vector embeddings via Ollama (nomic-embed-text)
- **Store** — SQLite with sqlite-vec (vector) + FTS5 (keyword) for hybrid search
- **LLM** — Anthropic Claude for synthesizing answers from retrieved ADRs
- **RAG Pipeline** — Orchestrates embed → hybrid search → rerank → expand → synthesize
- **Web UI** — Vanilla HTML/CSS/JS with marked.js for markdown rendering

## ADRs

This project's own [Architecture Decision Records](docs/adr/) serve as both documentation and the demo dataset:

| ADR | Decision |
|-----|----------|
| [ADR-001](docs/adr/ADR-001-why-go.md) | Why Go |
| [ADR-002](docs/adr/ADR-002-anthropic-for-synthesis.md) | Anthropic for Synthesis |
| [ADR-003](docs/adr/ADR-003-local-embeddings-ollama.md) | Local Embeddings with Ollama |
| [ADR-004](docs/adr/ADR-004-sqlite-vector-storage.md) | SQLite Vector Storage |
| [ADR-005](docs/adr/ADR-005-gitea-primary-github-mirror.md) | Gitea Primary, GitHub Mirror |
| [ADR-006](docs/adr/ADR-006-ncruces-sqlite-driver.md) | ncruces/WASM Driver (superseded) |
| [ADR-007](docs/adr/ADR-007-goldmark-markdown-parsing.md) | Goldmark Markdown Parsing |
| [ADR-008](docs/adr/ADR-008-section-based-chunking.md) | Section-Based Chunking |
| [ADR-009](docs/adr/ADR-009-structured-llm-output.md) | Structured LLM Output |
| [ADR-010](docs/adr/ADR-010-full-adr-expansion.md) | Full ADR Expansion |
| [ADR-011](docs/adr/ADR-011-stdlib-http-server.md) | stdlib HTTP Server |
| [ADR-012](docs/adr/ADR-012-marked-js-client-rendering.md) | marked.js Client Rendering |
| [ADR-013](docs/adr/ADR-013-go-embed-static-files.md) | go:embed Static Files |
| [ADR-014](docs/adr/ADR-014-docker-multistage-debian.md) | Docker Multi-Stage Debian |
| [ADR-015](docs/adr/ADR-015-mattn-cgo-sqlite-driver.md) | mattn/CGO SQLite Driver (supersedes ADR-006) |
| [ADR-016](docs/adr/ADR-016-llm-judge-evaluation.md) | LLM-as-Judge Evaluation |
| [ADR-017](docs/adr/ADR-017-fts5-hybrid-search.md) | FTS5 Hybrid Search |
| [ADR-018](docs/adr/ADR-018-llm-keyword-extraction.md) | LLM Keyword Extraction for FTS |
| [ADR-019](docs/adr/ADR-019-deterministic-retrieval-metrics.md) | Deterministic Retrieval Metrics |
| [ADR-020](docs/adr/ADR-020-adr-relationship-model.md) | ADR Relationship Model |
| [ADR-021](docs/adr/ADR-021-llm-relationship-classification.md) | LLM Relationship Classification |
| [ADR-022](docs/adr/ADR-022-chunking-strategy-validated.md) | Chunking Strategy Validated in Phase 2 |
| [ADR-023](docs/adr/ADR-023-alpine-js-reactive-ui.md) | Alpine.js for Reactive UI |
| [ADR-024](docs/adr/ADR-024-slog-structured-logging.md) | slog Structured Logging |

## Author

Tyler Colbert
