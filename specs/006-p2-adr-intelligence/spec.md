# Feature Specification: ADR Intelligence

**Feature Branch**: `006-p2-adr-intelligence`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "milestone 5"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Relationship-Aware Answers (Priority: P1)

A user asks a question that spans multiple related ADRs (e.g., "what happened with the SQLite driver decision?"). The system traces the relationship chain — ADR-004 (SQLite chosen) to ADR-006 (ncruces/WASM driver) to ADR-015 (superseded by mattn/CGO) — and synthesizes an answer that tells the story of the decision evolution, not just isolated facts from individual ADRs.

**Why this priority**: This is the core value proposition of M5. Without relationship context in synthesis, the system treats each ADR as an island and cannot answer questions about decision chains, which are the most interesting questions about architecture.

**Independent Test**: Ask "What happened with the SQLite driver decision?" and verify the answer traces the supersession chain from ADR-004 through ADR-006 to ADR-015, explaining why the driver changed.

**Acceptance Scenarios**:

1. **Given** the system has indexed ADRs with relationship metadata, **When** a user asks about a decision that spans multiple ADRs in a supersession chain, **Then** the answer describes the decision evolution in chronological order and cites all ADRs in the chain.
2. **Given** a query mentions a superseded ADR, **When** the system retrieves context for synthesis, **Then** the newer ADR is included alongside the superseded one, and the relationship is made explicit in the answer.
3. **Given** a query about "decisions affected by X," **When** the system searches, **Then** it follows relationship links to find downstream ADRs that depend on or were driven by the queried decision.

---

### User Story 2 - Relationship Parsing and Storage (Priority: P1)

During reindexing, the system parses "Related ADRs" sections from each ADR and stores the relationships as structured data. Relationships include supersession (supersedes/superseded by), dependency (depends on/drives), and topical relatedness.

**Why this priority**: This is the foundational infrastructure that all other stories depend on. Without parsed relationships, no other M5 feature can function.

**Independent Test**: After reindexing, query the relationship store and verify that ADR-015's relationship to ADR-006 is stored as "supersedes" and ADR-006's relationship to ADR-015 is stored as "superseded by."

**Acceptance Scenarios**:

1. **Given** ADR files with "Related ADRs" sections containing natural language descriptions, **When** reindex runs, **Then** relationships are extracted and stored with source ADR, target ADR, and relationship type.
2. **Given** an ADR whose Status field says "Superseded by ADR-NNN," **When** reindex runs, **Then** the supersession relationship is captured even if the Related ADRs section is incomplete.
3. **Given** the system is reindexed, **When** the relationship store is queried for a specific ADR, **Then** both inbound and outbound relationships are returned.

---

### User Story 3 - Relationship Display in UI (Priority: P2)

When viewing an ADR in the detail panel, the user sees a list of related ADRs with their relationship type and can click them to navigate directly.

**Why this priority**: Adds discoverability and browsing value to the UI, but the core answer quality improvement (US1) is more impactful for demonstrating engineering quality.

**Independent Test**: Open ADR-015 in the detail panel and verify it shows "Supersedes: ADR-006" and "Related: ADR-004, ADR-014" as clickable links that navigate to those ADRs.

**Acceptance Scenarios**:

1. **Given** an ADR with stored relationships, **When** the user views it in the detail panel, **Then** related ADRs are displayed grouped by relationship type with clickable links.
2. **Given** the user clicks a related ADR link, **When** the target ADR loads, **Then** it shows the reciprocal relationship back to the source ADR.
3. **Given** an ADR with no relationships, **When** the user views it, **Then** no relationship section is displayed (no empty state clutter).

---

### User Story 4 - Status-Aware Retrieval Boost (Priority: P3)

When searching, the system uses relationship data to boost retrieval quality. Superseded ADRs are deprioritized in favor of their successors, and relationship-linked ADRs are more likely to be co-retrieved when one is relevant.

**Why this priority**: Incremental improvement on top of the existing reranker from M4b. The reranker already penalizes superseded ADRs by content heuristic; this replaces the heuristic with authoritative relationship data.

**Independent Test**: Search for "SQLite driver" and verify ADR-015 ranks above ADR-006, with the ranking informed by the supersession relationship rather than just content heuristics.

