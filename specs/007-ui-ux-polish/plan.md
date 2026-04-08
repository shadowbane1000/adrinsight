# Implementation Plan: UI & UX Polish

**Branch**: `007-ui-ux-polish` | **Date**: 2026-04-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-ui-ux-polish/spec.md`

## Summary

Transform the ADR Insight web UI from a functional prototype into a polished, production-quality interface. Replace the split answer/detail layout with a unified content area and navigation stack. Add status badges, filtering, sorting, inline citations, query history, responsive layout, and visual polish. Introduce Alpine.js as a lightweight reactive framework to manage the increased UI complexity without a build step.

## Technical Context

**Language/Version**: Go 1.24+ (backend, unchanged) / JavaScript ES6+ (frontend)
**Primary Dependencies**: Alpine.js 3.x (CDN, ~17kb gzipped), marked.js (existing)
**Storage**: localStorage (query history, filter/sort preferences)
**Testing**: Manual visual testing, browser dev tools at 768px breakpoint
**Target Platform**: Modern browsers (Chrome, Firefox, Safari, Edge)
**Project Type**: Web service with embedded static frontend
**Performance Goals**: All interactions respond within 100ms; cached view navigation is instant
**Constraints**: No build step — all static files must be CDN-loadable or inline for go:embed
**Scale/Scope**: 22 ADRs, single-user, portfolio demo context

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Interfaces Everywhere | PASS | No new Go interfaces needed — this is a frontend milestone. Existing Store/Pipeline interfaces unchanged. |
| II. ADRs Are the Product | PASS | No ADR body edits. New ADR for Alpine.js framework decision. |
| II-A. ADR Immutability | PASS | No existing ADRs modified. |
| III. Skateboard First | PASS | Alpine.js is the simplest framework that handles the required complexity (nav stack, reactive state, transitions). Vanilla JS would require reinventing these patterns. |
| IV. Idiomatic Go | PASS | Minimal Go changes (possibly adding a date field to API responses). Frontend is JS. |
| V. Developer Experience | PASS | No new build steps, no new dependencies to install. Alpine.js loads from CDN. `docker-compose up` still works. |

All gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/007-ui-ux-polish/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (client-side state model)
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
web/static/
├── index.html           # Modified: Alpine.js directives, unified content area
├── app.js               # Modified: Alpine stores, navigation stack, view management
└── style.css            # Modified: status badges, responsive layout, animations, skeletons

internal/server/
├── handlers.go          # Minor: ensure date field on list endpoint if needed
└── handlers_test.go     # Updated if handler changes
```

**Structure Decision**: No new files needed. All changes are modifications to existing `web/static/` files. The Go backend requires at most minor adjustments to support frontend needs.

## ADR Impact

| Existing ADR | Impact | Action |
|-------------|--------|--------|
| None | No existing decisions are overridden | No superseding ADRs needed |

New ADR needed: Alpine.js framework selection (ADR-023).

## Complexity Tracking

No constitution violations to justify.
