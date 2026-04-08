# ADR-012: marked.js via CDN for Client-Side Markdown Rendering

**Status:** Accepted (note: ADR-023 amends the "no frontend framework" context — Alpine.js is now used alongside marked.js)  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight's Milestone 3 introduces a web UI that displays two types of
markdown content: synthesized answers from the LLM and full ADR documents.
Both must be rendered as formatted HTML (headings, bold, lists, code blocks).
The rendering could happen server-side (Go converts markdown to HTML before
sending) or client-side (browser converts markdown to HTML after receiving).

The project's constitution (Principle III: Skateboard First) requires the
simplest complete version, and the spec constrains the UI to vanilla
HTML/CSS/JS with no frontend framework, no build step, and no node_modules.

## Decision

Use marked.js loaded via CDN (`<script>` tag) for all client-side markdown
rendering. Usage: `marked.parse(markdownString)` returns rendered HTML.

## Rationale

- **Smallest footprint:** marked.js is ~16KB gzipped — the smallest of the
  mainstream JavaScript markdown libraries. A single `<script>` tag from a
  CDN is all that's needed, with no build step or bundler.
- **No server-side complexity:** Server-side rendering would require adding
  a Go markdown-to-HTML library (goldmark is already used for parsing, but
  adding HTML rendering to the API responses would couple the API format to
  the UI's display needs). Keeping rendering client-side keeps the API
  responses clean (raw markdown) and the server simple.
- **CDN delivery:** Loading from a CDN means zero impact on the Go binary
  size and leverages browser caching. The library loads in parallel with the
  page and is cached across visits.
- **Constitution compliance:** Principle III (Skateboard First) — this is the
  simplest way to get markdown rendering without a build step. Principle IV
  (Idiomatic Go) — keeps the Go server focused on data, not presentation.

## Consequences

### Positive
- No build step or node_modules
- API responses remain clean markdown (usable by other clients)
- Single `<script>` tag — trivial to add and remove
- Fast rendering — marked.js is one of the fastest JavaScript markdown parsers

### Negative
- CDN dependency — if the CDN is unreachable, markdown renders as raw text
- No server-side rendering for SEO (not relevant for this use case)
- Client-side rendering means a brief flash of unrendered content

### Mitigations
- For CDN failure: the `go:embed` approach (ADR-013) could bundle a local
  copy of marked.js as a fallback if needed in the future
- Flash of unrendered content is minimal given the small library size

## Alternatives Considered

### markdown-it
- **Not chosen because:** ~33KB gzipped — double the size of marked.js.
  Its extensibility (plugins for footnotes, abbreviations, etc.) solves
  problems this project doesn't have.

### showdown.js
- **Not chosen because:** ~24KB gzipped and the slowest of the three
  mainstream options. No meaningful advantages over marked.js for this
  use case.

### Server-side rendering (goldmark HTML output)
- **Not chosen because:** Would require the API to return pre-rendered
  HTML instead of markdown, coupling the API format to the web UI. Other
  potential API consumers (CLI, other UIs) would receive HTML they don't
  need. Also adds complexity to the Go handlers.

## Related ADRs

- ADR-007: Goldmark Markdown Parsing — goldmark parses ADR files server-side
  for chunking; marked.js renders markdown client-side for display. Different
  tools for different sides of the stack.
