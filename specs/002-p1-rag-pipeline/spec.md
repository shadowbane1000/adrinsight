# Feature Specification: RAG Pipeline and HTTP API

**Feature Branch**: `002-p1-rag-pipeline`  
**Created**: 2026-04-06  
**Status**: Draft  
**Input**: User description: "Phase 1, Milestone 2"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ask a Question and Get a Synthesized Answer (Priority: P1)

A developer has indexed their project's ADRs (using the reindex command from
Milestone 1). They now want to ask a natural-language question about the
project's architectural decisions and receive a synthesized answer that cites
specific ADRs. They send their question to an HTTP endpoint and receive a
response containing a coherent answer with references to the ADR numbers and
sections that informed the answer.

**Why this priority**: This is the core value proposition of the entire project.
Everything built so far (parser, embedder, store) exists to make this moment
work. Without synthesis, the tool is just a search engine.

**Independent Test**: Can be tested by indexing sample ADRs, sending a question
via the HTTP endpoint, and verifying the response contains a synthesized answer
with at least one ADR citation.

**Acceptance Scenarios**:

1. **Given** ADRs have been indexed, **When** a user sends a question to the
   query endpoint, **Then** the system returns a synthesized natural-language
   answer that references specific ADR numbers.

2. **Given** ADRs have been indexed, **When** a user asks a question that spans
   multiple ADRs (e.g., "what are the key technology choices?"), **Then** the
   response synthesizes information from multiple ADRs and cites each one.

3. **Given** ADRs have been indexed, **When** a user asks a question not covered
   by any ADR, **Then** the system responds indicating it couldn't find relevant
   information rather than hallucinating an answer.

4. **Given** ADRs have been indexed, **When** a user asks a vague or broad
   question, **Then** the system still returns a reasonable response drawing
   from the most relevant ADRs available.

---

### User Story 2 - Browse Indexed ADRs via API (Priority: P2)

A developer or front-end client wants to see what ADRs are available in the
index. They call an HTTP endpoint that returns a list of all indexed ADRs with
their metadata (number, title, status, date).

**Why this priority**: This enables the future web UI (Milestone 3) to show an
ADR browser/list view. It also lets developers verify what's in the index
without running a search.

**Independent Test**: Can be tested by indexing sample ADRs and calling the
list endpoint to verify all indexed ADRs appear with correct metadata.

**Acceptance Scenarios**:

1. **Given** ADRs have been indexed, **When** a client requests the ADR list
   endpoint, **Then** the system returns all indexed ADRs with their number,
   title, status, and date.

2. **Given** no ADRs have been indexed (empty database), **When** a client
   requests the ADR list endpoint, **Then** the system returns an empty list
   without errors.

---

### User Story 3 - Retrieve a Single ADR's Full Content (Priority: P3)

A developer sees a citation in a synthesized answer (e.g., "ADR-001") and wants
to read the full ADR. They call an HTTP endpoint with the ADR number to retrieve
the complete content.

**Why this priority**: Citations are only useful if the user can drill down into
the source material. This completes the query-to-source loop.

**Independent Test**: Can be tested by indexing sample ADRs and requesting a
specific ADR by number to verify the full content is returned.

**Acceptance Scenarios**:

1. **Given** ADRs have been indexed, **When** a client requests a specific ADR
   by number, **Then** the system returns the full ADR content with metadata.

2. **Given** a client requests an ADR number that doesn't exist in the index,
   **When** the request is made, **Then** the system returns a not-found
   response without errors.

### Edge Cases

- Synthesis service is unavailable when a query is submitted
- Embedding service is unavailable when a query is submitted
- Database is empty (no ADRs indexed) when a query is submitted
- Query text is empty or whitespace-only
- Query text is extremely long (exceeds reasonable input length)
- Retrieved ADR chunks are insufficient for a meaningful answer
- Multiple concurrent query requests
- Synthesis response is unexpectedly empty or malformed

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a natural-language question and return a
  synthesized answer with citations to specific ADRs
- **FR-002**: System MUST retrieve relevant ADR chunks via similarity search,
  then expand the results to include the full content of each matched ADR
  (all sections, not just the matched chunk) before sending to the synthesis
  service for answer generation
- **FR-003**: System MUST return synthesized responses as structured data with
  the answer text separate from a citations array, where each citation
  identifies an ADR by number, title, and the section used
- **FR-004**: System MUST provide an HTTP endpoint for submitting questions
  and receiving synthesized answers
- **FR-005**: System MUST provide an HTTP endpoint that lists all indexed ADRs
  with their metadata (number, title, status, date)
- **FR-006**: System MUST provide an HTTP endpoint that returns the full content
  of a single ADR by number
- **FR-007**: System MUST return a meaningful response when no relevant ADRs are
  found for a query, rather than fabricating information
- **FR-008**: System MUST return appropriate error responses when required
  services (synthesis, embedding) are unavailable
- **FR-009**: System MUST accept configuration for the synthesis service
  credentials via environment variable or flag

### Key Entities

- **Query**: A natural-language question submitted by a user, with the question
  text as its primary attribute
- **QueryResponse**: A synthesized answer containing the answer text and a list
  of citations, each referencing an ADR number, title, and the section used
- **Citation**: A reference to a specific ADR section used to inform the
  synthesized answer, including ADR number, title, and section name
- **ADRSummary**: A lightweight representation of an indexed ADR for list views,
  containing number, title, status, and date

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A question about a specific ADR topic returns a synthesized answer
  citing the correct ADR within 10 seconds
- **SC-002**: The synthesized answer for a known topic is factually consistent
  with the source ADR content (no hallucinated decisions)
- **SC-003**: The ADR list endpoint returns all indexed ADRs with complete
  metadata
- **SC-004**: The system handles unavailable services gracefully, returning
  error messages within 5 seconds rather than hanging
- **SC-005**: A developer can query the system using a single HTTP request with
  no authentication required

## Clarifications

### Session 2026-04-06

- Q: Should the query response return inline citations or structured data? → A: Structured — JSON with answer text plus a separate citations array
- Q: How much context to send to synthesis beyond matched chunks? → A: Full ADR expansion — when a chunk matches, send the entire ADR content to synthesis, not just the matched section

## Assumptions

- The ADR index has already been built using the reindex command from Milestone 1
- The synthesis service requires an API key provided via environment variable
  — the system does not manage key provisioning
- The system is single-user and does not require authentication for the HTTP API
- The synthesis service has rate limits and cost implications — the system does
  not implement caching or rate limiting in this milestone
- The HTTP API serves JSON responses only — no HTML rendering in this milestone
  (that's Milestone 3)
- The number of retrieved chunks sent to synthesis is configurable but defaults
  to a reasonable value (e.g., 5) — tuning is deferred to Phase 2
