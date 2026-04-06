# MCP Research

A structured research tool for AI assistants. Single binary — download, run, connect your AI, and start building organized knowledge bases with interviews, cross-referenced entries, tasks, and an interactive web UI.

An AI assistant (Claude, Cursor, ChatGPT, or any MCP-compatible client) designs the research structure, interviews you, writes structured entries, tracks tasks, and links everything together — all persisted in SQLite with a live web dashboard.

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

## Quick Start

Run the server with a persistent database:

```bash
./mcp-research --db research.db
```

That's it. The server is now running with:

- **MCP** over stdio (ready for local AI clients)
- **Web UI** at [http://localhost:8088](http://localhost:8088)
- **REST API** at the same port
- **LLMs.txt** documentation at [http://localhost:8088/llms.txt](http://localhost:8088/llms.txt)

## Connect Your AI

### Option 1: Local MCP Client (stdio)

MCP clients (Claude Desktop, Claude Code, Cursor) spawn the server as a local process. The AI gets 21 specialized tools and 2 research prompts.

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

With auth enabled locally, use `--default-user` to auto-create and scope to a user:

```json
{
  "mcpServers": {
    "mcp-research": {
      "command": "/path/to/mcp-research",
      "args": ["--db", "/path/to/research.db", "--auth-enabled", "--default-user", "dev@local.dev"]
    }
  }
}
```

Then ask your AI to use the `research/initialize` prompt to start a new research project.

### Option 2: Remote MCP (ChatGPT, Claude.ai)

Deploy the server with SSE transport and auth enabled. External MCP clients connect via OAuth2 with automatic setup — no manual configuration needed.

```bash
./mcp-research --transport sse --db research.db --auth-enabled --base-url "https://mcp.example.com"
```

**ChatGPT** — enter your server URL (e.g. `https://mcp.example.com/sse`) as the MCP Server URL in the Custom GPT or App settings.

**Claude.ai** — enter your server URL (e.g. `https://mcp.example.com`) in the MCP integration settings.

What happens automatically:

1. Client hits the server, gets 401 with `WWW-Authenticate` header
2. Client reads `/.well-known/oauth-protected-resource` to find the auth server
3. Client registers itself via Dynamic Client Registration (RFC 7591)
4. User sees a login page and authorizes access
5. Client receives an OAuth2 access token (with PKCE)
6. Client connects to MCP with the token — full toolset available

The server supports two MCP transports simultaneously:
- **Streamable HTTP** at `/mcp` and `/` — used by ChatGPT and Claude.ai
- **SSE** at `:8081/sse` — for legacy MCP clients

### Option 3: LLMs.txt (any AI, no MCP required)

If your AI doesn't support MCP, use the built-in documentation + REST API:

```
http://localhost:8088/llms.txt
```

This page describes the full API, data model, and research workflow in a format any LLM can understand. Feed it to any AI along with the [OpenAPI spec](http://localhost:8088/api/openapi.yaml).

To enable write access via REST API:

```bash
./mcp-research --db research.db --api-token my-secret-token
```

## Authentication

By default, auth is disabled for single-user / local usage. Enable it for multi-user or remote deployments:

```yaml
# config.yaml
auth_enabled: true
jwt_secret: "your-secret-here"
allow_registration: true
base_url: "https://mcp.example.com"
# default_user: "dev@local.dev"   # auto-create user for stdio (local dev)
```

With auth enabled:

- **Web UI** shows login/registration pages
- **Each user** sees only their own researches (enforced at service layer for all entities)
- **API keys** can be created in Settings for programmatic access
- **OAuth2** with PKCE and DCR enables external clients (ChatGPT, Claude.ai) to connect automatically
- **Default user** (`--default-user`): auto-created if not found, Web UI auto-logs in — zero-friction local dev

The first registered user automatically claims any pre-existing orphaned researches.

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
| Default User | `--default-user` | `MCP_RESEARCH_DEFAULT_USER` | — |
| Config File | `--config` | `MCP_RESEARCH_CONFIG` | `./config.yaml` |

Priority: CLI flags > env vars > config.yaml > defaults.

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

- `/` → `:8088` (Web UI, REST API, MCP Streamable HTTP, WebSocket, OAuth)
- `/sse`, `/message` → `:8081` (MCP SSE transport)

```bash
export SERVER_NAME=mcp.example.com
envsubst '${SERVER_NAME}' < deploy/nginx/mcp-research.conf > /etc/nginx/sites-enabled/mcp-research.conf
nginx -t && systemctl reload nginx
certbot --nginx -d mcp.example.com
```

## Building from Source

Requires Go 1.25+ and Node.js 20+ (for frontend).

```bash
git clone https://github.com/butschster/mcp-research.git
cd mcp-research
make build-all
./bin/mcp-research --db research.db
```

## License

MIT
