# Research: ADR Intelligence

## R1: Relationship Extraction from ADR Markdown

**Decision**: Parse "Related ADRs" H2 sections using the existing goldmark AST walker, extract ADR number references via regex (`ADR-(\d+)`), then classify each relationship using Haiku.

**Rationale**: The Related ADRs sections use natural language descriptions with varied phrasing. Regex reliably extracts the target ADR number, and LLM classification handles the type without brittle keyword matching. This mirrors the keyword extraction pattern already established in M4b.

**Alternatives considered**:
- Regex-only classification (match "supersedes", "drives", etc.) — rejected because phrasing varies and new ADRs may use unexpected wording.
- Parse entire ADR body for cross-references — rejected because Related ADRs sections are the canonical location, and body references are contextual, not structural.

## R2: Relationship Type Taxonomy

**Decision**: Five relationship types: `supersedes`, `superseded_by`, `depends_on`, `drives`, `related_to`. These are stored as directed edges.

**Rationale**: Covers all patterns observed in the current 17 ADRs:
- `supersedes` / `superseded_by`: ADR-015 supersedes ADR-006
- `depends_on`: ADR-008 depends on ADR-007 (goldmark provides AST walking)
- `drives`: ADR-004 drives ADR-014 (SQLite → CGO → Docker build)
- `related_to`: catch-all for topical relationships

**Alternatives considered**:
- Free-form relationship types — rejected because consistent types enable structured queries and UI grouping.
- Bidirectional edges only — rejected because directionality matters (A supersedes B is not the same as B supersedes A).

## R3: Relationship Storage Schema

**Decision**: New `adr_relationships` table in SQLite with columns: `id`, `source_adr`, `target_adr`, `rel_type`, `description`. Rebuilt during reindex (dropped and recreated alongside other tables in `Reset()`).

**Rationale**: Follows the existing pattern — all indexed data is derived and disposable, rebuilt from source ADR files. No migration needed. The table is small (dozens of rows) and queried by ADR number, so a simple B-tree index suffices.

**Alternatives considered**:
- In-memory only (Go map) — rejected because the data should be available to the serve command without re-parsing ADR files.
- Separate graph database — rejected per assumption: graph is tiny and fits in SQLite.

## R4: LLM Classification for Relationship Types

**Decision**: Use Haiku to classify each "Related ADRs" bullet into one of the five relationship types. Send the source ADR title, the bullet text, and the type options. Parse the response as a single type string.

**Rationale**: Haiku is fast and cheap (~$0.001 per 17 ADRs). The classification task is simple (pick one of five types given context) and doesn't need a larger model. This runs during reindex alongside keyword extraction — no additional latency at query time.

**Alternatives considered**:
- Use the same Haiku call for both keyword extraction and relationship classification — rejected because they operate on different inputs (whole ADR body vs individual relationship bullets).
- Local model (Ollama) — rejected based on M4b keyword extraction testing where the 3B model missed important nuances. Haiku is more reliable for classification.

## R5: Relationship Context in LLM Synthesis

**Decision**: When building the ADR context for synthesis, prepend a "Relationship Context" block before the ADR content that lists all relationships for the retrieved ADRs. Format: "ADR-015 supersedes ADR-006. ADR-004 drives ADR-014."

**Rationale**: The LLM already receives full ADR text which includes "Related ADRs" sections. But the LLM may not reliably extract relationship types from varied prose. A structured relationship summary gives the LLM explicit, unambiguous relationship data to reference in answers.

**Alternatives considered**:
- Rely on the LLM reading the "Related ADRs" section from the full ADR text — rejected because it already fails to trace chains (current behavior).
- Inject relationships into each ADR's content — rejected because it modifies the source text and could confuse the LLM about what's in the original ADR vs what's added.

## R6: Relationship-Aware Retrieval

**Decision**: After hybrid search returns initial results, expand the result set by adding directly-related ADRs (1 hop) that aren't already in the results. For supersession chains, walk the full chain. Cap expansion at topK * 2 total ADRs to bound LLM context size.

**Rationale**: The biggest retrieval gap is missing related ADRs. If a query matches ADR-015, the system should also retrieve ADR-006 (superseded) and ADR-004 (drives). This is more targeted than increasing topK globally.

**Alternatives considered**:
- Only expand supersession chains — rejected because depends_on/drives relationships are equally important for "affected by" queries.
- Expand at 2+ hops — rejected for now; 1-hop expansion with full supersession chain walking covers the known query patterns.

## R7: Standardized Relationship Format for New ADRs

**Decision**: Define a convention in the ADR template: each bullet in "Related ADRs" should include a bracketed type tag. Format: `- ADR-NNN: Title [type] — description`. Supported types: `[supersedes]`, `[superseded-by]`, `[depends-on]`, `[drives]`, `[related]`.

**Rationale**: Makes relationships machine-parseable without LLM classification for well-formatted ADRs. The LLM classifier serves as fallback for legacy or human-authored ADRs without tags. The tags are optional — the system works without them.

**Alternatives considered**:
- YAML frontmatter for relationships — rejected because it deviates from the established markdown-body approach and makes ADRs less readable.
- Separate relationship manifest file — rejected because it separates relationship data from the ADR it belongs to, creating sync issues.
