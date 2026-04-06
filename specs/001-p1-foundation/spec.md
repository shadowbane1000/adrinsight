# Feature Specification: Foundation — ADR Indexing Pipeline

**Feature Branch**: `001-p1-foundation`  
**Created**: 2026-04-06  
**Status**: Draft  
**Input**: User description: "phase1 milestone 1."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reindex ADRs from a Directory (Priority: P1)

A developer working on a project with Architecture Decision Records wants to
build a searchable index of those ADRs. They point the tool at a directory
containing markdown ADR files and run a single command. The tool parses each
file, extracts structured metadata and body content, generates vector
embeddings, and stores everything in a local database. The process completes
without requiring any external services beyond a local embedding model.

**Why this priority**: This is the core data pipeline. Nothing else in the
project works without parsed, embedded, and stored ADRs. Every downstream
feature (search, synthesis, web UI) depends on this.

**Independent Test**: Can be fully tested by running the reindex command against
a directory of sample ADR files and verifying that the database contains the
expected records with valid embeddings.

**Acceptance Scenarios**:

1. **Given** a directory containing valid markdown ADR files with standard
   frontmatter (title, date, status), **When** the user runs the reindex
   command, **Then** all ADRs are parsed, embedded, and stored in the local
   database with their metadata intact.

2. **Given** a directory containing a mix of valid ADR files and non-ADR
   markdown files, **When** the user runs the reindex command, **Then** only
   files matching the ADR format are processed, and non-matching files are
   skipped with a warning.

3. **Given** ADRs have already been indexed, **When** the user runs the reindex
   command again (e.g., after adding or editing ADRs), **Then** the index is
   rebuilt from scratch, reflecting the current state of the ADR directory.

---

### User Story 2 - Verify Indexed ADRs via Similarity Search (Priority: P2)

After indexing, a developer wants to confirm the pipeline works by searching
for ADRs related to a topic. They provide a natural-language query and the
system returns the most semantically relevant ADR chunks, ranked by similarity.

**Why this priority**: This validates that the full pipeline (parse → embed →
store → search) works end-to-end. Without verification, there's no confidence
that the index is usable for downstream synthesis.

**Independent Test**: Can be tested by indexing sample ADRs, running a known
query (e.g., "why did we choose Go?"), and verifying the top result is the
expected ADR.

**Acceptance Scenarios**:

1. **Given** ADRs have been indexed, **When** the user searches for a topic
   covered by one of the ADRs, **Then** the system returns the relevant ADR
   chunks ranked by semantic similarity.

2. **Given** ADRs have been indexed, **When** the user searches for a topic not
   covered by any ADR, **Then** the system returns an empty result set or
   low-relevance results without errors.

---

### User Story 3 - CI Validates Code Quality (Priority: P3)

A developer pushes code changes to the repository. The CI pipeline
automatically runs linting and tests to catch issues before they reach the main
branch. The pipeline also builds the binary to confirm it compiles.

**Why this priority**: CI is infrastructure that supports all future
development. It prevents regressions and enforces code quality standards from
day one, but it doesn't deliver direct user-facing functionality.

**Independent Test**: Can be tested by pushing a commit and verifying the CI
pipeline runs lint, test, and build steps successfully.

**Acceptance Scenarios**:

1. **Given** a developer pushes a commit, **When** CI runs, **Then** linting,
   tests, and a build are executed, and the result is reported as pass or fail.

2. **Given** a developer introduces a linting violation, **When** CI runs,
   **Then** the pipeline fails and reports the specific violation.

### User Story 4 - Run Quality Checks Locally (Priority: P3)

A developer wants to run linting, tests, and builds on their local machine
before pushing to the repository. They use a single standardized command to
execute the same quality checks that CI runs, getting fast feedback without
waiting for a remote pipeline.

**Why this priority**: Same priority tier as CI — it's developer infrastructure
that supports all future work. Developers should never have to push just to
find out if their code passes checks.

**Independent Test**: Can be tested by running the local quality command and
verifying it executes lint, test, and build steps, reporting pass or fail.

**Acceptance Scenarios**:

1. **Given** a developer has the project checked out locally, **When** they run
   the local quality check command, **Then** linting, tests, and a build are
   executed in sequence, and the result is reported as pass or fail.

2. **Given** a developer introduces a linting violation, **When** they run the
   local quality check command, **Then** the command fails and reports the
   specific violation before they need to push.

---

### Edge Cases

- ADR file with missing or malformed frontmatter (e.g., no title, no date)
- ADR file that is empty or contains only frontmatter with no body
- ADR directory that is empty (no files to process)
- ADR directory that does not exist (invalid path)
- Embedding service is unavailable when reindex is run
- Very large ADR body text that may need chunking before embedding
- ADR files with non-standard naming conventions

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST parse markdown ADR files from a specified directory,
  identifying ADRs by filename pattern (`ADR-*.md`, case-insensitive) and
  extracting title, date, status, and body text from each matching file
- **FR-002**: System MUST generate vector embeddings for ADR content using a
  local embedding service
- **FR-003**: System MUST store ADR metadata and vector embeddings in a local
  embedded database
- **FR-004**: System MUST support similarity search against stored embeddings,
  returning results ranked by relevance
- **FR-005**: System MUST provide a CLI command that orchestrates the full
  reindex pipeline (parse → embed → store)
- **FR-006**: System MUST rebuild the index from scratch on each reindex run
  (the database is a disposable, derived artifact)
- **FR-007**: System MUST skip files not matching the `ADR-*.md` filename
  pattern, logging skipped files at debug level
- **FR-008**: System MUST report meaningful errors when the ADR directory is
  missing or the embedding service is unavailable
- **FR-009**: CI MUST run lint, test, and build on every push

### Key Entities

- **ADR Document**: A single Architecture Decision Record with metadata (title,
  date, status, deciders) and body content (context, decision, rationale,
  consequences)
- **ADR Chunk**: A segment of an ADR document suitable for embedding (may be
  the full body or a subdivided section, depending on length)
- **Embedding**: A vector representation of an ADR chunk, used for similarity
  search
- **Index**: The complete collection of stored ADR metadata, chunks, and their
  embeddings

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The reindex command processes all valid ADR files in a directory
  and stores them with correct metadata
- **SC-002**: Similarity search for a known topic returns the relevant ADR as
  the top result
- **SC-003**: The reindex command completes in under 30 seconds for a corpus of
  up to 50 ADR files
- **SC-004**: CI pipeline passes on clean code and fails on linting violations
  or test failures
- **SC-005**: A developer can clone the repo, install dependencies, and run
  the reindex command with no manual configuration beyond having the embedding
  service available

## Clarifications

### Session 2026-04-06

- Q: How should the parser identify which files are ADRs? → A: Filename pattern `ADR-*.md` (case-insensitive)

## Assumptions

- A local embedding service (e.g., Ollama) is running and accessible when
  reindex is invoked — the tool does not manage the embedding service lifecycle
- ADR files follow a markdown format with a recognizable structure (title as
  H1, frontmatter-style metadata lines, standard sections like Context,
  Decision, Consequences)
- The ADR corpus is small (dozens to low hundreds of files) — performance
  optimization for large corpora is out of scope
- The database file is local and single-user — no concurrent write access is
  required
- Chunking strategy for long ADRs will use a simple approach (e.g., by section)
  for this milestone — refinement is deferred to Phase 2
- Embeddings are deterministic and independent per chunk — the same input text
  always produces the same vector, and changing one ADR does not affect the
  embeddings of other ADRs (incremental indexing is a possible future
  optimization but out of scope for this milestone)
