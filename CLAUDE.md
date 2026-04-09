# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build Go binary only (uses existing embedded frontend)
make build

# Full build: compile Nuxt frontend + embed into Go binary
make build-all

# Run with persistent SQLite database
make run                    # uses research.db

# Run with SSE transport (recommended for development with Web UI)
make run-sse                # SSE on :8081, REST+WebSocket on :8088

# Run tests (storage, service, tools layers)
make test                   # or: go test ./...

# Run specific test suites
go test ./internal/storage/ -v -run Code       # short code tests
go test ./internal/service/ -v -run CrossRef   # crossref parsing tests
go test ./internal/service/ -v -run Resolve    # code resolution tests
go test ./internal/service/ -v -run AccessControl  # user isolation tests

# Frontend development (hot-reload, separate from Go server)
make frontend-dev           # Nuxt dev on :3000, proxies API to :8088

# Rebuild embedded frontend assets
make frontend-embed
```

**Config path:** `./mcp-research --config /path/to/config.yaml`

## Testing Local HTTP API

When the service is running locally (e.g. `make run-sse` on port 8088), use curl to test the REST API.

### Step 1: Get auth token

If `auth_enabled` is true with a `default_user`, fetch the auto-login token:

```bash
# Get auto-login JWT from the auth info endpoint
curl -s http://localhost:8088/api/auth/info | python3 -m json.tool
# Response includes "auto_login_token": "eyJ..."

