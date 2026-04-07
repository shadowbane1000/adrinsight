# Feature Specification: Evaluation Harness

**Feature Branch**: `004-p2-eval-harness`  
**Created**: 2026-04-07  
**Status**: Draft  
**Input**: User description: "Milestone 4a"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Measure Current Answer Quality (Priority: P1)

A developer wants to understand how well the system answers questions before
making any changes to retrieval or synthesis. They run a single command that
executes a predefined set of test questions against the live system, scores
each answer on multiple dimensions, and produces a summary report showing
where the system performs well and where it falls short.

**Why this priority**: Without a baseline measurement, there is no way to know
whether future changes to chunking, search, or synthesis improve or regress
answer quality. This is the foundation for all Phase 2 retrieval work.

**Independent Test**: Can be tested by running the evaluation command and
verifying it produces scores for all test cases, with a human-readable summary.

**Acceptance Scenarios**:

1. **Given** the system is running with indexed ADRs, **When** a developer runs
   the evaluation command, **Then** each test question is submitted, the answer
   and citations are captured, and scores are produced for every test case.

2. **Given** the evaluation has completed, **When** the developer views the
   results, **Then** they see per-question scores for retrieval precision,
   retrieval recall, answer accuracy, and answer completeness, plus aggregate
   scores across all questions.

3. **Given** this is the first evaluation run, **When** no prior baseline exists,
   **Then** the results are saved as the baseline for future comparison.

---

### User Story 2 - Detect Quality Regressions (Priority: P1)

A developer makes a change to the retrieval or synthesis pipeline and wants
to know if the change improved or degraded answer quality. They run the
evaluation command, which automatically compares results against the saved
baseline and clearly flags any regressions.

**Why this priority**: This is the primary use case for the harness — gating
changes that make answers worse. Without regression detection, developers
can't safely experiment with retrieval improvements.

**Independent Test**: Can be tested by artificially degrading the system (e.g.,
reducing top-K to 1) and verifying the evaluation detects the regression.

**Acceptance Scenarios**:

1. **Given** a baseline exists and the developer has made changes, **When** they
   run the evaluation command, **Then** the results are compared against the
   baseline and any score drops are highlighted.

2. **Given** a test case's score drops below the baseline threshold, **When**
   the evaluation completes, **Then** the command exits with a non-zero exit
   code and the failing test cases are clearly listed.

3. **Given** all test cases meet or exceed the baseline, **When** the evaluation
   completes, **Then** the command exits with a zero exit code and reports
   success.

---

### User Story 3 - Run Evaluation in CI (Priority: P2)

The evaluation runs as part of the continuous integration pipeline so that
pull requests which degrade answer quality are automatically flagged before
merge. The CI step should complete in a reasonable time and produce clear
output in the build log.

**Why this priority**: Manual evaluation is unreliable. CI integration ensures
no regression slips through regardless of developer discipline.

**Independent Test**: Can be tested by triggering a CI build and verifying the
evaluation step appears in the build output with pass/fail status.

**Acceptance Scenarios**:

1. **Given** a CI pipeline is configured to run evaluation, **When** a PR is
   opened, **Then** the evaluation runs and its pass/fail status is visible
   in the build results.

2. **Given** the evaluation requires external services (embedding model, LLM),
   **When** running in CI, **Then** the evaluation either connects to those
   services or skips gracefully with a warning if they are unavailable.

---

### User Story 4 - Manage Test Cases (Priority: P2)

A developer wants to add, update, or review test cases in the evaluation
corpus. Test cases are stored in a human-readable format that makes it easy
to see what questions are tested, what the expected citations are, and what
key facts each answer should contain.

**Why this priority**: The test corpus will grow as new ADRs are added and
new failure modes are discovered. The format must be easy to maintain.

**Independent Test**: Can be tested by adding a new test case to the corpus
and verifying it appears in the next evaluation run.

**Acceptance Scenarios**:

1. **Given** a developer opens the test case file, **When** they read it,
   **Then** each test case clearly shows the question, expected citations,
   and key facts the answer should contain.

