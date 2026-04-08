# Research: Engineering Quality

## R1: Structured Logging Framework

**Decision**: Go stdlib `log/slog` (available since Go 1.21)

**Rationale**:
- Constitution Principle IV requires standard library preference. slog is Go's official structured logging package.
- Zero external dependencies — no `zap`, `zerolog`, or `logrus` needed.
- Native JSON handler (`slog.NewJSONHandler`) and text handler (`slog.NewTextHandler`) match the clarified requirement (JSON default, text for dev).
- Context-aware: `slog.With()` for adding fields, handler middleware for request IDs.
- Already the community consensus for new Go projects as of Go 1.21+.

**Alternatives considered**:
- **zerolog** (~zero alloc): Faster than slog in benchmarks, but adds an external dependency. Not justified for a portfolio project with low throughput.
- **zap** (Uber): Mature and fast, but heavyweight API and external dependency. Overkill.
- **logrus**: Effectively in maintenance mode. The Go community has moved to slog.

## R2: Configuration Pattern

**Decision**: Config struct loaded from environment variables at startup

**Rationale**:
- Single `config.Config` struct holds all settings. Loaded once in `main()`, passed to components.
- `os.Getenv` with fallback defaults — no third-party config libraries.
- Matches Constitution Principle V (Developer Experience): sensible defaults, override via env vars.
- No config file support in this milestone — YAGNI for single-binary deployment.

**Alternatives considered**:
- **Viper**: Full-featured config library (files, env, flags, remote). Massive dependency for what we need. Rejected.
- **envconfig** (kelseyhightower): Nice API but external dependency. The manual approach is ~30 lines of code.
- **Config file (YAML/TOML)**: Adds file parsing dependency and doesn't improve the Docker deployment story where env vars are native.

## R3: Request ID Strategy

**Decision**: UUID v4 generated per request in middleware, stored in context

**Rationale**:
- `google/uuid` is already a transitive dependency (via Anthropic SDK). No new dependency.
- Request ID set in middleware, stored in `context.Context`, extracted by slog handler.
- Returned in `X-Request-ID` response header for client correlation.
- If the client sends `X-Request-ID`, use theirs (enables end-to-end tracing).

**Alternatives considered**:
- **ULID**: Sortable, but adds a dependency for no practical benefit here.
- **Sequential counter**: Not unique across restarts. Rejected.

## R4: Health Check Design

**Decision**: `/health` endpoint with per-component status checks

**Rationale**:
- Returns `{"status": "healthy|degraded|unhealthy", "components": {...}}`.
- Database: `SELECT 1` ping. Ollama: HTTP HEAD to base URL. API key: check env var presence.
- Each component check has a 3-second timeout. Overall endpoint timeout 5 seconds.
- Returns 200 for healthy/degraded, 503 for unhealthy.
- Degraded = Ollama unreachable but DB works (queries will fail but browsing works).

## R5: Graceful Shutdown Pattern

**Decision**: `os/signal.NotifyContext` + `http.Server.Shutdown`

**Rationale**:
- `signal.NotifyContext` (Go 1.16+) creates a context cancelled on SIGTERM/SIGINT.
- `http.Server.Shutdown(ctx)` drains in-flight requests with timeout.
- Defer `store.Close()` for database cleanup.
- Replace `log.Fatal` calls in normal operation with error returns. Keep `log.Fatal` only for unrecoverable startup failures.

## R6: Integration Test Approach

**Decision**: Real SQLite (in-memory) + deterministic mock embedder + real parser

**Rationale**:
- `:memory:` SQLite database — fast, no cleanup, runs in CI.
- Mock embedder returns fixed vectors keyed by content hash — deterministic retrieval results.
- Real parser processes test fixture ADRs from `testdata/`.
- Tests verify: parse → store → search returns correct ADRs → handler returns correct JSON.
- No external services required (no Ollama, no Anthropic).
