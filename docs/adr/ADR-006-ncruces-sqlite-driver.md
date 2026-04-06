# ADR-006: mattn/go-sqlite3 (CGO) as SQLite Driver

**Status:** Accepted (amended)  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR-004 established SQLite with sqlite-vec as the storage layer. The next decision is which Go SQLite driver to use. The key constraint is that sqlite-vec is a native C extension — it must be loaded into the SQLite runtime.

The original plan was to use ncruces/go-sqlite3 (WASM-based, no CGO) with the official sqlite-vec WASM bindings. During implementation, this approach failed due to version incompatibilities between the sqlite-vec WASM binary and the ncruces driver — the WASM binary used atomic operations that the wazero runtime didn't support at compatible version ranges, causing crashes during KNN queries.

## Decision

Use mattn/go-sqlite3 (CGO-based) with the sqlite-vec CGO bindings at `github.com/asg017/sqlite-vec-go-bindings/cgo`. This requires `libsqlite3-dev` on the build system and `CGO_ENABLED=1`.

## Rationale

- **Proven stability:** mattn/go-sqlite3 is the most mature and widely-used Go SQLite driver (7k+ stars). The CGO sqlite-vec bindings are the primary, best-tested integration path.
- **sqlite-vec works reliably:** The CGO bindings auto-register the extension via `sqlite_vec.Auto()` — no manual loading needed. KNN queries, vector serialization, and all operations work correctly.
- **Pragmatic tradeoff:** The WASM approach (ncruces) would have been ideal for build simplicity, but version incompatibilities between the shipped WASM binary and the driver made it unreliable. The Store interface (Constitution Principle I) means we can revisit this when the WASM bindings stabilize.

## Consequences

### Positive
- Reliable, well-tested SQLite operations
- Official sqlite-vec integration path that works out of the box
- Best performance (native C, no WASM overhead)

### Negative
- Requires CGO (`CGO_ENABLED=1`) — complicates cross-compilation
- Requires `libsqlite3-dev` on the build system and in CI
- Docker builder stage needs a C toolchain

### Mitigations
- CI workflows install `libsqlite3-dev` as a build dependency
- Docker multi-stage build can use a builder image with the C toolchain
- The Store interface allows swapping back to ncruces/WASM when those bindings stabilize

## Alternatives Considered

### ncruces/go-sqlite3 (WASM, no CGO)
- **Originally chosen, then rejected:** The sqlite-vec WASM bindings shipped with atomic operations incompatible with the ncruces driver at available version ranges. KNN queries crashed with "out of bounds memory access" errors. Multiple version combinations (v0.17.1 through v0.22.0) were tested without success.

### modernc.org/sqlite (pure Go)
- **Rejected because:** Cannot load native C extensions like sqlite-vec. The sqlite-vec extension is central to the architecture (ADR-004).

### viant/sqlite-vec (pure Go reimplementation)
- **Not chosen:** A pure-Go reimplementation of vector search concepts using its own virtual table scheme (not `vec0`). Newer and less battle-tested than the official CGO bindings.

## Related ADRs

- ADR-004: SQLite with Vector Extension for Storage — establishes sqlite-vec as the storage layer
