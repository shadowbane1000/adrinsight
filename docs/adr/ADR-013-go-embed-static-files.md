# ADR-013: go:embed for Production Static Files with Disk-Based Dev Mode

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight's Milestone 3 adds a web UI consisting of static files
(HTML, CSS, JavaScript) served by the Go HTTP server. These files need to
be available to the server at runtime. The project must support two workflows:
production deployment (single binary, ideally via Docker) and local
development (fast iteration on UI changes without recompiling Go).

## Decision

Use Go's `go:embed` directive to bake static files into the production
binary. Provide a `--dev` flag that switches to disk-based file serving
for local development, allowing live editing of HTML/CSS/JS without
recompilation.

Implementation:
- `internal/server/embed.go` contains the `//go:embed` directive for
  `web/static/`
- Production mode: `http.FileServer(http.FS(embeddedFS))`
- Dev mode (`--dev`): `http.FileServer(http.Dir("web/static/"))`

## Rationale

- **Single-binary deployment:** `go:embed` bakes all static assets into the
  compiled binary. No need to copy files alongside the binary, manage paths,
  or worry about missing assets at runtime. This simplifies the Docker image
  (just the binary + runtime libraries) and eliminates deployment failures
  from missing files.
- **Version-locked assets:** The static files are compiled at the same time
  as the Go code. There's no risk of a binary running with stale or
  mismatched UI files.
- **Dev ergonomics:** A `--dev` flag switches to disk-based serving, so UI
  changes (HTML/CSS/JS) take effect on browser refresh without recompiling
  the Go binary. This keeps the development loop fast for frontend iteration.
- **Constitution compliance:** Principle IV (Idiomatic Go) — `go:embed` is
  the standard Go approach for bundling static assets since Go 1.16.
  Principle V (Developer Experience) — the dev flag ensures frontend
  iteration doesn't require a full rebuild.

## Consequences

### Positive
- Single binary contains everything needed to serve the UI
- Docker image needs only the binary and runtime libraries
- Assets are version-locked to the binary
- Dev mode enables fast UI iteration
- No external file paths to configure in production

### Negative
- Binary size increases by the size of the static files (HTML + CSS + JS,
  likely under 50KB — negligible)
- Developers must remember to rebuild the binary after UI changes for
  production testing (dev mode masks embedded staleness)
- Two code paths (embedded vs. disk) for file serving

### Mitigations
- Binary size impact is trivial for this project's static file footprint
- CI builds always produce a fresh binary with current assets
- The `--dev` flag is clearly documented; production always uses embedded

## Alternatives Considered

### Disk-only serving
- **Not chosen because:** Requires deploying static files alongside the
  binary. Complicates the Docker image (need to COPY files separately),
  introduces path configuration, and risks missing-file errors at runtime.

### Embed-only (no dev mode)
- **Not chosen because:** Every HTML/CSS/JS change would require
  `go build` before seeing the result in the browser. For UI iteration,
  this makes the development loop unnecessarily slow and frustrating.

### External web server (nginx, Caddy)
- **Not chosen because:** Adds another service to deploy and configure.
  For a small set of static files, Go's built-in file server is sufficient.
  Would violate Principle III (Skateboard First) by adding unnecessary
  infrastructure.

## Related ADRs

- ADR-011: Go Standard Library net/http — the HTTP server that serves
  these static files
- ADR-001: Why Go — `go:embed` is a Go-specific capability that reinforces
  the single-binary deployment story