# Save it for subsequent requests
TOKEN=$(curl -s http://localhost:8088/api/auth/info | python3 -c "import sys,json; print(json.load(sys.stdin).get('auto_login_token',''))")
```

If auth is disabled, skip the `Authorization` header in all requests.

### Step 2: Use the token in requests

```bash
# Read endpoints
curl -s http://localhost:8088/api/researches -H "Authorization: Bearer $TOKEN"
curl -s http://localhost:8088/api/roadmaps/{id} -H "Authorization: Bearer $TOKEN"

# Write endpoints (also need the token when auth is enabled)
curl -s -X POST http://localhost:8088/api/researches \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "description": "...", "goal": "..."}'
```

### Step 3: Typical test flow

1. **Create research** — `POST /api/researches` (returns `research_id` in `data.research.id` or `data.research_id`)
2. **Get research** — `GET /api/researches/{id}` (returns sections array with their IDs)
3. **Create entities** — entries (`POST /api/entries`), tasks (`POST /api/tasks`), sessions (`POST /api/sessions`)
4. **Create roadmap with refs** — `POST /api/roadmaps` with `ref_type`/`ref_id` on nodes
5. **GET roadmap** — `GET /api/roadmaps/{id}` verifies `ref_data` is populated (lazy sync)
6. **Update source entity** — e.g. `PUT /api/tasks/{id}` to change status
7. **GET roadmap again** — `ref_data` should reflect the updated status

### Response format notes

- Research create returns: `data.research_id` (not `data.id`)
- Entry create returns: `data.entry_id` (not `data.id`)
- Session create returns: `data.id` (nested under `data`)
- Task create returns: `data.id`
- Roadmap GET returns full `ref_data` for nodes with `ref_type`/`ref_id` (resolved at read time)

## Architecture

Single Go binary serving multiple protocols from one process:

```
MCP Client (Claude/Cursor)
    | stdio
    |
ChatGPT / Claude.ai
    | Streamable HTTP (:8088/mcp) via OAuth2
    |
Legacy MCP Client
    | SSE (:8081)
    |
Go Process
    |-- MCP Server (21 tools, 2 prompts)
    |-- REST API (:8088) -- read-only + write (bearer auth)
    |-- WebSocket (:8088/ws) -- real-time event push
    |-- OAuth2 endpoints (/auth/*)
    +-- Embedded Nuxt SPA (static files at /)
         |
      SQLite (file or in-memory)
```

### MCP Transports

Three MCP transports run simultaneously:

| Transport | Port | Path | Use case |
|-----------|------|------|----------|
| **stdio** | — | — | Local MCP clients (Claude Code, Cursor, Claude Desktop) |
| **Streamable HTTP** | 8088 | `/mcp` or `/` | Remote clients (ChatGPT, Claude.ai) via OAuth2 |
| **SSE** | 8081 | `/sse` | Legacy remote MCP clients |

The catch-all at `/` detects MCP traffic (POST with JSON body or MCP headers) and routes it to the Streamable HTTP handler. Browser GET requests get the Nuxt frontend.

### Layered Structure

```
cmd/mcp-research/main.go          -- bootstrap, DI wiring
internal/
  auth/                            -- JWT, bcrypt, API key hashing, context helpers, middleware
  config/                          -- YAML / env / CLI flag cascade
  domain/                          -- plain structs + status enums (no logic)
  storage/                         -- SQLite repos + embedded migrations (001-010)
  service/                         -- business logic, validation, event emission, access control
  mcp/                             -- MCP server wrapper, tool + prompt registration
    tools/                         -- 21 tool files (one per tool)
    prompts_data/                  -- embedded markdown prompt templates
  api/                             -- REST handlers, WebSocket hub, OAuth, static embedding
    ws/                            -- WebSocket hub + client + event notifier bridge
    handlers/                      -- HTTP endpoint handlers (read + write + auth + oauth)
    static/                        -- embedded Nuxt build output (generated by make frontend-embed)
frontend/                          -- Nuxt 4 SPA (separate npm project)
  composables/useAuth.ts           -- auth state, login/register/logout, auto-login
  middleware/auth.global.ts        -- route guard (redirect to /login when auth enabled)
  pages/login.vue, register.vue, settings.vue -- auth UI
  components/mindmap/              -- Vue Flow custom node components for mindmap view
deploy/
  nginx/                           -- nginx reverse proxy template
  prod/                            -- production config + docker-compose
```

### Data Model (8 entities)

**User** (multi-user auth, optional)
**Research** -> **Section** -> **Entry** (content hierarchy, scoped to user)
**Research** -> **Session** -> **Question** (interview workflow, multiple sessions per research)
**Research** -> **Task** (AI self-managed todo list)
**Entry** -> **CrossRef** (cross-references extracted from `[[...]]` patterns)
**Entry** -?> **Session** (optional link: `session_id` tracks which session produced the entry)

A research can have **multiple sessions** (e.g. initial exploration, deep-dive, follow-up). Each session has its own questions. Questions and answers may contain `[[...]]` cross-references just like entries. The frontend renders these references as clickable links everywhere: question text, answers, task titles/results, session notes.

### User Scoping & Access Control

When `auth_enabled` is true:
- `Research.UserID` links each research to its owner
- `ResearchService.Get()` checks ownership — returns `ErrNotFound` for cross-user access (no information leak)
- `ResearchService.List()` automatically filters by `auth.UserIDFromContext(ctx)`
- All downstream services (Entry, Section, Session, Task) call `validateResearchAccess()` before any operation
- Access control is tested in `internal/service/access_control_test.go` (24 tests covering all entities)

### Short Codes

- Researches get global codes: `R1`, `R2`, `R3` (auto-assigned on creation)
- Entries get codes scoped to research: `E1`, `E2`, `E3`
- Sessions get codes scoped to research: `SS1`, `SS2`, `SS3`
- Questions get codes scoped to session: `Q1`, `Q2`, `Q3`
- Cross-reference syntax: `[[E3]]` (same research), `[[R2:E5]]` (cross-research), `[[R2]]` (research link)
- Cross-references work in entry content, question text/answers, task results, session notes
- Cross-references are extracted and stored in `crossrefs` table on entry create/update
- `POST /api/researches/{id}/crossrefs/rebuild` re-scans all entries to fix stale references
- Frontend renders `[[...]]` as clickable links via `renderRefs()` composable (auto-imported by Nuxt)

### Event System

Services emit events after mutations via `EventNotifier` interface -> WebSocket Hub broadcasts to browser clients. This works regardless of MCP transport (stdio, SSE, or Streamable HTTP) because everything runs in one process.

## Configuration

Priority: CLI flags > env vars > config.yaml > defaults.

| Setting | CLI Flag | Env Var | YAML Key | Default |
|---------|----------|---------|----------|---------|
| Transport | `--transport` | `MCP_RESEARCH_TRANSPORT` | `transport` | `stdio` |
| MCP Port | `--mcp-port` | - | `mcp_port` | `8081` |
| Web Port | `--web-port` | - | `web_port` | `8088` |
| DB Path | `--db` | `MCP_RESEARCH_DB` | `db` | (in-memory) |
| Log Level | `--log-level` | `MCP_RESEARCH_LOG_LEVEL` | `log_level` | `info` |
| API Token | `--api-token` | `MCP_RESEARCH_API_TOKEN` | `api_token` | (write API disabled) |
| Auth Enabled | `--auth-enabled` | `MCP_RESEARCH_AUTH_ENABLED` | `auth_enabled` | `false` |
| JWT Secret | `--jwt-secret` | `MCP_RESEARCH_JWT_SECRET` | `jwt_secret` | (auto-generated) |
| Allow Registration | `--allow-registration` | `MCP_RESEARCH_ALLOW_REGISTRATION` | `allow_registration` | `true` |
| Base URL | `--base-url` | `MCP_RESEARCH_BASE_URL` | `base_url` | — |
| Default User | `--default-user` | `MCP_RESEARCH_DEFAULT_USER` | `default_user` | — |

## Authentication

When `auth_enabled` is true, multi-user auth is active:
- Users register/login via `/api/auth/register` and `/api/auth/login` (JWT)
- API keys for long-lived programmatic access (MCP SSE, REST API)
- OAuth2 Authorization Code flow with PKCE (RFC 7636) and DCR (RFC 7591) for external clients
- Researches are scoped to users — each user sees only their own
- Web UI has login/registration pages
- MCP SSE requires bearer token (JWT or API key) via header or `?token=` query param
- First registered user automatically claims any pre-existing orphaned researches
- `default_user`: auto-creates user if not found, auto-logs in Web UI (for local dev)

### Auth Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/auth/register` | Register new user |
| `POST` | `/api/auth/login` | Login, get JWT |
| `GET` | `/api/auth/me` | Current user info |
| `POST` | `/api/auth/api-keys` | Create API key |
| `GET` | `/api/auth/api-keys` | List API keys |
| `DELETE` | `/api/auth/api-keys/{id}` | Revoke API key |

### OAuth2 Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/.well-known/oauth-authorization-server` | OAuth2 server metadata (RFC 8414) |
| `GET` | `/.well-known/oauth-protected-resource` | Protected resource metadata (RFC 9728) |
| `POST` | `/auth/register` | Dynamic Client Registration (RFC 7591) |
| `GET/POST` | `/auth/authorize` | Authorization endpoint (HTML login page) |
| `POST` | `/auth/token` | Token exchange (accepts form + JSON) |

### OAuth2 Flow (ChatGPT / Claude.ai)

1. Client hits MCP endpoint → 401 with `WWW-Authenticate: Bearer resource_metadata="..."`
2. Client reads `/.well-known/oauth-protected-resource` → finds authorization server
3. Client reads `/.well-known/oauth-authorization-server` → finds endpoints
4. Client registers via `POST /auth/register` (DCR) → gets `client_id` + `client_secret`
5. Client redirects user to `GET /auth/authorize?...&code_challenge=...` (PKCE)
6. User logs in on HTML form → redirect back with authorization code
7. Client exchanges code at `POST /auth/token` with `code_verifier` → access token
8. Client connects to MCP endpoint with `Authorization: Bearer <token>`

## Write API

When `api_token` is configured, write endpoints are enabled with bearer token authentication.
All write endpoints require `Authorization: Bearer <token>` header.
Read-only endpoints remain unauthenticated (unless `auth_enabled`).

### Write Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/researches` | Create research + sections |
| `PUT` | `/api/researches/{id}` | Update research |
| `POST` | `/api/researches/{id}/sections` | Add section |
| `PUT` | `/api/sections/{sectionId}` | Update section |
| `POST` | `/api/entries` | Create entry |
| `PUT` | `/api/entries/{id}` | Update entry (supports text_replace) |
| `POST` | `/api/tasks` | Create task |
| `PUT` | `/api/tasks/{id}` | Update task |
| `DELETE` | `/api/tasks/{id}` | Delete task |
| `POST` | `/api/sessions` | Create session + questions |
| `POST` | `/api/researches/{id}/crossrefs/rebuild` | Rebuild cross-references |

### Read-only Endpoints (no auth)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/researches` | List all researches |
| `GET` | `/api/researches/{id}` | Get research + sections + active session |
| `GET` | `/api/researches/{id}/sections/{sectionId}/entries` | List entries in section |
| `GET` | `/api/entries/{id}` | Get entry with content |
| `GET` | `/api/researches/{id}/entries/by-code/{code}` | Resolve entry by short code |
| `GET` | `/api/resolve/research/{code}` | Resolve research by short code |
| `GET` | `/api/researches/{id}/crossrefs` | List cross-references |
| `GET` | `/api/researches/{id}/tasks` | List tasks |
| `GET` | `/api/researches/{id}/sessions` | List sessions |
| `GET` | `/api/researches/{id}/sessions/{sessionId}` | Get session + questions + progress (research-scoped) |
| `GET` | `/api/researches/{id}/roadmaps` | List roadmaps for research |
| `GET` | `/api/researches/{id}/roadmaps/{roadmapId}` | Get roadmap with nodes/edges (research-scoped, with ref_data) |
| `GET` | `/api/roadmaps/{id}` | Get roadmap (standalone, legacy) |
| `GET` | `/api/health` | Health check |

## Key Patterns

**MCP Tool registration** uses the Go SDK generic `mcp.AddTool`:
```go
mcp.AddTool(srv, &mcp.Tool{Name: "tool_name", Description: "..."},
    func(ctx, req, input InputStruct) (*mcp.CallToolResult, any, error) { ... })
```
- `jsonschema` struct tags are plain descriptions (not `required,description=`)
- Tools always return `(result, nil, nil)` -- never return Go errors (that's a protocol error)
- Use `successResult()` / `errorResult()` / `validationErrorResult()` helpers from `tools/helpers.go`

**SQLite:** Pure Go driver (`modernc.org/sqlite`), no CGo. `MaxOpenConns(1)`, WAL mode for file DBs, `foreign_keys=ON`.

**Frontend embedding:** `//go:embed all:static` (the `all:` prefix is required to include `_nuxt/` directory). MIME types registered manually in `embed.go` init().

**Auth context flow:** `auth.WithUser(ctx, user)` / `auth.UserFromContext(ctx)` propagates user identity through all layers. In stdio mode, `--default-user` injects user into the MCP context. In SSE/HTTP mode, middleware extracts from Bearer token.

**Access control:** `validateResearchAccess(ctx, repo, researchID)` helper in `service/errors.go` checks existence + ownership. Used by all services. Returns `ErrNotFound` (not 403) to prevent information leaks.

## Adding a New MCP Tool

1. Create `internal/mcp/tools/my_tool.go` with input struct + `RegisterMyTool()` function
2. Register in `internal/mcp/tools.go` -> `registerTools()`
3. If new entity: add domain struct, repo, service, migration, API handler, and wire in `main.go`
4. Ensure service methods call `validateResearchAccess()` for user scoping

## Adding a New REST Endpoint

1. Add handler method in `internal/api/handlers/`
2. Register route in `internal/api/server.go` using Go 1.22 pattern: `mux.HandleFunc("GET /api/path/{id}", handler)`
3. For write endpoints: wrap with `auth()` middleware and use `mux.Handle()` instead of `mux.HandleFunc()`
4. WebSocket endpoint uses `/ws` without method prefix (required for upgrade)

## Frontend

- SPA mode (`ssr: false`) for static embedding
- `useApi()` composable wraps `useFetch` with configurable base URL + auth token injection
- `useAuth()` composable manages auth state, login/register/logout, auto-login via `auto_login_token`
- `useRealtimeUpdates()` composable manages WebSocket connection with auto-reconnect
- `useResearchMindmap()` composable fetches all research data and builds dagre graph
- `middleware/auth.global.ts` — route guard, redirects to `/login` when auth enabled
- When `NUXT_PUBLIC_API_BASE` is empty, API calls use relative URLs (same-origin with Go server)
- `make frontend-dev` sets `NUXT_PUBLIC_API_BASE=http://localhost:8088` for cross-origin dev
- Vue Flow (`@vue-flow/core`) used for interactive mindmap visualization
- Custom node components in `components/mindmap/` (RootNode, SectionNode, EntryNode, etc.)

## Deployment

- `Dockerfile` — multi-stage: Node (frontend) → Go (binary) → Alpine (runtime)
- `docker-compose.yaml` — SSE transport, persistent SQLite volume, healthcheck
- `deploy/nginx/mcp-research.conf` — reverse proxy template with `${SERVER_NAME}`
- `deploy/prod/` — production-ready config, docker-compose, nginx for specific domain
