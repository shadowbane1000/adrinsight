# ADR-010: Full ADR Expansion from Filesystem for RAG Context

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

The RAG pipeline retrieves ADR chunks (individual sections) via similarity
search. However, sending only the matched chunks to the LLM for synthesis
loses context. For example, a search for "why did we choose Go?" might match
the "Decision" section ("Use Go") but miss the "Rationale" section that
explains why. The LLM needs enough context to produce a complete, accurate
answer with proper citations.

The question is how much context to send: just matched chunks, full ADRs for
matched chunks, or full ADRs plus related ADRs.

## Decision

When a chunk matches during similarity search, read the full original ADR
file from the filesystem (using the stored `ADRPath`) and send the complete
ADR content to the synthesis LLM. Deduplicate by ADR number so each ADR is
sent at most once, even if multiple sections matched.

## Rationale

- **Complete context**: Each ADR is a coherent document where Context, Decision,
  Rationale, and Consequences are interdependent. Sending the full ADR ensures
  the LLM has everything it needs to synthesize an accurate answer.
- **Corpus scale makes this free**: At the current scale (5-50 ADRs, ~500 words
  each), sending 3-5 full ADRs is roughly 1,500-2,500 words — well within
  Claude's context window and negligible compared to the model's capacity.
- **Simpler than store queries**: The filesystem already has the complete,
  authoritative content. Reading files avoids adding a "get all chunks by ADR
  number" method to the Store interface and avoids reassembling sections from
  the chunks table.
- **Skateboard first**: Constitution Principle III. This is the simplest
  approach that produces good synthesis quality. Optimization (selective
  context, related ADR expansion) is deferred to Phase 2.

## Consequences

### Positive
- LLM always has full context for every referenced ADR
- No information loss between retrieval and synthesis
- No new Store interface methods needed
- Implementation is trivial: `os.ReadFile(adrPath)`

### Negative
- Filesystem coupling: the RAG pipeline depends on ADR files being present
  on disk at the paths stored during indexing. If files move or are deleted
  after indexing, expansion fails.
- Token cost scales with ADR length, not relevance. A very long ADR where
  only one section is relevant still sends the full document.
- Doesn't follow "Related ADRs" references — a question about sqlite-vec
  gets ADR-004 but not ADR-006 (driver choice) unless it also matched.

### Mitigations
- The `reindex` command rebuilds from scratch, so stale paths are corrected
  on the next reindex run
- At current corpus scale, token cost is negligible
- Related ADR expansion is a natural Phase 2 enhancement — the architecture
  supports it without changes to the current design

## Alternatives Considered

### Send matched chunks only
- **Rejected because:** Real-world testing showed that matched chunks often
  miss critical context. A "Decision" chunk without its "Rationale" produces
  a thin, unsatisfying synthesis.

### Store method to retrieve all chunks by ADR number
- **Rejected because:** Adds complexity to the Store interface when the
  filesystem already has the data. The reassembled content would also lose
  the original markdown formatting and any content between sections.

### Related ADR expansion (follow references)
- **Deferred to Phase 2:** Parse "Related ADRs" sections and include
  referenced ADRs in the context. Good idea but adds complexity (cycle
  detection, depth limits) that isn't justified until we have quality
  feedback from real usage.

## Related ADRs

- ADR-004: SQLite with Vector Extension for Storage — chunks are the retrieval
  unit; this ADR explains why we expand beyond chunks for synthesis
- ADR-008: Section-Based Chunking — the chunking strategy that creates the
  "section-level retrieval, full-ADR synthesis" pattern
