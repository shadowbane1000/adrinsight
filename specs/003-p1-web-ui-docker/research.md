# Research: Web UI and Docker Deployment

## Client-Side Markdown Rendering

**Decision**: marked.js via CDN

**Rationale**: marked.js is the smallest (~16KB gzipped), fastest, and works
with a single `<script>` tag from a CDN — no build step. Constitution
Principle III (skateboard first) and the spec assumption (no frontend
framework, no node_modules) make this the natural choice.
Usage: `marked.parse(markdownString)` returns HTML.

**Alternatives considered**:
- markdown-it — ~33KB, more extensible but double the size for features we
  don't need
- showdown.js — ~24KB, slowest of the three

## Static File Serving

**Decision**: `go:embed` for production, disk serving for development

**Rationale**: `go:embed` bakes assets into the binary — single-binary
deployment, no file management, assets version-locked. During development,
serve from disk for live editing without recompilation. A flag or env var
switches between modes.

**Alternatives considered**:
- Disk-only serving — requires deploying assets alongside the binary,
  complicates Docker image
- Embed-only — works but makes UI iteration slow during development

## Ollama Docker Model Pre-Pull

**Decision**: Custom entrypoint script that pulls the model on first start

**Rationale**: `ollama pull` can't run during `docker build` because the
Ollama server isn't running. The entrypoint script starts Ollama in the
background, pulls the model, then restarts as the main process. The
container fails health checks until the model is ready, which is correct
behavior for orchestration.

**Alternatives considered**:
- Bake model into a custom Docker image — works but creates a large image
  (~2GB) and requires rebuilding when models update
- Rely on user to pull manually — violates Constitution Principle V
  (docker-compose up should be all it takes)

## Docker Build Strategy (CGO)

**Decision**: Multi-stage build with golang:bookworm (build) and
debian:bookworm-slim (runtime)

**Rationale**: The project uses mattn/go-sqlite3 with CGO (ADR-006). The
build stage needs gcc and libsqlite3-dev. The runtime stage needs only libc
and the sqlite3 shared library. Using bookworm (Debian) for both stages
ensures ABI compatibility.

**Alternatives considered**:
- Alpine-based build — would need musl-compatible sqlite, adds complexity
- Distroless — doesn't include sqlite3 shared library
- Scratch — can't use with CGO binaries that link to libc
