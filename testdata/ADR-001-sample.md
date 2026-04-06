# ADR-001: Use Go for the Project

**Status:** Accepted  
**Date:** 2026-01-15  
**Deciders:** Alice, Bob

## Context

We need to choose a programming language for the new service. The team has
experience with Python, Java, and Go. The service needs to handle concurrent
requests efficiently and deploy as a single binary.

## Decision

Use Go as the primary language.

## Rationale

Go produces static binaries with no runtime dependencies. Its goroutine model
handles concurrency well. The standard library covers most of our needs for
HTTP servers and CLI tools.

## Consequences

### Positive
- Fast compilation and single binary deployment
- Strong concurrency primitives

### Negative
- Less mature AI/ML ecosystem compared to Python
- Team needs to learn Go idioms
