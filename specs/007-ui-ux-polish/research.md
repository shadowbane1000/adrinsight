# Research: UI & UX Polish

## R1: Frontend Framework Selection

**Decision**: Alpine.js 3.x via CDN

**Rationale**:
- The feature set (navigation stack, reactive filter/sort, cached views, animated transitions, localStorage history, responsive sidebar toggle) requires state management beyond what vanilla JS handles cleanly. The current 170-line app.js would balloon to 600+ lines of imperative DOM manipulation.
- Alpine.js uses HTML directives (`x-show`, `x-bind`, `x-on`, `x-transition`) that enhance existing HTML rather than replacing it. The index.html remains the source of truth — no virtual DOM, no JSX, no build step.
- `Alpine.store()` provides a single reactive state object for filters, sort, navigation stack, and query history that auto-updates the DOM.
- `x-transition` handles animated view transitions (fade in/out) declaratively, avoiding manual CSS animation management.
- CDN setup is a single script tag: `<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>`
- 17kb gzipped — largest of the candidates but well within budget for a single-page app.

**Alternatives considered**:

| Framework | Size | Verdict |
|-----------|------|---------|
| Preact + HTM | ~5.3kb | Requires `type="module"` imports, component-oriented mental model, separate routing library. More architecture than needed for progressive enhancement of server-rendered HTML. |
| petite-vue | ~7kb | Effectively abandoned — issues disabled, 1.9k npm downloads/week, last release 0.4.1. Too risky. |
| Lit | ~6kb | Web components add ceremony (class definitions, shadow DOM, custom element registration) overkill for view toggling and filter state. Better suited for reusable component libraries. |
| Vanilla JS | 0kb | Zero cost but requires hand-rolling reactive state, view caching, transition animations, and declarative bindings. The feature list would produce fragile, hard-to-maintain imperative code. |

**Gotcha**: Alpine evaluates directives on DOMContentLoaded via `defer`. If state must be initialized from a fetch before Alpine starts, use `Alpine.start()` manually instead of `defer`.

## R2: Navigation Stack Pattern

**Decision**: Client-side stack array in Alpine store

**Rationale**:
- The navigation stack is an array of view objects: `[{type: 'answer', data: {...}}, {type: 'adr', number: 4, data: {...}}]`
- Current view is the top of the stack. Breadcrumbs render from the stack array.
- "Back" pops the stack. Clicking a breadcrumb segment slices the stack to that index.
- New query clears the stack and pushes the answer. Sidebar clicks push (if answer active) or replace (if browsing).
- All view data is cached in the stack entries, so backward navigation is instant (no re-fetch).
- No URL routing needed — this is a single-page app without shareable URLs for individual views.

**Alternatives considered**:
- Browser History API (pushState/popState): Adds URL routing complexity. URLs for individual answers don't make sense since they're ephemeral LLM responses. Overkill.
- Single active view variable: Can't represent the stack history needed for breadcrumb navigation.

## R3: Inline Citation Post-Processing

**Decision**: Scan rendered HTML for ADR-NNN patterns matching the citations array

**Rationale**:
- The LLM naturally references ADR numbers in its answers (e.g., "ADR-004 chose SQLite because...").
- After `marked.parse()` renders the markdown, scan the HTML text nodes for patterns like `ADR-\d+` or `ADR-\d{3}`.
- For each match that corresponds to an entry in the citations array, wrap it in a clickable `<a>` element that pushes onto the navigation stack.
- This avoids changing the LLM prompt or synthesis logic.
- Non-cited ADR references (ADR numbers not in the citations array) are left as plain text.

## R4: Responsive Sidebar Strategy

**Decision**: Collapsible sidebar with toggle button at 768px breakpoint

**Rationale**:
- Below 768px, the sidebar becomes a slide-out drawer triggered by a hamburger/toggle button in the header.
- The content area takes full width when the sidebar is collapsed.
- Alpine's `x-show` with `x-transition` handles the slide animation.
- The toggle state is managed in the Alpine store.
- Above 768px, the sidebar is always visible (current behavior).

**Alternatives considered**:
- Stacking (sidebar above content): Pushes content below the fold. Poor for the primary use case (reading an answer or ADR detail).
- Tab-based navigation: Loses the ability to see the ADR list while reading content. Sidebar is more natural for a list + detail layout.

## R5: Status Badge Color Mapping

**Decision**: CSS classes mapped to normalized status values

**Rationale**:
- Status values from the API: "Accepted", "Proposed", "Deprecated", "Superseded", and variations (e.g., "Accepted (validated by ADR-022)", "Superseded by ADR-015").
- Normalize by extracting the first word and lowercasing: "accepted", "proposed", "deprecated", "superseded".
- Map to CSS classes: `.badge-accepted` (green), `.badge-proposed` (blue), `.badge-deprecated` (amber), `.badge-superseded` (gray), `.badge-unknown` (neutral gray fallback).
- Badges render as small inline `<span>` elements with rounded corners and colored backgrounds.
