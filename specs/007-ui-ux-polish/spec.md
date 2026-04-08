# Feature Specification: UI & UX Polish

## Overview

ADR Insight's web interface is functional but minimal — it serves its purpose for demonstrating the RAG pipeline, but doesn't feel like a product someone would want to use regularly. This milestone transforms the UI from a developer prototype into a polished, production-quality interface that a hiring manager or non-technical stakeholder could navigate without guidance.

The current UI lacks visual cues for ADR status, has no filtering or sorting capabilities, presents citations as a flat list below the answer rather than inline, and breaks down on smaller screens. Loading and error states are bare-minimum. Most critically, the answer area and ADR detail panel compete for vertical space — long answers push the detail panel down, making it difficult to read both. This milestone addresses all of these gaps, starting with a unified content area that gives the full panel to whichever view the user is focused on.

## Actors

- **Interviewer / Hiring Manager** — evaluates the project as a portfolio piece; needs to understand value quickly without technical hand-holding
- **Developer** — uses the tool to explore ADR decisions; benefits from efficient navigation and filtering
- **Presenter** — demonstrates the system in a live setting; needs the UI to look polished and feel responsive

## Functional Requirements

### FR-001: Status Badges in ADR List and Detail
ADR status (Accepted, Proposed, Deprecated, Superseded) must be visually distinguished using color-coded badges in both the sidebar list and the detail panel header. Each status maps to a distinct color:
- Accepted: green
- Proposed: blue
- Deprecated: amber/orange
- Superseded: gray with muted styling

The badge must be visible without hovering or clicking.

### FR-002: ADR List Filtering by Status
Users can filter the sidebar ADR list by one or more statuses. The filter control appears above the ADR list. Selecting a status filter immediately updates the visible list. A "Show All" option clears all filters. Filter state persists when navigating between ADRs (clicking an ADR and returning to the list preserves the active filter).

### FR-003: ADR List Sorting
Users can sort the ADR list by number (ascending/descending) or by date (newest first / oldest first). Default sort is by number ascending (current behavior). Sort state persists during the session.

### FR-004: Unified Content Area with Navigation Stack
The answer view and ADR detail view share a single content panel (the main area to the right of the sidebar). Only one is visible at a time:
- When the user submits a query, the answer replaces whatever was in the content area and clears the navigation stack.
- When the user clicks a citation or relationship link, the ADR detail replaces the current view and pushes it onto the navigation stack.
- When the user clicks an ADR in the sidebar: if an answer is active in the stack, the ADR pushes onto the stack (preserving the path back to the answer). If no answer is active (just browsing), the sidebar click replaces the current view without growing the stack.
- A breadcrumb trail appears at the top showing the navigation path (e.g., "Answer > ADR-004 > ADR-015"). Each segment is clickable — the user can jump back to any point in the chain, not just the previous view.
- All views in the stack are cached in memory so navigating backward is instant (no re-fetch for the answer, no re-load for previously viewed ADRs).
- Clicking the header title returns to the about/landing page and clears the stack.

This eliminates the split-screen problem where long answers consume space needed for ADR reading, and supports deep exploration of decision chains (e.g., answer → cited ADR → related ADR → superseding ADR, then back to any step).

### FR-005: Inline Citation Markers
After the answer is rendered, ADR references (e.g., "ADR-004", "ADR-015") that match entries in the citations array are automatically converted into clickable links. Clicking an inline citation link navigates to that ADR in the content area (pushing onto the navigation stack). No changes to the LLM prompt are needed — the post-processing scans the rendered HTML for ADR-NNN patterns. The citation summary list below the answer is retained as a quick-reference.

### FR-006: Improved Loading States
While the ADR list is loading, a skeleton/placeholder UI appears instead of a blank area. While a query is processing, the loading indicator includes a subtle animation and disables the search form. If a query takes longer than 5 seconds, a "Still thinking..." message appears to reassure the user.

### FR-007: Error States with Retry
When a query fails, the error message includes a "Try Again" button that re-submits the last query. When the ADR list fails to load, a "Retry" button appears in the sidebar. Error messages are user-friendly (no raw error codes or stack traces).

