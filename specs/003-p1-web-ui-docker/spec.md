# Feature Specification: Web UI and Docker Deployment

**Feature Branch**: `003-p1-web-ui-docker`  
**Created**: 2026-04-06  
**Status**: Draft  
**Input**: User description: "Milestone 3 Phase 1"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ask a Question in the Browser (Priority: P1)

A developer opens the web UI in their browser, types a natural-language
question about their project's architectural decisions into a search bar, and
receives a synthesized answer with clickable citation links to the source ADRs.
The experience is simple: one text field, one submit button, one answer area.

**Why this priority**: This is the capstone of Phase 1. Everything built so far
(parser, embedder, store, RAG pipeline, HTTP API) culminates in this moment.
A hiring manager should be able to open a browser and immediately interact
with the system.

**Independent Test**: Can be tested by opening the web UI, typing a question,
and verifying the answer appears with clickable citations.

**Acceptance Scenarios**:

1. **Given** the server is running with ADR files in the configured directory,
   **When** a user opens the web UI and submits a question, **Then** the system
   automatically indexes any un-indexed ADRs and returns a synthesized answer
   with citations that link to the full ADR content.

2. **Given** a user receives an answer with citations, **When** they click a
   citation link, **Then** the full ADR content is displayed.

3. **Given** the system has no relevant ADRs for a question, **When** a user
   submits that question, **Then** the UI displays a message indicating no
   relevant information was found, without showing errors or blank content.

4. **Given** a query is in progress, **When** the user is waiting for a
   response, **Then** a loading indicator is visible so they know the system
   is working.

---

### User Story 2 - Browse All ADRs (Priority: P2)

A developer wants to see all the ADRs that have been indexed, browse their
titles and statuses, and read any ADR in full. The web UI provides a list
view of all indexed ADRs with links to view each one.

**Why this priority**: This complements the search experience — users should
be able to discover ADRs by browsing, not just by asking questions. It also
validates the GET /adrs and GET /adrs/{number} endpoints visually.

**Independent Test**: Can be tested by opening the ADR list view and verifying
all indexed ADRs appear with correct metadata and are viewable.

**Acceptance Scenarios**:

1. **Given** ADR files exist in the configured directory, **When** a user
   navigates to the ADR list view, **Then** all ADRs are displayed with their
   number, title, and status (indexing happens automatically on server start).

2. **Given** a user sees the ADR list, **When** they click on an ADR, **Then**
   the full content of that ADR is displayed.

3. **Given** no ADRs have been indexed, **When** a user navigates to the ADR
   list, **Then** a message indicates no ADRs are available.

---

### User Story 3 - Run Everything with Docker Compose (Priority: P3)

A developer clones the repository and wants to try it out. They run a single
command to bring up the entire system — the Go application, the embedding
service with its model pre-loaded, and any required infrastructure. The system
automatically indexes the ADR files on startup. After the system starts, they
open a browser and begin asking questions.

**Why this priority**: This is the developer experience payoff. Constitution
Principle V says "docker-compose up and one API key should be all it takes."
Without containerization, the setup requires installing Go, Ollama, pulling
models, building the binary, and running multiple commands.

**Independent Test**: Can be tested by running docker-compose up on a clean
machine (with Docker installed) and verifying the web UI is accessible and
functional.

**Acceptance Scenarios**:

1. **Given** a developer has Docker and an Anthropic API key, **When** they
   run docker-compose up, **Then** the application starts with all services,
   automatically indexes the ADR files, and the web UI is accessible and
   queryable in a browser without any manual reindex step.

2. **Given** the system is running via docker-compose, **When** a user asks a
   question in the web UI, **Then** they receive a synthesized answer (the
   full pipeline works end-to-end in containers).

3. **Given** a developer runs docker-compose up for the first time, **When**
   the embedding model needs to be downloaded, **Then** the system handles
   the download automatically and becomes ready without manual intervention.

4. **Given** the system has already indexed ADRs from a previous run, **When**
   docker-compose is restarted, **Then** the system detects the existing index
   and skips re-indexing (startup is fast on subsequent runs).

---

### User Story 4 - Discover the Project from the Web UI (Priority: P2)

A visitor arrives at a publicly hosted instance of ADR Insight — perhaps linked
from a resume, portfolio, or shared by a colleague. They have no prior context
about the project. The web UI provides a visible "About" section or link that
explains what ADR Insight is, how it was built, and where to find the source
code. Specifically, the visitor learns that:

- ADR Insight is an AI-powered tool for querying Architecture Decision Records
- The project was built using spec-driven development with a modified version
  of GitHub's spec-kit that includes ADR generation as part of the workflow
- Any project that maintains ADRs in the standard markdown format can use
  ADR Insight to provide natural-language Q&A over their architecture decisions
- A link to the GitHub repository is provided for those who want to explore
  the code, ADRs, or run their own instance

**Why this priority**: This is the portfolio play. When a hiring manager visits
the live instance, the "about" content is the first thing that frames the
project's purpose and Tyler's approach. Without it, the tool is impressive but
the story behind it is invisible.

