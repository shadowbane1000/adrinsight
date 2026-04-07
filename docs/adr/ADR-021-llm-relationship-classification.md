# ADR-021: LLM-Based Relationship Classification

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

ADR "Related ADRs" sections describe relationships in natural language with
varied phrasing. ADR-015 says "superseded by this ADR", ADR-008 says
"provides the AST walking that makes section splitting straightforward",
ADR-014 says "the CGO requirement drives the multi-stage build." A new
human-authored ADR might say "this replaces", "made obsolete by", or "built
on top of."

The system needs to classify each relationship into one of five types
(ADR-020) to enable structured queries, retrieval expansion, and UI grouping.
Keyword matching would miss varied phrasing; the classification must be
robust to natural language variation.

## Decision

Use an LLM (Anthropic Haiku) to classify each relationship bullet into one
of the five relationship types during reindex. The prompt provides the
source ADR title, the relationship bullet text, and the allowed type values.
The LLM returns a single type string.

Additionally, define a standardized format convention for new ADRs: each
bullet in "Related ADRs" should include an optional bracketed type tag
(e.g., `[supersedes]`). The LLM classifier serves as fallback for ADRs
without tags.

## Rationale

- **Handles varied phrasing**: The LLM correctly classifies "superseded by
  this ADR", "this replaces", and "made obsolete by" as `supersedes` without
  a brittle keyword list. This is critical for human-authored and legacy ADRs.
- **Proven pattern**: M4b already uses Haiku for keyword extraction during
  reindex (ADR-018). Relationship classification follows the same pattern —
  LLM at index time, structured data at query time.
- **Cheap and fast**: Haiku classifies all relationships for 17 ADRs in
  under 5 seconds at ~$0.001 cost. The classification task is simple (pick
  one of five options given context) and doesn't need a larger model.
- **Standardized format reduces LLM dependency over time**: As new ADRs
  use the `[type]` tag convention, fewer bullets need LLM classification.
  The LLM remains as fallback for backward compatibility.

## Consequences

### Positive
- Robust classification without brittle keyword matching
- Works with any natural language phrasing
- Standardized convention improves consistency in new ADRs
- No new runtime dependency — classification happens at index time only

### Negative
- Reindex requires Anthropic API key for relationship classification
- Classification is non-deterministic — different runs may produce different
  types for ambiguous relationships
- Adds latency to reindex (bounded: ~5s for 17 ADRs)

### Mitigations
- Without API key, relationships are stored with `related_to` as default type
- Ambiguous relationships (e.g., "related to" vs "depends on") have low
  impact — the most important type (supersedes/superseded_by) is the least
  ambiguous
- Standardized format tags bypass LLM entirely for well-formatted ADRs

## Alternatives Considered

### Keyword matching (regex on "supersedes", "drives", etc.)
- **Rejected because:** Phrasing varies across ADRs and would be even more
  varied in human-authored ADRs. A keyword list would need constant
  maintenance and would miss novel phrasing. This was the specific concern
  raised during specification review.

### Local LLM (Ollama)
- **Rejected because:** M4b keyword extraction testing showed local models
  (qwen2.5:0.5b, llama3.2:3b) miss nuances and hallucinate. For a simple
  5-way classification, Haiku is more reliable and the cost is negligible.

### Parse standardized format only (no LLM)
- **Rejected because:** The 17 existing ADRs don't use the standardized
  format. Requiring retroactive format changes to existing ADRs would violate
  Constitution Principle II-A (ADR Immutability). The LLM classifier handles
  existing ADRs without modification.

## Related ADRs

- ADR-020: ADR Relationship Model — defines the five types this classifier
  produces
- ADR-018: LLM Keyword Extraction — established the pattern of using Haiku
  at index time for domain-specific classification
