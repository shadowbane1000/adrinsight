# ADR-009: Structured JSON Output via Anthropic OutputConfig

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight's RAG pipeline needs to return a synthesized answer alongside
structured citations (ADR number, title, section). The LLM generates the answer
and must identify which ADRs it drew from. The system needs to parse the LLM
response into a clean JSON structure for the HTTP API — the web UI in Milestone 3
will render citations as clickable links, so they must be reliably extracted.

The challenge is that LLMs produce free-form text by default. Getting structured
data out requires either parsing the output or constraining the output format.

## Decision

Use the Anthropic Messages API `OutputConfig` with `JSONOutputFormatParam` and a
JSON schema to constrain Claude's output to the exact structure needed (answer
text + citations array).

## Rationale

- **Guaranteed valid JSON**: Anthropic's constrained decoding ensures every
  response matches the provided schema. No parsing failures, no retry loops,
  no regex extraction.
- **Schema-driven contract**: The JSON schema defines the exact response structure
  (`answer` string + `citations` array of objects). This acts as a contract
  between the LLM layer and the HTTP handler — if the schema is wrong, it's
  caught at compile time, not at runtime.
- **No prompt engineering fragility**: Alternative approaches (system prompt
  instructions asking for JSON) work most of the time but fail unpredictably.
  Constrained decoding eliminates this class of bugs entirely.
- **Minimal latency overhead**: First request with a new schema has ~100-300ms
  compilation overhead (cached for 24 hours). Subsequent requests have no
  additional cost.

## Consequences

### Positive
- Response parsing is trivial — `json.Unmarshal` into a Go struct
- No retry logic needed for malformed LLM output
- Citation extraction is deterministic, not heuristic
- Schema change is a code change, not a prompt change

### Negative
- Ties the implementation to Anthropic's OutputConfig feature — other LLM
  providers may not support constrained decoding
- Schema compilation adds a small latency hit on cold starts
- The LLM is constrained in how it structures its response — it can't add
  unexpected but useful fields

### Mitigations
- The LLM interface (Constitution Principle I) abstracts this — a different
  provider could use tool-use or prompt-based JSON extraction behind the same
  interface
- Cold start latency is negligible compared to the ~2-5 second LLM generation
  time

## Alternatives Considered

### System prompt instructions only
- **Rejected because:** No guarantee of valid JSON. Claude usually complies
  but occasionally produces malformed output, requires retry logic, and makes
  the system less reliable.

### Tool use / function calling
- **Rejected because:** Designed for agent loops where the LLM decides which
  tool to call. Overkill for "always return this exact structure." Adds
  complexity without benefit for a single-response pattern.

### Regex extraction from free text
- **Rejected because:** Fragile. Requires maintaining regex patterns that
  match the LLM's citation format, which can change between model versions.
  Loses structured data (e.g., section names) that don't fit neatly into
  regex groups.

## Related ADRs

- ADR-002: Anthropic Claude for LLM Synthesis — establishes Anthropic as the
  synthesis provider; this ADR selects how to get structured output from it
