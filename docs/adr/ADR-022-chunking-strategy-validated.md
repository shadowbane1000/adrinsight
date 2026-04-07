# ADR-022: Section-Based Chunking Validated in Phase 2

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None (amends ADR-008)

## Context

ADR-008 chose H2-section-based chunking as the initial strategy for Milestone 1
and explicitly deferred alternatives to Phase 2: "This decision will be revisited
in Phase 2 after real-world experience with retrieval quality." With the M4b
evaluation harness in place and real retrieval metrics available, we ran
controlled experiments comparing three chunking strategies across the full
evaluation suite.

## Decision

Retain H2-section-based chunking as the production strategy. The Phase 2
experiments confirmed it outperforms the alternatives tested. No change to the
chunking implementation.

### Experiment Results

| Strategy | Precision | Recall | F1 Score |
|----------|-----------|--------|----------|
| H2 sections (current) | 0.58 | 0.69 | 0.63 |
| Preamble only (title + status + context) | 0.58 | 0.60 | 0.59 |
| Whole document (one chunk per ADR) | 0.44 | 0.55 | 0.49 |

All experiments used the same evaluation cases, hybrid search weights (0.6
vector / 0.4 keyword), and LLM keyword vocabulary.

## Rationale

- **Sections produce the best retrieval quality:** F1 of 0.63 vs 0.59 (preamble)
  and 0.49 (whole-doc). The improvement comes primarily from recall — section
  chunks allow the retrieval system to surface the specific part of an ADR that
  answers the question.
- **Whole-doc dilutes semantic signal:** Embedding an entire ADR blends the
  Context, Decision, Rationale, and Consequences into a single vector. A query
  about "why was X chosen" matches less precisely when the vector also encodes
  unrelated consequences and alternatives.
- **Preamble loses detail:** Embedding only the title, status, and context
  section captures what the ADR is about but misses the decision rationale and
  consequences — the sections most likely to answer "why" and "what are the
  tradeoffs" questions.
- **No overlap still acceptable:** ADR-008 flagged lack of overlap between
  chunks as a known limitation. In practice, the hybrid search (vector +
  FTS5 keyword) compensates — keyword search catches exact-term matches that
  fall in a different section than the one vector search finds.

## Consequences

### Positive
- Closes the Phase 2 deferral from ADR-008 with empirical data
- Confirms the current implementation requires no changes
- Provides a baseline for future chunking experiments

### Negative
- Fixed-size token windows remain untested — they may outperform sections for
  ADRs with unusually long or short sections
- Sliding window with overlap was not tested — it may improve cross-section
  query matching
- Results are specific to this corpus (16 ADRs); a larger corpus could shift
  the balance

### Mitigations
- The `--chunk-strategy` flag added in M4b makes it easy to test new strategies
  without code changes
- The evaluation harness provides automated regression detection if a future
  strategy change hurts retrieval quality

## Alternatives Considered

### Whole-document embedding
- **Tested, rejected:** F1 of 0.49 — a 22% drop from sections. Semantic signal
  dilution confirmed the concern ADR-008 raised.

### Preamble-only embedding
- **Tested, rejected:** F1 of 0.59 — better than whole-doc but 6% worse than
  sections. Loses the detail in Decision/Rationale/Consequences sections.

### Fixed-size token windows with overlap
- **Deferred:** Not tested in this round. Would require token counting
  infrastructure. May be worth revisiting if the corpus grows significantly
  or if retrieval quality degrades with more ADRs.

### Semantic chunking (split by topic shifts)
- **Deferred:** Requires an additional embedding or classification pass per
  document. Overkill for structured ADR documents that already have clear
  section boundaries.

## Related ADRs

- ADR-008: Section-Based Chunking [related] — original Phase 1 decision; this ADR validates it with Phase 2 data
- ADR-017: FTS5 Hybrid Search [related] — hybrid search compensates for the no-overlap limitation of section chunking
- ADR-018: LLM Keyword Extraction [related] — keyword vocabulary improves FTS signal, which was part of the experiment environment
