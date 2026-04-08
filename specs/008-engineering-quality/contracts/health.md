# Contract: Health Check Endpoint

## `GET /health`

### Request
No body. No parameters.

### Response (200 OK — healthy or degraded)
```json
{
  "status": "healthy",
  "components": {
    "database": { "status": "healthy" },
    "ollama": { "status": "healthy" },
    "anthropic_key": { "status": "healthy" }
  }
}
```

### Response (503 Service Unavailable — unhealthy)
```json
{
  "status": "unhealthy",
  "components": {
    "database": { "status": "unhealthy", "error": "connection refused" },
    "ollama": { "status": "healthy" },
    "anthropic_key": { "status": "healthy" }
  }
}
```

### Response Headers
- `Content-Type: application/json`
- `X-Request-ID: <uuid>`

### Status Logic
- **healthy**: All components report healthy
- **degraded**: Non-critical components (ollama, anthropic_key) unhealthy, database healthy
- **unhealthy**: Database is unreachable

### Timeout
- Overall endpoint: 5 seconds
- Per-component check: 3 seconds
