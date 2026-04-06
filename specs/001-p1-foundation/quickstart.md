# Quickstart: Foundation — Integration Test Scenarios

## Prerequisites

- Go 1.22+ installed
- Ollama running locally with `nomic-embed-text` model pulled:
  ```
  ollama pull nomic-embed-text
  ```

## Scenario 1: Reindex the project's own ADRs

```bash
# Build
go build -o adr-insight ./cmd/adr-insight

# Reindex
./adr-insight reindex --adr-dir ./docs/adr

# Expected: all ADR-*.md files parsed, embedded, stored
# Verify: adr-insight.db file created with data
```

**Verify with**:
```bash
./adr-insight search "why did we choose Go?"
# Expected: ADR-001 appears as top result
```

## Scenario 2: Search for a specific topic

```bash
./adr-insight search "vector storage"
# Expected: ADR-004 (SQLite Vector Storage) appears as top result
```

## Scenario 3: Search for an unrelated topic

```bash
./adr-insight search "microservice deployment"
# Expected: low-relevance results or empty output, no errors
```

## Scenario 4: Reindex with no ADRs

```bash
mkdir /tmp/empty-adrs
./adr-insight reindex --adr-dir /tmp/empty-adrs
# Expected: completes with "0 ADRs, 0 chunks" message
```

## Scenario 5: Invalid directory

```bash
./adr-insight reindex --adr-dir /nonexistent/path
# Expected: exit code 1, meaningful error message
```

## Scenario 6: Ollama unavailable

```bash
./adr-insight reindex --adr-dir ./docs/adr --ollama-url http://localhost:99999
# Expected: exit code 1, error about connection refused
```

## Scenario 7: Local quality checks

```bash
make lint    # Run golangci-lint
make test    # Run all tests
make build   # Compile binary
make check   # Run all three in sequence
```
