# UI Layout Contract: ADR Insight Web UI

## Page Structure

```
┌─────────────────────────────────────────────────────┐
│  ADR Insight                              [search bar] [Go]  │
├─────────────────────────────────────────────────────┤
│  Answer Area (hidden until query submitted)          │
│  ┌───────────────────────────────────────────────┐  │
│  │ Synthesized answer (markdown rendered)         │  │
│  │ Citations: [ADR-001] [ADR-003] (clickable)    │  │
│  └───────────────────────────────────────────────┘  │
├────────────────────┬────────────────────────────────┤
│  ADR List (left)   │  Detail Panel (right)           │
│                    │                                 │
│  ADR-001: Why Go   │  About (default)                │
│  ADR-002: Anthro.. │  ─────────────────              │
│  ADR-003: Local E..│  ADR Insight is an AI-powered   │
│  ADR-004: SQLite.. │  tool for querying Architecture │
│  ADR-005: Gitea..  │  Decision Records...            │
│  ...               │                                 │
│                    │  Built with spec-driven dev...   │
│                    │                                 │
│                    │  [GitHub →]                      │
│                    │                                 │
│                    │  ─── OR (when ADR clicked) ───  │
│                    │                                 │
│                    │  # ADR-001: Why Go              │
│                    │  **Status:** Accepted            │
│                    │  (full markdown rendered)        │
│                    │                                 │
└────────────────────┴────────────────────────────────┘
```

## Interactions

| User Action | Result |
|-------------|--------|
| Type query + submit | Answer area appears with loading indicator, then rendered answer + citation links |
| Click citation link | Right panel switches to that ADR's full content (markdown rendered) |
| Click ADR in list | Right panel switches to that ADR's full content (markdown rendered) |
| Page load | Answer area hidden, right panel shows About content, ADR list populated |
| Click header title | Right panel returns to About content (from any ADR detail view) |

## About Panel Content

The default right-side panel displays:

- **Project name**: ADR Insight
- **Description**: AI-powered search and reasoning over Architecture Decision Records
- **How it was built**: Built using spec-driven development with a modified version of GitHub's spec-kit that includes ADR generation as part of the development workflow
- **Universality**: Any project that maintains ADRs in the standard markdown format can use ADR Insight to provide natural-language Q&A over their architecture decisions
- **GitHub link**: Link to the repository
- **Author**: Tyler Colbert

## API Calls

| UI Action | Endpoint | When |
|-----------|----------|------|
| Page load | `GET /adrs` | On load — populate ADR list |
| Submit query | `POST /query` | User clicks Go or presses Enter |
| Click ADR/citation | `GET /adrs/{number}` | User clicks ADR link |