### FR-008: Search Clear and Reset
A visible clear button (X) appears in the search input when text is present. Clicking it clears the input, clears the navigation stack, and returns to the about page. The cached answer is discarded. The Escape key also clears the search input.

### FR-009: Responsive Layout
The layout adapts to screens as narrow as 768px (tablet). Below the breakpoint, the sidebar collapses to a toggleable drawer or stacks above the detail panel. The content area remains usable on smaller screens. The UI does not need to be fully mobile-optimized but must not be broken or unusable on tablets.

### FR-010: Visual Refinement
Consistent spacing and typography hierarchy across all views. Subtle transitions on interactive elements (hover states, panel transitions, answer appearance). The header, sidebar, and detail panel have clear visual separation. The overall color scheme remains cohesive (current dark header + white content is retained, with refinements to spacing and contrast).

### FR-011: Recent Query History
The system stores the last 10 queries in the browser's local storage. A small dropdown or list appears below the search input when focused (or when clicked), showing recent queries. Clicking a recent query populates the search input and submits it. Users can clear their query history. History persists across page reloads and browser sessions.

### FR-012: Answer Appearance Animation
When a new answer loads, it fades in smoothly rather than appearing abruptly. Citations appear after the answer with a slight stagger delay. This creates a polished feel during the most-watched interaction.

## User Stories & Scenarios

### US-1: Interviewer Browses ADR List (P1)
An interviewer opens ADR Insight and sees the ADR list in the sidebar with color-coded status badges. They can quickly identify which decisions are current (Accepted) versus replaced (Superseded). They click an ADR and the detail panel shows the full content with a status badge in the header.

**Acceptance**: Status badges visible in list and detail. No additional clicks needed to see status.

### US-2: Developer Filters and Sorts ADRs (P1)
A developer wants to find only the active decisions. They select "Accepted" from the filter control, and the list shows only accepted ADRs. They switch the sort to "Date (newest)" to see recent decisions first. Navigating to an ADR detail and clicking back preserves both the filter and sort.

**Acceptance**: Filter and sort controls work independently. State persists across navigation. "Show All" clears filters.

### US-3: User Explores Decision Chain from Answer (P1)
A user asks a question. The answer loads in the content area with inline citation markers like [ADR-004]. They click a citation to read ADR-004 — the breadcrumb shows "Answer > ADR-004". From ADR-004, they click a relationship link to ADR-015 — the breadcrumb shows "Answer > ADR-004 > ADR-015". They click "ADR-004" in the breadcrumb to jump back, then click "Answer" to return to the original answer instantly.

**Acceptance**: Answer and ADR detail share the same content area. Breadcrumb trail shows full navigation path. Each breadcrumb segment is clickable. All cached views restore instantly. Navigation works from citations, relationship links, and sidebar clicks.

### US-4: User Encounters Error and Retries (P2)
A user submits a query but the server is temporarily unavailable. An error message appears with a "Try Again" button. The user clicks it and the query re-fires without needing to retype.

**Acceptance**: Error includes retry action. Retry re-submits the exact previous query. Works for both query errors and ADR list load failures.

### US-5: User on a Tablet (P3)
A user opens ADR Insight on an iPad or similar tablet-sized screen. The layout adjusts: the sidebar either collapses to a toggle or stacks above the detail content. The search bar remains accessible. The answer area and detail panel are readable without horizontal scrolling.

**Acceptance**: Layout is usable at 768px width. No content is cut off or overlapping. Core functionality (search, browse, read) works.

### US-6: User Revisits a Previous Query (P2)
A user asked a question yesterday and wants to ask it again. They click into the search input and see a dropdown of their recent queries. They click one and it fires immediately. Later, they clear their history.

**Acceptance**: Recent queries appear on search focus. Clicking one submits it. History persists across reloads. Clear history option works.

### US-7: User Clears Search (P3)
A user has submitted a query and is reading the answer. They want to start fresh. They click the X button in the search bar (or press Escape), and the content area returns to the about page. The cached answer is discarded.

**Acceptance**: Clear button appears when text is present. Clicking it or pressing Escape clears input and returns to the about page.

