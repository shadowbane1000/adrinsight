# Research: Abuse Hardening for AI Cost Control

## Rate Limiting Algorithm

**Decision**: Fixed-window counter per IP address
**Rationale**: Simplest approach that meets requirements. The window resets after the configured duration. Slightly burstier than sliding window at window boundaries, but acceptable for a demo with generous limits (10 req/min).
**Alternatives considered**:
- Sliding window log: More accurate but requires storing individual request timestamps. Unnecessary complexity for this scale.
- Token bucket: Better for smoothing bursts. Overkill for a demo app with low traffic.
- Leaky bucket: Similar to token bucket. Same conclusion.

## IP Extraction

**Decision**: Use `X-Forwarded-For` header first, fall back to `RemoteAddr`
**Rationale**: The app runs behind Nginx, which sets `X-Forwarded-For`. Without this, all requests would share the proxy's IP and rate limit collectively.
**Alternatives considered**:
- `X-Real-IP` header: Some proxies use this instead. Could support both, but Nginx is the known deployment target and uses `X-Forwarded-For`.
- Always use `RemoteAddr`: Would break rate limiting behind any reverse proxy.

## State Storage

**Decision**: In-memory `map[string]*rateLimitEntry` protected by `sync.Mutex`
**Rationale**: Single-instance deployment, state loss on restart is acceptable per spec. No external dependencies. Periodic cleanup of stale entries prevents unbounded memory growth.
**Alternatives considered**:
- SQLite table: Persistent but adds write contention on the DB used for ADR queries. Unnecessary.
- Redis: External dependency, violates constitution principle V (developer experience). Overkill.

## Stale Entry Cleanup

**Decision**: Lazy cleanup during rate limit checks — remove entries older than the window duration when accessed. No background goroutine needed at this scale.
**Rationale**: With at most tens of unique IPs, memory usage is negligible. A background sweeper would add complexity for no practical benefit.
