# Data Model: Evaluation Harness

## Entities

### TestCase

A single evaluation question with ground truth expectations.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier (kebab-case slug, e.g., "sqlite-tradeoffs") |
| question | string | The natural-language question to submit |
| expected_adrs | []int | ADR numbers that should appear in citations |
| key_facts | []string | Short assertions the answer should contain |

Stored in: `testdata/eval/cases.json` (array of TestCase)

### EvalResult

The scored output for a single test case in one evaluation run.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Matches TestCase.id |
| question | string | The question that was asked |
| answer | string | The system's raw answer text |
| returned_adrs | []int | ADR numbers the system cited |
| precision | float64 | |returned ∩ expected| / |returned| (0.0-1.0) |
| recall | float64 | |returned ∩ expected| / |expected| (0.0-1.0) |
| f1 | float64 | 2 × (precision × recall) / (precision + recall) (0.0-1.0) |
| accuracy | float64 | LLM judge score normalized to 0.0-1.0 (judge scores 0-5 internally, divided by 5) |
| completeness | float64 | LLM judge score normalized to 0.0-1.0 (judge scores 0-5 internally, divided by 5) |
| accuracy_reason | string | Brief justification from the LLM judge |
| completeness_reason | string | Brief justification from the LLM judge |

### Baseline

A snapshot of EvalResults from a known-good state, used for regression
detection.

| Field | Type | Description |
|-------|------|-------------|
| created_at | string | ISO 8601 timestamp when baseline was captured |
| delta_threshold | float64 | Maximum allowed per-question score drop (default: 0.2) |
| results | []EvalResult | Per-question scores at baseline time |

Stored in: `testdata/eval/baseline.json`

### RunReport

Aggregate output from a single evaluation run.

| Field | Type | Description |
|-------|------|-------------|
| timestamp | string | ISO 8601 timestamp of this run |
| results | []EvalResult | Per-question scores |
| aggregate | AggregateScores | Averages across all questions |
| regressions | []Regression | List of test cases that regressed |
| new_cases | []string | Test case IDs with no baseline entry |

### AggregateScores

| Field | Type | Description |
|-------|------|-------------|
| avg_precision | float64 | Mean precision across all questions |
| avg_recall | float64 | Mean recall across all questions |
| avg_f1 | float64 | Mean F1 across all questions |
| avg_accuracy | float64 | Mean LLM accuracy score (0.0-1.0) |
| avg_completeness | float64 | Mean LLM completeness score (0.0-1.0) |

### Regression

| Field | Type | Description |
|-------|------|-------------|
| id | string | Test case ID |
| dimension | string | Which score regressed (e.g., "accuracy", "recall") |
| baseline_score | float64 | Score in the baseline |
| current_score | float64 | Score in this run |
| delta | float64 | Drop amount |
