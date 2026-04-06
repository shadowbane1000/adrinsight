# ADR-002: Use SQLite for Storage

**Status:** Accepted  
**Date:** 2026-02-01  
**Deciders:** Alice

## Context

The service needs a lightweight storage solution for metadata and vector
embeddings. The expected data volume is small (hundreds of records).

## Decision

Use SQLite with the sqlite-vec extension.

## Rationale

SQLite is embedded, requires no server, and the database file is easily
disposable and regenerable. sqlite-vec adds vector similarity search.

## Consequences

### Positive
- No database server to manage
- Single file, easy to backup or delete

### Negative
- Single writer limitation at scale
- sqlite-vec is a young extension
