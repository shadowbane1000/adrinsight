# ADR-003: Local Embeddings via Ollama

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight needs to convert ADR text into vector embeddings for semantic search. When a user asks a question, the question is also embedded and compared against the stored ADR vectors to retrieve the most relevant documents before passing them to the LLM for synthesis.

The embedding step has different requirements than synthesis:

- It runs at indexing time (once per ADR, re-run on change) and at query time (once per question)
- Embedding quality needs to be "good enough" to rank a small corpus (dozens to hundreds of documents), not web-scale
- It heavily influences the developer experience — if a user has to sign up for a second API service just to generate embeddings, the onboarding friction doubles

## Decision

Use Ollama running locally (in Docker) with the `nomic-embed-text` model for all embedding operations.

## Rationale

- **Zero external dependencies for embeddings:** Users run `docker-compose up` and the embedding service starts automatically. No API key, no account, no cost. Combined with ADR-002 (Anthropic for synthesis), the project requires exactly one external API key total.
- **Quality is sufficient for scope:** `nomic-embed-text` produces 768-dimensional embeddings and benchmarks competitively with OpenAI's `text-embedding-3-small` on retrieval tasks. For a corpus of dozens of ADRs, the quality difference is negligible.
- **Docker-native:** Ollama publishes official Docker images, making it easy to include in `docker-compose.yml` alongside the main application.
- **Demonstrates architectural judgment:** Using a local model for embeddings and a cloud API for synthesis shows intentional tool selection — picking the right tool for each job rather than defaulting to one provider for everything.
- **Offline development:** Embedding generation works without internet access, which is convenient during development.

## Alternatives Considered

- **OpenAI Embeddings API:** Higher benchmark scores on some retrieval tasks, but adds a second API key requirement, a second vendor dependency, and per-request cost. The quality advantage doesn't matter at this corpus scale.
- **Anthropic Embeddings:** Not available — Anthropic does not offer an embeddings API.
- **Sentence-transformers (Python sidecar):** Would produce good embeddings but introduces a Python dependency into an otherwise pure-Go project, complicating the build and Docker setup.

## Consequences

### Positive
- Clean `docker-compose up` experience with no additional account setup for embeddings
- No per-request cost for embedding operations
- Demonstrates thoughtful multi-model architecture (local embeddings + cloud synthesis)
- Works offline for development and indexing

### Negative
- First run requires downloading the embedding model (~275MB), adding to initial startup time
- Ollama in Docker increases the overall memory footprint of the local development environment
- Embedding quality ceiling is lower than the best commercial APIs (irrelevant at this scale, but worth noting)

### Mitigations
- Document the first-run model download in the README and consider a setup script that pre-pulls the model
- Include memory requirements in documentation
- The LLM provider interface abstraction (ADR-002) extends to embeddings — swapping to a cloud embedding provider later is a configuration change, not a rewrite
