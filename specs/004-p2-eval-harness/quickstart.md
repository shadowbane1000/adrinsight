# Quickstart: Evaluation Harness — Integration Test Scenarios

## Prerequisites

- ADRs indexed (`./adr-insight reindex` or auto-reindex via serve)
- Ollama running with `nomic-embed-text` model
- `ANTHROPIC_API_KEY` environment variable set

## Scenario 1: First Evaluation Run (Create Baseline)

```bash
./adr-insight eval --save-baseline

# Expected: Runs all 6 test cases, prints scores for each, saves baseline
# Output ends with: "Baseline saved to ./testdata/eval/baseline.json"
# Exit code: 0
```

## Scenario 2: Subsequent Evaluation (Compare Against Baseline)

```bash
./adr-insight eval

# Expected: Runs all test cases, compares against baseline
# If all pass: "RESULT: PASS" and exit code 0
# If regression: "RESULT: FAIL (N regressions detected)" and exit code 1
```

## Scenario 3: Detect a Regression

```bash
# Artificially degrade: reduce top-K to 1 in the pipeline
# Then run eval
./adr-insight eval

# Expected: Several questions regress (fewer citations returned)
# Output shows REGRESSED status for affected questions
# Exit code: 1
```

## Scenario 4: Add a New Test Case

```bash
# Edit testdata/eval/cases.json, add a new entry
# Run eval
./adr-insight eval

# Expected: New test case is scored and reported as "NEW (no baseline)"
# Does not trigger regression failure
# Exit code: 0 (assuming no other regressions)
```

## Scenario 5: Missing Services

```bash
# Stop Ollama, then run eval
./adr-insight eval

# Expected: Warning message about unavailable embedding service
# Skips evaluation gracefully
# Exit code: 0 (not a failure — services unavailable)
```

## Scenario 6: Missing Baseline

```bash
# Delete baseline.json, then run eval
./adr-insight eval

# Expected: Warning that no baseline exists
# Runs evaluation and reports scores
# Suggests running with --save-baseline to create one
# Exit code: 0 (no baseline to regress against)
```
