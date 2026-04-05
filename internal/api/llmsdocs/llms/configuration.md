# Configuration

MCP Research is configured via config file, environment variables, and CLI flags. Priority: CLI flags > env vars > config.yaml > defaults.

## config.yaml

```yaml
transport: stdio       # stdio or sse
mcp_port: 8081         # MCP SSE port (only when transport=sse)
web_port: 8088         # REST API + Web UI port
db: ""                 # SQLite path (empty = in-memory, data lost on restart)
log_level: info        # debug, info, warn, error
api_token: ""          # Bearer token for write API (empty = write API disabled)
```

Place `config.yaml` in the working directory or specify path with `--config /path/to/config.yaml`.

## CLI Flags

```
--config       Path to config.yaml (default: ./config.yaml)
--transport    stdio or sse (default: stdio)
--mcp-port     MCP SSE port (default: 8081)
--web-port     Web/API port (default: 8088)
--db           SQLite database path (default: in-memory)
--log-level    Log level (default: info)
--api-token    Bearer token for write API (default: disabled)
--version      Print version and exit
```

## Environment Variables

```
MCP_RESEARCH_CONFIG       Path to config.yaml
MCP_RESEARCH_TRANSPORT    stdio or sse
MCP_RESEARCH_DB           SQLite database path
MCP_RESEARCH_LOG_LEVEL    debug, info, warn, error
MCP_RESEARCH_API_TOKEN    Bearer token for write API
```

## Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 8088 | HTTP | REST API, WebSocket, embedded Web UI |
| 8081 | HTTP | MCP SSE transport (only when transport=sse) |

## Database

SQLite with pure Go driver (no CGo). Two modes:

- **In-memory** (default): fast, data lost on restart. Good for testing.
- **File-based**: `--db research.db`. Persistent. WAL mode enabled automatically.

## Transport Modes

- **stdio** (default): MCP communication over stdin/stdout. Used when the MCP client spawns the process directly.
- **sse**: MCP communication over HTTP Server-Sent Events on `--mcp-port`. Used when the server runs independently and clients connect via URL.

Both modes always start the REST API + Web UI on `--web-port`.
