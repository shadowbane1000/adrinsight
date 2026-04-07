# ADR-020: ADR Relationship Model and Storage

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

ADR Insight treats each ADR as an isolated document. However, ADRs naturally
form a graph: ADR-015 supersedes ADR-006, ADR-008 depends on ADR-007 for AST
walking, ADR-004 drives ADR-014's Docker build requirements. These
relationships are documented in "Related ADRs" sections but not parsed or
used by the system.

Without structured relationship data, the system cannot answer questions
about decision chains ("what happened with the SQLite driver decision?"),
cannot reliably deprioritize superseded ADRs, and cannot expand retrieval
to include related documents when one is relevant.

## Decision

Model ADR relationships as directed edges with five typed relationship
categories. Store them in a new `adr_relationships` table in SQLite,
rebuilt during reindex alongside chunks and embeddings.

The five relationship types:

| Type | Meaning |
|------|---------|
| `supersedes` | This ADR replaces the target's decision |
| `superseded_by` | This ADR's decision was replaced by the target |
| `depends_on` | This ADR's decision relies on the target |
| `drives` | This ADR's decision caused or motivated the target |
| `related_to` | Topically related, no directional dependency |

Graph traversal is unlimited depth with visited-node tracking to prevent
loops on circular references.

## Rationale

- **Fixed types over free-form**: Five types cover all patterns observed in
  the current 17 ADRs. Fixed types enable structured queries ("find all ADRs
  superseded by X"), UI grouping by relationship type, and programmatic
  reranker logic. Free-form labels would require ad-hoc parsing at query time.
- **Directed edges**: Directionality matters — "A supersedes B" is different
  from "B supersedes A." Each relationship is stored as a directed edge from
  source to target. Bidirectional queries work by searching both source and
  target columns.
- **SQLite table over in-memory only**: The relationship data should be
  available to the serve command without re-parsing ADR files. SQLite
  provides persistence with zero additional infrastructure.
- **Rebuilt during reindex**: Relationships are derived from ADR source files,
  following the existing pattern where all indexed data is disposable and
  regenerated via `reindex`.
- **Unlimited traversal with loop prevention**: The corpus is small (17 ADRs,
  ~20 edges). Depth limits would complicate the code without a real
  performance benefit. Visited-node tracking prevents infinite loops on any
  circular references.

## Consequences

### Positive
- Enables relationship-aware synthesis (trace decision chains in answers)
- Enables relationship expansion in retrieval (co-retrieve related ADRs)
- Provides authoritative supersession data for the reranker
- Supports clickable relationship navigation in the web UI
- Small, queryable graph fits naturally in SQLite

### Negative
- Adds a table and index to the database schema
- Relationship parsing adds time to reindex (bounded by LLM classification)
- Relationship types may need expansion as ADR patterns evolve

### Mitigations
- Table is small and rebuilt during reindex — no migration burden
- LLM classification is batched and uses Haiku (fast, cheap)
- Adding new relationship types is a one-line constant addition

## Alternatives Considered

### Free-form relationship labels
- **Rejected because:** Every consumer (reranker, UI, retrieval expansion)
  would need to interpret free-form text, leading to inconsistent behavior.
  Fixed types are a closed set that can be exhaustively handled.

### In-memory only (Go map)
- **Rejected because:** The serve command would need to re-parse ADR files
  on startup to reconstruct relationships. SQLite persistence follows the
  established pattern and is trivial to implement.

### Graph database (Neo4j, etc.)
- **Rejected because:** Massive overkill for ~20 edges. Adds a server
  dependency that violates Constitution Principle V (Developer Experience).
  SQLite with two indexes handles the query patterns needed.

### Bidirectional edges only
- **Rejected because:** Loses directionality. "A drives B" and "B drives A"
  are meaningfully different relationships that affect retrieval expansion
  and synthesis differently.

## Related ADRs

- ADR-004: SQLite with Vector Extension for Storage — the storage layer
  extended with the relationships table
- ADR-017: FTS5 Hybrid Search — relationships complement keyword and vector
  search with graph-aware retrieval expansion
