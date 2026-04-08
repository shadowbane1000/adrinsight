# Feature Specification: Engineering Quality

**Feature Branch**: `008-engineering-quality`
**Created**: 2026-04-07
**Status**: Draft

## Overview

ADR Insight was built rapidly across Milestones 1-6 with a "skateboard first" approach. The result is a fully functional system, but one that carries the engineering debt of fast iteration: unstructured logging, inconsistent error handling, hardcoded configuration, no health checks, no graceful shutdown, and no integration tests against real infrastructure.

This milestone brings the codebase to production-grade quality — the kind of code a reviewer would expect from a well-maintained service. It's the "would I ship this?" milestone: structured observability, clean error propagation, centralized configuration, operational readiness (health checks, graceful shutdown), and comprehensive testing.

## Actors

- **Code Reviewer** — reads the codebase to evaluate engineering quality; expects consistent patterns, clean error handling, and observable behavior
- **Operations Engineer** — deploys and monitors the service; needs health checks, structured logs, and graceful shutdown
- **Developer (Tyler)** — maintains the codebase; benefits from centralized configuration, actionable error messages, and idiomatic patterns

## Functional Requirements

### FR-001: Structured Logging
All logging throughout the application must use structured, leveled output instead of plain text. Log entries must include:
- Timestamp
- Level (debug, info, warn, error)
- Message
- Contextual fields where relevant (request ID, ADR number, duration, query text)

Log output format must be JSON by default (machine-parseable for Docker log drivers and cloud aggregators). A text/console format must be available for local development, selectable via environment variable. Log level must be configurable at startup (default: info). Debug logging must be available for troubleshooting without code changes.

### FR-002: Request ID Tracing
Each HTTP request must be assigned a unique request ID. The request ID must appear in all log entries generated during that request's lifecycle. The request ID must be returned in the response headers so clients can correlate responses to server-side logs.

### FR-003: Request Timing
All HTTP requests must log their duration. Slow requests (over a configurable threshold) must be logged at warn level. The timing information must be available in structured log fields, not embedded in the message string.

