---
name: "speckit-adr"
description: "Generate or update Architecture Decision Records (ADRs) from planning artifacts or standalone decisions."
argument-hint: "Decision topic or 'from-plan' to extract decisions from current feature's planning artifacts"
compatibility: "Requires spec-kit project structure with .specify/ directory"
metadata:
  author: "adrinsight-custom"
  source: "custom skill for ADR governance"
user-invocable: true
disable-model-invocation: true
---


## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## ADR Directory

All ADRs live in `docs/adr/` at the project root (NOT per-feature). This is a project-wide governance artifact.

## Modes of Operation

This skill operates in two modes based on user input:

### Mode 1: Extract from Planning Artifacts (`from-plan` or invoked as post-plan hook)

1. **Setup**: Run `.specify/scripts/bash/check-prerequisites.sh --json` from repo root to get FEATURE_DIR and AVAILABLE_DOCS.

2. **Load planning artifacts**: Read from FEATURE_DIR:
   - **Required**: `plan.md` (technical context, constitution check, structure decisions)
   - **If exists**: `research.md` (decisions with rationale and alternatives)
   - **If exists**: `data-model.md` (entity and storage decisions)
   - **If exists**: `contracts/` (interface design decisions)

3. **Scan for existing ADRs**: Read all files in `docs/adr/` to determine:
   - Next available ADR number
   - Existing decisions (to avoid duplicates)
   - Potential supersession or amendment targets

4. **Extract architectural decisions** from the loaded artifacts. A decision is ADR-worthy if it:
   - Chooses between multiple viable alternatives
   - Has significant consequences (positive or negative) for the project
   - Would be non-obvious to a new team member
   - Affects the system's structure, dependencies, or key quality attributes
   - Is NOT a routine implementation detail (e.g., variable naming, file organization within an established pattern)

5. **For each extracted decision**, generate an ADR using the template at `.specify/templates/adr-template.md`:
   - Pull context from `research.md` (why the decision was needed)
   - Pull rationale and alternatives from `research.md` or `plan.md`
   - Pull consequences from the planning discussion
   - Cross-reference related existing ADRs
   - Set status to **Accepted** (it was decided during planning)

6. **Write ADRs** to `docs/adr/ADR-[NNN]-[kebab-case-title].md`

7. **Report**: List all generated ADRs with titles and a one-line summary of each decision.

### Mode 2: Standalone ADR (specific topic provided)

1. **Scan for existing ADRs**: Read all files in `docs/adr/` to determine next number and existing decisions.

2. **Interactive decision capture**: Based on the user's topic, ask structured questions:
   - What is the context/problem?
   - What alternatives were considered?
   - What was decided and why?
   - What are the expected consequences?

   If the user provides enough context in their input, skip redundant questions and draft directly.

3. **Generate ADR** using `.specify/templates/adr-template.md`

4. **Write** to `docs/adr/ADR-[NNN]-[kebab-case-title].md`

5. **Check for related ADRs**: If any existing ADRs are related (same domain, superseded, or depended upon), note the relationship in the Related ADRs section and suggest updating the related ADR's status if appropriate.

## ADR Numbering

- ADRs use a three-digit sequential number: `ADR-001`, `ADR-002`, etc.
- Scan `docs/adr/` for the highest existing number and increment by 1
- Never reuse a number, even if an ADR is deprecated

## ADR File Naming

Format: `ADR-[NNN]-[kebab-case-title].md`

Examples:
- `ADR-001-why-go.md`
- `ADR-005-chunking-strategy.md`

## Quality Checks

Before writing each ADR, verify:
- [ ] Context explains WHY the decision was needed (not just what)
- [ ] At least one alternative was considered (even if it was "do nothing")
- [ ] Consequences include both positive AND negative impacts
- [ ] Rationale connects the decision back to project goals or constraints
- [ ] Status is appropriate (Proposed if uncertain, Accepted if decided)

## Key Rules

- ADRs are project-wide, stored in `docs/adr/` — never in feature-specific directories
- Use absolute paths when reading/writing files
- Match the existing ADR style in the project (check existing ADRs for tone and depth)
- Don't generate ADRs for trivial decisions — focus on decisions that matter
- When extracting from planning artifacts, err on the side of fewer, higher-quality ADRs
