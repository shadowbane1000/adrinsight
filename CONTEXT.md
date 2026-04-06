# ADR Insight — Project Context

## What Is This?

ADR Insight is an AI-powered tool that ingests a repository's markdown Architecture Decision Records (ADRs), builds a semantic index, and lets users ask natural-language questions about their team's architectural decisions through a web UI.

**GitHub repo description:** "AI-powered search and reasoning over your Architecture Decision Records"

## Why This Project Exists

This is a portfolio project for Tyler Colbert, targeting Principal Engineer / Software Engineering Manager / Staff+ roles. It's designed to be maximally impressive to hiring managers by demonstrating:

1. **Leadership thinking** (top priority) — The project's own ADRs document every significant decision with context, tradeoffs, and consequences. The ADR collection *is* the most impressive part.
2. **AI/ML depth** — Thoughtful RAG implementation, not just "call an LLM API." Intentional tool selection (local embeddings vs cloud synthesis).
3. **System architecture & scalability** — Clean interfaces, swappable components, documented tradeoffs, and clear answers to "what would you change at scale?"
4. **Cloud-native & DevOps maturity** — Docker Compose, CI/CD from day one, infrastructure as code thinking.

**The meta-play:** The project's own ADRs are the demo dataset. A hiring manager opens it, asks "why did you choose Go?" and gets a synthesized answer from Tyler's own architectural reasoning.

## Key Decisions (see docs/adr/ for full ADRs)

- **Language:** Go (new to Tyler — demonstrates adaptability; aligns with platform engineering market)
- **LLM Synthesis:** Anthropic Claude API (existing account, strong instruction-following, long context window)
- **Embeddings:** Local via Ollama with `nomic-embed-text` model (zero external dependency for embeddings, clean `docker-compose up` experience)
- **Vector Storage:** SQLite with `sqlite-vec` extension (embedded, no server, vectors + metadata in one place, disposable/regenerable index)
- **CI/CD:** Gitea Actions (primary dev environment) + GitHub Actions (public mirror), both running lint + test + build

## Architecture Overview

```
User Question
     │
     ▼
 [Web UI] ──► [HTTP API]
                  │
                  ▼
            [RAG Pipeline]
             /          \
   [Embed Question]   [Retrieve Top-K]
   (Ollama local)     (SQLite + sqlite-vec)
                          │
                          ▼
                   [Synthesize Answer]
                   (Anthropic Claude)
                          │
                          ▼
                   [Answer + Citations]
```

**Data flow at index time:**
1. Parse markdown ADRs from a directory (extract title, date, status, body)
2. Chunk and embed via Ollama (`nomic-embed-text`)
3. Store vectors + metadata in SQLite with `sqlite-vec`

**Data flow at query time:**
1. User types a question in the web UI
2. Question is embedded via Ollama
3. Similarity search in SQLite returns top-K relevant ADR chunks
4. Retrieved chunks + question sent to Anthropic Claude for synthesis
5. Claude returns a natural-language answer with citations to specific ADRs
6. Web UI displays answer with clickable citation links

## Project Structure

```
adr-insight/
├── cmd/
│   └── adr-insight/
│       └── main.go              # entrypoint
├── internal/
│   ├── parser/                  # ADR markdown parsing
│   ├── embedder/                # embedding interface + Ollama implementation
│   ├── store/                   # SQLite + sqlite-vec storage
│   ├── rag/                     # retrieval + synthesis pipeline orchestration
│   ├── llm/                     # LLM interface + Anthropic implementation
│   └── server/                  # HTTP API + serves web UI
├── web/
│   └── static/                  # HTML/CSS/JS for the web UI
├── docs/
│   └── adr/                     # the project's own ADRs (also the demo dataset)
├── testdata/                    # sample ADRs for testing
├── .gitea/
│   └── workflows/
│       └── ci.yaml              # lint + test + build
├── .github/
│   └── workflows/
│       └── ci.yaml              # mirror of Gitea CI for GitHub
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── Makefile
├── README.md
└── .golangci.yml
```

