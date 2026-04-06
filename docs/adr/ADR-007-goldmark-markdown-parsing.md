# ADR-007: goldmark for Markdown Parsing

**Status:** Accepted  
**Date:** 2026-04-06  
**Deciders:** Tyler Colbert

## Context

ADR Insight needs to parse markdown ADR files to extract structured metadata (title, date, status) and body content split by section. The parser must handle YAML-style frontmatter, walk the document's heading structure, and extract section-level content for chunking. This is the entry point of the entire data pipeline — every downstream component depends on the parser producing correct, structured output.

## Decision

Use goldmark with the goldmark-meta extension for markdown parsing and frontmatter extraction.

## Rationale

- **CommonMark compliant:** goldmark is the most standards-compliant Go markdown parser, reducing the risk of parsing edge cases with unusual markdown formatting in ADR files.
- **Extensible AST:** goldmark's `ast.Walk` API provides structured traversal of the parsed document tree. This is the right abstraction for extracting headings and their associated content — much more robust than regex-based approaches.
- **goldmark-meta extension:** Adds YAML frontmatter support as a first-class feature, parsing the `---` delimited block into a map. This handles the ADR metadata lines (Status, Date, Deciders) cleanly.
- **Active maintenance:** goldmark has 4.7k+ stars and is actively maintained as of 2026, with regular releases.
- **Go ecosystem standard:** goldmark is used by Hugo and other major Go projects, making it the de facto standard for markdown processing in Go.

## Consequences

### Positive
- Robust handling of markdown edge cases via CommonMark compliance
- Clean separation of parsing (goldmark) and structure extraction (AST walking)
- YAML frontmatter handled by a dedicated extension rather than custom parsing
- Well-documented, widely-used library with strong community support

### Negative
- Adds two external dependencies (goldmark + goldmark-meta) where simpler string splitting could work for well-formatted ADRs
- AST walking has a learning curve for someone new to Go and goldmark's node types
- Slightly heavier than a minimal regex approach for the simple case

### Mitigations
- goldmark is a stable, well-maintained dependency — low risk of churn
- The Parser interface (Constitution Principle I) means the implementation can be swapped if a simpler approach proves sufficient
- Table-driven tests with fixture files in `testdata/` will validate parsing behavior and make the AST walking code easy to verify

## Alternatives Considered

### blackfriday
- **Rejected because:** Older library, largely in maintenance mode. Its v2 API is less extensible for structured AST walking. No built-in frontmatter support.

### Manual regex/string parsing
- **Rejected because:** Fragile for anything beyond perfectly formatted files. Doesn't handle markdown edge cases (nested headings, inline code in headings, varied frontmatter formatting). Would require more code and more test cases to achieve the same robustness.

## Related ADRs

- ADR-001: Why Go — establishes Go as the language, making goldmark the natural choice from Go's markdown ecosystem
