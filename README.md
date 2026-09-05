# MCP Research

**Your AI asks the questions, and the answers stop disappearing into a chat log.**

MCP Research is a self-hosted research workspace that an AI assistant drives
through the [Model Context Protocol](https://modelcontextprotocol.io). The
assistant designs the structure of a topic, interviews you, writes up what it
learns as cross-referenced entries, tracks its own todo list, draws roadmaps —
and every bit of it lands in a web UI you can browse, share and export, backed by
a single SQLite file you own, or your PostgreSQL/MySQL database.

One Go binary. No account, no cloud, no vendor. Download it, point Claude, Cursor
or ChatGPT at it, and start.

![Research overview](docs/images/research.webp)

---

## Table of contents

- [The problem](#the-problem)
- [What you get](#what-you-get)
- [How a research actually runs](#how-a-research-actually-runs)
- [Install](#install)
- [Connect your AI](#connect-your-ai)
- [Documentation for the AI: llms.txt](#documentation-for-the-ai-llmstxt)
- [Teams, roles and share links](#teams-roles-and-share-links)
- [Configuration](#configuration)
- [MCP tools](#mcp-tools)
- [Deployment](#deployment)
- [Building from source](#building-from-source)

---

## The problem

You spend an hour with an AI working out which vector store to use, how payments
flow through your company, or what actually broke last Tuesday. It is a good
hour. Then the conversation scrolls away.

Next week you ask the same assistant the same question and it has no idea what
you decided, or why. The reasoning was never written down anywhere a person — or
the next session — could find it.

**MCP Research gives the conversation somewhere to land.** The assistant is no
longer producing chat messages; it is producing a structured document with
sections, entries, sources, open questions and decisions, kept in a database, in
front of you, updating live while it works.

What that changes in practice:

| Without it | With it |
|---|---|
| Findings live in scrollback | Findings live in sections and entries with short codes (`E3`, `R2:E5`) |
| Every session starts from zero | Working instructions and memory persist per research, and the AI can ask what is still unfinished |
| "Didn't we decide this already?" | Decision entries, revision history, and who wrote what |
| Copy-paste into Notion afterwards | Markdown, PDF, an Obsidian vault, or portable JSON on demand |
| The AI's plan is invisible | A task board and roadmaps the AI maintains itself |
| Disagreeing means retyping the paragraph into a chat | You mark the sentence itself, and the AI works that queue |

---

## What you get

### An AI-run interview, recorded

Sessions are the interview loop. The assistant proposes the questions, you
answer, and it keeps asking until the section is actually covered — questions can
be answered, deferred or skipped, follow-ups branch off the ones that need it,
and progress is visible the whole way. A research can hold as many sessions as
the topic needs: scoping, deep dive, follow-up three weeks later.

![Session with questions and answers](docs/images/session.webp)

### Entries that are documents, not notes

Every finding is an entry inside a section, written in markdown — tables, code,
quotes, and **Mermaid diagrams rendered as interactive SVG** (pan, zoom,
fullscreen, reopen in mermaid.live).

Cross-references use a wiki syntax the assistant is taught to write: `[[E3]]`
links to another entry, `[[R2:E5]]` reaches into another research, `[[RM1]]`
points at a roadmap, `[[T4]]` opens that task on the board. They render as real
links everywhere — entries, answers, task results, session notes — and they build
a graph you can query.

![Entry with a table, cross-references and a Mermaid diagram](docs/images/entry.webp)

### Block documents and HTML artifacts

When prose is the wrong shape, an entry can be a **block document** instead:
typed blocks — paragraphs, callouts, tables, checklists, Mermaid, code — that an
agent can patch one block at a time without rewriting the page. A checklist keeps
the boxes a human ticked even when the agent rewrites the text around them.

Two of the blocks reach outside the document. A **task list** projects tasks that
already exist onto the page: tick one there and the task moves on the board, so
there is no second todo list drifting out of step with the first. A
**transcript** holds a call or a meeting that happened somewhere else, kept as
turns with a speaker and a timestamp rather than pasted into a code block — so a
line is searchable by who said it, carries `[[E3]]` references like any other
prose, and can be marked like any other sentence. Fourteen block types in all:
`/llms/blocks.md`.

![Block document with a callout, table and checklist](docs/images/blocks.webp)

And when the finding *is* a picture, the assistant writes an **artifact**: a
self-contained HTML document rendered in a sandboxed iframe — a chart, a
comparison grid, a small interactive view — sized to its content and printable
with the rest of the research.

![HTML artifact rendering a cost chart](docs/images/artifact.webp)

### The sentence you do not believe

You read what the assistant wrote, reach a claim you do not buy, and your options
are to describe the paragraph back into a chat window or to open a task called
"the bit about p99". Instead you **select the sentence and mark it**: `verify`
(find a source, or say plainly you could not), `dig` (this deserves its own
document), `disagree` (do not smooth this over — record both positions).

**The AI cannot make a mark and cannot close one.** It reads the queue
(`annotation_list`), does the work, and records what it did
(`annotation_answer`) — `answered` is as far as it may go. You accept a pass of
answers, or send one back with a reason. There is deliberately no tool that
creates a mark: the whole point is that a person read this specific text.

Marks are anchored by the quote, not by an offset, so they survive the agent
rewriting the document around them — and when the paragraph you doubted is gone,
the mark says so instead of quietly disappearing, with a diff from the revision
it was made against. Open marks are counted in the research header, and
`/research/R1/annotations` is the queue: grouped by document, filtered by kind
or state, with the answers reviewed and accepted a pass at a time. Full
reference: `/llms/annotations.md`.

### A todo list the AI keeps for itself

Tasks are how the assistant tracks its own work: what it has verified, what is in
flight, what is still open. Statuses, priorities, and a result written back when
a task closes.

![Task board](docs/images/tasks.webp)

### Roadmaps: deliberate graphs

A roadmap is a directed graph you design on purpose — a migration plan, a
learning path, a decision tree with branches that actually branch. Nodes carry
their own statuses, can reference an entry or a task, and show its live state.

![Roadmap of a migration plan](docs/images/roadmap.webp)

### The whole research as one map

The mindmap is generated, not drawn: every section, entry, session, question and
task in one canvas, collapsible, with cross-reference edges. There is also a
knowledge-graph view filtered by node and edge type when you want to see the
links rather than the hierarchy.

![Auto-generated mindmap of the research](docs/images/mindmap.webp)

### Export that produces a document, not a dump

One page renders the entire research — table of contents, every section and
entry, every session with its answers, every task with its result — ready to
**print to PDF**, **download as markdown**, or **download as an Obsidian vault**
(a zip of folders and linked notes, `[[E3]]` intact). Portable JSON moves a whole
research to another server, and imports back.

![Export page with the full research rendered](docs/images/export.webp)

### The first five minutes are not improvised

The quality of a research is decided before a single entry exists, by whatever
structure the model invented that day — and nothing kept it. **Twenty-five
methodologies ship with the binary**: technology comparison, competitive
landscape, incident postmortem, literature review, win/loss analysis, churn
diagnosis, raise-or-not, and more.

Crucially, a methodology does not dictate a shape. It is prose the model reads
first, and the model still designs the research — so a comparison gets its
criteria fixed before any candidate is named, and a postmortem is asked why
detection took as long as it did, without either being poured into a template's
mould. Browse them at `/templates`; a team can write its own.

### Skills the agent opens when it needs them

A research's `instruction` says what *this research* is, and it is loaded on
every call. Methodology is the wrong size for that: how to grade a source or run
a structured interview runs to a page each, a research needs one or two of them
at the moment of use, and none of it is reusable next time.

So a **skill** is a document with a trigger line. The skills a research follows
arrive as an index — a name and one line saying *when to use this* — and the
bodies stay one `skill_load` call away, read when the work actually happens.
They belong to a team and are reused across its researches, with a
research-private tier for a rule that applies to one and a built-in tier that
ships with the app.

### Picking it up next week, in a new chat

The context is gone, the scrollback is gone, and the honest first question is
"where were we?". The assistant asks the research itself: **`research_resume`**
returns the queue rather than the archive — the tasks in flight, blocked and
waiting, the open questions of the session it is continuing, the marks you left,
the documents that changed most recently, and up to three candidate next actions,
each with the reason for it and whether it is waiting on the agent or on you.

The same picture is the **Continue** block at the top of the research page, so
you can see what it is about to pick up before it does. Nothing here is a guess
and nothing here is a write: every group reports its real total beside the few it
shows, so "nothing left" is something the summary says rather than something you
infer from a short list — and a response the size cap shortened says so outright;
a document last edited by a *person* is flagged as one, because that is a
correction to build on rather than stale work to redo; and with two sessions open
it asks which one you mean instead of choosing for you. No session is started and nothing is marked as read until
somebody says so.

### Everything else that makes it a product

- **A personal Updates queue for every research** — new documents and revisions
  since you last opened each one, with `New` and `Changed` badges wherever you
  browse them. Review the exact before-and-after diff, or mark the displayed
  updates as seen in one click; another reader's queue stays their own.
- **Revision history on every entry** — who wrote it (agent, human, import,
  restore), during which session, what changed, and a diff. Restore any revision.
  A session page can show everything that session changed.
- **Live updates over WebSocket** — the UI moves while the AI works, in every
  transport mode, because it all runs in one process.
- **Search** across researches, and inside one research.
- **Read-only share links** for people without an account.
- **Teams and roles** when more than one person is involved.
- **Multiple researches**, tagged, filtered, with global short codes (`R1`, `R2`).

![Research list](docs/images/list.webp)

---

## How a research actually runs

```mermaid
flowchart TD
    A["You: 'research which vector store we should use'"] --> B["research/initialize prompt"]
    B --> B2["template_list — is there a methodology for this?<br/>template_get — read it"]
    B2 --> C["AI proposes goal + sections, following it<br/>you correct what is wrong"]
    C --> D["research_create — R1 exists, web UI live"]
    D --> E["research/conduct prompt<br/>session SS1 opens"]
    E --> F{Interview loop}
    Z["Next week, a new chat: 'continue R1'"] --> Z2["research_get — the constraints<br/>research_resume — what is still open"]
    Z2 --> F
    F -->|answer| G["question_update — answer recorded"]
    F -->|follow-up| H["question_create — new question"]
    G --> I["entry_create — finding written to a section"]
    H --> F
    I --> J["task_create / task_update — AI tracks its own work"]
    I --> P["You mark a sentence you do not believe<br/>verify · dig · disagree"]
    P --> Q["annotation_list — AI works the queue<br/>annotation_answer — answered, never closed"]
    Q --> R["You accept the pass, or send it back"]
    R --> L
    J --> K["roadmap_create — plan drawn as a graph"]
    K --> L{Covered?}
    L -->|no| F
    L -->|yes| M["Sections and research marked completed"]
    M --> N["Export: PDF · markdown · Obsidian vault · JSON<br/>or share a read-only link"]
```

In a client, it is about this short:

```
You:  Use the research/initialize prompt. Topic: which vector store for our
      support assistant.

AI:   [reads the Technology comparison methodology]
      [asks what disqualifies a candidate outright, and who signs off]
      [proposes: Constraints & disqualifiers, Criteria & weights, Candidates,
       Evidence & spikes, Head-to-head, Decision]
      Created R1. Open http://localhost:8088/research/R1

You:  Good. Conduct it.

AI:   [opens session SS1, asks 5 questions, records your answers]
      [writes E1 What retrieval has to do, E2 The shortlist...]
      [opens tasks: run the benchmark harness, price the managed tiers]
      Section "Context & constraints" is complete. Ready for the benchmark section?

You:  I marked three sentences in E2. Work them.

AI:   [annotation_list — 3 open marks in E2: two verify, one disagree]
      [A1: cites the vendor's pricing page. A2: could not confirm, asked [[Q7]]]
      [A3: records both positions in the document instead of rewriting it]
      Three marks answered. Accept them, or send any back.
```

A week later, in a chat that knows none of the above:

```
You:  Continue R1.

AI:   [research_get, then research_resume]
      Session SS2 "Migration planning". Open: 1 task in flight (benchmark
      harness), 2 questions unanswered, 1 mark you still have to accept.
      E3 changed since SS1 — you edited it, so I will read it first.
      I would start with the harness. Go?
```

Everything the assistant writes appears in the browser as it happens. You are
reading a document being built, not waiting for a summary at the end.

---

## Install

### Download a binary

Single static binary, no dependencies, no CGo.

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

Run it with a database file:

```bash
./mcp-research --db research.db
```

That is the whole setup. You now have:

| | |
|---|---|
| Web UI | [http://localhost:8088](http://localhost:8088) |
| REST API | same port, `/api/...` |
| AI documentation | [http://localhost:8088/llms.txt](http://localhost:8088/llms.txt) |
| API reference | [http://localhost:8088/api-docs](http://localhost:8088/api-docs) — every route, in the browser |
| OpenAPI spec | [http://localhost:8088/api/openapi.yaml](http://localhost:8088/api/openapi.yaml) (or `.json`) — the same thing as a file |
| MCP over stdio | ready for local clients |

With SQLite and no database path/DSN configured, everything runs in memory and
disappears on exit. Existing `--db` and `MCP_RESEARCH_DB` settings keep working.

PostgreSQL (16+) and MySQL (8.4+) are also supported through Bun. Create an empty
database and configure its connection; schema migrations run at startup:

```bash
MCP_RESEARCH_DB_DRIVER=postgres \
MCP_RESEARCH_DB_DSN='postgres://user:password@localhost:5432/research?sslmode=require' \
./mcp-research

MCP_RESEARCH_DB_DRIVER=mysql \
MCP_RESEARCH_DB_DSN='user:password@tcp(localhost:3306)/research' \
./mcp-research
```

Switching the driver selects another database; it does not transfer existing
SQLite data. See [database development and testing](docs/databases.md) for the
migration rules, test matrix and recovery behavior.

### Docker

Published image, tagged per release:

```bash
docker run -d --name mcp-research \
  -p 8088:8088 -p 8081:8081 \
  -v mcp-data:/data \
  -e MCP_RESEARCH_DB=/data/research.db \
  -e MCP_RESEARCH_TRANSPORT=sse \
  ghcr.io/butschster/mcp-research:latest
```

Or with compose and a config file:

```bash
git clone https://github.com/butschster/mcp-research.git
cd mcp-research
cp config.yaml.example config.yaml     # set base_url, jwt_secret, auth_enabled
docker compose up -d
```

`docker-compose.yaml` builds from source by default; swap `build: .` for
`image: ghcr.io/butschster/mcp-research:latest` to pull the published image
instead.

---

## Connect your AI

Three ways in, depending on what your client supports. All of them talk to the
same process and the same database.

### 1. Local MCP client over stdio

Claude Code, Claude Desktop and Cursor spawn the binary themselves. The assistant
gets **51 tools** and **2 research prompts**.

**Claude Code** — `~/.claude/mcp.json`:

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

**Claude Desktop** — `~/.claude/claude_desktop_config.json`, and **Cursor** —
`~/.cursor/mcp.json`, take the same block.

For a local setup with auth on, `--default-user` creates the user if missing and
auto-logs the web UI in, so there is no login page in your own workflow:

```json
{
  "mcpServers": {
    "mcp-research": {
      "command": "/path/to/mcp-research",
      "args": ["--db", "/path/to/research.db",
               "--auth-enabled", "--default-user", "you@local.dev"]
    }
  }
}
```

Then ask for the `research/initialize` prompt to start.

### 2. Remote MCP — ChatGPT, Claude.ai

Deploy the server with auth enabled and a public base URL. External clients set
themselves up over OAuth2 — Dynamic Client Registration (RFC 7591) and PKCE
(RFC 7636) — so there is nothing to configure by hand beyond the URL.

```bash
./mcp-research --transport sse --db research.db \
  --auth-enabled --base-url "https://mcp.example.com"
```

- **ChatGPT** — MCP server URL: `https://mcp.example.com/sse`
- **Claude.ai** — integration URL: `https://mcp.example.com`

What happens on first connect:

1. Client hits the endpoint, gets `401` with a `WWW-Authenticate` header
2. Reads `/.well-known/oauth-protected-resource` → finds the authorization server
3. Reads `/.well-known/oauth-authorization-server` → finds the endpoints
4. Registers itself via `POST /auth/register` (DCR)
5. You log in on an HTML page and authorize it
6. It exchanges the code with PKCE for an access token
7. It connects to MCP with that token — full toolset

Two remote transports run at once: **Streamable HTTP** at `/mcp` and `/` (ChatGPT,
Claude.ai) and **SSE** at `:8081/sse` for legacy clients. Browser traffic to `/`
still gets the web UI — MCP requests are detected and routed.

### 3. No MCP at all — llms.txt plus the REST API

Any assistant that can read a URL and make HTTP calls can use the product without
MCP. See below.

---

## Documentation for the AI: llms.txt

The server documents itself, for models rather than for people, at
[`/llms.txt`](http://localhost:8088/llms.txt) — the
[llms.txt convention](https://llmstxt.org/). It is not a marketing page: it is
the data model, the short-code and cross-reference syntax, the entry types, the
access rules, and an index of deeper guides that are served as plain markdown at
`/llms/<name>.md`:

| Document | What it covers |
|---|---|
| `/llms.txt` | Entry point — data model, short codes, `[[refs]]`, entry types, index |
| `/llms/mcp-client-guide.md` | All 51 tools, the input-schema contract, nullable fields, common pitfalls |
| `/llms/domain-guide.md` | Every entity in full: fields, statuses, lifecycle, the role matrix, the real-time event stream |
| `/llms/conducting-research.md` | The workflow itself — initialize, interview, write entries, complete |
| `/llms/tasks.md` | When to open a task, statuses, tasks vs questions |
| `/llms/roadmaps.md` | Node and edge types, custom statuses, building a graph step by step |
| `/llms/templates.md` | The methodologies — what a template is, the built-in set, writing your own |
| `/llms/skills.md` | Skills: the tiers, the index, `skill_load`, and how to write a trigger line |
| `/llms/blocks.md` | The block document format and all block types |
| `/llms/artifacts.md` | Writing the HTML that goes inside an artifact, and the sandbox rules |
| `/llms/metadata.md` | Typed fields a section declares and its documents carry, and the completion gate |
| `/llms/annotations.md` | Marks a person leaves on a sentence: the three kinds, the anchor states, what the agent may and may not do |
| `/llms/revisions.md` | History, diffs, restore, and what does not create a revision |
| `/llms/export.md` | Every export form and endpoint |
| `/api/openapi.yaml` | OpenAPI 3.1 spec for the REST API — every route the server registered, and which credential each one wants on this instance. `/api/openapi.json` is the same document |

Point any assistant at `https://your-server/llms.txt` and it can drive the whole
product over REST. MCP clients get the same guides — they are what the tool
descriptions link to when a model needs more than a one-line schema.

### Every route, and what it wants from you

The spec is generated from the routes the server registered, not written beside
them, so what it lists is what the server actually serves — and every route in it
says which credential it expects.

**Read it in the browser at [`/api-docs`](http://localhost:8088/api-docs)** —
grouped by area, searchable, and every route has a panel that sends a real
request to the instance you are looking at. `/api/openapi.yaml` and
`/api/openapi.json` are the same document as a file, which is what a code
generator wants. Both ship inside the binary: no account, no internet
connection, no CDN. `/docs`, `/swagger`, `/redoc` and `/openapi` land on the
same page, because those are the addresses people guess; the app itself links to
it from the footer of every page and from the API keys card in Settings.

That is one document to read before writing an integration, and there are two
credentials in it, which are not interchangeable.

**A person's bearer token.** A JWT from `POST /api/auth/login`, an API key made
in Settings, or an OAuth2 access token: any of the three works on any route, and
what you may do is decided by your role in the team that owns the research. This
is what a script, a non-MCP assistant or a remote MCP client presents to a server
with `auth_enabled`.

**The instance `api_token`.** Configured on the server, belonging to no user and
no team — it identifies whoever runs the thing.

```bash
./mcp-research --db research.db --api-token my-secret-token
```

With no accounts on the instance, that token is what turns the write API on:
reads stay open, writes need `Authorization: Bearer my-secret-token`. With
`auth_enabled`, ordinary writes want a person instead, and the operator token is
accepted only on the template routes — which is the one thing no account can do:
adding a **kickoff methodology that every team on this server sees**.

```bash
curl -X POST https://your-server/api/templates \
  -H "Authorization: Bearer my-secret-token" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "House diligence",
    "when_to_use": "Use when a supplier or partnership is assessed here and the answer goes to the committee.",
    "when_not_to_use": "Not for a technical evaluation.",
    "body": "## Before you propose anything\n\nAsk who signs it off. ...",
    "skills": ["evidence-grading"]
  }'
```

A methodology is what an AI reads *before* it starts a research: what to ask you
first, what structure to propose, when the work is finished. Twenty-five ship with the
app; this adds yours next to them. It survives every upgrade — what ships is
refreshed from the binary, what you write here is not touched.

A team can write its own too, at `POST /api/teams/{id}/templates`, and that one
stays private to the team. No role reaches the server-wide list — however senior
the account, the answer is `operator_required`; on a local run with neither
accounts nor a token there is no boundary to cross and the write just happens.
Everybody can browse the result at `/templates` in the web UI. Full reference:
`/llms/templates.md` on your own server.

---

## Teams, roles and share links

Auth is **off by default** — a single user on a laptop never meets any of this.
Turn it on for anything shared:

```yaml
# config.yaml
auth_enabled: true
jwt_secret: "a-long-random-string"
allow_registration: true
base_url: "https://mcp.example.com"
```

**Teams own researches; your role in the team is what grants access.** Everyone
gets a personal team on registration, so the concept stays invisible until you
actually collaborate.

| Role | Can |
|---|---|
| `viewer` | read and export |
| `editor` | + create and edit content |
| `owner` | + manage members, move researches between teams |

Members join by invite link. A non-member is told the research does not exist
rather than that it is forbidden — confirming a research exists is itself
information.

**Share links** hand a research to someone with no account at all: a revocable
token, optionally password-protected and time-limited, that opens a read-only
copy at `/s/{token}`. You choose what the link includes — sessions, tasks,
roadmaps, export. Working process is never shared: instructions, memory, entry
provenance, revision history, the marks people left on sentences and the
continuation summary stay behind the login.

Other auth features:

- **API keys** for long-lived programmatic access, created in Settings
- **OAuth2** with PKCE + DCR for external MCP clients
- The first registered user adopts any pre-existing researches from an
  unauthenticated database

---

## Configuration

Priority: **CLI flags > env vars > `config.yaml` > defaults.**

| Setting | CLI flag | Env var | Default |
|---|---|---|---|
| Transport | `--transport` | `MCP_RESEARCH_TRANSPORT` | `stdio` |
| MCP port (SSE) | `--mcp-port` | — | `8081` |
| Web port | `--web-port` | — | `8088` |
| Database driver | `--db-driver` | `MCP_RESEARCH_DB_DRIVER` | `sqlite` |
| Database DSN | `--db-dsn` | `MCP_RESEARCH_DB_DSN` | — |
| SQLite path | `--db` | `MCP_RESEARCH_DB` | in-memory |
| Log level | `--log-level` | `MCP_RESEARCH_LOG_LEVEL` | `info` |
| Write API token | `--api-token` | `MCP_RESEARCH_API_TOKEN` | unset — with no accounts, writes are off |
| ↳ also the **operator** credential, which is what a server-wide methodology is written with | | | |
| Auth | `--auth-enabled` | `MCP_RESEARCH_AUTH_ENABLED` | `false` |
| JWT secret | `--jwt-secret` | `MCP_RESEARCH_JWT_SECRET` | auto-generated |
| Registration | `--allow-registration` | `MCP_RESEARCH_ALLOW_REGISTRATION` | `true` |
| Public base URL | `--base-url` | `MCP_RESEARCH_BASE_URL` | — |
| Default user | `--default-user` | `MCP_RESEARCH_DEFAULT_USER` | — |
| Revision limit | `--revision-limit` | `MCP_RESEARCH_REVISION_LIMIT` | `0` — keep every revision |
| Config file | `--config` | `MCP_RESEARCH_CONFIG` | `./config.yaml` |

Data model, in short:

```
Team — owns researches; a role in it is the whole access model
  Research (R1, R2...)
    ├── Section (S1, S2...) → Entry (E1, E2...) → Revision (1, 2, 3...)
    │                         ↳ Annotation (A1, A2...) — a mark on a sentence
    ├── Session (SS1, SS2...) → Question (Q1, Q2...)
    ├── Task (T1, T2...)
    ├── Roadmap (RM1, RM2...) → Node (N1, N2...) + edges
    └── Share — a revocable read-only link
```

Short codes are assigned automatically and accepted wherever an id is:
`/api/researches/R1/sessions/SS1/export` resolves like the UUIDs do.

---

## MCP tools

51 tools, 2 prompts.

| Category | Tools |
|---|---|
| **Research** | `research_create`, `research_get`, `research_resume`, `research_list`, `research_update`, `research_add_section` |
| **Sections** | `section_list`, `section_update` |
| **Entries** | `entry_create`, `entry_read`, `entry_list`, `entry_update`, `entry_patch`, `entry_delete`, `entry_history`, `entry_diff` |
| **Sessions** | `session_create`, `session_get`, `session_update` |
| **Questions** | `question_create`, `question_update`, `question_list` |
| **Tasks** | `task_create`, `task_update`, `task_list`, `task_delete` |
| **Annotations** | `annotation_list`, `annotation_answer` |
| **Roadmaps** | `roadmap_create`, `roadmap_get`, `roadmap_list`, `roadmap_update`, `roadmap_delete`, `roadmap_add_nodes`, `roadmap_update_node`, `roadmap_remove_nodes` |
| **Templates** | `template_list`, `template_get` |
| **Skills** | `skill_list`, `skill_load`, `skill_attach`, `skill_detach`, `skill_create`, `skill_update`, `skill_fork`, `skill_copy`, `skill_promote`, `skill_delete` |
| **Teams** | `team_list` |
| **Transfer** | `research_export`, `research_import` |

**Prompts:** `research/initialize` (design a new research with you) and
`research/conduct` (run the interview and write it up).

Share links have no tool on purpose: creating or revoking one is a human act, done
in the web UI or over REST. Annotations have no create tool and no close tool for
the same reason — a mark is a person pointing at a sentence, and only a person
accepts the work done about it.

---

## Deployment

### Docker Compose

```bash
cp config.yaml.example config.yaml   # set base_url, jwt_secret, auth_enabled
docker compose up -d
```

Ports: `8088` (web UI, REST, MCP Streamable HTTP, WebSocket, OAuth) and `8081`
(MCP SSE). The SQLite database lives in the `mcp-data` volume — that file is your
entire installation, so back it up like one.

### Behind nginx

A template lives in `deploy/nginx/mcp-research.conf`:

```bash
export SERVER_NAME=mcp.example.com
envsubst '${SERVER_NAME}' < deploy/nginx/mcp-research.conf \
  > /etc/nginx/sites-enabled/mcp-research.conf
nginx -t && systemctl reload nginx
certbot --nginx -d mcp.example.com
```

It proxies `/` to `:8088` (including the WebSocket upgrade) and `/sse`,
`/message` to `:8081`. Set `base_url` to the public HTTPS URL — the OAuth
metadata documents are generated from it. The API document does not need it:
with `base_url` unset it publishes a relative `/`, which resolves against
whatever address it was fetched from, proxy and all.

---

## Building from source

Requires Go 1.25+ and Node.js 22 with npm 11 (the repository ships no lockfile,
and npm 10 fails to resolve the frontend tree — `npm install -g npm@11` first).

```bash
git clone https://github.com/butschster/mcp-research.git
cd mcp-research

make build-all     # build the Nuxt frontend, embed it, compile the binary
make run           # run against research.db
make run-sse       # SSE on :8081, REST + WebSocket on :8088
make test          # go test ./cmd/... ./internal/...
make frontend-dev  # Nuxt dev server on :3000, proxying the API to :8088
```

Architecture: one Go process serving MCP (stdio + Streamable HTTP + SSE), a REST
API, a WebSocket hub, OAuth2 endpoints, and the embedded Nuxt 4 SPA — over Bun
with SQLite (pure-Go driver), PostgreSQL, or MySQL. `CLAUDE.md` documents the
layering for contributors.

---

## License

MIT
