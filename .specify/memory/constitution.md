# ADR Insight Constitution

## Core Principles

### I. Interfaces Everywhere
Every external dependency (LLM, embedder, store) sits behind a Go interface. Components must be swappable without changing callers. This enables testing, demonstrates "what would you change at scale" thinking, and keeps the architecture clean.

### II. ADRs Are the Product
Every significant architectural decision gets an ADR in `docs/adr/`. ADRs serve triple duty: project documentation, demo dataset for the RAG pipeline, and interview talking points. If a decision involves choosing between viable alternatives with meaningful consequences, it gets an ADR.

### III. Skateboard First
Build the simplest complete version before expanding. The walking skeleton should be rough but functional end-to-end. No speculative abstractions, no premature optimization, no features beyond what the current phase requires.

### IV. Idiomatic Go
Write Go the way Go is meant to be written: explicit error handling, short variable names in small scopes, table-driven tests, `cmd/` + `internal/` layout. Flag non-idiomatic patterns during review. Lean on the standard library before reaching for third-party packages.

### V. Developer Experience
`docker-compose up` and one API key should be all it takes. The database is disposable (regenerable via `reindex`). Configuration should have sensible defaults. README should get someone running in under 5 minutes.

## Quality Gates

- Lint with golangci-lint must pass (CI enforced)
- Tests must pass (CI enforced)
- Architectural decisions without ADRs are flagged during planning and task generation
- Constitution compliance is checked during `/speckit-plan`

## Governance

- This constitution governs all development on ADR Insight
- Amendments require an ADR documenting the change and its rationale
- Constitution principles are checked during the planning phase — violations must be justified in the Complexity Tracking section of plan.md

**Version**: 1.0.0 | **Ratified**: 2026-04-06 | **Last Amended**: 2026-04-06
