# Data Model: Foundation — ADR Indexing Pipeline

## Entities

### ADR (parsed from markdown)

| Field     | Type     | Source                           | Notes                           |
|-----------|----------|----------------------------------|---------------------------------|
| FilePath  | string   | File system path                 | Relative to ADR directory       |
| Number    | int      | Parsed from filename (ADR-NNN)   | Used for display and references |
| Title     | string   | H1 heading or filename           | e.g., "Why Go"                  |
| Date      | string   | Frontmatter `Date:` line         | ISO format YYYY-MM-DD           |
| Status    | string   | Frontmatter `Status:` line       | Accepted, Proposed, Deprecated  |
| Deciders  | string   | Frontmatter `Deciders:` line     | Optional                        |
| Body      | string   | Full markdown body below metadata| Used for chunking and embedding |

**Validation rules**:
- FilePath MUST not be empty
- Title MUST not be empty (fall back to filename if H1 missing)
- Status defaults to "Unknown" if not present
- Date defaults to empty string if not parseable

### Chunk (derived from ADR)

| Field      | Type     | Source                      | Notes                              |
|------------|----------|-----------------------------|------------------------------------|
| ADRID      | int      | ADR number                  | Links chunk back to source ADR     |
| SectionKey | string   | Heading text (e.g., "Context") | Identifies which ADR section    |
| Content    | string   | Section body text           | The text that gets embedded        |

**Chunking strategy** (Milestone 1 — simple):
- Split ADR body by H2 headings (## Context, ## Decision, etc.)
- Each section becomes one chunk
- If an ADR has no H2 headings, the entire body is one chunk
- Prefix each chunk with `search_document: ` before embedding

### StoredChunk (persisted in SQLite)

| Column     | Type          | Table       | Notes                           |
|------------|---------------|-------------|---------------------------------|
| id         | INTEGER PK    | chunks      | Auto-increment                  |
| adr_number | INTEGER       | chunks      | FK to ADR number                |
| adr_title  | TEXT          | chunks      | Denormalized for query results  |
| adr_status | TEXT          | chunks      | Denormalized for filtering      |
| adr_path   | TEXT          | chunks      | Denormalized for citations      |
| section    | TEXT          | chunks      | Section heading                 |
| content    | TEXT          | chunks      | Raw section text                |

### VectorEntry (sqlite-vec virtual table)

| Column    | Type          | Table       | Notes                           |
|-----------|---------------|-------------|---------------------------------|
| rowid     | INTEGER       | vec_chunks  | Matches chunks.id               |
| embedding | float[768]    | vec_chunks  | nomic-embed-text vector         |

## Relationships

```
ADR (file) ──1:N──► Chunk (section)
                        │
                        ▼
              StoredChunk (chunks table)
                        │
                  rowid = rowid
                        │
                        ▼
              VectorEntry (vec_chunks virtual table)
```

## SQLite Schema

```sql
CREATE TABLE IF NOT EXISTS chunks (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    adr_number INTEGER NOT NULL,
    adr_title  TEXT    NOT NULL,
    adr_status TEXT    NOT NULL DEFAULT 'Unknown',
    adr_path   TEXT    NOT NULL,
    section    TEXT    NOT NULL,
    content    TEXT    NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(
    embedding float[768]
);

-- rowid in vec_chunks corresponds to id in chunks
```

## Reindex Behavior

On each reindex run:
1. DROP TABLE chunks (if exists)
2. DROP TABLE vec_chunks (if exists)
3. Recreate both tables
4. Parse all ADR files → chunks → embed → insert

This ensures the index always reflects the current state of the ADR
directory. The database is a disposable, derived artifact.
