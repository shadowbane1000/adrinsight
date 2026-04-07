# Quickstart: Web UI and Docker Deployment — Integration Test Scenarios

## Prerequisites

- Docker and Docker Compose v2 installed
- Anthropic API key (`ANTHROPIC_API_KEY` environment variable)
- ADR markdown files in `docs/adr/` (project's own ADRs work as demo data)

## Scenario 1: Docker Compose Full Stack

```bash
# Start everything
ANTHROPIC_API_KEY=sk-... docker compose up

# Expected: app and ollama services start
# Ollama pulls nomic-embed-text model on first run (may take a few minutes)
# App auto-reindexes ADRs on first start (skips if database already populated)
# App becomes ready on http://localhost:8081
```

**Verify**: Open `http://localhost:8081` in a browser. The web UI loads with:
- Search bar at top
- ADR list on the left
- About panel on the right

## Scenario 2: Ask a Question via Web UI

1. Open `http://localhost:8081`
2. Type "Why did we choose Go?" in the search bar
3. Click Go (or press Enter)

**Expected**: Loading indicator appears, then a synthesized answer with citation
links like [ADR-001]. Clicking [ADR-001] shows the full ADR content in the
right panel.

## Scenario 3: Browse ADRs

1. Open `http://localhost:8081`
2. Click any ADR in the left-side list (e.g., "ADR-001: Why Go")

**Expected**: The right panel switches from the About content to the full
rendered markdown of that ADR.

## Scenario 4: No Relevant ADRs

1. Type "microservice deployment with Kubernetes" in the search bar
2. Submit the query

**Expected**: A message indicating no relevant information was found. No errors
or blank content.

## Scenario 5: Dev Mode (Local Development)

```bash
# Build and run with --dev flag for live static file editing
go build -o adr-insight ./cmd/adr-insight
./adr-insight serve --dev

# Expected: serves static files from disk (web/static/) instead of embedded
# Changes to HTML/CSS/JS are reflected on browser refresh without recompilation
```

## Scenario 6: Docker Without API Key

```bash
docker compose up
# (without setting ANTHROPIC_API_KEY)

# Expected: app starts but query requests return a clear error message
# ADR browsing still works (doesn't require the API key)
```

## Scenario 7: About Panel Content

1. Open `http://localhost:8081`
2. Look at the right-side panel (default view)

**Expected**: Displays project description including:
- Project name (ADR Insight)
- Description of AI-powered ADR search
- Spec-driven development methodology note
- Universality statement (any project with ADRs can use it)
- GitHub link
- Author (Tyler Colbert)
