# ADR-006: ncruces/go-sqlite3 (WASM) as SQLite Driver

**Status:** Superseded by ADR-015  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR-004 established SQLite with sqlite-vec as the storage layer. The next decision is which Go SQLite driver to use. The key constraint is that sqlite-vec is a native C extension — it must be loaded into the SQLite runtime.

There are three main Go SQLite drivers:

- **ncruces/go-sqlite3** — WASM-based, no CGO required. Uses wazero as a WASM runtime. sqlite-vec provides WASM bindings.
- **mattn/go-sqlite3** — CGO-based, wraps the C SQLite library directly. Requires a C compiler and `libsqlite3-dev`. sqlite-vec provides CGO bindings.
- **modernc.org/sqlite** — Pure Go transpilation of the C source. Cannot load native C extensions like sqlite-vec.

## Decision

Use ncruces/go-sqlite3 (WASM-based, no CGO) with the sqlite-vec WASM bindings. This avoids requiring a C toolchain on the build system and simplifies cross-compilation.

## Rationale

- **No CGO:** Simplifies the build pipeline — no C compiler needed, no `libsqlite3-dev`, trivial cross-compilation. Constitution Principle V (Developer Experience).
- **sqlite-vec WASM bindings exist:** The sqlite-vec project publishes WASM bindings specifically for the ncruces driver.
- **wazero runtime:** Pure Go WASM runtime with no system dependencies.
- **Skateboard first:** Constitution Principle III — if the WASM approach works, it's simpler than CGO.

## Consequences

### Positive
- No CGO, no C toolchain requirement
- Simple cross-compilation
- Cleaner Docker builds (no gcc needed)

### Negative
- WASM overhead compared to native C
- sqlite-vec WASM bindings are newer and less tested than the CGO bindings
- wazero compatibility with sqlite-vec's WASM binary is not guaranteed across versions

### Mitigations
- The Store interface (Constitution Principle I) allows swapping drivers if WASM proves unreliable
- Performance overhead is negligible at this corpus scale

## Alternatives Considered

### mattn/go-sqlite3 (CGO)
- **Not chosen initially:** Requires CGO, a C compiler, and `libsqlite3-dev`. Adds build complexity. Kept as a fallback if WASM proves unreliable.

### modernc.org/sqlite (pure Go)
- **Rejected because:** Cannot load native C extensions like sqlite-vec. The sqlite-vec extension is central to the architecture (ADR-004).

## Related ADRs

- ADR-004: SQLite with Vector Extension for Storage — establishes sqlite-vec as the storage layer
- ADR-015: Switch to mattn/go-sqlite3 (CGO) — supersedes this ADR after WASM incompatibilities
