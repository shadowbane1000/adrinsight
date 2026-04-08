# Quickstart: UI & UX Polish

## Verification Scenarios

### Scenario 1: Status Badges
1. Run `./adr-insight serve --dev`
2. Open browser to `http://localhost:8081`
3. Verify: Each ADR in the sidebar shows a colored status badge
4. Click ADR-008 (status: "Accepted (validated by ADR-022)") — badge shows "Accepted" in green
5. Click ADR-006 — badge shows "Superseded" in gray

### Scenario 2: Filtering and Sorting
1. Click the status filter control above the ADR list
2. Select "Accepted" — list shows only accepted ADRs
3. Change sort to "Date (newest)" — order updates
4. Click an ADR to view details, then click another — filter and sort persist
5. Click "Show All" — full list returns

### Scenario 3: Unified Content Area with Navigation
1. Submit query: "What are all the ways SQLite is used in this system?"
2. Verify: Answer appears in the content area (full panel width)
3. Answer text contains clickable ADR-004, ADR-015, ADR-017 references
4. Click ADR-004 in the answer — content switches to ADR-004 detail
5. Breadcrumb shows: "Answer > ADR-004"
6. Click a relationship link on ADR-004 (e.g., ADR-015)
7. Breadcrumb shows: "Answer > ADR-004 > ADR-015"
8. Click "ADR-004" in breadcrumb — returns to ADR-004 instantly
9. Click "Answer" in breadcrumb — returns to answer instantly (no re-query)

### Scenario 4: Error and Retry
1. Stop the server
2. Submit a query — error message appears with "Try Again" button
3. Restart the server
4. Click "Try Again" — query re-fires and answer loads

### Scenario 5: Query History
1. Submit a few queries
2. Click into the search input — dropdown shows recent queries
3. Click a previous query — it fires immediately
4. Reload the page — history persists
5. Clear history — dropdown is empty

### Scenario 6: Responsive Layout
1. Open browser dev tools, toggle device toolbar
2. Set viewport to 768px width
3. Verify: Sidebar collapses, toggle button appears in header
4. Click toggle — sidebar slides out as drawer
5. All functionality works at this width

### Scenario 7: Search Clear
1. Submit a query and read the answer
2. Click the X button in the search bar
3. Verify: Content area returns to the about page, search input is empty
4. Press Escape while typing — same behavior

### Scenario 8: Visual Polish
1. Verify: Answer fades in smoothly when loaded
2. Verify: Hover states on buttons, links, and list items have subtle transitions
3. Verify: No jarring content jumps when switching views
4. Verify: Typography is consistent across answer, ADR detail, and about page
