# ADR-004: SQLite with Vector Extension for Storage

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight needs to store vector embeddings alongside ADR metadata (title, date, status, file path, relationships) and support similarity search at query time. The storage solution must:

- Support vector similarity search (cosine or dot product)
- Store structured metadata for browsing and filtering
- Run locally with minimal setup (no external database server)
- Fit naturally into a `docker-compose` workflow
- Handle a corpus of dozens to low hundreds of documents (not millions)

## Decision

Use SQLite with the `sqlite-vec` extension for both vector storage and structured metadata.

## Rationale

- **Single dependency:** SQLite is embedded — it's a file, not a server. No additional container in `docker-compose.yml`, no connection management, no port conflicts. The database is a single file on disk that can be deleted and regenerated from the source ADRs at any time.
- **Vectors and metadata together:** `sqlite-vec` adds vector similarity search to SQLite, allowing a single query to filter by metadata (e.g., status = "accepted") and rank by semantic similarity. This avoids the complexity of coordinating between a vector store and a separate metadata database.
- **Scale-appropriate:** SQLite comfortably handles the read-heavy, low-write workload of dozens to hundreds of ADRs. There is no concurrent write contention to worry about — indexing is a batch operation, and queries are read-only.
- **Go ecosystem support:** `modernc.org/sqlite` provides a pure-Go SQLite driver (no CGO required), keeping the build simple and cross-platform. `sqlite-vec` can be loaded as an extension.
- **Disposable index:** The SQLite database is a derived artifact — it can be regenerated from source ADR files at any time. This simplifies backup, versioning, and the development workflow.

## Alternatives Considered

- **ChromaDB:** Purpose-built vector database with good Python support. However, it requires running a separate server (another Docker container), and its Go client ecosystem is immature. Overkill for this corpus size.
- **Qdrant:** Production-grade vector database, excellent API. Same problem — it's a server with operational overhead that isn't justified for dozens of documents.
- **pgvector (PostgreSQL):** Strong option for production systems, but PostgreSQL is a heavy dependency for a project that could use an embedded database.
- **In-memory (Go maps/slices):** Simplest possible approach — load all vectors into memory, do brute-force cosine similarity. Actually viable at this scale. Rejected because SQLite adds very little complexity while providing persistence, SQL-based metadata queries, and a more realistic architecture to discuss in interviews.
- **FAISS:** Excellent vector search library, but Python-native with no Go bindings. Would require a sidecar service.

## Consequences

### Positive
- Zero-server architecture — the entire storage layer is an embedded file
- Single data model for vectors and metadata
- Trivially disposable and regenerable from source files
- Simple backup (copy one file)
- No Docker container needed for the database

### Negative
- `sqlite-vec` is a relatively young extension — less battle-tested than pgvector or dedicated vector databases
- SQLite's concurrency model (single writer) would be a bottleneck at production scale, but is irrelevant for this use case
- Pure-Go SQLite drivers can be slower than CGO-based ones for large datasets

### Mitigations
- The storage layer will be behind a Go interface, allowing swap to pgvector or a dedicated vector DB if the project scales beyond its current scope
- Include a `reindex` command that rebuilds the database from source ADRs, providing a simple recovery path if the database is corrupted or the schema changes
- Document the architecture decision clearly so that in an interview context, the conversation naturally leads to "what would you change at scale?" — which is exactly the kind of question a Principal Engineer should be able to answer well
