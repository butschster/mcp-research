---
name: local-api-testing
description: Drive the local mcp-research REST API with curl — grab the auto-login JWT, call read and write endpoints, run an end-to-end research → session → entry → roadmap flow. Use to check an endpoint for real, reproduce an API bug, confirm ref_data resolves, or look up a response shape.
---

# Testing the local HTTP API

Start the service with `make run-sse`: SSE on :8081, REST + WebSocket on :8088.

## Step 1 — get a token

When `auth_enabled: true` and `default_user` is set, the service hands out an
auto-login token:

```bash
TOKEN=$(curl -s http://localhost:8088/api/auth/info \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('auto_login_token',''))")
```

With auth disabled, drop the `Authorization` header everywhere.

## Step 2 — make requests

```bash
curl -s http://localhost:8088/api/researches -H "Authorization: Bearer $TOKEN"

curl -s -X POST http://localhost:8088/api/researches \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "description": "...", "goal": "..."}'
```

When `auth_enabled: true`, read endpoints need the token too.

## Step 3 — end-to-end flow

1. `POST /api/researches` — create a research
2. `GET /api/researches/{id}` — read back the sections and their ids
3. Create entities: `POST /api/entries`, `POST /api/tasks`, `POST /api/sessions`
4. `POST /api/roadmaps` with `ref_type` / `ref_id` on nodes
5. `GET /api/roadmaps/{id}` — check `ref_data` was populated (lazy sync)
6. `PUT /api/tasks/{id}` — change the source entity's status
7. `GET /api/roadmaps/{id}` again — `ref_data` should reflect the new status

A fresh research has no sections unless you pass them at creation or add one with
`POST /api/researches/{id}/sections`.

## Response shapes — where the id lives

The key differs per endpoint, which is a recurring source of confusion:

| Create | id in response |
|---|---|
| Research | `data.research_id` |
| Entry | `data.entry_id` |
| Session | `data.id` |
| Task | `data.id` |

Short codes work in place of ids on scoped routes:
`GET /api/researches/R1/sessions/SS1/export` resolves the same as full UUIDs.

`GET /api/roadmaps/{id}` returns full `ref_data` for nodes carrying
`ref_type` / `ref_id` — resolved at read time, not stored.

## The full route list

Do not work from memory or from the spec: routes are registered in
`internal/api/server.go`, and the spec is served at `GET /api/openapi.yaml`
(source: `internal/api/openapi.go`).

The spec lags behind `server.go` — several routes are missing from it. When
changing routes, check the code.
