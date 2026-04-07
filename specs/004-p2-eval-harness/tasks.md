# Tasks: Evaluation Harness

**Input**: Design documents from `/specs/004-p2-eval-harness/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/eval-cli.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the eval package structure, data types, and test case corpus.

- [x] T001 Create directory structure: `internal/eval/` for evaluation logic and `testdata/eval/` for test data
- [x] T002 Define Go types for TestCase, EvalResult, Baseline, RunReport, AggregateScores, and Regression in `internal/eval/types.go` per data-model.md
- [x] T003 Create the initial test case corpus `testdata/eval/cases.json` with the 6 sample queries from the About page, each annotated with expected ADR numbers and key facts

**Checkpoint**: Types compile, test case JSON is valid and loadable.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Mechanical scoring and JSON I/O — needed by all user stories.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T004 Implement test case loading in `internal/eval/eval.go` — read `cases.json`, unmarshal into []TestCase, validate required fields (id, question, expected_adrs)
- [x] T005 Implement mechanical scoring in `internal/eval/scoring.go` — given returned ADR numbers and expected ADR numbers, compute precision, recall, and F1. Handle edge cases: empty returned set (precision=0), empty expected set (recall=1.0), both empty (all 1.0)
- [x] T006 Implement baseline I/O in `internal/eval/eval.go` — load baseline from JSON file, save baseline to JSON file. Handle missing file gracefully (return nil baseline, not error)

**Checkpoint**: Can load test cases, compute retrieval scores for mock data, read/write baseline files.

---

## Phase 3: User Story 1 — Measure Current Answer Quality (Priority: P1) 🎯 MVP

**Goal**: Run all test cases against the live system, score each one, produce a summary report.

**Independent Test**: `./adr-insight eval --save-baseline` runs all 6 questions, prints per-question scores, saves baseline JSON.

### Implementation for User Story 1

- [x] T007 [US1] Implement LLM judge in `internal/eval/judge.go` — define a Judge interface with `Score(ctx, question, expectedADRContent, answer string) (accuracy float64, completeness float64, accuracyReason string, completenessReason string, err error)`. Implement AnthropicJudge that sends the rubric prompt to Claude with OutputConfig JSON schema (same pattern as ADR-009). Rubric: accuracy 0-5 (factual correctness), completeness 0-5 (addresses all aspects). Judge returns integer scores 0-5; normalize to 0.0-1.0 by dividing by 5 before returning. All scores stored and compared on the 0.0-1.0 scale.
- [x] T008 [US1] Implement the eval orchestrator in `internal/eval/eval.go` — for each test case: call Pipeline.Query, extract returned ADR numbers from citations, compute mechanical scores (T005), read expected ADR full text from disk, call LLM judge (T007), assemble EvalResult. Collect all results into RunReport with aggregate scores.
- [x] T009 [US1] Implement report formatting in `internal/eval/eval.go` — produce human-readable stdout output per contracts/eval-cli.md: per-question block (citations, precision/recall/F1, accuracy/completeness with reasons, status), summary line (questions/passed/regressed/new), aggregate scores, final RESULT line.
- [x] T010 [US1] Add `eval` subcommand to `cmd/adr-insight/main.go` — parse flags per contracts/eval-cli.md (--cases, --baseline, --save-baseline, --delta, --db, --ollama-url, --adr-dir, --model). Wire up: create embedder, store, pipeline, judge. Load test cases, run orchestrator, print report. When --save-baseline is set, save results as baseline. Handle missing ANTHROPIC_API_KEY (required for both pipeline and judge).
- [x] T011 [US1] Handle unavailable services in `cmd/adr-insight/main.go` eval command — if Ollama is unreachable (embedding fails on first test case), print warning "Evaluation skipped: embedding service unavailable" and exit 0. If Anthropic API fails, print warning and exit 0.

**Checkpoint**: `./adr-insight eval --save-baseline` runs all 6 questions, prints scores, saves `testdata/eval/baseline.json`.

---

## Phase 4: User Story 2 — Detect Quality Regressions (Priority: P1)

**Goal**: Compare current results against baseline, flag regressions with per-question delta.

**Independent Test**: Modify top-K to 1, run `./adr-insight eval`, see regressions flagged and exit code 1.

### Implementation for User Story 2

- [x] T012 [US2] Implement regression detection in `internal/eval/eval.go` — given current RunReport and loaded Baseline, compare each test case's scores (precision, recall, f1, accuracy, completeness — all 0.0-1.0 floats) against baseline. A regression is any dimension dropping by more than delta_threshold (default 0.2). Populate RunReport.regressions. Test cases in corpus but not in baseline are added to RunReport.new_cases (no regression check).
- [x] T013 [US2] Wire regression detection into the eval command in `cmd/adr-insight/main.go` — after running orchestrator, if baseline exists, run regression detection. Print regressions in report output. Exit code 1 if any regressions found, exit code 0 otherwise. If no baseline exists, print suggestion to run with --save-baseline.

**Checkpoint**: Artificially degrading the system (e.g., top-K=1) triggers regression detection with exit code 1.

---

## Phase 5: User Story 3 — Run Evaluation in CI (Priority: P2)

**Goal**: CI pipeline runs eval and surfaces pass/fail.

**Independent Test**: Push a change, CI runs eval step, result visible in build output.

### Implementation for User Story 3

- [x] T014 [US3] Add `eval` target to `Makefile` — runs `./adr-insight eval` with default paths. Depends on build target.
- [x] T015 [US3] Add eval step to `.gitea/workflows/ci.yaml` and `.github/workflows/ci.yaml` — run after build step. Allow failure gracefully if services unavailable (the eval command itself handles this with exit 0 + warning).

**Checkpoint**: CI build includes eval step; passes when services available, skips gracefully when not.

---

## Phase 6: User Story 4 — Manage Test Cases (Priority: P2)

**Goal**: Test cases are easy to read, add, and update.

**Independent Test**: Add a 7th test case to cases.json, run eval, see it scored as "NEW".

### Implementation for User Story 4

- [x] T016 [US4] Validate test case format in `internal/eval/eval.go` — on load, check each test case has non-empty id (unique), non-empty question, and at least one expected_adr. Report validation errors with line context. Ensure new test cases added to corpus but absent from baseline are reported as "NEW (no baseline)" in output.

**Checkpoint**: Adding a new entry to cases.json shows it scored and labeled "NEW" in the next eval run.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Error handling, edge cases, lint, and documentation.

- [x] T017 Handle edge case in `internal/eval/judge.go`: LLM judge returns malformed JSON — retry once, then fall back to score 0 with reason "Judge scoring failed"
- [x] T018 Handle edge case in `internal/eval/scoring.go`: test case with empty expected_adrs list — treat as "no relevant ADRs expected", precision based on returned set being empty
- [x] T019 Handle edge case in `internal/eval/eval.go`: baseline file is corrupted JSON — log warning, treat as no baseline
- [x] T020 Run `make lint` and fix any lint errors across all modified files
- [ ] T021 Run quickstart.md scenarios to validate end-to-end functionality
- [x] T022 Update `docs/architecture.md` with eval package description and data flow
- [x] T023 Update `README.md` with eval command documentation (usage, flags, example output)

---

## ADR Handling During Implementation

- **NEVER edit the body** of an existing ADR (Context, Decision, Rationale,
  Consequences, Alternatives Considered sections) during task implementation
- If a task requires changing a previously recorded decision, **create a new
  superseding ADR** instead (use the ADR template with the `Supersedes` field)
- When a new ADR supersedes an old one, you MAY update the old ADR's
  **Status** (e.g., `Superseded by ADR-NNN`) and **Related ADRs** fields only
- This preserves decision history — see Constitution Principle II-A

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational — core eval functionality
- **US2 (Phase 4)**: Depends on US1 — needs working eval to compare against baseline
- **US3 (Phase 5)**: Depends on US1 — CI runs the eval command
- **US4 (Phase 6)**: Depends on Foundational — test case validation
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Phase 2 (Foundational) — no other story dependencies
- **User Story 2 (P1)**: Depends on US1 — needs eval results to compare against baseline
- **User Story 3 (P2)**: Depends on US1 — CI runs the eval command
- **User Story 4 (P2)**: Depends on Phase 2 (Foundational) — test case loading and validation

### Parallel Opportunities

- T002 and T003 can run in parallel (types and test data are independent files)
- T004, T005, and T006 can run in parallel (loading, scoring, and I/O are independent)
- T014 and T015 can run in parallel (Makefile and CI configs are independent)
- US3 and US4 can proceed in parallel after US1 is complete

---

## Parallel Example: Foundational Phase

```bash
# Launch all three in parallel:
Task: "Implement test case loading" (T004)
Task: "Implement mechanical scoring" (T005)
Task: "Implement baseline I/O" (T006)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T006)
3. Complete Phase 3: User Story 1 (T007-T011)
4. **STOP and VALIDATE**: `./adr-insight eval --save-baseline` produces scores for all 6 questions
5. This alone captures the baseline needed for M4b

### Incremental Delivery

1. Setup + Foundational → Types, test cases, scoring functions ready
2. Add US1 (eval orchestrator) → Baseline captured
3. Add US2 (regression detection) → Changes are gated
4. Add US3 (CI) + US4 (validation) → Fully automated
5. Polish → Production-ready

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- The eval command reuses existing Pipeline, Store, and Embedder — no new external dependencies
- The LLM judge reuses the Anthropic SDK patterns from internal/llm/
- Test case corpus starts small (6 questions) and grows organically
