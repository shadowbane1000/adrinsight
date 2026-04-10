# Quickstart: Abuse Hardening

## Verification Scenarios

### Scenario 1: Rate Limiting
1. Start the server: `ANTHROPIC_API_KEY=sk-... ./adr-insight serve --dev`
2. Submit a query via the UI or curl
3. Rapidly submit 10+ more queries within 1 minute
4. Verify: After the limit is reached, responses return HTTP 429 with `Retry-After` header
5. Verify: The UI shows a friendly "too many requests" message with retry timing
6. Wait for the window to expire, submit another query — it should succeed

### Scenario 2: Different IPs Not Affected
1. From one client, exhaust the rate limit
2. From a different IP (or using a different `X-Forwarded-For` value), submit a query
3. Verify: The second client's query succeeds

### Scenario 3: Query Length Limit
1. Submit a normal-length query — succeeds
2. Submit a query exceeding 500 characters (or configured limit)
3. Verify: Returns HTTP 400 with "query too long" error
4. Verify: No AI services were called (check logs — no embedding or synthesis entries)

### Scenario 4: Configuration
1. Start with `RATE_LIMIT_REQUESTS=2 RATE_LIMIT_WINDOW_S=60 MAX_QUERY_LENGTH=100`
2. Verify: Rate limit triggers after 2 queries
3. Verify: Queries over 100 characters are rejected

### Scenario 5: Non-Query Endpoints Unaffected
1. Exhaust the rate limit with queries
2. Verify: `GET /adrs`, `GET /adrs/1`, `GET /health`, and page loads still work
