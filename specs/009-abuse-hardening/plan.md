# Implementation Plan: Abuse Hardening for AI Cost Control

**Branch**: `009-abuse-hardening` | **Date**: 2026-04-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-abuse-hardening/spec.md`

## Summary

Add rate limiting and query length validation to the query endpoint to prevent abuse that could drive up Anthropic API costs. Per-IP rate limiting with configurable thresholds, query length cap before AI services are invoked, and user-friendly error feedback in the web UI. All in-memory, no external dependencies.

## Technical Context

**Language/Version**: Go 1.24+ (same as existing project)
**Primary Dependencies**: `net/http` (existing), `sync` (stdlib for concurrent map access)
**Storage**: In-memory (no persistence required)
**Testing**: `go test` with `-tags fts5` (existing CI pipeline)
**Target Platform**: Linux server (Docker container behind Nginx)
**Project Type**: Web service (existing)
**Performance Goals**: Rate limit check adds <1ms latency per request
**Constraints**: No external dependencies for rate limiting; in-memory state acceptable for single-instance deployment
**Scale/Scope**: Single instance, tens of concurrent users at most (public demo)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | PASS | Rate limiter will be middleware, no new external dependencies requiring interfaces |
| II. ADRs Are the Product | PASS | No ADR-worthy decisions — standard rate limiting pattern, no alternatives with meaningful tradeoffs |
| II-A. ADR Immutability | N/A | No existing ADRs affected |
| III. Skateboard First | PASS | Simplest viable implementation — in-memory fixed-window rate limiting, no token buckets or distributed state |
| IV. Idiomatic Go | PASS | stdlib `sync.Mutex` for concurrent map access, standard middleware pattern |
| V. Developer Experience | PASS | New env vars have sensible defaults, no additional setup required |

## Project Structure

### Documentation (this feature)

```text
specs/009-abuse-hardening/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/
├── server/
│   ├── middleware.go     # Add rateLimitMiddleware (existing file has requestID + logging middleware)
│   ├── handlers.go       # Add query length validation in handleQuery
│   └── server.go         # Wire rate limit middleware, add config fields
├── config/
│   └── config.go         # Add RateLimitRequests, RateLimitWindow, MaxQueryLength fields
web/
└── static/
    └── app.js            # Handle 429 responses in query store
```

**Structure Decision**: All changes go into existing files. No new packages or files needed — rate limiting is middleware in `internal/server/`, configuration extends the existing `Config` struct, and UI error handling extends the existing Alpine.js query store.

## ADR Impact

No existing ADRs are affected by this feature. No new ADRs needed — in-memory per-IP rate limiting with fixed windows is a standard, well-understood pattern with no meaningful alternatives at this scale.

## Complexity Tracking

No constitution violations to justify.
