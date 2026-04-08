# ADR-023: Alpine.js for Reactive UI State Management

**Status:** Accepted  
**Date:** 2026-04-07  
**Deciders:** Tyler Colbert  
**Supersedes:** None

## Context

Milestone 6 (UI & UX Polish) adds significant frontend complexity beyond what
the existing vanilla JavaScript can cleanly handle. The feature set includes:

- A navigation stack with cached views and breadcrumb trail (answer → ADR → 
  related ADR, with instant backward navigation to any point)
- Reactive filter and sort controls for the ADR sidebar list
- A recent query history dropdown backed by localStorage
- Animated transitions between views (fade in/out on navigation)
- A responsive sidebar that collapses to a drawer on tablet-sized screens
- Loading skeletons and error states with retry actions

The current `app.js` (~170 lines of vanilla JS) uses imperative DOM
manipulation — `classList.add/remove`, `innerHTML` assignments, and manual
event listeners. Extending this approach to support reactive state, cached
view switching, and declarative transitions would require hand-rolling a
state management system and view layer, ballooning the code to 600+ lines
of fragile imperative logic.

The constraint from ADR-012 (marked.js) and ADR-013 (go:embed) still
applies: the framework must be CDN-loadable via a single `<script>` tag
with no build step, no node_modules, and no bundler. Static files are
embedded into the Go binary.

## Decision

Use Alpine.js 3.x loaded via CDN for reactive UI state management and
declarative DOM interactions. Setup is a single script tag:

```html
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

Alpine.js adds HTML directives (`x-show`, `x-bind`, `x-on`, `x-for`,
`x-transition`) that make the existing HTML declaratively reactive. Shared
state lives in `Alpine.store()` objects (navigation stack, ADR list with
filter/sort, query state, UI toggles).

## Rationale

- **HTML-first approach:** Alpine enhances existing HTML with directives
  rather than replacing it with a virtual DOM or JSX. The `index.html`
  remains the single source of truth for layout and structure — matching
  the project's existing pattern and making the UI readable without
  understanding a component tree.
- **No build step:** A single `<script defer>` tag from a CDN. Compatible
  with `go:embed` static file embedding. Consistent with the ADR-012
  precedent (marked.js via CDN).
- **Built-in transitions:** `x-transition` provides declarative CSS
  transitions for view switching — fade, slide, scale — without manual
  CSS animation management. This directly satisfies the animation
  requirements (FR-012).
- **Reactive state management:** `Alpine.store()` provides reactive global
  state that auto-updates the DOM when values change. Filter the ADR list?
  Set `store.filter = 'accepted'` and the list re-renders. Push onto the
  nav stack? The breadcrumb updates automatically.
- **Right-sized complexity:** At ~17kb gzipped, Alpine is the largest of
  the CDN-loadable options but the best fit for the directive-based,
  HTML-enhancement pattern this project uses. Smaller alternatives either
  require a different mental model (Preact/Lit) or are unmaintained
  (petite-vue).

## Consequences

### Positive
- Declarative UI logic replaces imperative DOM manipulation — more readable,
  less error-prone
- Navigation stack, filter/sort, and query history are simple reactive store
  operations
- View transitions are one-line directives instead of manual CSS animation code
- No build pipeline — CDN tag matches existing marked.js approach
- Large community, active maintenance, good documentation
- Learning curve is low — HTML directives, not a new language or paradigm

### Negative
- Adds a 17kb gzipped dependency to the frontend (largest of the alternatives)
- CDN dependency — if jsDelivr is unreachable, Alpine doesn't load and the
  UI becomes non-functional (worse than marked.js failure, which just shows
  raw markdown)
- Alpine evaluates directives on DOMContentLoaded — data that must be
  fetched before first render requires manual `Alpine.start()` instead of
  the `defer` attribute
- Team members unfamiliar with Alpine need to learn the directive syntax

### Mitigations
- 17kb is acceptable for a single-page portfolio app — total page weight
  is still well under 100kb
- A local copy of Alpine could be bundled via `go:embed` if CDN reliability
  becomes a concern (same fallback path as marked.js per ADR-012)
- Initial data fetch (ADR list) can use `x-init` to trigger loading after
  Alpine starts — no manual `Alpine.start()` needed for this use case

## Alternatives Considered

### Vanilla JavaScript (no framework)
- **Not chosen because:** The required features (nav stack with cached views,
  reactive filter/sort, animated transitions, history dropdown) would require
  hand-rolling a state management and view layer. The resulting 600+ lines
  of imperative DOM manipulation would be harder to maintain than the feature
  warrants.

### Preact + HTM (~5.3kb)
- **Not chosen because:** Requires `type="module"` imports and a
  component-oriented mental model (virtual DOM, JSX-like syntax via HTM).
  This is a fundamentally different approach than enhancing existing HTML —
  it would require rewriting the UI as a component tree. Also needs a
  separate routing library for the navigation stack.

### petite-vue (~7kb)
- **Not chosen because:** Effectively abandoned. GitHub issues are disabled,
  npm downloads are ~1.9k/week (vs Alpine's 446k/week), and the last
  meaningful release was 0.4.1. Building on an unmaintained dependency
  creates risk.

### Lit (~6kb)
- **Not chosen because:** Web components add ceremony (class definitions,
  shadow DOM, custom element registration) that is overkill for toggling
  views and managing filter state. Better suited for building reusable
  component libraries than for application-level UI enhancement.

## Related ADRs

- ADR-012: marked.js Client Rendering [related] — established the CDN-loadable, no-build-step pattern for frontend dependencies; Alpine.js follows the same approach. ADR-012's context stated "no frontend framework" — Alpine.js amends that constraint while preserving the no-build-step requirement.
- ADR-013: go:embed Static Files [related] — static files are embedded into the Go binary; Alpine.js is CDN-loaded alongside them
