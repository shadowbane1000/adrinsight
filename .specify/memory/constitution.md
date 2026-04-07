<!--
  Sync Impact Report
  ===================
  Version change: 1.0.0 → 1.1.0
  Bump rationale: Added ADR Immutability principle (II-A)

  Principles established:
    I.    Interfaces Everywhere — Go interfaces for all external dependencies
    II.   ADRs Are the Product — project-wide ADRs in docs/adr/ serve as docs, demo data, and talking points
    II-A. ADR Immutability — ADR bodies are immutable once accepted; changes require a new superseding ADR
    III.  Skateboard First — simplest complete version before expanding
    IV.   Idiomatic Go — Go conventions, standard library preference, cmd/ + internal/ layout
    V.    Developer Experience — docker-compose up, one API key, disposable database

  Added sections:
    - Core Principles (5 principles + 1 sub-principle)
    - Quality Gates (CI enforcement, ADR coverage checks, ADR body edit detection)
    - Governance (amendment process, compliance checking)

  Removed sections: (none)

  Templates updated:
    ✅ plan-template.md — Added ADR Impact section for tracking superseding ADRs
    ✅ tasks-template.md — Added ADR Handling During Implementation section
    ✅ adr-template.md — Added Supersedes field to header
    ✅ spec-template.md — Technology-agnostic, no conflicts
    ✅ checklist-template.md — Generic, no conflicts
    ✅ agent-file-template.md — Placeholder, filled at plan time

  Follow-up TODOs: None
-->

# ADR Insight Constitution

## Core Principles

### I. Interfaces Everywhere
Every external dependency (LLM, embedder, store) MUST sit behind a Go
interface. Components MUST be swappable without changing callers. This
enables testing, demonstrates "what would you change at scale" thinking,
and keeps the architecture clean.

### II. ADRs Are the Product
Every significant architectural decision MUST get an ADR in `docs/adr/`.
ADRs serve triple duty: project documentation, demo dataset for the RAG
pipeline, and interview talking points. A decision is ADR-worthy if it
involves choosing between viable alternatives with meaningful consequences.

### II-A. ADR Immutability
Once an ADR is accepted, its body (Context, Decision, Rationale,
Consequences, Alternatives Considered sections) MUST NOT be edited during
task implementation. If a decision needs to change, create a new ADR that
supersedes the original. The superseding ADR MUST reference the original
and explain why the decision changed. When creating a superseding ADR,
the overridden ADR's **Status** and **Related ADRs** fields MAY be updated
to reflect the supersession (e.g., `Status: Superseded by ADR-NNN`), but
its body MUST remain untouched. This preserves the decision history that
is core to the project's value as a portfolio piece and RAG dataset.

### III. Skateboard First
Build the simplest complete version before expanding. The walking skeleton
MUST be rough but functional end-to-end. No speculative abstractions, no
premature optimization, no features beyond what the current phase requires.

### IV. Idiomatic Go
Write Go the way Go is meant to be written: explicit error handling, short
variable names in small scopes, table-driven tests, `cmd/` + `internal/`
layout. Non-idiomatic patterns MUST be flagged during review. Lean on the
standard library before reaching for third-party packages.

### V. Developer Experience
`docker-compose up` and one API key MUST be all it takes. The database is
disposable (regenerable via `reindex`). Configuration MUST have sensible
defaults. README MUST get someone running in under 5 minutes.

## Quality Gates

- Lint with golangci-lint MUST pass (CI enforced)
- Tests MUST pass (CI enforced)
- Architectural decisions without ADRs are flagged during planning and
  task generation
- Constitution compliance is checked during `/speckit-plan`

## Governance

- This constitution governs all development on ADR Insight
- Amendments require an ADR documenting the change and its rationale
- Constitution principles are checked during the planning phase —
  violations MUST be justified in the Complexity Tracking section of
  plan.md
- ADR body edits during task implementation are flagged as violations
  of Principle II-A

**Version**: 1.1.0 | **Ratified**: 2026-04-06 | **Last Amended**: 2026-04-07
