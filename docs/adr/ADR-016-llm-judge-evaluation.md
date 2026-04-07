# ADR-016: LLM-as-Judge for Answer Quality Evaluation

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert

## Context

ADR Insight's Phase 2 focuses on improving retrieval and answer quality.
Before making any changes to chunking, search, or synthesis, the project
needs a way to measure whether answers are getting better or worse. This
requires evaluating two things:

1. **Retrieval quality** — did the system find the right ADRs?
2. **Answer quality** — is the synthesized answer accurate and complete?

Retrieval quality can be measured mechanically by comparing returned citations
against a ground truth set. Answer quality is harder — it's inherently
subjective. A human could score answers, but that's not repeatable or
automatable. The question is how to score answer quality in an automated,
repeatable way that can run in CI.

## Decision

Use an LLM-as-judge to evaluate answer quality. A separate Anthropic Claude
call receives the question, the full text of the expected ADRs (ground truth),
and the system's answer, then scores on a structured rubric:

- **Accuracy (0-5)**: Does the answer correctly represent what the ADRs say?
- **Completeness (0-5)**: Does the answer address all aspects of the question?

The judge returns structured JSON (same OutputConfig pattern as ADR-009)
with integer scores and brief text justifications per dimension.

Mechanical retrieval metrics (precision, recall, F1) are computed separately
without LLM involvement, by comparing returned ADR numbers against expected
ADR numbers.

Regression is detected using a per-question delta model: any individual
question dropping more than a configurable threshold (default: 1 point) on
any dimension triggers a failure.

## Rationale

- **Automated and repeatable:** An LLM judge produces consistent-enough
  scores across runs to detect regressions. Minor variance (±0.5 points) is
  expected; the delta threshold absorbs this noise.
- **Two dimensions capture distinct failure modes:** An answer can be
  factually accurate but incomplete (misses sub-questions), or complete
  but inaccurate (hallucinated claims). A single overall score would mask
  which failure mode is present.
- **Ground truth anchoring:** The judge receives the actual ADR text, not
  just the system's answer. This grounds scoring in what the ADRs actually
  say, reducing the judge's tendency to accept plausible-sounding but
  incorrect answers.
- **Structured output guarantees parseable results:** Using Anthropic's
  OutputConfig with a JSON schema (same approach as ADR-009) ensures the
  judge always returns valid, parseable scores.
- **Constitution compliance:** Principle III (Skateboard First) — two
  scoring dimensions are the minimum that capture meaningful quality
  differences. More dimensions (relevance, coherence, citation formatting)
  can be added later if needed.

## Consequences

### Positive
- Answer quality is measurable and trackable across changes
- Regressions are caught automatically before merge (CI integration)
- Scoring rubric is explicit and auditable (the prompt is in code)
- Mechanical metrics (precision/recall) provide LLM-independent validation
- The evaluation harness becomes a forcing function for test case curation

### Negative
- Each evaluation run costs Anthropic API credits (one judge call per
  test case, on top of the synthesis call)
- LLM judge scores have inherent variance — the same answer may score
  differently across runs (mitigated by the delta threshold)
- The rubric quality depends on prompt engineering — a poorly written
  rubric produces unreliable scores
- Judge and synthesis use the same LLM provider, creating a potential
  bias (Claude judging Claude's output)

### Mitigations
- API cost is bounded by the test corpus size (6 initial questions =
  6 judge calls per run, ~$0.10-0.30 per evaluation)
- The delta threshold (default: 1 point) absorbs typical LLM variance
  while catching genuine regressions
- The rubric prompt will be iterated based on early results — if scores
  are inconsistent, the rubric needs tightening
- Same-provider bias is acknowledged but acceptable at this stage. A
  cross-provider judge (e.g., OpenAI judging Claude) could be added later
  if bias proves problematic.

## Alternatives Considered

### Human evaluation only
- **Rejected because:** Not automated or repeatable. Requires a human to
  read every answer after every change. Doesn't scale to CI integration.
  Useful as a periodic sanity check but not as the primary evaluation method.

### Automated metrics only (BLEU, ROUGE, BERTScore)
- **Rejected because:** These metrics measure surface-level text similarity,
  not factual accuracy or completeness. A rephrased correct answer would
  score poorly; a fluent but wrong answer could score well. They're designed
  for translation and summarization, not RAG evaluation.

### Single overall quality score (0-10)
- **Rejected because:** Collapses distinct failure modes into one number.
  An answer that's accurate but incomplete looks the same as one that's
  complete but inaccurate. Two dimensions are the minimum for actionable
  feedback.

### More scoring dimensions (relevance, coherence, citation quality, etc.)
- **Deferred:** Additional dimensions add evaluation cost and complexity.
  Starting with two dimensions (accuracy + completeness) is the skateboard.
  If these prove insufficient to diagnose quality issues during M4b retrieval
  experiments, more dimensions can be added.

## Related ADRs

- ADR-009: Structured LLM Output — the OutputConfig/JSON schema pattern
  reused for the judge's structured response
- ADR-008: Section-Based Chunking — the chunking strategy this evaluation
  harness will measure and help improve in M4b
