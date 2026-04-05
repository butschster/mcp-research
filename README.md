# MCP Research

A structured research tool for AI assistants. Single binary — download, run, connect your AI, and start building organized knowledge bases with interviews, cross-referenced entries, tasks, and an interactive web UI.

An AI assistant (Claude, Cursor, or any MCP-compatible client) designs the research structure, interviews you, writes structured entries, tracks tasks, and links everything together — all persisted in SQLite with a live web dashboard.

## Install

Download the latest binary for your OS:

```bash
# macOS (Apple Silicon)
curl -L -o mcp-research \
  https://github.com/butschster/mcp-research/releases/latest/download/mcp-research-darwin-arm64
chmod +x mcp-research

# macOS (Intel)
curl -L -o mcp-research \
  https://github.com/butschster/mcp-research/releases/latest/download/mcp-research-darwin-amd64
chmod +x mcp-research

# Linux (x86_64)
curl -L -o mcp-research \
  https://github.com/butschster/mcp-research/releases/latest/download/mcp-research-linux-amd64
chmod +x mcp-research

# Linux (ARM64)
curl -L -o mcp-research \
  https://github.com/butschster/mcp-research/releases/latest/download/mcp-research-linux-arm64
chmod +x mcp-research

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/butschster/mcp-research/releases/latest/download/mcp-research-windows-amd64.exe -OutFile mcp-research.exe
```

Verify it works:

```bash
./mcp-research --version
```

### Docker

```bash
git clone https://github.com/butschster/mcp-research.git
cd mcp-research
cp config.yaml.example config.yaml
# Edit config.yaml — set base_url, jwt_secret, etc.

docker compose up -d
```

The server will be available at `http://localhost:8088` (Web UI + API) and `http://localhost:8081/sse` (MCP SSE).

## Quick Start

Run the server with a persistent database:

```bash
./mcp-research --db research.db
```

That's it. The server is now running with:

