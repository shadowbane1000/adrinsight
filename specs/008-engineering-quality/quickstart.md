# Quickstart: Engineering Quality

## Verification Scenarios

### Scenario 1: Structured Logging
1. Run `LOG_FORMAT=text LOG_LEVEL=debug ./adr-insight serve --dev`
2. Verify: Log output is human-readable text with timestamp, level, message, and fields
3. Run `LOG_FORMAT=json ./adr-insight serve --dev`
4. Verify: Log output is JSON, one object per line, parseable by `jq`
5. Submit a query — verify log entries include request_id and duration_ms fields

### Scenario 2: Request Tracing
1. Start the server and submit a query
2. Check the response headers for `X-Request-ID`
3. Search the log output for that request ID — all log entries for that request share the same ID
4. Send a request with a custom `X-Request-ID` header — verify the server uses it

### Scenario 3: Configuration
1. Run `PORT=9090 ./adr-insight serve` — verify server starts on port 9090
2. Run with no env vars — verify all defaults work (port 8081, etc.)
3. Run with `LOG_LEVEL=warn` — verify debug and info messages are suppressed
4. Check README for configuration table — all env vars documented

### Scenario 4: Health Check
1. Run `./adr-insight serve --dev`
2. `curl http://localhost:8081/health` — verify JSON response with all components healthy
3. Stop Ollama — `curl /health` shows ollama as unhealthy, overall status degraded
4. Verify response includes `X-Request-ID` header

### Scenario 5: Graceful Shutdown
1. Start the server: `./adr-insight serve --dev`
2. Send SIGTERM: `kill -TERM <pid>`
3. Verify: Server logs "shutting down", drains requests, logs shutdown duration, exits 0
4. During shutdown: new requests get 503, in-flight requests complete

### Scenario 6: Error Handling
1. Run with invalid DB path: `DB_PATH=/nonexistent/path ./adr-insight serve`
2. Verify: Clear error message about database path, exits with non-zero code
3. Check that no `log.Printf` or `log.Println` calls remain: `grep -rn 'log.Printf\|log.Println' internal/ cmd/`

### Scenario 7: Integration Tests
1. Run `go test -tags fts5 ./...`
2. Verify: Integration tests pass without Ollama or Anthropic running
3. Tests exercise: parse fixtures → store in SQLite → search → verify correct ADRs returned

### Scenario 8: API Documentation
1. Open `docs/api.md`
2. Verify: All endpoints documented (POST /query, GET /adrs, GET /adrs/{number}, GET /health)
3. Each endpoint has request/response examples and error codes
