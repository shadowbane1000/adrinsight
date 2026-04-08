# ADR-024: slog for Structured Logging

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

ADR Insight's codebase uses Go's `log` package throughout — `log.Printf`,
`log.Println`, and `log.Fatalf` — for all diagnostic output. This produces
unstructured, unleveled text that is difficult to parse programmatically,
impossible to filter by severity, and lacks contextual fields (request IDs,
timing, ADR numbers).

As the system matures toward production quality (Milestone 7), operators need
machine-parseable logs for Docker log drivers and cloud aggregators, developers
need filterable log levels for debugging, and request tracing requires
structured fields that flow through the entire request lifecycle.

The project constitution (Principle IV: Idiomatic Go) requires leaning on the
standard library before reaching for third-party packages.

## Decision

Use Go's standard library `log/slog` package (available since Go 1.21) for
all structured logging. Output JSON by default via `slog.NewJSONHandler`,
with a text/console format available via `slog.NewTextHandler` for local
development. The format and level are configurable via environment variables
(`LOG_FORMAT`, `LOG_LEVEL`).

All existing `log.Printf`, `log.Println`, and `log.Fatal` calls will be
replaced with `slog.Info`, `slog.Warn`, `slog.Error`, and `slog.Debug`
calls with structured key-value fields.

## Rationale

- **stdlib first:** `log/slog` is Go's official structured logging package,
  shipped in the standard library since Go 1.21. Using it requires zero
  external dependencies — fully aligned with Constitution Principle IV.
- **JSON + Text handlers built in:** `slog.NewJSONHandler` produces
  machine-parseable output for Docker/cloud environments. `slog.NewTextHandler`
  produces human-readable output for development. No custom formatters needed.
- **Context-aware:** `slog.With()` adds persistent fields (request ID, ADR
  number) to a logger that flows through the call chain. Handler middleware
  can inject fields from `context.Context` automatically.
- **Level filtering:** `slog.LevelDebug`, `slog.LevelInfo`, `slog.LevelWarn`,
  `slog.LevelError` — configurable at runtime via `slog.HandlerOptions`.
  Debug logging is available without code changes.
- **Community consensus:** As of Go 1.21+, slog is the recommended logging
  approach for new Go projects. Third-party libraries (zerolog, zap) are
  converging on slog compatibility via handler adapters.

## Consequences

### Positive
- Zero new external dependencies
- All log output is structured and filterable by level
- JSON output works with Docker log drivers and cloud aggregators out of the box
- Request IDs and timing can be injected as structured fields, enabling
  request tracing through the entire pipeline
- Consistent logging API across all packages

### Negative
- slog's API is slightly more verbose than `log.Printf` — every call requires
  key-value field pairs instead of format strings
- slog is newer (Go 1.21) — some Go developers may be more familiar with
  third-party alternatives
- No built-in log rotation or file output — relies on external tooling
  (Docker log drivers, systemd journal, logrotate)

### Mitigations
- Verbosity is offset by structured fields being more useful than formatted
  strings for debugging and monitoring
- The project already requires Go 1.24+ — slog availability is not a concern
- Log rotation is handled by the container runtime in the Docker deployment
  (the primary deployment model)

## Alternatives Considered

### zerolog
- **Not chosen because:** Adds an external dependency (`github.com/rs/zerolog`).
  While slightly faster than slog (zero-allocation design), the performance
  difference is irrelevant at this project's throughput. Constitution
  Principle IV favors stdlib.

### zap (Uber)
- **Not chosen because:** Adds an external dependency with a heavier API
  surface (sugared vs structured loggers, encoder configuration). Well-suited
  for high-throughput services, overkill for a portfolio project.

### logrus
- **Not chosen because:** Effectively in maintenance mode since 2020. The
  Go community has moved to slog. Adding a deprecated dependency would be
  a negative signal in a portfolio project.

### Keep log.Printf (do nothing)
- **Not chosen because:** Unstructured text cannot be filtered by level,
  parsed by log aggregators, or enriched with contextual fields. Every
  operational improvement (request tracing, slow query detection, error
  correlation) requires structured logging as a foundation.

## Related ADRs

- ADR-001: Why Go [related] — Go's standard library richness (including slog) is part of the rationale for choosing Go
- ADR-011: stdlib HTTP Server [related] — same stdlib-first principle applied to HTTP; this ADR applies it to logging