**Go conventions:**
- `cmd/` for the entrypoint
- `internal/` signals packages aren't meant to be imported externally
- Interfaces in each package so components are swappable (embedder, LLM, store)

## Phased Roadmap

### Phase 1 — Walking Skeleton (weeks 1–3)
Goal: `docker-compose up`, open browser, ask a question, get a synthesized answer with citations.

**Week 1 — Foundation**
1. Initialize Go module, set up directory structure
2. ADR markdown parser — read a directory, extract frontmatter/metadata (title, date, status) and body text
3. Ollama integration — call the embedding API, get vectors back
4. SQLite + sqlite-vec setup — store ADR content and vectors, run similarity queries
5. `reindex` command: parse → embed → store
6. Gitea Actions + GitHub Actions CI (lint + test + build)

**Week 2 — The Brain**
7. Anthropic integration — send retrieved context + user question, get synthesized answer
8. RAG pipeline end-to-end: question → embed → retrieve top-K → synthesize → return with citations
9. HTTP API (at minimum: `POST /query` and `GET /adrs`)

**Week 3 — The Face**
10. Minimal web UI — search bar, results with citations linking to source ADRs, ADR browse/list view
11. Dockerfiles (Go app + Ollama with model pre-pulled)
12. `docker-compose.yml` that brings it all up
13. README with setup instructions

### Phase 2 — Make It Good (weeks 4–6)
- Improve chunking strategy (experiment, document results in an ADR)
- Hybrid search (keyword + semantic), reranking
- Web UI polish: ADR browsing/filtering, citation links, status badges
- ADR metadata awareness (understands supersedes/deprecated relationships)
- Solid error handling, logging, configuration
- Write ADRs for every significant decision along the way

### Phase 3 — Make It Impressive (weeks 7–9)
- Architecture diagram in the README
- Comprehensive ADR collection for the project itself
- Meaningful test strategy (integration tests on RAG pipeline, not coverage for its own sake)
- CI/CD via GitHub Actions (already in place from Phase 1, expand as needed)
- Clean, documented API for extensibility
- Blog post or detailed README walkthrough of design philosophy

### Phase 4 — Optional Stretch
- Decision dependency graph visualization in the UI
- Staleness detection ("this ADR is 2 years old and references a deprecated service")
- "Onboarding mode" — summarize the N most important decisions for new engineers
- Multi-LLM backend support (OpenAI, Ollama for synthesis, etc.)
- Multi-repo / org-wide federation

## Design Principles

- **Interfaces everywhere:** Every external dependency (LLM, embedder, store) sits behind a Go interface. This makes testing easy and shows "what would you change at scale" thinking.
- **ADRs are the product:** Every significant decision gets an ADR. These are simultaneously documentation, demo data, and interview talking points.
- **Skateboard first:** Do one thing well before expanding. The walking skeleton should be rough but complete end-to-end.
- **Developer experience matters:** `docker-compose up` and one API key should be all it takes. No complex setup.
- **The database is disposable:** SQLite file is a derived artifact, regenerable from source ADRs via `reindex` at any time.

## Technical Notes

- **Go proficiency:** Tyler is learning Go through this project. Use AI-assisted coding. Revisit early code in Phase 2 to improve idiom compliance.
- **Anthropic has no embeddings API** — that's why embeddings are handled separately via Ollama.
- **sqlite-vec** is relatively young but appropriate for this scale. The interface abstraction means it can be swapped later.
- **Ollama Docker image** is official and well-maintained. First run downloads `nomic-embed-text` (~275MB).
- **Gitea** is the primary dev environment with Actions runners already configured. GitHub is the public-facing mirror.

## Tyler's Background (relevant to project decisions)

- 25+ years in software, most recently Senior Software Architect at Hexagon
- Strongest in C#, C++, Java, Python — Go is new
- Deep cloud experience (Azure, AWS, Kubernetes, Docker, IaC)
- 14+ year tenure at Disney/Avalanche, 8+ years at Hexagon — values long-term commitment
- 3 US patents, Disney corporate inventor awards
- Has built an AI inference pipeline (audio capture → speech-to-text → LLM inference)
- Exploring spec-driven development
- Located in Syracuse, UT