**Acceptance Scenarios**:

1. **Given** an ADR is marked as superseded in the relationship store, **When** search results include it, **Then** its score is reduced relative to the superseding ADR.
2. **Given** a query matches an ADR that has related ADRs, **When** the system retrieves results, **Then** strongly related ADRs are boosted in the result set.

---

### Edge Cases

- What happens when a Related ADRs section references an ADR number that doesn't exist in the collection?
- How does the system handle circular relationships (ADR-A relates to ADR-B which relates to ADR-A)? → Traversal tracks visited nodes to prevent infinite loops.
- What happens when an ADR is superseded by an ADR that is itself superseded (chain of three or more)?
- How does the system handle malformed Related ADRs sections with inconsistent formatting?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST parse "Related ADRs" sections from ADR markdown files and extract structured relationship data (source ADR, target ADR, relationship type, description). Relationship type classification MUST use an LLM to handle varied natural language phrasing in both legacy and human-authored ADRs.
- **FR-002**: System MUST recognize supersession relationships from both the "Related ADRs" section content and the ADR "Status" field (e.g., "Superseded by ADR-015").
- **FR-010**: The project MUST define a standardized format convention for the "Related ADRs" section in new ADRs (e.g., explicit relationship type tags), documented in the ADR template and project instructions.
- **FR-003**: System MUST store relationships in a persistent store that survives server restarts and is rebuilt during reindexing.
- **FR-004**: System MUST support querying relationships for a given ADR, returning both inbound and outbound relationships.
- **FR-005**: System MUST include relationship metadata when providing ADR context to the LLM for synthesis, so answers can reference decision chains.
- **FR-006**: System MUST expose relationship data via the existing HTTP API so the web UI can display it.
- **FR-007**: System MUST display related ADRs in the detail panel as clickable navigation links grouped by relationship type.
- **FR-008**: System MUST use authoritative relationship data (from parsing) to inform the reranker's supersession penalty, replacing the current content-heuristic approach.
- **FR-009**: System MUST handle missing or malformed relationship references gracefully (log a warning, skip the relationship, continue processing).

### Key Entities

- **ADRRelationship**: Represents a directed relationship between two ADRs — source ADR number, target ADR number, relationship type (supersedes, superseded_by, depends_on, drives, related_to), and a free-text description.
- **RelationshipGraph**: The complete set of relationships across all ADRs, queryable by ADR number with unlimited-depth traversal for chains (e.g., walk supersession chain to find the current active ADR). Traversal MUST track visited nodes to prevent infinite loops on circular relationships.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Questions about decision chains (e.g., "what happened with the SQLite driver?") produce answers that trace the full supersession chain and cite all ADRs in the chain.
- **SC-002**: The eval harness shows no regression in retrieval metrics (precision, recall, F1) compared to the M4b baseline.
- **SC-003**: All ADR detail views display correct relationship links, verified by manually checking ADR-015 (3 relationships), ADR-006 (2 relationships), and ADR-004 (multiple dependents).
- **SC-004**: The reranker uses authoritative supersession data — searching for "SQLite driver" ranks ADR-015 above ADR-006 based on relationship data, not content heuristics.
- **SC-005**: Reindexing with relationship parsing completes in under 30 seconds for the current 17-ADR corpus (no significant performance regression from M4b).

## Clarifications

### Session 2026-04-07

- Q: Graph traversal depth limit for relationship queries? → A: No limit — traverse the full graph. Track visited nodes to prevent infinite loops on circular relationships.

## Assumptions

- The "Related ADRs" section format is generally consistent across existing ADRs (H2 heading, bulleted list with "ADR-NNN: Title — description" pattern), but human-authored or legacy ADRs may use varied phrasing.
- An LLM is used to classify relationship types from natural language descriptions during reindex. This handles varied phrasing ("supersedes", "replaces", "obsoletes", "built on top of") without brittle keyword matching.
- A standardized format convention will be defined for new ADRs going forward, but the LLM classifier ensures backward compatibility with any phrasing.
- The relationship graph is small (dozens of edges) and can be loaded entirely into memory. No graph database is needed.
- The existing eval test cases are sufficient for regression testing. New test cases for relationship-specific queries will be added.
- The web UI relationship display uses the existing vanilla JS approach (no framework needed).
