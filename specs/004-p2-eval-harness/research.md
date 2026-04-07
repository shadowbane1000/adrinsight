# Research: Evaluation Harness

## LLM-as-Judge Rubric Design

**Decision**: Structured JSON rubric with two dimensions scored 0-5

**Rationale**: The judge receives the question, the full text of the expected
ADRs (ground truth), and the system's answer. It scores on two dimensions:

- **Accuracy (0-5)**: Does the answer correctly represent what the ADRs say?
  Penalize hallucinated claims, misattributed decisions, or factual errors
  about what was decided and why.
- **Completeness (0-5)**: Does the answer address all aspects of the question?
  Penalize missing key facts, unanswered sub-questions, or superficial
  treatment of complex topics.

The judge returns structured JSON (using Anthropic's OutputConfig, same
pattern as ADR-009) with scores and brief justifications.

**Alternatives considered**:
- Single overall score — too coarse; a factually accurate but incomplete
  answer would score the same as a complete but inaccurate one
- More dimensions (relevance, coherence, citation quality) — adds complexity
  without clear value at this stage; can be added later
- Human evaluation — not scalable or repeatable; LLM judge provides consistent
  baseline even if imperfect

## Mechanical Scoring Approach

**Decision**: Standard information retrieval metrics on citation sets

**Rationale**: Precision, recall, and F1 are well-understood IR metrics.
They measure citation quality without any LLM involvement:

- **Precision**: |returned ∩ expected| / |returned| — are the citations relevant?
- **Recall**: |returned ∩ expected| / |expected| — are all relevant ADRs cited?
- **F1**: 2 × (precision × recall) / (precision + recall) — balanced measure

Comparison is by ADR number (integer match), not by title or content.

**Alternatives considered**:
- NDCG (normalized discounted cumulative gain) — considers ranking order,
  but our citations are unordered sets. Deferred.
- MAP (mean average precision) — same issue, ranking-aware metric for an
  unordered output

## Test Case Format

**Decision**: JSON array with structured fields per test case

**Rationale**: JSON is natively supported in Go (encoding/json), human-readable
with formatting, and easy to validate. Each test case includes:

```json
{
  "id": "sqlite-tradeoffs",
  "question": "What tradeoffs were made by choosing SQLite...",
  "expected_adrs": [4, 6, 15],
  "key_facts": [
    "SQLite was chosen for single-file simplicity",
    "sqlite-vec enables vector search",
    "CGO required due to WASM incompatibilities"
  ]
}
```

Key facts are short assertions the answer should contain, used by the LLM
judge to evaluate completeness.

**Alternatives considered**:
- YAML — slightly more readable but requires a third-party Go parser; JSON
  is stdlib-only (Constitution Principle IV)
- TOML — similar issue, less common for array-heavy data
- Go test tables — not easily editable by non-developers, not inspectable
  outside of code

## Baseline Storage and Comparison

**Decision**: JSON file with per-question scores, delta-based regression

**Rationale**: The baseline is a JSON file (`testdata/eval/baseline.json`)
containing the score snapshot from a known-good state. Each subsequent eval
run compares per-question scores against the baseline. A regression is
detected when any individual question's score drops by more than a
configurable delta (default: 1 point on any dimension).

New test cases (present in corpus but absent from baseline) are scored and
reported but do not trigger regression failure.

**Alternatives considered**:
- Database storage — overkill for 6-50 test cases
- Git-tracked results with diff — interesting for history but harder to
  automate threshold checking
- Rolling average baseline — smooths LLM judge variance but complicates
  the "first run" bootstrapping

## CI Integration Strategy

**Decision**: Add `eval` target to Makefile, optional in CI based on service availability

**Rationale**: The eval command requires Ollama (for embeddings) and Anthropic
API (for both the RAG pipeline and the LLM judge). In CI, these may not be
available. The eval command should detect missing services and skip gracefully
with a warning and exit code 0 (not a failure). When services are available,
it runs fully and fails on regressions.

The Makefile gets an `eval` target. CI config gets an optional eval step.

**Alternatives considered**:
- Mock services in CI — defeats the purpose of end-to-end evaluation
- Require services in CI — makes CI dependent on external services,
  fragile and slow
- Separate eval pipeline — adds operational complexity for a small project
