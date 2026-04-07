# Data Model: Retrieval Improvements

## Schema Changes

### New: FTS5 Virtual Table

```sql
CREATE VIRTUAL TABLE fts_chunks USING fts5(content);
```

Populated alongside `chunks` and `vec_chunks` during `StoreChunks`. Uses
matching rowids for cross-referencing with the `chunks` metadata table.

### Unchanged Tables

- `chunks` — metadata table (adr_number, adr_title, adr_status, adr_path, section, content)
- `vec_chunks` — sqlite-vec virtual table for vector similarity search

## New Store Methods

### SearchFTS

Full-text keyword search using FTS5 BM25 ranking.

| Parameter | Type | Description |
|-----------|------|-------------|
| query | string | Search terms (FTS5 MATCH syntax) |
| topK | int | Maximum results to return |
| **Returns** | []SearchResult | Same type as vector Search, with BM25-normalized score |

### HybridSearch

Combines vector and FTS5 results with weighted merge.

| Parameter | Type | Description |
|-----------|------|-------------|
| queryVec | []float32 | Embedded query vector |
| queryText | string | Raw query text for FTS5 |
| topK | int | Maximum final results |
| vecWeight | float64 | Weight for vector score (default 0.7) |
| kwWeight | float64 | Weight for keyword score (default 0.3) |
| **Returns** | []SearchResult | Merged, deduplicated, ranked by combined score |

## Reranking Types

### RerankConfig

| Field | Type | Description |
|-------|------|-------------|
| TitleBoost | float64 | Score boost for ADRs whose title matches query words (default 0.2) |
| StatusPenalty | float64 | Score penalty for superseded/deprecated ADRs (default 0.1) |
| SectionBoost | float64 | Score boost for section-relevant matches (default 0.1) |

### Reranker Interface

```
Rerank(query string, results []SearchResult, config RerankConfig) []SearchResult
```

Takes search results and re-scores them based on heuristics, returning
results in new rank order.
