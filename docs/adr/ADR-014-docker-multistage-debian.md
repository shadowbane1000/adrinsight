# ADR-014: Multi-Stage Docker Build with Debian Bookworm

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight's Milestone 3 containerizes the application with Docker for
one-command deployment via `docker compose up`. The Go application uses
mattn/go-sqlite3 with CGO (ADR-006), which means the build requires a C
compiler and the resulting binary links dynamically against libc and sqlite3.
The Docker image must support this CGO requirement while keeping the final
image reasonably small.

## Decision

Use a multi-stage Docker build:
- **Build stage:** `golang:bookworm` — includes Go toolchain, gcc, and
  access to Debian packages for `libsqlite3-dev`
- **Runtime stage:** `debian:bookworm-slim` — minimal Debian with libc and
  the ability to install `libsqlite3-0` for the sqlite3 shared library

Both stages use Debian Bookworm to ensure ABI compatibility between the
build-time and runtime C libraries.

## Rationale

- **CGO requires matching libc:** The mattn/go-sqlite3 driver (ADR-006) uses
  CGO, producing a dynamically linked binary. The C library (glibc) used at
  build time must be ABI-compatible with the one at runtime. Using the same
  Debian release (Bookworm) for both stages guarantees this.
- **Multi-stage keeps images small:** The build stage includes the full Go
  toolchain and C compiler (~1GB), but only the compiled binary and runtime
  libraries are copied to the final image. The runtime stage
  (`debian:bookworm-slim`) is ~80MB.
- **Debian over Alpine:** Alpine uses musl libc instead of glibc. CGO
  binaries built against glibc crash or behave unexpectedly on musl. While
  it's possible to build for musl, it adds complexity (cross-compilation
  flags, musl-dev packages, potential sqlite3 compatibility issues) with no
  meaningful benefit for this project.
- **Constitution compliance:** Principle V (Developer Experience) —
  `docker compose up` just works, no special build flags or compatibility
  workarounds. Principle III (Skateboard First) — Debian is the
  straightforward choice that avoids musl complications.

## Consequences

### Positive
- Build and runtime environments are ABI-compatible by construction
- Runtime image is reasonably small (~80MB base + application)
- Standard Debian package manager available for runtime dependencies
- No musl/glibc compatibility issues to debug

### Negative
- Runtime image is larger than Alpine (~80MB vs ~5MB base) or distroless
- Debian Bookworm will eventually reach end-of-life (June 2028 for LTS),
  requiring a base image update
- Full Debian runtime includes more packages than strictly needed

### Mitigations
- The ~75MB difference is acceptable for a self-hosted application; this
  is not a microservice deployed at scale where image size compounds
- Base image updates are a routine maintenance task tracked by Dependabot
  or similar tools
- `bookworm-slim` variant strips unnecessary files, reducing the gap

## Alternatives Considered

### Alpine-based build and runtime
- **Not chosen because:** Alpine uses musl libc. CGO binaries built with
  glibc (the default in `golang:latest`) are not compatible with musl.
  Building for musl requires `CGO_ENABLED=1 CC=musl-gcc` and
  musl-compatible sqlite3, adding complexity. The mattn/go-sqlite3 driver
  has known edge cases with musl that would require additional testing.

### Google Distroless
- **Not chosen because:** Distroless images don't include a package manager
  or shared libraries beyond glibc. The sqlite3 shared library
  (`libsqlite3-0`) is not available in distroless, and there's no way to
  install it. Statically linking sqlite3 into the binary is possible but
  adds build complexity.

### Scratch image
- **Not chosen because:** Scratch images contain nothing — no libc, no
  certificates, no timezone data. CGO binaries require at minimum a
  dynamically linked libc. While it's technically possible to copy libc
  and dependencies into a scratch image, this is fragile and hard to
  maintain.

### Static compilation (CGO_ENABLED=0)
- **Not chosen because:** The mattn/go-sqlite3 driver requires CGO — it
  wraps the C sqlite3 library. Disabling CGO is not an option without
  switching to a pure-Go SQLite driver, which was evaluated and rejected
  in ADR-006 due to sqlite-vec compatibility issues.

## Related ADRs

- ADR-006: mattn/go-sqlite3 CGO Driver — the CGO requirement that drives
  this Docker build strategy
- ADR-004: SQLite Vector Storage — the storage layer that requires sqlite3
  in the runtime image
