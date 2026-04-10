# Feature Specification: Abuse Hardening for AI Cost Control

**Feature Branch**: `009-abuse-hardening`  
**Created**: 2026-04-09  
**Status**: Draft  
**Input**: User description: "Hardening against abuse to avoid AI costs going out of control."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rate Limiting on Query Endpoint (Priority: P1)

ADR Insight is deployed as a public-facing demo. The query endpoint triggers two AI services per request: a local embedding call and a paid Anthropic Claude API call. Without rate limiting, a bot or curious visitor could fire hundreds of queries and run up significant API costs in minutes. The system must limit how frequently queries can be submitted from a single source.

**Why this priority**: This is the primary cost vector. A single unthrottled client could exhaust an API budget in hours. Rate limiting is the minimum viable protection.

**Independent Test**: Can be tested by sending rapid queries from a single IP and verifying that excess requests are rejected with a clear message.

**Acceptance Scenarios**:

1. **Given** a visitor submits a query, **When** they submit another query within the rate limit window, **Then** the second query is rejected with a "too many requests" response and a message indicating when they can try again.
2. **Given** a visitor has been rate-limited, **When** the rate limit window expires, **Then** their next query succeeds normally.
3. **Given** two visitors are querying simultaneously from different IPs, **When** one is rate-limited, **Then** the other is unaffected.

---

### User Story 2 - Query Length Limits (Priority: P2)

Excessively long queries waste embedding compute and send large prompts to Claude, increasing token cost. The system must reject queries that exceed a reasonable length.

**Why this priority**: Prevents a simple abuse vector (pasting huge text blobs as queries) and limits per-request cost. Less critical than rate limiting because a single oversized query costs less than hundreds of normal-sized ones.

**Independent Test**: Can be tested by submitting queries of increasing length and verifying the cutoff behavior.

**Acceptance Scenarios**:

1. **Given** a visitor submits a query within the length limit, **When** processed, **Then** the query succeeds normally.
2. **Given** a visitor submits a query exceeding the length limit, **When** submitted, **Then** the system returns an error indicating the query is too long, without calling any AI services.

---

### Edge Cases

- What happens when a rate-limited user refreshes the page? The rate limit applies to the query endpoint only — page loads and ADR browsing are unaffected.
- How does the system behave behind a reverse proxy (Nginx)? Rate limiting must use the real client IP from the `X-Forwarded-For` header, not the proxy's IP.
- What happens if the rate limit store is unavailable or corrupted? The system should default to allowing requests (fail-open) rather than blocking all users.
- What if multiple users share an IP (office NAT, VPN)? Accepted tradeoff — the rate limit is generous enough that normal use from shared IPs is unlikely to trigger it.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST enforce a per-IP rate limit on the query endpoint (`POST /query`), rejecting excess requests with HTTP 429 and a `Retry-After` header.
- **FR-002**: System MUST reject queries exceeding a configurable maximum character length before calling any AI services, returning HTTP 400 with a descriptive error message.
- **FR-003**: Rate limit parameters (requests per window, window duration) MUST be configurable via environment variables with sensible defaults.
- **FR-004**: Maximum query length MUST be configurable via environment variable with a sensible default.
- **FR-005**: Rate limiting MUST use the client's real IP address, respecting `X-Forwarded-For` headers for proxied deployments.
- **FR-006**: Rate limiting MUST NOT affect non-query endpoints (ADR listing, ADR detail, health check, static files).
- **FR-008**: Rate limit state MUST be stored in-memory (no external dependencies). State is lost on restart, which is acceptable.
- **FR-009**: The web UI MUST display a user-friendly message when a query is rate-limited, including when the user can retry.

### Key Entities

- **Rate Limit Entry**: Tracks request count per IP address within a time window. Attributes: IP address, request count, window start time.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A single IP address cannot submit more than the configured number of queries within the configured time window.
- **SC-002**: Queries exceeding the maximum length are rejected before any AI service is invoked, adding zero cost.
- **SC-003**: Rate-limited responses include a clear indication of when the user can retry.
- **SC-004**: Legitimate users querying at a normal pace (a few queries per minute) are never rate-limited.

## Assumptions

- The system runs behind Nginx, which provides the `X-Forwarded-For` header for real client IP identification.
- In-memory rate limiting is sufficient — the system is single-instance and state loss on restart is acceptable (restarts are infrequent).
- A default rate limit of 10 queries per minute per IP is generous enough for legitimate demo use while preventing automated abuse. This can be tuned via environment variables.
- A default maximum query length of 500 characters is sufficient for natural language questions while preventing abuse via large text payloads.
- No authentication or API keys are required for users — this is a public demo. Rate limiting by IP is the appropriate access control mechanism.
