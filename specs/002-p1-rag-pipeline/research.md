# Research: RAG Pipeline and HTTP API

## Anthropic Go SDK

**Decision**: Use the official Anthropic Go SDK (`github.com/anthropics/anthropic-sdk-go`)

**Rationale**: Anthropic maintains an official Go SDK. It reads `ANTHROPIC_API_KEY`
from env automatically, provides typed model constants (`ModelClaudeSonnet4_5`,
etc.), and handles the Messages API cleanly. No reason to use raw HTTP.

**Alternatives considered**:
- Raw HTTP to Messages API — works but reinvents auth, serialization, error handling
- Third-party SDKs — unnecessary given official SDK exists

## HTTP Server Framework

**Decision**: Use Go 1.22+ `net/http` stdlib ServeMux

**Rationale**: Go 1.22 added method-based routing and path parameters to the
standard library's ServeMux (`mux.HandleFunc("GET /adrs/{number}", handler)`).
This covers all three endpoints without any dependency. Constitution Principle IV
says "lean on the standard library before reaching for third-party packages."

**Alternatives considered**:
- Chi — excellent router with middleware, but overkill for 3 endpoints
- Echo/Gin — feature-rich frameworks, unnecessary complexity for this scope
- gorilla/mux — maintenance mode, and stdlib now covers its key features

## Structured JSON Output from Claude

**Decision**: Use `OutputConfig` with `JSONOutputFormatParam` and a JSON schema

**Rationale**: The Anthropic API supports constrained decoding that guarantees
valid JSON matching a provided schema. This eliminates the need to parse
free-form LLM output or retry on malformed responses. Define a schema with
`answer` (string) and `citations` (array of objects), and Claude is constrained
to produce exactly that structure.

**Alternatives considered**:
- Tool use / function calling — more complex, designed for agent loops not structured output
- System prompt instructions only — no guarantee of valid JSON, requires parsing and retries
- Regex extraction from free text — fragile, loses structure

## RAG Context Expansion Strategy

**Decision**: Full ADR expansion — read original markdown files for matched ADRs

**Rationale**: Search returns chunks (individual sections), but synthesis needs
full context. When a chunk matches, read the original ADR file from disk using
the stored `ADRPath` to get the complete ADR content. At current corpus scale
(5-50 ADRs, ~500 words each), sending 3-5 full ADRs fits easily within Claude's
context window. This avoids building a separate "get full ADR" query into the
store layer.

**Alternatives considered**:
- Send matched chunks only — loses context (e.g., Decision without Rationale)
- Store method to retrieve all chunks by ADR number — adds complexity to the
  store interface when the filesystem already has the complete data
- Related ADR expansion — follow "Related ADRs" references for deeper context.
  Deferred to Phase 2.

## Anthropic Model Selection

**Decision**: Default to `claude-sonnet-4-5` with model configurable via flag/env

**Rationale**: Sonnet is the best balance of quality and cost for synthesis tasks.
The model should be configurable so users can upgrade to Opus or downgrade to
Haiku based on their needs and budget.

**Alternatives considered**:
- Hard-code Opus — highest quality but most expensive, overkill for ADR synthesis
- Hard-code Haiku — cheapest but may produce lower quality citations