2. **Given** a developer adds a new test case, **When** they run the evaluation,
   **Then** the new test case is included in the results and scored.

3. **Given** a new ADR is added to the project, **When** it affects existing
   test cases, **Then** the developer can update expected citations and key
   facts without changing the evaluation infrastructure.

### Edge Cases

- Evaluation run when embedding service is not available
- Evaluation run when LLM judge API key is not set
- Test case where the expected citation list is empty (no relevant ADRs)
- Test case where the system returns zero citations
- LLM judge returns malformed or unparseable scores
- Baseline file is missing or corrupted
- New test cases added since the last baseline (score and report, but skip
  regression check — no baseline to compare against)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an evaluation command that runs a predefined
  set of test questions against the live system and produces scored results
- **FR-002**: Each test case MUST include the question, a list of expected ADR
  citations (ground truth), and key facts the answer should contain
- **FR-003**: System MUST compute retrieval metrics mechanically: precision
  (fraction of returned citations in expected set), recall (fraction of
  expected citations returned), and F1 score
- **FR-004**: System MUST use an LLM-as-judge to score answer accuracy (does
  the answer correctly represent what the ADRs say) and completeness (does it
  address all aspects of the question). All scores MUST be normalized to a
  0.0-1.0 scale for uniform comparison and threshold checking
- **FR-005**: System MUST save evaluation results so they can be compared
  across runs
- **FR-006**: System MUST compare current results against a saved baseline and
  report regressions using a per-question delta — any individual question
  dropping more than a configurable threshold (default 0.2 on the 0.0-1.0
  scale) from its baseline score on any dimension is a regression
- **FR-007**: The evaluation command MUST exit with a non-zero code when any
  individual test case regresses beyond the per-question delta threshold
- **FR-008**: The initial test corpus MUST include the six sample queries from
  the web UI's About page, each with human-annotated expected citations and
  key facts
- **FR-009**: The evaluation command MUST produce a human-readable summary
  showing per-question scores and aggregate metrics
- **FR-010**: System MUST handle unavailable external services gracefully —
  if the embedding service is down, skip evaluation with a clear warning
  rather than crashing

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The evaluation command runs all test cases and produces a
  complete score report in under 5 minutes (for the initial 6-question corpus)
- **SC-002**: A developer can determine whether a change improved or degraded
  answer quality by running a single command
- **SC-003**: The evaluation detects a known regression (e.g., reducing
  retrieval results from 5 to 1) with 100% reliability
- **SC-004**: Adding a new test case requires editing only the test case file,
  not any evaluation infrastructure code
- **SC-005**: The evaluation can run in the CI pipeline and its pass/fail
  status is visible in build results

## Clarifications

### Session 2026-04-07

- Q: How should regression be defined — absolute threshold, per-question delta, aggregate, or combined? → A: Per-question delta — any question dropping more than N points from its baseline score fails.
- Q: How should new test cases with no baseline entry be handled? → A: Score and report, but skip regression check — no baseline to compare against.
- Q: What scale should scores use? → A: All scores normalized to 0.0-1.0 floats. LLM judge scores 0-5 internally, divided by 5. Mechanical scores (precision/recall/F1) are naturally 0-1. Delta threshold is a float (default 0.2).

## Assumptions

- The evaluation runs against a live system with indexed ADRs — it is not a
  unit test but an integration/system test
- The LLM-as-judge uses the same LLM provider API key as the synthesis
  endpoint, consuming additional API credits per evaluation run
- The initial test corpus is small (6 questions from the About page) and will
  grow over time as new failure modes are discovered
- Retrieval metrics (precision/recall) can be computed mechanically without
  an LLM by comparing returned citations against the expected list
- The LLM judge scoring rubric is deterministic enough to produce consistent
  results across runs (minor variation is acceptable; large swings indicate
  a rubric problem)
- The evaluation command requires the same dependencies as the serve command
  (embedding service for queries, LLM API for judge scoring)
- Baseline thresholds will need tuning after the first run — the initial
  baseline captures current performance as-is, warts and all
