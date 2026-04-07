# Data Model: RAG Pipeline and HTTP API

## New Entities

### QueryRequest (HTTP input)

| Field | Type   | Source     | Notes                          |
|-------|--------|------------|--------------------------------|
| Query | string | HTTP body  | Natural-language question text |

**Validation**: Query MUST not be empty or whitespace-only.

### QueryResponse (HTTP output)

| Field     | Type       | Source      | Notes                              |
|-----------|------------|-------------|------------------------------------|
| Answer    | string     | LLM output  | Synthesized natural-language answer |
| Citations | []Citation | LLM output  | Structured references to ADRs      |

### Citation (part of QueryResponse)

| Field     | Type   | Source     | Notes                              |
|-----------|--------|------------|------------------------------------|
| ADRNumber | int    | LLM output | e.g., 1 for ADR-001               |
| Title     | string | LLM output | e.g., "Why Go"                     |
| Section   | string | LLM output | e.g., "Rationale" (optional)       |

### ADRSummary (list endpoint output)

| Field  | Type   | Source         | Notes                       |
|--------|--------|----------------|-----------------------------|
| Number | int    | Store/chunks   | ADR number                  |
| Title  | string | Store/chunks   | ADR title                   |
| Status | string | Store/chunks   | Accepted, Proposed, etc.    |
| Path   | string | Store/chunks   | File path for retrieval     |

### ADRDetail (single ADR endpoint output)

| Field   | Type   | Source          | Notes                      |
|---------|--------|-----------------|----------------------------|
| Number  | int    | Store/chunks    | ADR number                 |
| Title   | string | Store/chunks    | ADR title                  |
| Status  | string | Store/chunks    | Status                     |
| Date    | string | Parser          | From frontmatter           |
| Content | string | Filesystem read | Full markdown body          |

## LLM Interface

```
LLM interface:
  Synthesize(ctx, query string, adrContents []ADRContext) → (QueryResponse, error)
```

### ADRContext (input to LLM)

| Field   | Type   | Source           | Notes                          |
|---------|--------|------------------|--------------------------------|
| Number  | int    | Search result    | ADR number                     |
| Title   | string | Search result    | ADR title                      |
| Content | string | Filesystem read  | Full markdown body of the ADR  |

## RAG Pipeline Flow

```
QueryRequest
    │
    ▼
Embed query (Embedder, "search_query:" prefix)
    │
    ▼
Search (Store, top-K chunks)
    │
    ▼
Deduplicate by ADR number
    │
    ▼
Read full ADR files from disk (using ADRPath)
    │
    ▼
Build ADRContext[] for each unique ADR
    │
    ▼
Synthesize (LLM, query + ADRContexts → QueryResponse)
    │
    ▼
Return QueryResponse (answer + citations)
```

## Existing Entities (from M1, unchanged)

- **SearchResult** — returned by Store.Search(), contains ADRNumber, ADRPath
  used to identify which ADRs to expand
- **ChunkRecord** — storage format, unchanged

## Store Extension

The Store interface gains one method for the list/detail endpoints:

```
ListADRs(ctx) → ([]ADRSummary, error)
```

This queries the chunks table for distinct ADR metadata. The single-ADR
detail endpoint reads from the filesystem using the ADR path, not the store.
