# Quickstart: RAG Pipeline — Integration Test Scenarios

## Prerequisites

- Go 1.24+ installed
- Ollama running with `nomic-embed-text` model
- ADRs indexed (`./adr-insight reindex --adr-dir ./docs/adr`)
- `ANTHROPIC_API_KEY` environment variable set

## Scenario 1: Ask about a specific ADR topic

```bash
./adr-insight serve &
curl -s -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Why did we choose Go?"}' | jq .

# Expected: answer mentions Go, citations include ADR-001
```

## Scenario 2: Ask about multiple ADRs

```bash
curl -s -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What are the key technology choices?"}' | jq .

# Expected: answer covers multiple decisions, citations from multiple ADRs
```

## Scenario 3: Ask about an unrelated topic

```bash
curl -s -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{"query": "How do we deploy to Kubernetes?"}' | jq .

# Expected: answer indicates no relevant ADRs found, no hallucinated info
```

## Scenario 4: Empty query

```bash
curl -s -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{"query": ""}' | jq .

# Expected: 400 error with "query is required"
```

## Scenario 5: List all ADRs

```bash
curl -s http://localhost:8081/adrs | jq .

# Expected: array of all indexed ADRs with number, title, status, path
```

## Scenario 6: Get a single ADR

```bash
curl -s http://localhost:8081/adrs/1 | jq .

# Expected: full ADR-001 content with metadata
```

## Scenario 7: Get nonexistent ADR

```bash
curl -s http://localhost:8081/adrs/999 | jq .

# Expected: 404 with "ADR 999 not found"
```

## Scenario 8: Synthesis service unavailable

```bash
ANTHROPIC_API_KEY=invalid ./adr-insight serve --port 8081 &
curl -s -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{"query": "test"}' | jq .

# Expected: 500 error about synthesis service, within 5 seconds
```