## Edge Cases & Error Handling

- **All ADRs have the same status**: Filter shows all or none; no empty-state confusion
- **ADR with missing or empty status field**: Display an "Unknown" badge in a neutral color (gray)
- **Answer text contains markdown that looks like citation markers**: Only system-generated citations are interactive; raw markdown brackets are rendered as text
- **Filter active + no results**: Show "No ADRs match the current filter" with a link to clear the filter
- **Very long answer text**: The content area scrolls; inline citations must remain clickable
- **User clicks ADR in sidebar while viewing an answer**: ADR pushes onto the stack; breadcrumb still allows returning to the answer
- **User clicks ADR in sidebar while just browsing (no answer active)**: ADR replaces the current view; no stack growth
- **User submits new query while viewing an ADR from a previous answer**: New answer replaces everything; navigation stack is cleared
- **Deep navigation stack (5+ levels)**: Breadcrumb trail truncates gracefully (e.g., "Answer > ... > ADR-006 > ADR-015") while still allowing jumps to any point
- **User clicks same ADR that's already on the stack**: Navigates back to that point in the stack rather than pushing a duplicate
- **Query while previous query is loading**: New query replaces the in-flight one; no stale answer appears

## Success Criteria

### Measurable Outcomes

- **SC-001**: A first-time user can identify the status of any ADR within 2 seconds of seeing the list (status badges visible without interaction)
- **SC-002**: Filtering the ADR list to a single status takes no more than 1 click
- **SC-003**: All interactive elements (buttons, links, citations) respond within 100ms of user action (visual feedback)
- **SC-004**: The UI is usable (no overlapping content, no horizontal scroll, core functions work) at 768px viewport width
- **SC-005**: Error states always include a user-actionable recovery path (retry button or clear action)
- **SC-006**: Page transitions and answer loading include visible smooth animation (no jarring content jumps)
- **SC-007**: Navigating from answer to a cited ADR and back takes no more than 1 click each way, with no visible loading delay on return

## Scope & Boundaries

### In Scope
- All frontend changes (HTML, CSS, JavaScript)
- Minor backend changes if needed to support new frontend features (e.g., adding a field to an API response)
- Unified content area replacing the separate answer area and detail panel
- Breadcrumb navigation between answer and ADR detail views
- Status badge color mapping
- Filter and sort controls
- Inline citation markers
- Loading skeletons and error retry
- Recent query history (localStorage)
- Responsive layout at 768px breakpoint
- Visual polish (spacing, transitions, typography)

### Out of Scope
- Search-as-you-type / autocomplete suggestions (deferred — requires backend support)
- Recent query history beyond the last 10 queries
- Dark mode (deferred — nice-to-have for a later milestone)
- Full mobile optimization below 768px
- Backend performance optimization
- New API endpoints (unless minimally required for a frontend feature)

## Dependencies

- Milestone 5 (ADR Intelligence) must be merged — the relationship display in the detail panel is already implemented and should not regress
- The existing API responses (`/adrs`, `/adrs/{number}`, `/query`) provide sufficient data for all planned features — status, date, relationships, and citations are already returned

## Clarifications

### Session 2026-04-07
- Q: Should sidebar clicks always push onto the navigation stack, or only when an answer is active? → A: Push only when answer is active; otherwise replace (Option B). May revisit if it feels awkward in practice.
- Q: How should inline citation markers be generated? → A: Post-process rendered answer HTML — scan for ADR-NNN patterns matching the citations array and wrap as clickable links. No LLM prompt changes.

## Assumptions

- The current color scheme (dark header #1a1a2e, accent #e94560, white content panels) is retained and refined, not redesigned
- A lightweight frontend framework may be introduced if research determines it provides significant value for the navigation stack, state management, and view caching requirements. The framework must work without a build step (CDN-loadable via script tag) to preserve the current `go:embed` static file approach. Framework selection is a planning-phase research question.
- The ADR list size (currently 22 ADRs) is small enough that client-side filtering and sorting is appropriate
- Citation markers are generated by post-processing rendered answer HTML, not by changing the LLM prompt
