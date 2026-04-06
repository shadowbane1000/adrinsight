# ADR-005: Gitea as Primary Dev Environment, GitHub as Public Mirror

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight needs a CI/CD platform and a source control host. Tyler runs a self-hosted Gitea instance with Actions runners already configured. The project also needs public visibility on GitHub for portfolio purposes — hiring managers and interviewers will evaluate the repo there.

This creates a dual-hosting question: where does development actually happen, and how does the code get to the public-facing repository?

## Decision

Use Gitea as the primary development environment with CI via Gitea Actions. Mirror the repository to GitHub for public access. Maintain parallel CI configurations (Gitea Actions + GitHub Actions) so that both platforms run lint + test + build independently.

## Rationale

- **Existing infrastructure:** Gitea instance and Actions runners are already operational. No setup cost, no per-minute billing, no rate limits on CI runs during heavy development.
- **Full control:** Self-hosted runners mean no waiting in shared queues, no restrictions on Docker-in-Docker or resource usage, and complete control over the build environment.
- **Portfolio visibility:** GitHub is where hiring managers look. The mirror ensures the repo, commit history, and CI status badges are all visible where they matter.
- **CI redundancy:** Running CI on both platforms catches environment-specific issues early. If the Gitea Actions workflow file syntax drifts from GitHub Actions (they're compatible but not identical), having both configs keeps them honest.

## Consequences

### Positive
- Zero CI cost during development (self-hosted runners)
- No external dependency for the inner development loop
- Public GitHub presence for portfolio and discoverability
- CI runs on two independent environments, increasing confidence

### Negative
- Two CI configurations to maintain (`.gitea/workflows/` and `.github/workflows/`)
- Mirror sync adds a step to the workflow — pushes to Gitea must propagate to GitHub
- External contributions can only happen on GitHub since Gitea is not publicly exposed, meaning PRs must be pulled back into the Gitea workflow

### Mitigations
- Keep CI workflow files as similar as possible — Gitea Actions is largely GitHub Actions-compatible, so most of the YAML is shared structure
- Set up Gitea's built-in repository mirroring to push to GitHub automatically on every push, minimizing manual sync burden
- Accept contributions via GitHub PRs and pull them into the Gitea workflow as needed — this is a portfolio project so external contribution volume is expected to be near zero
