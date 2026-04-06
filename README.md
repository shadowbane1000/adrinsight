# adrinsight

AI-powered search and reasoning over your Architecture Decision Records

This project was built and is maintained using spec driven development. A modified version of github speckit has been used. The speckit modifications support tracking and maintaining a directory of Architecture Decision Records (ADRs). These records then make up the sample data for this application. These ADRs can then be queried against to determine the reasons behind decisions that were made, and help in making future decisions.

## Prerequisites

- **Go 1.22+**
- **GCC / C toolchain** — required for CGO (SQLite + sqlite-vec)
- **libsqlite3-dev** — SQLite development headers
  ```
  # Debian/Ubuntu
  sudo apt-get install libsqlite3-dev
  ```
- **golangci-lint** (optional, for linting)
  ```
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **Ollama** with `nomic-embed-text` (for reindex/search, not needed for tests)
  ```
  ollama pull nomic-embed-text
  ```

## Running Tests

```
go test -short ./...
```

The `-short` flag skips integration tests that require a running Ollama instance. All unit tests run without external dependencies.

## Local Quality Checks

A Makefile provides the same checks that CI runs:

```
make check    # lint + test + build (all three in sequence)
make lint     # golangci-lint only
make test     # tests only
make build    # compile binary only
make clean    # remove binary and database file
```

## Usage

### Reindex ADRs

```
go run ./cmd/adr-insight reindex --adr-dir ./docs/adr
```

This parses all `ADR-*.md` files, generates embeddings via Ollama, and stores them in `adr-insight.db`.

### Search

```
go run ./cmd/adr-insight search "why did we choose Go?"
```

| Flag | Default | Description |
|------|---------|-------------|
| `--adr-dir` | `./docs/adr` | Directory containing ADR files (reindex only) |
| `--db` | `./adr-insight.db` | Path to SQLite database |
| `--ollama-url` | `http://localhost:11434` | Ollama API base URL |
| `--top-k` | `5` | Number of search results (search only) |

## Project Documentation

- [Architecture](docs/architecture.md) — system overview, data flows, project structure
- [Roadmap](docs/roadmap.md) — phased development plan
- [ADRs](docs/adr/) — architecture decision records (also the demo dataset)
