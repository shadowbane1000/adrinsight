# Research: Retrieval Improvements

## Hybrid Search via SQLite FTS5

**Decision**: Add FTS5 full-text search alongside sqlite-vec vector search,
combine results with configurable weighted merge.

**Rationale**: FTS5 is built into mattn/go-sqlite3 (no additional dependency).
It uses BM25 ranking and supports boolean queries, phrase matching, and
tokenization. FTS5 and sqlite-vec coexist as independent virtual tables in
the same database — no conflicts.

Implementation approach:
- Add `fts_chunks` virtual table: `CREATE VIRTUAL TABLE fts_chunks USING fts5(content)`
- Populate during `StoreChunks` alongside `chunks` and `vec_chunks` (same rowid)
- Query: `SELECT rowid, rank FROM fts_chunks WHERE fts_chunks MATCH ? ORDER BY rank`
- BM25 returns negative floats (closer to 0 = better match)
- Normalize BM25 scores to 0-1 range for combination with vector scores
- Weighted merge: `combined = w_vec * vec_score + w_kw * kw_score` (default 0.7/0.3)

**Alternatives considered**:
- External search engine (Elasticsearch, Meilisearch) — overkill for 16 ADRs,
  violates Principle III (Skateboard First)
- LIKE queries on content column — no ranking, poor relevance
- Trigram search — SQLite doesn't support natively; FTS5 covers the use case

## Chunking Strategy Experiments

**Decision**: Evaluate three alternatives against the M4a baseline, select
the best performer.

**Strategies to evaluate**:

1. **Current: H2-section splitting** (ADR-008 baseline)
   - Pros: Natural ADR boundaries, semantically meaningful
   - Cons: No overlap, very short/long sections, queries spanning sections miss

2. **H2-section + whole-document embedding** (hybrid)
   - Add a whole-ADR embedding alongside section chunks
   - Pros: Catches broad queries; section chunks catch specific queries
   - Cons: Increases chunk count by ~16 (one per ADR), more storage

3. **H2-section with overlap preamble**
   - Prepend each section chunk with the ADR title + first line of Context
   - Pros: Every chunk carries enough context to identify which ADR it's from
   - Cons: Increases embedding input size, may dilute section-specific signal

4. **Fixed-size token windows with overlap** (deferred from ADR-008)
   - Split into ~200 token windows with 50 token overlap
   - Pros: Uniform chunk sizes, handles varying section lengths
   - Cons: Breaks semantic boundaries, harder to map back to sections

**Evaluation approach**: For each strategy, reindex, run eval harness, record
scores. Compare precision, recall, F1, and judge scores. Document results in
`testdata/eval/experiments/`.

**Alternatives considered (not evaluating)**:
- Semantic paragraph grouping — requires an LLM or embedding similarity
  to determine paragraph boundaries; adds too much complexity for this corpus
- Recursive character splitting (LangChain-style) — designed for arbitrary
  documents, not structured ADRs with clear headings

## Score Normalization for Hybrid Merge

**Decision**: Normalize both vector and FTS5 scores to 0-1 range before
combining.

**Rationale**: Vector distance (sqlite-vec) and BM25 rank (FTS5) are on
completely different scales. Normalization ensures the weights are meaningful.

- **Vector scores**: sqlite-vec returns distance (lower = better). Normalize
  using `1 - (distance / max_distance)` from the result set, or use a fixed
  max distance threshold.
- **FTS5 scores**: BM25 returns negative floats (closer to 0 = better).
  Normalize using `1 - (abs(rank) / max_abs_rank)` from the result set.

Min-max normalization within each result set is simple and effective for a
small corpus. For larger corpora, a fixed normalization range would be more
stable.

## Reranking Heuristics

**Decision**: Simple, deterministic reranking applied after hybrid search.

**Heuristics**:

1. **Title match boost**: If the query contains words that appear in the ADR
   title, boost the score by a configurable factor (e.g., +0.2). Catches
   cases like "SQLite driver" matching ADR-015's title.

2. **Status deprioritization**: ADRs with status "Superseded" or "Deprecated"
   get a score penalty (e.g., -0.1). Their successors should rank higher.
   This uses the adr_status field already stored in chunks.

3. **Section relevance**: For queries containing "why" or "rationale", boost
   chunks from Rationale/Decision sections. For "alternative" queries, boost
   Alternatives Considered sections.

**Alternatives considered**:
- Cross-encoder reranking (e.g., a separate embedding model that scores
  query-document pairs) — too complex for this scale, violates Principle III
- LLM-based reranking — expensive, slow, and overkill for 16 ADRs
