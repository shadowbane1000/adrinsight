# ADR-011: Go Standard Library net/http for HTTP Server

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight's Milestone 2 introduces an HTTP API with three endpoints:
`POST /query`, `GET /adrs`, and `GET /adrs/{number}`. The project needs an
HTTP server that supports method-based routing and path parameters. Go has
a rich ecosystem of HTTP frameworks (Chi, Echo, Gin, gorilla/mux) as well
as the standard library's `net/http` package.

The project's constitution (Principle IV: Idiomatic Go) states "lean on the
standard library before reaching for third-party packages."

## Decision

Use Go's standard library `net/http` package with the Go 1.22+ enhanced
`ServeMux` for all HTTP routing and serving.

## Rationale

- **Go 1.22 changed the equation:** Prior to Go 1.22, `ServeMux` only
  supported path-prefix matching — no method routing, no path parameters.
  Frameworks like Chi existed to fill this gap. Go 1.22 added
  `mux.HandleFunc("GET /adrs/{number}", handler)` syntax, making the
  standard library sufficient for most REST APIs.
- **Three endpoints don't justify a framework:** The API surface is small.
  Chi's middleware grouping, Echo's context binding, and Gin's validator
  integration all solve problems this project doesn't have.
- **Zero dependencies:** Using stdlib means no third-party HTTP dependency
  to version, audit, or worry about maintenance status. gorilla/mux went
  into maintenance mode in 2022 — a cautionary tale for framework
  dependencies in small projects.
- **Constitution compliance:** Principle IV explicitly calls for standard
  library preference. Using a framework here would require justification
  in the Complexity Tracking section.
- **Portfolio signal:** Choosing stdlib over a framework demonstrates that
  the developer understands Go's standard library capabilities and makes
  technology choices based on actual needs rather than habit.

## Consequences

### Positive
- No HTTP framework dependency
- Idiomatic Go that any Go developer can read without learning a framework
- Routing is visible in one place (the ServeMux setup)
- Standard `http.Handler` and `http.HandlerFunc` signatures — easy to test
  with `httptest`

### Negative
- No built-in middleware chaining (must implement manually if needed)
- No request body binding/validation helpers — must unmarshal JSON manually
- No built-in structured error response helpers
- If the API grows significantly (10+ endpoints, auth, rate limiting),
  a framework would reduce boilerplate

### Mitigations
- For 3 endpoints, manual JSON marshaling is trivial
- Middleware can be added as a simple function wrapper if needed later
- If the API grows in Phase 2+, migrating to Chi is straightforward since
  Chi uses standard `http.Handler` interfaces

## Alternatives Considered

### Chi
- **Not chosen because:** Chi is excellent and would be the first choice if
  the API were larger. For 3 endpoints, it adds a dependency without solving
  a real problem. Chi remains the recommended upgrade path if the API grows.

### Echo
- **Not chosen because:** Echo uses its own `echo.Context` instead of
  standard `http.Request`/`http.ResponseWriter`, which diverges from stdlib
  conventions and makes handler code less portable.

### Gin
- **Not chosen because:** Gin is feature-rich (validation, binding, rendering)
  but heavyweight for 3 JSON endpoints. Its custom context and middleware
  model add concepts that aren't needed here.

### gorilla/mux
- **Not chosen because:** gorilla/mux entered maintenance mode in 2022 and
  its key feature (method routing + path params) is now in the standard
  library as of Go 1.22.

## Related ADRs

- ADR-001: Why Go — establishes Go as the language; this ADR selects how to
  use Go's HTTP capabilities
