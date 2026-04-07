# Feature Specification: Retrieval Improvements

**Feature Branch**: `005-p2-retrieval-improvements`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Milestone M4b"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Better Answers to Sample Queries (Priority: P1)

A developer asks the six sample queries from the About page and receives
answers that correctly cite all relevant ADRs. The evaluation harness scores
improve measurably over the M4a baseline, demonstrating that retrieval changes
translate into better answers.

**Why this priority**: The sample queries are the project's demo — they must
produce high-quality answers. The M4a harness provides the measurement
framework; this story delivers the actual improvement.

**Independent Test**: Run `./adr-insight eval` and verify all test case scores
meet or exceed the M4a baseline, with measurable improvement in aggregate
recall and F1.

**Acceptance Scenarios**:

1. **Given** the M4a baseline is saved, **When** a developer runs the evaluation
   after retrieval improvements, **Then** aggregate recall improves by at least
   0.15 over baseline and no individual test case regresses.

2. **Given** a query uses exact technical terms (e.g., "sqlite-vec", "goldmark",
   "mattn"), **When** the system searches for relevant ADRs, **Then** ADRs
   containing those exact terms are retrieved even if the vector similarity
   alone would not rank them highly.

3. **Given** a query spans multiple ADR topics (e.g., "single binary
   constraint"), **When** the system retrieves results, **Then** ADRs from
   all relevant topics are included, not just the single closest vector match.

---

### User Story 2 - Evaluate Chunking Alternatives (Priority: P1)

A developer experiments with alternative chunking strategies and measures each
against the evaluation harness to determine which produces the best retrieval
quality. The winning strategy is documented in an ADR that supersedes or amends
ADR-008.

**Why this priority**: The current H2-section chunking was explicitly deferred
as "skateboard first" in ADR-008. Three alternatives were identified for Phase 2
evaluation. This is the research deliverable that informs the chunking decision.

**Independent Test**: At least two alternative chunking strategies are evaluated,
with before/after scores documented. An ADR records the decision.

**Acceptance Scenarios**:

1. **Given** the current H2-section chunking strategy, **When** a developer runs
   the evaluation harness, **Then** baseline scores are recorded for comparison.

2. **Given** an alternative chunking strategy is implemented, **When** the
   developer runs the evaluation harness, **Then** scores for the alternative
   are recorded alongside the baseline for direct comparison.

3. **Given** multiple strategies have been evaluated, **When** the developer
   selects the best strategy, **Then** an ADR is created documenting the
   experiment results, the chosen strategy, and why alternatives were rejected.

---

### User Story 3 - Keyword-Aware Search (Priority: P2)

A developer asks a question using specific technical terminology (library
names, acronyms, configuration flags) and the system finds ADRs containing
those exact terms, even when vector similarity would rank other ADRs higher.
This complements the existing vector search rather than replacing it.

**Why this priority**: Vector search excels at semantic similarity but misses
exact keyword matches. Searching for "ncruces" should surface ADR-006 and
ADR-015 regardless of semantic distance from the query embedding.

**Independent Test**: Search for "ncruces" and verify ADR-006/015 appear in
results. Search for "goldmark" and verify ADR-007 appears.

**Acceptance Scenarios**:

1. **Given** a query contains an exact technical term present in an ADR,
   **When** the system searches, **Then** the ADR containing that term is
   included in results even if it would not rank in the top-K by vector
   similarity alone.

2. **Given** a query matches both semantically and by keyword, **When** results
   are ranked, **Then** the ADR that matches on both dimensions ranks higher
   than one matching on only one dimension.

3. **Given** a query contains no specific technical terms, **When** the system
   searches, **Then** results are equivalent to pure vector search (keyword
   search adds no noise).

---

### User Story 4 - Result Reranking (Priority: P3)

After initial retrieval, results are re-scored using heuristics that improve
relevance. For example, exact title matches are boosted and deprecated/
superseded ADRs are deprioritized in favor of their successors.

**Why this priority**: Simple heuristics can significantly improve result
ordering without the complexity of a separate ML model. This is low-effort,
high-signal polish.

**Independent Test**: Query "SQLite driver" and verify ADR-015 (active) ranks
above ADR-006 (superseded).

**Acceptance Scenarios**:

1. **Given** a query matches an ADR title closely, **When** results are ranked,
   **Then** that ADR is boosted higher in the results.

2. **Given** a superseded ADR and its successor are both retrieved, **When**
   results are ranked, **Then** the active ADR ranks above the superseded one.

3. **Given** reranking is applied, **When** the evaluation harness runs,
   **Then** scores do not regress compared to pre-reranking results.

### Edge Cases

- Query consists entirely of a library name with no surrounding context
- ADR has very short sections that produce thin embeddings
- Query matches many ADRs equally (broad question like "what decisions were made?")
- Keyword search returns results with zero vector similarity
- All retrieved ADRs are superseded (no active successor in the result set)
- Chunking strategy change alters the database schema (requires full reindex)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support combining keyword-based and vector-based
  search results using a weighted merge — retrieve from both methods
  independently, combine scores with configurable weights (e.g., 0.7 vector
  + 0.3 keyword), deduplicate by ADR, and rank by combined score
- **FR-002**: At least two alternative chunking strategies MUST be evaluated
  against the M4a evaluation harness with documented before/after scores
- **FR-003**: The chosen chunking strategy MUST be documented in an ADR that
  references ADR-008 (supersedes or amends)
- **FR-004**: System MUST support reranking retrieved results using configurable
  heuristics (title match boosting, status-aware deprioritization)
- **FR-005**: Each retrieval change MUST be measured against the M4a baseline
  before being shipped — changes that regress scores MUST NOT be merged
- **FR-006**: The system MUST maintain backward compatibility — the reindex
  command rebuilds the search index for the new strategy without requiring
  manual migration
- **FR-007**: Keyword search MUST NOT degrade results for queries that contain
  no specific technical terms (pure semantic queries should perform at least
  as well as before)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Aggregate recall across the 6-question test corpus improves by
  at least 0.15 over the M4a baseline
- **SC-002**: No individual test case regresses beyond the configured delta
  threshold (0.2)
- **SC-003**: At least two alternative chunking strategies are evaluated with
  documented scores
- **SC-004**: A query for a specific library name (e.g., "goldmark") retrieves
  the corresponding ADR within the top 3 results
- **SC-005**: The evaluation harness passes after all changes are applied

## Clarifications

### Session 2026-04-07

- Q: How should keyword and vector results be combined? → A: Weighted merge — retrieve from both methods independently, combine scores with configurable weights (e.g., 0.7 vector + 0.3 keyword), deduplicate by ADR, rank by combined score.

## Assumptions

- The M4a evaluation harness and baseline are in place and functional —
  all improvements are measured against this baseline
- Chunking strategy changes require a full reindex — this is acceptable
  since the reindex command already exists and the database is a derived
  artifact
- The current ADR corpus (16 ADRs) is small enough that keyword search
  overhead is negligible
- Hybrid search combines results from two search methods, not replaces one
  with the other — vector search remains the primary retrieval mechanism
- Reranking heuristics are simple, deterministic functions — no ML models
  or external services
- Search query preprocessing (abbreviation expansion, question pattern
  handling) is optional and only implemented if time permits — the higher
  priority items are chunking, hybrid search, and reranking
