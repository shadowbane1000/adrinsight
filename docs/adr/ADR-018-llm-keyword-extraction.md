# ADR-018: LLM Keyword Extraction for FTS Query Filtering

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

Milestone 4b added FTS5 hybrid search (ADR-017) to combine keyword matching
with vector similarity. However, FTS5 MATCH queries with natural language
input produced noisy results. The initial approach — stripping stop words and
joining remaining terms with OR — kept too many generic English words
("decisions", "affected", "choice") that matched across all ADRs equally,
drowning out domain-specific terms like "sqlite" or "pgvector" that provided
real signal.

The core problem: FTS is most valuable when queries contain domain-specific
terms (technology names, library names, architectural patterns), but natural
language questions are mostly common English words. We needed a way to filter
FTS queries down to only the terms that matter.

## Decision

During reindex, use an LLM (Anthropic Haiku) to extract domain-specific
search terms from each ADR. Store these terms in a `keywords` table as a
vocabulary. At query time, filter the user's query to only include terms
present in the keyword vocabulary before sending to FTS5.

Multi-word keywords and keywords with punctuation (e.g., "net/http",
"docker-compose", "natural language") are tokenized into individual
alphanumeric tokens so they match the same tokenization applied to query
words.

## Rationale

- **Precision over recall for FTS**: The keyword vocabulary acts as an
  allow-list. Only domain-relevant terms reach FTS, so every FTS hit is
  meaningful. This is the inverse of the stop-word approach — instead of
  removing known-bad words, we keep only known-good words.
- **LLM extraction handles varied terminology**: Haiku extracts terms like
  "pgvector", "postgresql", "chromadb" from ADR-004's Alternatives section
  — terms that are critical for search but would be missed by simple
  heuristics. Testing showed Haiku captured 22 terms per ADR vs 2-10 for
  a local 3B model.
- **Index-time cost only**: The LLM runs during reindex (once), not at
  query time. For 17 ADRs, Haiku extraction takes ~25 seconds and costs
  ~$0.01. The vocabulary is stored in SQLite and loaded at query time with
  no API calls.
- **Graceful fallback**: If no vocabulary is stored (reindex without API
  key), FTS falls back to stop-word removal. The system works either way,
  just with lower FTS precision.

## Consequences

### Positive
- FTS queries contain only domain-specific terms, eliminating noise
- The sqlite-tradeoffs query went from R=0.50 to R=1.00 (both expected ADRs retrieved)
- Overall eval recall improved from 0.77 to 0.89
- Works with any phrasing — the LLM handles varied terminology without brittle keyword lists

### Negative
- Reindex now requires an Anthropic API key for optimal results
- Keywords are non-deterministic — different LLM runs may produce slightly different vocabularies
- Very rare or novel terms in a query may not be in the vocabulary

### Mitigations
- Fallback to stop-word removal when no vocabulary exists
- Vocabulary is rebuilt on every reindex, so new terms are captured when ADRs change
- The vocabulary is large enough (~300 tokens for 17 ADRs) to cover most query terms

## Alternatives Considered

### Stop-word removal only
- **Rejected because:** Generic English words like "decisions", "affected",
  "choice" passed through and matched everywhere, producing noisy FTS
  results that degraded hybrid search quality.

### Local LLM extraction (Ollama qwen2.5:0.5b, llama3.2:3b)
- **Rejected because:** Testing showed the 0.5B model produced only 2-8
  terms per ADR with hallucinations ("SQLAlchemy" for a Go project). The
  3B model was better (10 terms) but missed alternative technologies and
  used inconsistent casing ("PG Vector" vs "pgvector").

## Related ADRs

- ADR-017: FTS5 Hybrid Search — this ADR improves the FTS query quality
  that ADR-017 introduced
- ADR-016: LLM-as-Judge Evaluation — the eval harness that measured the
  improvement from keyword extraction
