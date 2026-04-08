# Data Model: Engineering Quality

No new database tables or persistent storage changes. This milestone introduces Go structs for configuration and health check responses.

## Config Struct (`internal/config/config.go`)

| Field | Type | Env Var | Default | Description |
|-------|------|---------|---------|-------------|
| Port | int | `PORT` | 8081 | HTTP server port |
| DBPath | string | `DB_PATH` | `./adr-insight.db` | SQLite database file path |
| ADRDir | string | `ADR_DIR` | `./docs/adr` | ADR markdown files directory |
| OllamaURL | string | `OLLAMA_URL` | `http://localhost:11434` | Ollama API base URL |
| AnthropicKey | string | `ANTHROPIC_API_KEY` | (none) | Anthropic API key |
| EmbedModel | string | `EMBED_MODEL` | `nomic-embed-text` | Ollama embedding model name |
| LogLevel | string | `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| LogFormat | string | `LOG_FORMAT` | `json` | Log output format: json, text |
| SlowRequestThreshold | duration | `SLOW_REQUEST_MS` | `2000` (ms) | Threshold for slow request warnings |
| ShutdownTimeout | duration | `SHUTDOWN_TIMEOUT_S` | `10` (seconds) | Graceful shutdown drain timeout |

## Health Response

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

### Status Values

| Status | Meaning | HTTP Code |
|--------|---------|-----------|
| healthy | All components operational | 200 |
| degraded | Non-critical component unreachable (e.g., Ollama down but DB works) | 200 |
| unhealthy | Critical component unreachable (e.g., DB inaccessible) | 503 |

### Component Checks

| Component | Check | Timeout | Critical? |
|-----------|-------|---------|-----------|
| database | `SELECT 1` on SQLite connection | 3s | Yes |
| ollama | HTTP HEAD to OllamaURL | 3s | No |
| anthropic_key | Check env var is non-empty | instant | No |