### FR-004: Centralized Configuration
All configurable settings must be manageable through environment variables with sensible defaults. Settings include:
- Server port (default: 8081)
- Database path (default: ./adr-insight.db)
- ADR directory path (default: ./docs/adr)
- Ollama URL (default: http://localhost:11434)
- Log level (default: info)
- Slow request threshold (default: 2s)
- Embedding model name (default: nomic-embed-text)
- Log format (default: json; alternative: text)

All configuration options must be documented in the README. A single configuration source must be used throughout the application — no scattered environment variable lookups.

### FR-005: Consistent Error Handling
All errors must propagate context using wrapped errors so the full error chain is available at the top level. Error messages must be actionable — they should tell the operator what went wrong and what to check. No silent error swallowing. All user-facing error responses must be friendly (no stack traces, no internal details).

### FR-006: Health Check Endpoint
A `GET /health` endpoint must verify system readiness by checking:
- Database is accessible and queryable
- Ollama is reachable (if configured)
- Anthropic API key is present (not validated — just present)

The endpoint must return a structured response indicating overall status and per-component status. It must respond within 5 seconds (timeout on external checks). Docker health checks must use this endpoint.

### FR-007: Graceful Shutdown
The application must handle termination signals (SIGTERM, SIGINT) gracefully:
- Stop accepting new requests
- Drain in-flight requests (with a configurable timeout, default: 10s)
- Close database connections cleanly
- Log the shutdown reason and duration

The application must not exit with a hard `log.Fatal` during normal operation. Fatal exits are only acceptable during startup if the application cannot proceed (missing required configuration, database initialization failure).

### FR-008: Go Idiom Review
The codebase must be audited for non-idiomatic patterns. Specific areas:
- Variable naming (short names in small scopes, descriptive names for exports)
- Error handling style (consistent `if err != nil` patterns, no naked returns after error checks)
- Package organization (no circular dependencies, clean package boundaries)
- Comment style (package comments, exported function documentation where non-obvious)

### FR-009: Integration Test Suite
Tests must exercise the full pipeline using real components (not just mocks):
- Parse real ADR files from the test fixtures
- Embed using a mock embedder that returns deterministic vectors
- Store in a real SQLite database (in-memory or temp file)
- Search and verify correct ADRs are retrieved
- Verify the HTTP layer returns correct responses

These tests must run in CI without requiring external services (no Ollama, no Anthropic API).

### FR-010: API Documentation
All HTTP API endpoints must be documented with:
- Request method and path
- Request body format (with example)
- Response body format (with example for success and error cases)
- Status codes and their meaning

Documentation must live in the repository (not an external service).

## User Stories & Scenarios

### US-1: Code Reviewer Evaluates Logging (P1)
A code reviewer opens the codebase and sees structured, leveled logging throughout. They can trace a single request through the system by following the request ID. They see timing information on every HTTP handler. They find no `log.Printf` or `log.Println` calls — all logging uses the structured logger.

**Acceptance**: Zero unstructured log calls in the codebase. All HTTP requests produce structured log entries with request ID and duration. Log level is configurable.

### US-2: Operator Configures the Service (P1)
An operator deploying ADR Insight can configure all settings through environment variables. They read the README and find a complete table of configuration options with defaults. They override the port and log level for their environment. The application starts with their custom settings.

**Acceptance**: All settings configurable via environment variables. README documents all options. Application respects overrides. No hardcoded values in business logic.

### US-3: Operator Monitors Health (P2)
An operator configures a Docker health check pointing at `/health`. The endpoint returns a structured response showing database, Ollama, and API key status. When Ollama goes down, the health check reflects the degraded state.

**Acceptance**: `GET /health` returns structured JSON with per-component status. Docker health check configured. Response within 5 seconds even when components are unreachable.

### US-4: Operator Shuts Down Cleanly (P2)
An operator sends SIGTERM to the running process (or Docker stops the container). The application logs that shutdown has started, finishes any in-flight requests, closes the database, and exits cleanly with code 0.

**Acceptance**: SIGTERM triggers graceful shutdown. In-flight requests complete. Database closed. Exit code 0. Shutdown logged with duration.

### US-5: Developer Fixes Error Handling (P1)
A developer reviews all error paths and ensures every error is wrapped with context, every error response is user-friendly, and no errors are silently ignored. The developer can trace any error from the HTTP response back to its origin through the wrapped error chain.

**Acceptance**: All `fmt.Errorf` calls use `%w` for wrapping. No silent error swallowing. User-facing errors are friendly. Error chain is traceable.

### US-6: Developer Runs Integration Tests (P2)
A developer runs `go test ./...` and the integration tests exercise the full pipeline — parsing, embedding, storage, retrieval, and HTTP response — using real SQLite and test fixtures, without requiring external services.

**Acceptance**: Integration tests pass in CI. Tests use real SQLite, mock embedder with deterministic vectors, real parser. No Ollama or Anthropic required.

### US-7: Developer Reads API Docs (P3)
A developer or API consumer reads the API documentation and understands all available endpoints, request/response formats, and error codes without reading the source code.

**Acceptance**: All endpoints documented with examples. Documentation lives in the repository.

## Edge Cases & Error Handling

- **Ollama unreachable at startup**: Application starts but logs a warning. Health check reports Ollama as unhealthy. Queries that require embedding return a clear error.
- **Database file locked**: Application logs an actionable error and exits at startup (not a silent hang).
- **Invalid environment variable values**: Application logs the invalid value, the expected format, and exits at startup with a clear message.
- **Shutdown timeout exceeded**: After the drain timeout, forcefully close remaining connections and exit. Log which requests were interrupted.
- **Health check during shutdown**: Return 503 Service Unavailable.
- **Concurrent health checks**: Must not deadlock or block request handling.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Zero instances of unstructured logging (`log.Printf`, `log.Println`, `log.Fatal` outside startup) in the codebase
- **SC-002**: 100% of HTTP requests produce structured log entries with request ID and duration
- **SC-003**: All configurable settings documented in README with defaults and environment variable names
- **SC-004**: Health check endpoint responds within 5 seconds under all conditions
- **SC-005**: Graceful shutdown completes in-flight requests and exits with code 0 on SIGTERM
- **SC-006**: Integration tests achieve the same pipeline coverage as the mock-based unit tests, using real storage
- **SC-007**: A code reviewer finds no non-idiomatic Go patterns flagged during the idiom review

## Scope & Boundaries

### In Scope
- Replace all unstructured logging with structured leveled logging
- Add request ID middleware and request timing
- Centralize configuration in a config struct loaded from environment variables
- Add `/health` endpoint with component checks
- Implement graceful shutdown with signal handling
- Audit and fix error handling (wrap with context, no silent swallowing)
- Audit and fix Go idiom issues
- Add integration tests using real SQLite and test fixtures
- Document the HTTP API
- Update Docker health check configuration

### Out of Scope
- Distributed tracing (OpenTelemetry) — deferred to a future milestone
- Metrics/Prometheus endpoint — deferred
- Rate limiting — deferred
- Authentication/authorization — not needed for this portfolio project
- Performance optimization — separate concern
- OpenAPI/Swagger generation — simple markdown documentation is sufficient

## Dependencies

- Milestone 6 (UI & UX Polish) must be merged — this milestone modifies the server, handlers, and main.go
- No new external dependencies expected (structured logging uses Go stdlib slog)

## Clarifications

### Session 2026-04-07
- Q: What log output format should be used? → A: JSON by default, with a text/console option for local development (configurable via env var).

## Assumptions

- The structured logging solution uses Go's built-in `log/slog` package (available since Go 1.21), consistent with the project's stdlib-first principle
- Configuration uses environment variables only — no config file support in this milestone (YAGNI for a single-binary deployment)
- Integration tests use in-memory SQLite and deterministic mock embeddings — no external services required
- API documentation is a markdown file in `docs/` — no auto-generation tooling
- The idiom review focuses on patterns a Go reviewer would flag, not stylistic preferences
