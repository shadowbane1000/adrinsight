# Data Model: ADR Intelligence

## Extended Types

### ADR (extended — `internal/parser/parser.go`)

Existing fields unchanged. New field:

| Field | Type | Description |
|-------|------|-------------|
| RelatedADRs | []RawRelationship | Parsed from "Related ADRs" H2 section. Contains target ADR number and raw description text. |

### RawRelationship (new — `internal/parser/parser.go`)

| Field | Type | Description |
|-------|------|-------------|
| TargetADR | int | ADR number referenced (extracted via regex) |
| Description | string | Full text of the relationship bullet (e.g., "superseded by this ADR") |

### ADRRelationship (new — `internal/store/store.go`)

| Field | Type | Description |
|-------|------|-------------|
| SourceADR | int | ADR number that declares the relationship |
| TargetADR | int | ADR number being referenced |
| RelType | string | One of: `supersedes`, `superseded_by`, `depends_on`, `drives`, `related_to` |
| Description | string | Free-text description from the ADR |

### RelationshipType Constants

| Value | Meaning | Example |
|-------|---------|---------|
| `supersedes` | This ADR replaces the target ADR's decision | ADR-015 supersedes ADR-006 |
| `superseded_by` | This ADR's decision was replaced by the target | ADR-006 superseded_by ADR-015 |
| `depends_on` | This ADR's decision relies on the target's decision | ADR-008 depends_on ADR-007 |
| `drives` | This ADR's decision caused or motivated the target | ADR-004 drives ADR-014 |
| `related_to` | Topically related, no directional dependency | General cross-reference |

## Database Schema

### New Table: `adr_relationships`

```sql
CREATE TABLE adr_relationships (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    source_adr  INTEGER NOT NULL,
    target_adr  INTEGER NOT NULL,
    rel_type    TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT ''
)
```

Created in `Reset()` alongside existing tables. Dropped and rebuilt during reindex.

### Index

```sql
CREATE INDEX idx_rel_source ON adr_relationships(source_adr)
CREATE INDEX idx_rel_target ON adr_relationships(target_adr)
```

## Store Interface Extensions

New methods on the `Store` interface:

```
StoreRelationships(ctx, relationships []ADRRelationship) error
GetRelationships(ctx, adrNumber int) ([]ADRRelationship, error)
GetAllRelationships(ctx) ([]ADRRelationship, error)
```

- `StoreRelationships`: Bulk insert after reindex parsing. Clears existing relationships first.
- `GetRelationships`: Returns all relationships where the given ADR is either source or target.
- `GetAllRelationships`: Returns the full graph. Used by the reranker to build an in-memory lookup.

## API Extensions

### Enhanced `GET /adrs/{number}` Response

Add a `relationships` field to the existing `adrDetailResponse`:

```json
{
  "number": 15,
  "title": "Switch to mattn/go-sqlite3 (CGO) as SQLite Driver",
  "status": "Accepted",
  "date": "2026-04-06",
  "content": "...",
  "relationships": [
    {
      "target_adr": 6,
      "target_title": "ncruces/go-sqlite3 (WASM) as SQLite Driver",
      "rel_type": "supersedes",
      "description": "superseded by this ADR"
    },
    {
      "target_adr": 4,
      "target_title": "SQLite with Vector Extension for Storage",
      "rel_type": "depends_on",
      "description": "establishes sqlite-vec as the storage layer"
    }
  ]
}
```

## LLM Classification Prompt

Input to Haiku for each relationship bullet:

```
Classify the relationship between these two ADRs.

Source ADR: "{source_title}"
Related ADR bullet: "{bullet_text}"

Respond with exactly one of: supersedes, superseded_by, depends_on, drives, related_to
```

Expected output: a single word from the allowed set.

## Reranker Enhancement

The `DefaultReranker` currently detects superseded ADRs by scanning chunk content for the string "superseded" or "deprecated" (lines 53-57 of rerank.go). This will be replaced with a lookup against the relationship store:

- Accept a `relationships map[int][]ADRRelationship` parameter (or load it once at startup)
- For each result, check if the ADR has a `superseded_by` relationship
- If so, apply the status penalty (existing -0.1 default)
- Additionally, if both the superseded and superseding ADR are in results, boost the superseding one

## Pipeline Enhancement

In `Pipeline.Query()`, after retrieval and reranking:

1. Load relationships for all retrieved ADRs
2. Expand the result set: for each retrieved ADR, add directly-related ADRs not already present (1-hop expansion). For supersession chains, walk the full chain.
3. Cap at `topK * 2` total ADRs
4. Build a relationship summary block prepended to the LLM context:
   ```
   ## Relationship Context
   - ADR-015 supersedes ADR-006
   - ADR-004 drives ADR-014
   - ADR-006 superseded_by ADR-015
   ```
5. Continue with existing full-ADR expansion and synthesis
