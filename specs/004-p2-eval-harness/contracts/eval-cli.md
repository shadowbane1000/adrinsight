# CLI Contract: eval Command

## Usage

```
adr-insight eval [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--cases` | `./testdata/eval/cases.json` | Path to test case corpus |
| `--baseline` | `./testdata/eval/baseline.json` | Path to baseline file |
| `--save-baseline` | `false` | Save this run's results as the new baseline |
| `--delta` | `0.2` | Maximum allowed per-question score drop (0.0-1.0 scale) |
| `--db` | `./adr-insight.db` | Path to SQLite database |
| `--ollama-url` | `http://localhost:11434` | Ollama API base URL |
| `--adr-dir` | `./docs/adr` | ADR directory for full content reads |
| `--model` | `claude-sonnet-4-5` | Anthropic model for judge |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All test cases pass (or skipped due to unavailable services) |
| 1 | One or more test cases regressed below baseline threshold |
| 2 | Fatal error (missing test cases file, invalid JSON, etc.) |

## Output Format (stdout)

```
Evaluation Report
=================

Question: What tradeoffs were made by choosing SQLite...
  Citations: [4, 6, 15] (expected: [4, 6, 15])
  Precision: 1.00  Recall: 1.00  F1: 1.00
  Accuracy:  0.80 — "Correctly identifies SQLite tradeoffs..."
  Completeness: 0.60 — "Misses migration path discussion..."
  Status: PASS

Question: How do the embedding and synthesis components...
  Citations: [2, 3, 9] (expected: [2, 3, 9, 10])
  Precision: 1.00  Recall: 0.75  F1: 0.86
  Accuracy:  0.80 — "..."
  Completeness: 0.80 — "..."
  Status: PASS

...

Summary
-------
Questions: 6  Passed: 5  Regressed: 1  New: 0
Avg Precision: 0.89  Avg Recall: 0.83  Avg F1: 0.86
Avg Accuracy: 0.76  Avg Completeness: 0.70

RESULT: FAIL (1 regression detected)
```

## Baseline Save

When `--save-baseline` is passed:
1. Run the full evaluation
2. Save results to the baseline file path
3. Report "Baseline saved to {path}" in output
4. Exit code based on whether the run itself had errors (not regressions,
   since there's nothing to compare against)
