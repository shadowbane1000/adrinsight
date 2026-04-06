# ADR-001: Why Go

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight is a portfolio project intended to demonstrate principal-engineer-level architectural thinking, AI integration, and system design. The project needs a primary language that:

- Is relevant to the roles being targeted (Principal Engineer, Engineering Manager, Staff+)
- Is widely used in the platform engineering and developer tooling space
- Produces clean, readable code that hiring managers can quickly evaluate
- Has a strong ecosystem for HTTP servers, CLI tools, and concurrency

My professional background is strongest in C#, C++, Java, and Python. Go is a new language for me, which introduces learning risk but also demonstrates adaptability — a key attribute for senior technical leadership.

## Decision

Use Go as the primary language for ADR Insight.

## Rationale

- **Market alignment:** Go dominates the cloud-native and developer tooling ecosystem (Kubernetes, Docker, Terraform, and most modern CLI tools are written in Go). Demonstrating Go proficiency signals alignment with this space.
- **Simplicity as a feature:** Go's opinionated simplicity (limited language features, explicit error handling, strong standard library) makes code easy to review — important for a portfolio piece where someone may spend 10 minutes skimming the repo.
- **Single binary deployment:** Go compiles to a static binary with no runtime dependencies, making Docker images small and `docker-compose up` fast. This supports the "clone and run" developer experience goal.
- **Concurrency model:** Goroutines and channels are well-suited for the eventual architecture (parsing files, generating embeddings, querying LLMs concurrently).
- **Learning signal:** Choosing a new language for a serious project and writing idiomatic code demonstrates the rapid learning ability expected at the Principal/Staff level.

## Consequences

### Positive
- Aligns the project with the platform engineering and developer tooling job market
- Produces a small, fast, easily deployable artifact
- Forces clean, explicit code that reads well in a portfolio context

### Negative
- Learning curve will slow initial development; early code may not be fully idiomatic
- Go's AI/ML library ecosystem is less mature than Python's — LLM integration will rely on HTTP API calls rather than rich SDKs
- No generics-heavy abstractions or metaprogramming; some patterns from C# or Java won't translate directly

### Mitigations
- Use AI-assisted coding to accelerate Go learning and catch non-idiomatic patterns
- Lean on Go's strong standard library and well-established third-party packages (e.g., Chi or Echo for HTTP, Cobra for CLI)
- Revisit early code during Phase 2 to improve idiom compliance
