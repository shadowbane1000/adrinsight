# ADR-008: Section-Based Chunking for Milestone 1

**Status:** Accepted (validated by ADR-022)  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

Before embedding, ADR documents must be split into chunks — the units that get individually embedded and retrieved during similarity search. The chunking strategy directly affects retrieval quality: chunks that are too large dilute the semantic signal, while chunks that are too small lose context. This decision will be revisited in Phase 2 after real-world experience with retrieval quality.

## Decision

Split ADR documents by H2 headings (## Context, ## Decision, ## Rationale, etc.). Each section becomes one chunk. If an ADR has no H2 headings, the entire body is a single chunk.

## Rationale

- **Natural boundaries:** ADR sections are semantically distinct — Context describes the problem, Decision states the choice, Rationale explains why, Consequences describes impact. These are the natural units a user would want retrieved.
- **Right granularity for this corpus:** ADR sections are typically 1-5 paragraphs. This fits comfortably within nomic-embed-text's 8192 token context window and produces vectors with focused semantic content.
- **Simple to implement:** goldmark's AST walking already identifies H2 headings. Splitting by heading requires no additional dependencies, heuristics, or configuration.
- **Skateboard first:** Constitution Principle III. This is the simplest strategy that could work. Experimenting with overlap, sliding windows, or recursive splitting is deferred to Phase 2.

## Consequences

### Positive
- Zero-configuration — no chunk size parameters to tune
- Semantically meaningful chunks aligned with ADR structure
- Trivial to implement and test
- Each chunk naturally carries context about what aspect of the decision it represents (via the section heading)

### Negative
- No overlap between chunks — a query about something that spans Context and Decision may not match either chunk perfectly
- Very short sections (one sentence) produce thin vectors with less semantic content
- Very long sections are not subdivided — a rambling Consequences section could dilute the embedding
- Not generalizable to non-ADR documents without modification

### Mitigations
- At the current corpus scale (5 ADRs, ~20 chunks), retrieval quality can be evaluated manually and quickly
- Phase 2 will revisit chunking with real retrieval data and document the comparison in a superseding ADR
- The chunking logic is isolated in the parser package, making it easy to swap strategies

## Alternatives Considered

### Fixed-size token windows
- **Deferred to Phase 2:** More sophisticated, handles varying section lengths, but adds complexity (token counting, overlap management) that isn't justified until we have retrieval quality data to improve.

### Whole-document embedding
- **Rejected because:** A full ADR (500-2000 words) produces a vector that blends all topics. A query about "why Go" would match the ADR-001 vector, but the retrieval can't point to the specific section — losing the citation granularity needed for Milestone 2's synthesis.

### Paragraph-level splitting
- **Deferred to Phase 2:** Too granular for ADRs where paragraphs within a section are tightly coupled. Would increase chunk count without clear retrieval benefit at this scale.

## Related ADRs

- ADR-007: goldmark for Markdown Parsing — provides the AST walking that makes section splitting straightforward
