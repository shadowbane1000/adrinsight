# Data Model: UI & UX Polish

This milestone is frontend-only. No new database tables or backend entities. All state lives in the browser.

## Client-Side State (Alpine Store)

### Navigation Store (`Alpine.store('nav')`)

| Field | Type | Description |
|-------|------|-------------|
| stack | Array<ViewEntry> | Navigation history stack. Top = current view. |
| currentView | computed | Returns `stack[stack.length - 1]` or null |

**ViewEntry**:

| Field | Type | Description |
|-------|------|-------------|
| type | string | `"about"`, `"answer"`, `"adr"` |
| data | object | View-specific cached data |
| label | string | Breadcrumb display label (e.g., "Answer", "ADR-004") |

For `type: "answer"`:
- `data.query` — the original query string
- `data.answer` — rendered answer HTML
- `data.citations` — array of `{adr_number, title, section}`

For `type: "adr"`:
- `data.number` — ADR number
- `data.title` — ADR title
- `data.content` — raw markdown content
- `data.status` — status string
- `data.date` — date string
- `data.relationships` — array of relationship objects

### ADR Store (`Alpine.store('adrs')`)

| Field | Type | Description |
|-------|------|-------------|
| all | Array<ADRSummary> | Full ADR list from `/adrs` endpoint |
| filtered | computed | `all` filtered by active status filter, then sorted |
| filter | string or null | Active status filter (e.g., "accepted") or null for all |
| sort | string | Sort key: `"number-asc"` (default), `"number-desc"`, `"date-desc"`, `"date-asc"` |
| loading | boolean | True while fetching ADR list |
| error | string or null | Error message if list fetch failed |

**ADRSummary** (from API):

| Field | Type | Description |
|-------|------|-------------|
| number | int | ADR number |
| title | string | ADR title |
| status | string | Status string (e.g., "Accepted") |
| path | string | File path |
| normalizedStatus | computed | First word, lowercased (e.g., "accepted") — for badge CSS class |

### Query Store (`Alpine.store('query')`)

| Field | Type | Description |
|-------|------|-------------|
| text | string | Current search input value |
| loading | boolean | True while query is in-flight |
| error | string or null | Error message if query failed |
| lastQuery | string | Last submitted query (for retry) |
| history | Array<string> | Recent queries from localStorage (max 10) |
| showHistory | boolean | True when history dropdown is visible |

### UI Store (`Alpine.store('ui')`)

| Field | Type | Description |
|-------|------|-------------|
| sidebarOpen | boolean | Sidebar visibility on small screens (default: true on desktop) |

## localStorage Keys

| Key | Value Type | Description |
|-----|-----------|-------------|
| `adr-insight-query-history` | JSON array of strings | Last 10 queries, newest first |
| `adr-insight-filter` | string or null | Persisted status filter |
| `adr-insight-sort` | string | Persisted sort preference |

## No Backend Changes Required

The existing API responses already include all needed fields:
- `/adrs` — returns `{number, title, status, path}` per ADR
- `/adrs/{number}` — returns `{number, title, status, date, content, relationships}`
- `/query` — returns `{answer, citations: [{adr_number, title, section}]}`
