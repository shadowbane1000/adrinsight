# ADR-002: Anthropic Claude for LLM Synthesis

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight retrieves relevant ADR documents via semantic search and then synthesizes a natural-language answer with citations. This synthesis step requires a large language model capable of:

- Reasoning over multiple retrieved documents
- Producing accurate, well-cited summaries
- Following structured output instructions (e.g., returning citations in a parseable format)

The major LLM API providers considered were OpenAI (GPT-4+), Anthropic (Claude), and Google (Gemini). Local models via Ollama were also considered for synthesis but are addressed separately from their role in embeddings (see ADR-003).

## Decision

Use Anthropic's Claude API for the LLM synthesis/answer generation step.

## Rationale

- **Existing account and familiarity:** I have an active Anthropic account and have used Claude extensively for AI-assisted development work. Reducing friction on the LLM integration lets me focus learning energy on Go.
- **Strong instruction-following:** Claude handles structured output well, which matters for returning answers with precise citations back to specific ADRs.
- **Long context window:** Claude's large context window accommodates passing multiple full ADR documents alongside the query, reducing the need for aggressive chunking or summarization before synthesis.
- **Sufficient for scope:** For a corpus of dozens to low hundreds of ADRs, Claude's capabilities are more than adequate. This isn't a cost-sensitive, high-volume production workload — it's a portfolio project with occasional interactive queries.

## Alternatives Considered

- **OpenAI GPT-4:** Equally capable, but would require setting up a new account and API key. No meaningful technical advantage for this use case.
- **Local LLM via Ollama:** Eliminates the external API dependency entirely, which is appealing for the "clone and run" story. However, synthesis quality from local models (7B–13B parameter range) is noticeably worse for multi-document reasoning tasks. The embedding step (ADR-003) already uses Ollama, so the project still demonstrates local model integration.
- **Google Gemini:** Capable, but less personal familiarity and no existing account.

## Consequences

### Positive
- Fast integration — well-documented HTTP API, familiar tool
- High-quality synthesis with accurate citation behavior
- Large context window reduces retrieval pipeline complexity

### Negative
- External API dependency — users need their own Anthropic API key to run the synthesis step
- Cost per query (minimal at portfolio-project scale, but nonzero)
- Vendor lock-in on the synthesis path

### Mitigations
- Abstract the LLM provider behind a Go interface from the start, making it straightforward to add OpenAI, Ollama, or other backends later
- Document API key setup clearly in the README
- The roadmap includes multi-provider support as a Phase 4 stretch goal