**Independent Test**: Can be tested by opening the web UI and verifying the
project description and GitHub link are visible without needing to search for
them.

**Acceptance Scenarios**:

1. **Given** a visitor opens the web UI for the first time, **When** they look
   at the page, **Then** a project description and GitHub link are visible
   without additional navigation.

2. **Given** a visitor reads the project description, **When** they want to
   learn more, **Then** they can click a link to the GitHub repository.

---

### User Story 5 - Setup from README (Priority: P3)

A developer discovers the project on GitHub and wants to understand what it
does and how to run it. The README provides clear instructions that get them
from clone to running in under 5 minutes (excluding Docker image download
time).

**Why this priority**: The README is the first thing a hiring manager sees.
Constitution Principle V requires that the README gets someone running in under
5 minutes.

**Independent Test**: Can be tested by following the README instructions from
a fresh clone and verifying the system is running.

**Acceptance Scenarios**:

1. **Given** a developer reads the README, **When** they follow the setup
   instructions, **Then** they can have the system running and queryable
   within 5 minutes (excluding download time).

### Edge Cases

- Browser with JavaScript disabled (web UI should degrade gracefully or
  indicate JavaScript is required)
- Very long synthesized answers that exceed the viewport
- Multiple rapid queries submitted before previous results return
- Docker compose started without the API key environment variable set (server
  should start, serve ADR browsing, and return a clear error on query attempts)
- Ollama model download interrupted mid-way
- Server started without a populated index (auto-reindex should handle this)
- Ollama not yet ready when auto-reindex runs (server should retry or wait)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST serve a web UI that allows users to type questions
  and view synthesized answers with citations
- **FR-002**: Citations in the web UI MUST be clickable, navigating to the
  full content of the referenced ADR
- **FR-012**: The answer area and the ADR detail panel MUST render markdown
  content as formatted HTML (headings, bold, lists, code blocks, etc.)
- **FR-003**: Web UI MUST display a loading indicator while waiting for a
  query response
- **FR-011**: The answer area MUST be hidden when no question has been asked,
  and MUST size to fit the answer content when displayed
- **FR-004**: Web UI MUST provide a list view of all indexed ADRs with their
  number, title, and status
- **FR-005**: Web UI MUST allow users to view the full content of any indexed
  ADR in the right-side detail panel, replacing the about content
- **FR-006**: System MUST be deployable via a single docker-compose command
  with only an API key as external configuration
- **FR-007**: Docker setup MUST include the embedding service with its model
  pre-loaded or auto-downloaded on first start
- **FR-013**: The serve command MUST automatically run reindex on startup if
  the database is empty or missing, so that no manual reindex step is needed
- **FR-008**: README MUST contain complete setup instructions for both local
  development and Docker deployment
- **FR-009**: Web UI MUST handle error states gracefully, showing user-friendly
  messages when services are unavailable or no results are found
- **FR-010**: Web UI MUST display a project description explaining what ADR
  Insight is, that it was built using spec-driven development with a modified
  spec-kit, that any project with standard-format ADRs can use it, and a link
  to the GitHub repository

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can ask a question and see a synthesized answer with
  citations within 15 seconds of submitting (including network + LLM latency)
- **SC-002**: A user can navigate from a citation to the full ADR content
  in one click
- **SC-003**: docker-compose up brings the system to a queryable state with
  no manual steps beyond providing the API key
- **SC-004**: A developer following the README can go from clone to running
  system in under 5 minutes (excluding download time)
- **SC-005**: The web UI is functional in current versions of Chrome, Firefox,
  and Safari without framework dependencies

## Clarifications

### Session 2026-04-06

- Q: Should the web UI be single page or multi-page? → A: Single page — search bar at top, answer below (hidden until asked, sized to content), lower section split with ADR list on left and about/ADR detail panel on right. About is default right-side content; clicking an ADR or citation replaces it with that ADR's full content.
- Q: How is the database populated in Docker? → A: The serve command auto-reindexes on startup if the database is empty or missing. No manual reindex step required. Skips reindex if an existing index is detected.

## Assumptions

- The web UI is simple HTML/CSS/JS served by the Go HTTP server — no frontend
  framework, no build step, no node_modules (Constitution Principle III)
- The web UI calls the existing HTTP API endpoints (POST /query, GET /adrs,
  GET /adrs/{number}) — no new backend endpoints are needed
- The web UI is a single page with this layout: search bar at top, answer
  area directly below (hidden until a question is asked, sized to fit
  content), lower section split into ADR list on the left and about/ADR
  detail panel on the right. The about panel is the default right-side
  content; clicking an ADR or citation replaces it with that ADR's full
  content
- Docker Compose v2 is the target (the `docker compose` command, not the
  legacy `docker-compose` binary)
- The Anthropic API key is provided via environment variable, not baked into
  any container image
- The embedding model download on first start may take several minutes
  depending on network speed — this is acceptable and should be documented
- The web UI does not need to be responsive/mobile-optimized for this milestone
  — desktop browser is the target