- **MCP** over stdio (ready for AI clients)
- **Web UI** at [http://localhost:8088](http://localhost:8088)
- **REST API** at the same port
- **LLMs.txt** documentation at [http://localhost:8088/llms.txt](http://localhost:8088/llms.txt)

## Connect Your AI

### Option 1: MCP Client (recommended)

MCP clients (Claude Desktop, Claude Code, Cursor, etc.) communicate with the server over the MCP protocol. The AI gets 21 specialized tools and 2 research prompts to work with.

**Claude Code** — add to `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "mcp-research": {
      "command": "/path/to/mcp-research",
      "args": ["--db", "/path/to/research.db"]
    }
  }
}
```

**Claude Desktop** — add to `~/.claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mcp-research": {
      "command": "/path/to/mcp-research",
      "args": ["--db", "/path/to/research.db"]
    }
  }
}
```

**Cursor** — add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "mcp-research": {
      "command": "/path/to/mcp-research",
      "args": ["--db", "/path/to/research.db"]
    }
  }
}
```

Then ask your AI to use the `research/initialize` prompt to start a new research project.

**SSE mode** — if your client supports URL-based MCP connections:

```bash
./mcp-research --transport sse --mcp-port 8081 --db research.db
```

```json
{
  "mcpServers": {
    "mcp-research": {
      "url": "http://localhost:8081/sse"
    }
  }
}
```

### Option 2: LLMs.txt (any AI, no MCP required)

If your AI doesn't support MCP, you can use the built-in documentation + REST API. After starting the server, open:

```
http://localhost:8088/llms.txt
```

This page describes the full API, data model, and research workflow in a format any LLM can understand. Feed it to ChatGPT, Gemini, or any other AI along with the [OpenAPI spec](http://localhost:8088/api/openapi.yaml), and it can interact with the server through the REST API.

To enable write access via REST, start the server with an API token:

```bash
./mcp-research --db research.db --api-token my-secret-token
```

### Option 3: ChatGPT (via OAuth2)

When authentication is enabled, ChatGPT and other external clients can connect via the standard OAuth2 flow:

1. Point ChatGPT to your server's MCP SSE URL
2. It auto-discovers OAuth endpoints via `/.well-known/oauth-authorization-server`
3. It registers a client automatically via Dynamic Client Registration (RFC 7591)
4. Users see a login page and authorize access
5. ChatGPT receives an access token and can use the full MCP toolset

No manual setup required — everything is automatic.

## Authentication

By default, auth is disabled for single-user / local usage. Enable it for multi-user deployments:

```bash
./mcp-research --db research.db --auth-enabled --jwt-secret "your-secret-here"
```

Or in `config.yaml`:

```yaml
auth_enabled: true
jwt_secret: "your-secret-here"
allow_registration: true
base_url: "https://mcp.example.com"
```

With auth enabled:

- **Web UI** shows login/registration pages
- **Each user** sees only their own researches
- **API keys** can be created in Settings for programmatic access (MCP SSE, REST API)
- **OAuth2** enables external clients (ChatGPT) to connect with user authorization
- **MCP SSE** requires a bearer token (JWT or API key) via header or `?token=` query param

The first registered user automatically claims any pre-existing researches.

## How It Works

### Research Workflow

1. **Initialize** — the AI designs a research structure: sections, goals, tags
2. **Conduct** — the AI interviews you, records answers, writes structured entries
3. **Complete** — sections and research are marked done; everything is browsable in the web UI

### Data Model

```
Research (R1, R2, ...)
├── Section (S1, S2, ...) → Entry (E1, E2, ...)
├── Session (SS1, SS2, ...) → Question (Q1, Q2, ...)
└── Task (T1, T2, ...)
```

Every entity gets an auto-assigned **short code**. Entries support cross-references with `[[E3]]` (same research) or `[[R2:E5]]` (cross-research) syntax, stored and rendered as navigable links.

### Web UI

The embedded web UI at `:8088` provides:

- Research list with status and tag filters
- Research detail with sections, entries, and progress tracking
- Interactive **mindmap** view with collapsible sections and cross-reference edges
- Session view with grouped questions and progress counters
- Real-time updates via WebSocket
- PDF export

## Configuration

| Setting | CLI Flag | Env Var | Default |
|---------|----------|---------|---------|
| Transport | `--transport` | `MCP_RESEARCH_TRANSPORT` | `stdio` |
| MCP Port | `--mcp-port` | — | `8081` |
| Web Port | `--web-port` | — | `8088` |
| DB Path | `--db` | `MCP_RESEARCH_DB` | in-memory |
| Log Level | `--log-level` | `MCP_RESEARCH_LOG_LEVEL` | `info` |
| API Token | `--api-token` | `MCP_RESEARCH_API_TOKEN` | disabled |
| Auth Enabled | `--auth-enabled` | `MCP_RESEARCH_AUTH_ENABLED` | `false` |
| JWT Secret | `--jwt-secret` | `MCP_RESEARCH_JWT_SECRET` | auto-generated |
| Allow Registration | `--allow-registration` | `MCP_RESEARCH_ALLOW_REGISTRATION` | `true` |
| Base URL | `--base-url` | `MCP_RESEARCH_BASE_URL` | — |
| Config File | `--config` | `MCP_RESEARCH_CONFIG` | `./config.yaml` |

You can also use a `config.yaml` file. Priority: CLI flags > env vars > config.yaml > defaults.

## MCP Tools

| Category | Tools |
|----------|-------|
| **Research** | `research_create`, `research_get`, `research_list`, `research_update`, `research_add_section` |
| **Sections** | `section_list`, `section_update` |
| **Entries** | `entry_create`, `entry_list`, `entry_read`, `entry_update` |
| **Sessions** | `session_create`, `session_get`, `session_update` |
| **Questions** | `question_create`, `question_update`, `question_list` |
| **Tasks** | `task_create`, `task_update`, `task_list`, `task_delete` |

**MCP Prompts:** `research/initialize` (design a new research) and `research/conduct` (run an interview session).

## Deployment

### Docker Compose

```bash
cp config.yaml.example config.yaml
# Edit config.yaml
docker compose up -d
```

### Nginx Reverse Proxy

A template config is provided in `deploy/nginx/mcp-research.conf`. It proxies:

- `/` → `:8088` (Web UI, REST API, WebSocket, OAuth)
- `/sse`, `/message` → `:8081` (MCP SSE transport)

Install it:

```bash
export SERVER_NAME=mcp.example.com
envsubst '${SERVER_NAME}' < deploy/nginx/mcp-research.conf > /etc/nginx/sites-enabled/mcp-research.conf
nginx -t && systemctl reload nginx
```

Then use certbot for HTTPS:

```bash
certbot --nginx -d mcp.example.com
```

## Building from Source

Requires Go 1.25+ and Node.js 20+ (for frontend).

```bash
git clone https://github.com/butschster/mcp-research.git
cd mcp-research

# Full build: frontend + Go binary
make build-all

# Run
./bin/mcp-research --db research.db
```

## License

MIT
