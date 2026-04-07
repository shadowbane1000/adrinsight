# ADR-017: FTS5 Hybrid Search with Weighted Merge

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

ADR Insight's retrieval relies solely on vector similarity search via
sqlite-vec. While effective for semantic queries ("what tradeoffs were made
by choosing SQLite?"), pure vector search misses exact keyword matches. A
query for "goldmark" may not retrieve ADR-007 (Goldmark Markdown Parsing) if
the embedding of "goldmark" doesn't cluster near the ADR's embedding. This
was observed during M4a evaluation — technical terms, library names, and
acronyms produced lower recall than expected.

The system needs a way to find ADRs by exact keyword match in addition to
semantic similarity, without replacing vector search or adding external
dependencies.

## Decision

Add SQLite FTS5 full-text search as a second retrieval path alongside
sqlite-vec vector search. Results from both methods are combined using a
configurable weighted merge:

1. Run vector search (sqlite-vec) → normalized scores
2. Run keyword search (FTS5 BM25) → normalized scores
3. Merge results: `combined = w_vec * vec_score + w_kw * kw_score`
4. Deduplicate by ADR number, keep highest combined score
5. Return top-K by combined score

Default weights: 0.7 vector + 0.3 keyword. Both vector and BM25 scores
are normalized to 0-1 range before combination.

## Rationale

- **FTS5 is built in:** mattn/go-sqlite3 includes FTS5 by default with CGO
  enabled. No new dependencies, no external services. A new `fts_chunks`
  virtual table is added alongside existing `chunks` and `vec_chunks` tables.
- **Complementary strengths:** Vector search finds semantically related
  content even with different terminology. Keyword search finds exact term
  matches that embeddings may miss. The weighted merge captures both.
- **Configurable weights:** The 0.7/0.3 default can be tuned based on
  evaluation harness results. If keyword search proves less useful, reduce
  its weight without removing the capability.
- **Same transaction:** FTS5 table is populated in the same `StoreChunks`
  transaction as `chunks` and `vec_chunks`, keeping all three in sync.
- **Constitution compliance:** Principle III (Skateboard First) — FTS5 is
  the simplest keyword search that works. Principle I (Interfaces Everywhere)
  — the Store interface is extended with hybrid search methods.

## Consequences

### Positive
- Queries with exact technical terms (library names, acronyms) reliably
  find the corresponding ADRs
- No new external dependencies — FTS5 is part of SQLite
- Weights are tunable per evaluation harness feedback
- Falls back gracefully — if FTS5 returns no matches, vector results are
  used alone (keyword weight effectively becomes 0)

### Negative
- Database size increases slightly (FTS5 index alongside vec index)
- `StoreChunks` writes to three tables instead of two (marginal performance
  impact at this scale)
- Score normalization adds complexity — BM25 and vector distance are on
  different scales and must be carefully mapped to 0-1
- Two search paths means two potential failure modes to handle

### Mitigations
- Database size is negligible for 16 ADRs / ~80 chunks
- Score normalization uses min-max within each result set — simple and
  effective at this scale
- If FTS5 query fails, fall back to vector-only (degraded but functional)
- The evaluation harness catches any regression from the hybrid approach

## Alternatives Considered

### External search engine (Elasticsearch, Meilisearch)
- **Rejected because:** Adds an external service dependency, Docker container,
  and operational complexity. Overkill for 16 ADRs. Violates Principle III
  (Skateboard First) and Principle V (Developer Experience).

### LIKE queries on content column
- **Rejected because:** No relevance ranking. `LIKE '%goldmark%'` returns
  matches but can't score or rank them. FTS5 provides BM25 ranking for free.

### Keyword filter on vector results (not independent retrieval)
- **Rejected because:** Only works when vector search already retrieves the
  right ADRs. The whole point is catching ADRs that vector search misses.

### Replace vector search with FTS5 entirely
- **Rejected because:** FTS5 can't handle semantic similarity. "What are the
  tradeoffs of our storage choice?" wouldn't match ADR-004 unless it
  contains the word "tradeoff." Vector search is essential for natural
  language queries.

## Related ADRs

- ADR-004: SQLite with Vector Extension — establishes sqlite-vec as the
  vector search layer; this ADR adds FTS5 alongside it
- ADR-015: mattn/go-sqlite3 (CGO) — the driver that provides FTS5 support
