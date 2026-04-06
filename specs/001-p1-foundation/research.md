# Research: Foundation — ADR Indexing Pipeline

## Markdown Parsing Library

**Decision**: goldmark with goldmark-meta extension

**Rationale**: goldmark is the most actively maintained Go markdown parser,
CommonMark-compliant, and designed for extensibility via AST walking. The
`ast.Walk` API allows structured extraction of headings and sections —
exactly what we need for parsing ADR structure. goldmark-meta adds YAML
frontmatter support. blackfriday is older and in maintenance mode with a
less extensible API.

**Alternatives considered**:
- blackfriday — older, maintenance mode, less extensible AST API
- Manual regex/string parsing — fragile, doesn't handle markdown edge cases

## SQLite Driver (No CGO)

**Decision**: ncruces/go-sqlite3 (WASM-based via wazero)

**Rationale**: This is the only Go SQLite driver that supports loading
native extensions like sqlite-vec without requiring CGO. It runs SQLite
compiled to WASM, with the sqlite-vec author providing official Go bindings
at `github.com/asg017/sqlite-vec-go-bindings/ncruces`. Just importing the
package auto-registers the extension — no manual loading needed. This keeps
the build simple and cross-platform.

**Alternatives considered**:
- mattn/go-sqlite3 — requires CGO, complicates cross-compilation and Docker builds
- modernc.org/sqlite — pure Go but cannot load native C extensions like sqlite-vec
- viant/sqlite-vec — pure Go reimplementation of sqlite-vec, newer and less tested

## Ollama Go Client

**Decision**: Official ollama/ollama/api Go package

**Rationale**: Ollama provides an official Go client. Use `Client.Embed()`
(the newer batch-capable method) rather than the deprecated
`Client.Embeddings()`. No need for raw HTTP calls.

**Alternatives considered**:
- Raw HTTP to `POST /api/embed` — works but reinvents what the official client already provides

## Embedding Model Details

**Decision**: nomic-embed-text with 768 dimensions (default)

**Rationale**: 768 dimensions provides the best quality. The model supports
Matryoshka dimensionality reduction (64–512) but there's no reason to
reduce quality at this corpus scale.

**Key implementation detail**: The model requires input prefixes:
- Index time: prefix ADR chunk text with `search_document: `
- Query time: prefix query text with `search_query: `
These prefixes are part of the model's training and are required for
correct similarity results.

**Context length**: 8192 tokens — sufficient for even long ADR sections
without further subdivision in most cases.

## Vector Serialization

**Decision**: Use sqlite-vec's `SerializeFloat32()` helper

**Rationale**: sqlite-vec stores vectors as little-endian float32 blobs.
The official Go bindings provide `sqlite_vec.SerializeFloat32()` for
serialization. Vectors must be serialized before insert and for query
parameters.

## sqlite-vec Table Design

**Decision**: Use vec0 virtual table with float[768]

**Rationale**: The virtual table syntax is
`CREATE VIRTUAL TABLE vec_chunks USING vec0(embedding float[768])`.
Queries use `WHERE embedding MATCH ? ORDER BY distance LIMIT k` with
serialized query vectors. rowid links to the metadata table.
