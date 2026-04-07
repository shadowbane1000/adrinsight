# ADR-019: Deterministic Retrieval Metrics

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

The M4a evaluation harness (ADR-016) measured retrieval quality by comparing
ADR numbers cited in the LLM's synthesized answer against expected ADR
numbers. This made retrieval metrics (precision, recall, F1) non-deterministic
— the same query produced different scores across runs because the LLM
synthesis step chooses which ADRs to cite based on its generation, not
on what was retrieved.

During M4b testing, the `single-binary-constraint` test case showed recall
varying from 0.40 to 1.00 across runs with identical retrieval results.
The retrieval was consistently finding the right ADRs, but the LLM sometimes
omitted citations for ADRs it received. This made it impossible to measure
whether retrieval changes actually improved retrieval — the noise from
synthesis obscured the signal.

## Decision

Separate retrieval measurement from synthesis measurement. The RAG pipeline
now records which ADRs were retrieved (before synthesis) in a `RetrievedADRs`
field on the query response. The eval harness computes precision, recall, and
F1 against `RetrievedADRs` rather than the LLM's `Citations`.

The LLM citations are still recorded as `CitedADRs` for synthesis quality
tracking, but they no longer affect retrieval metrics.

## Rationale

- **Deterministic measurement**: Retrieval is deterministic (same query
  embedding + same FTS index = same results). Measuring against retrieval
  output eliminates the LLM's non-determinism from retrieval scores.
- **Separate concerns**: Retrieval quality ("did we find the right ADRs?")
  and synthesis quality ("did the LLM cite them correctly?") are different
  questions that should be measured independently.
- **Actionable metrics**: When retrieval scores change, it's because
  retrieval changed — not because the LLM had a different generation.
  This makes A/B experiments (chunking, FTS weights, reranking) reliable.
- **Backward compatible**: The `CitedADRs` field preserves the old behavior.
  The `ReturnedADRs` field is kept for backward compatibility with existing
  baseline files.

## Consequences

### Positive
- Retrieval metrics are now fully deterministic — same results every run
- Can reliably A/B test retrieval changes (chunking, weights, reranking)
- Clear separation between retrieval and synthesis quality
- Eval report shows both Retrieved and Cited sets for each question

### Negative
- Retrieval precision appears lower because the system returns 5 ADRs
  regardless of relevance, and expected sets are often 1-4 ADRs
- Old baselines that mixed retrieval and synthesis scores are not directly
  comparable to new baselines

### Mitigations
- Save a fresh baseline after the metric change
- Precision can be improved by relationship-aware retrieval (M5) that
  returns more targeted results

## Alternatives Considered

### Run eval multiple times and average
- **Rejected because:** Adds cost (multiple LLM synthesis calls per question)
  and still doesn't isolate retrieval quality. The average would reduce
  variance but not eliminate it.

### Set LLM temperature to 0
- **Rejected because:** Anthropic's API doesn't guarantee determinism even
  at temperature 0 due to batching and infrastructure changes. Also, forcing
  temperature 0 for eval would not match the production behavior.

### Measure retrieval at the store layer (before pipeline)
- **Rejected because:** This would require duplicating the embed + search
  logic in the eval harness. Using the pipeline's own retrieval output is
  simpler and tests the actual code path.

## Related ADRs

- ADR-016: LLM-as-Judge Evaluation — this ADR refines the evaluation
  methodology introduced in ADR-016
- ADR-017: FTS5 Hybrid Search — the retrieval changes that motivated the
  need for deterministic measurement
