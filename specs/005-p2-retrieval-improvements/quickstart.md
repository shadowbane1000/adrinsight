# Quickstart: Retrieval Improvements — Integration Test Scenarios

## Prerequisites

- ADRs indexed (`./adr-insight reindex`)
- Ollama running with `nomic-embed-text`
- `ANTHROPIC_API_KEY` set
- M4a baseline saved (`testdata/eval/baseline.json`)

## Scenario 1: Baseline Comparison After Changes

```bash
# Reindex with updated chunking/search
./adr-insight reindex --adr-dir ./docs/adr

# Run eval and compare against baseline
./adr-insight eval

# Expected: All test cases pass, aggregate recall improved
```

## Scenario 2: Keyword Search Verification

```bash
# Search for a specific library name
./adr-insight search "goldmark"
# Expected: ADR-007 in top 3 results

./adr-insight search "ncruces"
# Expected: ADR-006 and/or ADR-015 in results

./adr-insight search "mattn"
# Expected: ADR-015 in top results
```

## Scenario 3: Hybrid Search Beats Pure Vector

```bash
# Query with exact technical term
# In web UI: "What role does sqlite-vec play?"
# Expected: ADR-004 appears in citations (keyword match on "sqlite-vec")

# Compare: pure vector might miss this if the embedding of "sqlite-vec"
# doesn't cluster near ADR-004's embedding
```

## Scenario 4: Reranking — Active Over Superseded

```bash
# Query about SQLite driver
# In web UI: "Which SQLite driver does the project use?"
# Expected: ADR-015 (Accepted) ranks above ADR-006 (Superseded)
```

## Scenario 5: Chunking Experiment Comparison

```bash
# For each chunking strategy:
# 1. Modify parser chunking logic
# 2. Reindex
# 3. Run eval with --output
./adr-insight eval --output testdata/eval/experiments/strategy-name.json

# Compare scores across experiments to select winner
```

## Scenario 6: No Regression After All Changes

```bash
./adr-insight eval
# Expected: RESULT: PASS
# No individual test case regresses beyond delta threshold
```
