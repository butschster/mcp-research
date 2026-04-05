# MCP Research

A self-contained MCP server for structured AI-driven research sessions. Built in Go with an embedded Nuxt 4 web UI.

The server enables Claude (or any MCP client) to autonomously design research structures, conduct interactive Q&A sessions, manage tasks, and track progress — all persisted in SQLite.

## Quick Start

```bash
# Build (Go binary only)
make build

# Build with embedded frontend (single binary)
make build-all

# Run with persistent storage
./bin/mcp-research --db research.db

# Open http://localhost:8088 for the web UI
```

## Configuration

Configuration is loaded from `config.yaml` → environment variables → CLI flags (highest priority).

### config.yaml

```yaml
transport: stdio    # stdio or sse
mcp_port: 8081      # MCP SSE port (only used when transport=sse)
web_port: 8088      # REST API + Web UI port
db: ""              # SQLite path (empty = in-memory)
log_level: info     # debug, info, warn, error
api_token: ""       # Bearer token for write API (empty = write API disabled)
```

### CLI Flags

```
--transport    stdio or sse (default: stdio)
--mcp-port     MCP SSE port (default: 8081)
--web-port     Web/API port (default: 8088)
--db           SQLite database path (default: in-memory)
--log-level    Log level (default: info)
--api-token    Bearer token for write API (default: disabled)
--version      Print version and exit
```

### Environment Variables

```
MCP_RESEARCH_TRANSPORT
MCP_RESEARCH_DB
MCP_RESEARCH_LOG_LEVEL
MCP_RESEARCH_API_TOKEN
MCP_RESEARCH_CONFIG    # path to config.yaml (default: ./config.yaml)
```

## MCP Client Setup

### Claude Desktop

Add to `~/.claude/claude_desktop_config.json`:

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

### Cursor

Add to `~/.cursor/mcp.json`:

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

### SSE Mode

Run the server first, then connect via URL:

```bash
./bin/mcp-research --transport sse --mcp-port 8081 --db research.db
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

## MCP Tools (21)

### Research

| Tool | Description |
|------|-------------|
| `research_create` | Create a research project with initial sections |
| `research_get` | Get full context: research + sections + entry counts + active session |
| `research_list` | List research projects with optional status filter |
| `research_update` | Update research fields, memory, instructions |
| `research_add_section` | Add a section to existing research |

### Sections

| Tool | Description |
|------|-------------|
| `section_list` | List sections with entry counts |
| `section_update` | Update section status/properties |

### Entries

| Tool | Description |
|------|-------------|
| `entry_create` | Create entry with auto-generated title/description |
| `entry_list` | List entries (without content, for token efficiency) |
| `entry_read` | Read single entry with full markdown content |
| `entry_update` | Update entry, supports `text_replace` for surgical edits |

### Sessions & Questions

| Tool | Description |
|------|-------------|
| `session_create` | Create session with initial questions (atomic) |
| `session_get` | Get session with grouped questions + progress |
| `session_update` | Update session notes (supports append) |
| `question_create` | Batch create questions with parent/child support |
| `question_update` | Update question status and answer |
| `question_list` | List questions with status/area/priority filters |

### Tasks (Todo List)

| Tool | Description |
|------|-------------|
| `task_create` | Create a task in the research todo list |
| `task_update` | Update task status, priority, result |
| `task_list` | List tasks with filters + progress counters |
| `task_delete` | Remove a task |

Task statuses: `pending`, `in_progress`, `blocked`, `completed`, `failed`, `deferred`

## MCP Prompts (2)

| Prompt | Description |
|--------|-------------|
| `research/initialize` | Interactive workflow to design and create a new research project |
| `research/conduct` | Systematic research execution: interview user, create entries, track progress |

## Architecture

```
Claude/Cursor ←→ stdio/SSE ←→ MCP Server ←→ SQLite
                                    ↕
                              REST API + Web UI (:8088)
```

### Data Model

- **Research** — top-level project with goal, instructions, memory, tags
- **Section** — grouping within research (slug-based naming)
- **Entry** — markdown content within a section (auto-title, auto-description, short code `E1`)
- **Session** — Q&A interview session with focus area
- **Question** — individual question with priority, status, parent/child nesting (max 3 levels)
- **Task** — self-managed todo item for AI planning
- **CrossRef** — cross-references between entries, extracted from `[[E3]]` / `[[R2:E5]]` patterns in content

### Layers

```
internal/
├── domain/     # Data structs and status constants
├── storage/    # SQLite repositories + migrations
├── service/    # Business logic and validation
├── mcp/        # MCP server, tools, prompts
└── api/        # REST API + embedded frontend
```

## Write API

When `api_token` is configured, write endpoints mirror all MCP tools via REST:

```bash
# Enable write API
./bin/mcp-research --db research.db --api-token my-secret-token

# Create a research
curl -X POST http://localhost:8088/api/researches \
  -H "Authorization: Bearer my-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Research","goal":"Investigate X","sections":[{"name":"overview","display_name":"Overview"}]}'

# Create an entry
curl -X POST http://localhost:8088/api/entries \
  -H "Authorization: Bearer my-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"research_id":"...","section_id":"...","content":"# Title\n\nContent here. See [[E1]] for details."}'
```

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/researches` | Create research + sections |
| `PUT` | `/api/researches/{id}` | Update research |
| `POST` | `/api/researches/{id}/sections` | Add section |
| `PUT` | `/api/sections/{sectionId}` | Update section |
| `POST` | `/api/entries` | Create entry |
| `PUT` | `/api/entries/{id}` | Update entry (supports `text_replace`) |
| `POST` | `/api/tasks` | Create task |
| `PUT` | `/api/tasks/{id}` | Update task |
| `DELETE` | `/api/tasks/{id}` | Delete task |
| `POST` | `/api/sessions` | Create session + questions |
| `POST` | `/api/researches/{id}/crossrefs/rebuild` | Rebuild cross-references |

### Cross-References

Entries auto-receive short codes (`E1`, `E2`, ...) and researches get global codes (`R1`, `R2`, ...). Use `[[E3]]` syntax in entry content to link entries within the same research, or `[[R2:E5]]` for cross-research links. References are stored in the database on every entry create/update. Use the rebuild endpoint to re-scan all entries when needed.

## Web UI

The Nuxt 4 frontend is embedded into the Go binary. It provides:

- Research list with status and tag filters
- Research detail with section sidebar, entry cards, tags panel
- Entry detail with rendered markdown
- Session detail with Questions + Tasks tabs and progress tracking
- Interactive mindmap view (Vue Flow) with collapsible sections
- PDF export via `window.print()`

### Development Mode

Run the frontend separately with hot-reload:

```bash
# Terminal 1: Go server
make run

# Terminal 2: Nuxt dev server (proxies API to Go server)
make frontend-dev
```

## Makefile

```
make build           # Build Go binary only
make build-all       # Build frontend + embed + Go binary
make run             # Build and run with file DB
make run-memory      # Build and run with in-memory DB
make run-sse         # Build and run with SSE transport
make test            # Run Go tests
make clean           # Remove build artifacts
make frontend-install  # Install npm dependencies
make frontend-dev    # Run Nuxt dev server
make frontend-build  # Build Nuxt for static output
make frontend-embed  # Build + copy to Go embed directory
```

## Dependencies

- **Go 1.25+**
- **Node.js 20+** (for frontend build only)
- **SQLite** via `modernc.org/sqlite` (pure Go, no CGo)
- **MCP SDK** via `github.com/modelcontextprotocol/go-sdk`

No CGo required. Cross-compilation works out of the box with `CGO_ENABLED=0`.
