# ADR-015: Switch to mattn/go-sqlite3 (CGO) as SQLite Driver

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR-006 chose ncruces/go-sqlite3 (WASM) as the SQLite driver to avoid CGO
complexity. During Milestone 1 implementation, the WASM approach failed due
to version incompatibilities between the sqlite-vec WASM binary and the
ncruces driver.

Specifically, the sqlite-vec WASM binary used atomic operations that the
wazero runtime didn't support at compatible version ranges. Multiple version
combinations were tested:

- **ncruces v0.17.1 + sqlite-vec WASM:** Crashed with "out of bounds memory
  access" during KNN queries
- **ncruces v0.18.4 + sqlite-vec WASM:** Same crash, different stack trace
- **ncruces v0.22.0 + sqlite-vec WASM:** WASM function signature mismatch
  errors at extension load time

All failures occurred specifically during vector similarity search operations
(the core use case for sqlite-vec), not during basic SQLite operations. The
WASM bindings work for simple queries but fail under the KNN workload.

## Decision

Switch to mattn/go-sqlite3 (CGO-based) with the sqlite-vec CGO bindings at
`github.com/asg017/sqlite-vec-go-bindings/cgo`. This requires `libsqlite3-dev`
on the build system and `CGO_ENABLED=1`.

This supersedes ADR-006.

## Rationale

- **Proven stability:** mattn/go-sqlite3 is the most mature and widely-used
  Go SQLite driver (7k+ stars). The CGO sqlite-vec bindings are the primary,
  best-tested integration path.
- **sqlite-vec works reliably:** The CGO bindings auto-register the extension
  via `sqlite_vec.Auto()` — no manual loading needed. KNN queries, vector
  serialization, and all operations work correctly.
- **Pragmatic tradeoff:** The WASM approach would have been ideal for build
  simplicity, but the incompatibilities are in the sqlite-vec WASM binary
  itself — not something we can fix. The Store interface (Constitution
  Principle I) means we can revisit this when the WASM bindings stabilize.

## Consequences

### Positive
- Reliable, well-tested SQLite operations
- Official sqlite-vec integration path that works out of the box
- Best performance (native C, no WASM overhead)

### Negative
- Requires CGO (`CGO_ENABLED=1`) — complicates cross-compilation
- Requires `libsqlite3-dev` on the build system and in CI
- Docker builder stage needs a C toolchain (addressed in ADR-014)

### Mitigations
- CI workflows install `libsqlite3-dev` as a build dependency
- Docker multi-stage build uses a builder image with the C toolchain (ADR-014)
- The Store interface allows swapping back to ncruces/WASM if those bindings
  stabilize in the future

## Alternatives Considered

### Keep ncruces/go-sqlite3 (WASM)
- **Rejected:** After testing three version combinations across two months of
  ncruces releases, the sqlite-vec WASM bindings remain incompatible with KNN
  queries. The failure mode (crashes during vector search) makes this unusable
  for the core use case.

### modernc.org/sqlite (pure Go)
- **Rejected because:** Cannot load native C extensions like sqlite-vec
  (same reason as in ADR-006).

### viant/sqlite-vec (pure Go reimplementation)
- **Not chosen:** A pure-Go reimplementation of vector search concepts using
  its own virtual table scheme (not `vec0`). Newer and less battle-tested
  than the official CGO bindings.

## Related ADRs

- ADR-006: ncruces/go-sqlite3 (WASM) — superseded by this ADR
- ADR-004: SQLite with Vector Extension for Storage — establishes sqlite-vec
  as the storage layer; this ADR selects the driver to access it
- ADR-014: Docker Multi-Stage Debian Build — the CGO requirement drives the
  Docker build strategy
