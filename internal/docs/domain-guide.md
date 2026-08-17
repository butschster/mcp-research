# Domain Guide

Complete reference for every entity in the Research system.

## Entities

### Research

Top-level container for an investigation project. Owned by a [team](#team) when authentication is enabled — your role in that team is what decides whether you may read it, write to it, or move it.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Project name |
| `goal` | string | Single declarative sentence — success criteria |
| `description` | string | What this research covers |
| `instruction` | string | Working methodology — tone, depth, rules for entries |
| `memory` | string[] | Key insights persisted across sessions |
| `tags` | string[] | Categorization tags |
| `status` | enum | `active` / `completed` / `archived` |
| `code` | string | Auto-assigned: `R1`, `R2`... (global scope) |
| `team_id` | string | The owning team — the only thing permissions consult |
| `user_id` | string | Who created it. A record, never a permission |
| `team_name` | string | Read-only, resolved per request |
| `team_is_personal` | bool | Read-only: true when the owning team is a single user's own |
| `role` | enum | Read-only: **your** role on this research — `viewer` / `editor` / `owner` |

**Key rules:**
- One research per topic. Don't mix unrelated investigations.
- `instruction` governs all future sessions — set it during initialization. It says what **this research** is; how a **kind of work** is done belongs in a [skill](#skill), which other researches can follow too.
- `memory` survives across sessions. Use `add_memory` to append, not replace.
- `goal` is a success criterion, not a question. "Identify top 3 competitive threats" not "What are the threats?"
- A new research lands in the creator's personal team. `research_create` has no team parameter; `POST /api/researches` takes an optional `team_id`, and `research_import` / `POST /api/researches/import?team=` take one too.
- `team_name`, `team_is_personal` and `role` are computed per request and returned by `research_get`, `GET /api/researches/{id}` and `GET /api/researches`. They are not stored and are ignored on input.
- Moving a research between teams is `POST /api/researches/{id}/transfer` — there is no MCP tool for it, and it is the only way a research changes audience.

---

### Team

Who may do what. A team owns researches; membership in that team is the whole of the access model, and every user gets a personal team when they register — a solo user never has to think about any of this.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | The creator's own name for a personal team, otherwise whatever was typed |
| `personal` | bool | Created alongside its user. Cannot be renamed away from its owner, deleted, or left |
| `role` | enum | Read-only: the caller's role in this team |
| `member_count` | int | Read-only summary |
| `research_count` | int | Read-only summary |

**Roles** (`viewer` → `editor` → `owner`, each adding to the one before):

| Role | May |
|------|-----|
| `viewer` | Read everything in the team's researches — entries, sessions, questions, tasks, roadmaps, revisions — and export them |
| `editor` | Everything above, plus create and change entries, sections, sessions, questions, tasks and roadmaps |
| `owner` | Everything above, plus manage members and invitations, rename or delete the team, and transfer a research into or out of it |

**What a failure looks like:**

- **Not a member of the owning team** → the research does not exist for you: `not found` from a tool, `404` from REST. Confirming that someone else's research exists is itself a leak, so this is deliberate.
- **A member without the right** — a `viewer` writing — → `your role in this team does not allow this`, `403` from REST. Hiding a research from someone who can already read it would protect nothing.
- With `auth_enabled: false` there is no caller and no check: everything lives in one local team and every operation is permitted, exactly as before teams existed.
- The WebSocket at `/ws` needs the same credential as everything else when auth is on, and delivery is decided **per event, per connection**: a research event reaches only those who may read it, a team event only its members, and losing access stops the updates on a socket already open. See [Real-time Events](#real-time-events) below.
- That local team survives if auth is later turned on. It has no members, so its researches are **readable by every signed-in user and writable by none** until the first registration claims them — a deliberate compromise between stranding them behind a team nobody can join and letting the first caller take them for themselves.

**A [share link](#share) is the one way to read a research without a role.** It grants no membership and names no person: it is a capability over one research, resolved to `viewer` on that research and nothing else, and every write made through one is refused before any role is consulted.

**Cross-references resolve for the reader, not the author.** `[[R2:E5]]` is stored resolved whatever the writer may see, and every read path strips the targets its reader cannot follow — so a reference into a colleague's research works for whoever is entitled to it, and reads back as plain text for whoever is not.

**Invitations are links, never emails.** The server sends nothing; an owner creates an invitation and hands the link over themselves.

- `POST /api/teams/{id}/invites` returns the token **once** — it is stored hashed and cannot be read back.
- The invitation carries the role it was created with and expires after 14 days.
- `GET /api/invites/{token}` previews it without consuming it and works without a session (`status`: `pending` / `expired` / `revoked` / `accepted` / `unknown`); signed in, it also reports whether you are already a member and whether the address matches.
- `POST /api/invites/{token}/accept` joins the signed-in user. An address mismatch is allowed — the token is the capability — and accepting a second time reports that you are already a member.
- `DELETE /api/invites/{id}` revokes an unused one. A revoked link then previews as `revoked` and nothing else — no team, no sender — because withdrawing an invitation that went to the wrong person should withdraw what it disclosed.
- A personal team accepts no invitations: it refuses every removal, so a member could never be taken out again.
- `POST /api/auth/register` accepts an `invite_token`. A valid one gets the account created **even where `allow_registration` is false** and joins the team in the same request: the invitation is the authorization. It follows that a leaked link is a way onto a closed server, for one person, until it is used or expires.

**REST routes** (the OpenAPI spec does not cover teams — `internal/api/server.go` is the source of truth):

| Method | Path | Needs |
|--------|------|-------|
| `GET` | `/api/teams` | a signed-in user — returns your teams with your role and counts |
| `GET` | `/api/teams/{id}` | membership |
| `POST` | `/api/teams` | a signed-in user |
| `PUT` | `/api/teams/{id}` | owner; refused on a personal team |
| `DELETE` | `/api/teams/{id}` | owner; refused while it still holds researches, and on a personal team |
| `GET` | `/api/teams/{id}/members` | membership |
| `PUT` | `/api/teams/{id}/members/{userId}` | owner; cannot demote the last owner |
| `DELETE` | `/api/teams/{id}/members/{userId}` | owner, or yourself to leave; cannot remove the last owner |
| `GET` | `/api/teams/{id}/invites` | owner — open invitations, expired ones included |
| `POST` | `/api/teams/{id}/invites` | owner — returns the link once |
| `DELETE` | `/api/invites/{id}` | owner of that team |
| `GET` | `/api/invites/{token}` | nobody — public preview |
| `POST` | `/api/invites/{token}/accept` | a signed-in user |
| `POST` | `/api/researches/{id}/transfer` | owner of the source team, write access in the target |

Every `/api/teams` and `/api/invites` route except the public preview needs a session: with `auth_enabled: false` they answer `401 sign in to manage teams`, because there are no users to put in a team. The transfer route is the exception — with no caller there is nothing to check, and it moves the research.

`GET /api/researches` lists every research across all your teams and takes `?team={id}` to narrow it to one, and each item carries `team_id`, `team_name`, `team_is_personal` and `role`.

---

### Share

A revocable, read-only capability over **one** research, addressed by an unguessable token rather than by who is holding it. Not an account, not an invitation, not a role: nobody signs in, the link names no user, and a visitor holding one is never a member of anything.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Identifies the link — for the manage list and for revocation. Not the token |
| `research_id` | string | The single research it reaches. Everything else is out of scope by construction |
| `scope` | enum | `research` — the only value issued. `session` / `entry` / `roadmap` exist in the schema for narrower forms and are refused today |
| `label` | string | Free text for the owner: "Client review, March" |
| `include` | object | `sessions`, `tasks`, `roadmaps`, `export` — four booleans, described below |
| `has_password` | bool | Whether a password is set. The hash is returned to nobody |
| `expires_at` | string | `null` for a link that never expires on its own |
| `revoked_at` | string | Once set, the link answers exactly as if it never existed |
| `last_seen_at` | string | When it was last opened |
| `view_count` | int | Counted once per page load, not once per fetch |
| `created_by_name` | string | Read-only, resolved per request — who handed the link out |
| `research_code` | string | Read-only, so a client can build the URL it already knows |

**No short code.** A share is never referenced from content, so there is no `[[…]]` form for one; it is addressed by its token and by `id`.

**The token** is 256 random bits behind an `mrs_` prefix, stored as SHA-256 alongside API keys and team invites. It appears in exactly one response — the one that created it — and no route can read it back. A leaked database hands out no working links.

**Include flags.** The zero value is the safe one: a flag has to be set to reveal something, so nothing leaks by being forgotten. `roadmaps` defaults to true at creation because roadmaps are part of the research proper; `sessions`, `tasks` and `export` default to false, being working state rather than a result.

**Owner routes.** Creating, listing and revoking all need **write** access — an editor or an owner. A viewer cannot republish what the team owns; an editor could already export the whole research to a file and send it, so a link is not a capability they lacked.

| Method | Path | Returns |
|--------|------|---------|
| `POST` | `/api/researches/{id}/shares` | `201` with `share`, `token` (the only time it exists) and `url` — the page a visitor opens, `https://<host>/s/<token>` |
| `GET` | `/api/researches/{id}/shares` | every share of that research, live and dead, metadata only |
| `DELETE` | `/api/shares/{id}` | revokes it |

The create body takes `label`, `include` (any subset of the four booleans), `expires_in_days` (omit for a link that never expires; clamped to 1–3650) and `password` (optional, at least 6 characters). `GET /api/researches/{id}` carries `active_share_count` beside the research — `0` for anyone who could not manage the links anyway, a share visitor reading through one of them included.

**Visitor routes** live under their own prefix, `/api/shared/{token}/…`, behind their own middleware. The token is checked once, at the prefix; there is no path by which it reaches a route built for an owner. Inside the prefix the paths are the ones the authenticated API uses, so `/api/shared/{token}/researches/R1/entries` is served by the ordinary entries handler.

| Under `/api/shared/{token}` | Method | Needs |
|---|---|---|
| (the prefix itself) | `GET` | — the share payload: what this link is, the research, its sections |
| `/unlock` | `POST` | — exchanges a password for the unlock value |
| `/researches/{id}`, `/researches/{id}/entries`, `/researches/{id}/sections/{sectionId}/entries`, `/researches/{id}/tags` | `GET` | — |
| `/researches/{id}/entries/{entryId}`, `/researches/{id}/entries/by-code/{code}`, `/entries/{id}`, `/entries/{id}/related` | `GET` | — |
| `/researches/{id}/crossrefs`, `/entries/{id}/crossrefs`, `/researches/{id}/links`, `/entries/{id}/links` | `GET` | — |
| `/researches/{id}/roadmaps`, `/researches/{id}/roadmaps/{roadmapId}`, `/roadmaps/{id}` | `GET` | `include.roadmaps` |
| `/researches/{id}/sessions`, `/researches/{id}/sessions/{sessionId}` | `GET` | `include.sessions` |
| `/researches/{id}/tasks` | `GET` | `include.tasks` |
| `/researches/{id}/export` | `GET` | `include.export` |

That list is the whole surface. Anything else under the prefix — another method, another path, a route whose flag is off — answers the same `404 this link is no longer available` that a revoked, expired or invented token gets. A link without sessions is meant to look like a research that has none.

**What a visitor never sees:**

- `instruction` and `memory` — stripped from the research on every read path, the export included. They are the agent's working notes about how to conduct the research, not a result, and their author did not publish them by sending a link to the findings.
- `user_id`, `team_id`, `team_name`, `team_is_personal` — a share is about one research, not about the organisation behind it. `role` survives and is always `viewer`.
- Any other research. There is no list route under the prefix, and the listing service itself answers empty for a share rather than falling through to "no user in context, so no filter" — which would have returned every research on the server.
- The Obsidian vault and the portable JSON. The vault builds its payload from the repository rather than from the redacted read path, so it is refused outright; the portable route is not mounted.
- Revision history, the knowledge graph, the mindmap and search — none of those routes exist under the prefix.
- Every write, without exception. `Access.Write` refuses a share context before it looks at any role, so this does not depend on the `viewer` it resolves to.

**Cross-references** out of the shared research (`[[R2:E5]]`) render as inert text — not a link, not a 404, because a share must not confirm that R2 exists. An *incoming* reference from a research the visitor cannot open is dropped from the list rather than blanked: the stripped row would still announce that something unseen cites this entry.

**Password.** Optional, bcrypt. `POST /api/shared/{token}/unlock` returns an `unlock` value, sent back on every later request as `X-Share-Unlock` — or as `?unlock=…` where a header cannot be set, which is the WebSocket and a download the browser navigates to. A locked link answers `401` with `reason: password_required`; a wrong password answers `401` with `reason: invalid_password`, the one place the distinction is made, since whoever is standing at the prompt already knows the link is real. The unlock value is derived from the share, not stored: it dies with the link and stops working the moment the password changes. It is not a session, and there is nobody to have one.

**Rate limits.** Reads under the prefix: 600 per minute per address — one page load is many fetches. Unlock: 10 per minute counted against the link itself as well as the address, because spreading guesses across addresses must not spread the budget. Both answer `429` with `Retry-After: 60`.

**Revocation** takes effect on the next request; every layer consults the share per request. The exception is an open WebSocket, which re-resolves the link on its own timer and closes within the minute — see [Real-time Events](#real-time-events).

**No MCP tool creates, lists or revokes a share.** Handing out a public link is a human act; the tool list is unchanged. A share token is a REST credential only — it never reaches an MCP endpoint. Shares also work with `auth_enabled: false`, where the row simply records no creator.

---

### Section

Logical division within a research. Organizes entries by topic.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Slug identifier (e.g., `market-analysis`) |
| `display_name` | string | Human-readable name |
| `description` | string | One sentence scope definition |
| `position` | int | Sort order (0-based) — investigation sequence |
| `status` | enum | `draft` / `active` / `completed` / `archived` |
| `code` | string | Auto-assigned: `S1`, `S2`... (per research) |

**Key rules:**
- 3-7 sections per research. Non-overlapping scopes.
- Order by investigation logic (context → current state → gaps → recommendations), not alphabetically.
- Requires at least one entry before marking `completed`.

---

### Entry

A document containing research findings. Lives in a section. Holds either markdown
or, when `entry_type` is `blocks`, a JSON document of typed blocks — prose next to
tables, callouts, checklists, mermaid diagrams and self-contained HTML rendered in
a sandboxed iframe. See [Block Documents](/llms/blocks.md). `entry_type: artifact`
is an input alias for a block document holding one `html` block; it is never
stored, and such an entry reads back as `blocks`. A `code`, `mermaid` or `html`
body is not searched as text and contributes no cross-references.

| Field | Type | Description |
|-------|------|-------------|
| `entry_type` | enum | `markdown` (default) / `blocks` (`artifact` accepted on input, stored as `blocks`) |
| `title` | string | Entry title (auto-generated from content if omitted; for a block document from its first heading, and the call fails when there is nothing to take one from) |
| `content` | string | Markdown, a block document, or a full HTML document for the `artifact` alias |
| `description` | string | Short summary (auto-generated if omitted) |
| `section_id` | string | Parent section |
| `session_id` | string | Optional: session that produced this entry |
| `tags` | string[] | Categorization tags for filtering |
| `status` | enum | `draft` / `active` / `completed` / `archived` |
| `code` | string | Auto-assigned: `E1`, `E2`... (per research) |

**Key rules:**
- One entry per topic. Synthesize multiple answers into one coherent document.
- Use `[[E3]]` syntax in content to cross-reference other entries.
- Tags enable filtering on the research page. Use consistent taxonomy.
- Title/description auto-generated from content if not provided.
- `session_id` tracks provenance. When you leave it empty, the server links the entry to the research's active session automatically; the session export (`/research/{code}/session/{sessionCode}/export`) lists entries by this link.
- Every write that changes an entry appends a [revision](#revision); `entry_history` says who wrote each one and `entry_diff` what it changed. Read them before rewriting an entry a previous session produced.
- Deleting an entry (`entry_delete`) also deletes its cross-references, its extracted external links, and its whole revision history.
- A blocks entry is edited whole with `entry_update` or block by block with `entry_patch`; `text_replace` is refused on it.
- URLs found in entry content are extracted into an external-links index, readable at `GET /api/entries/{id}/links` and `GET /api/researches/{id}/links`.

**Cross-reference syntax in content:**
- `[[E3]]` — entry E3 in same research
- `[[R2:E5]]` — entry E5 in research R2
- `[[R2]]` — link to research R2 itself

---

### Session

Interactive Q&A interview workflow. How knowledge enters the system.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Session title |
| `focus` | string | Specific topic for this session |
| `notes` | string | Accumulated observations and pivots |
| `status` | enum | `active` / `completed` / `archived` |
| `code` | string | Auto-assigned: `SS1`, `SS2`... (per research) |

**Key rules:**
- **One active session at a time** per research. Complete before starting another.
- **Multiple sessions are normal** — each focuses on different aspects (initial exploration, deep-dive, follow-up).
- Use `add_note` to log decisions and pivots during the session. Notes support markdown and `[[...]]` cross-references.
- Create entries when enough material accumulates from Q&A. Set `session_id` on entries to track provenance.

**Typical session progression:**
1. Create session with 3-8 initial questions
2. Work through questions one at a time
3. Record answers, create follow-ups as needed
4. When enough answers accumulate → create entry in appropriate section
5. Mark session `completed` when all questions addressed

---

### Question

Structured Q&A prompt within a session.

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | The question |
| `area` | string | Topic area (matches section topics) |
| `rationale` | string | Why this question matters |
| `priority` | enum | `high` / `medium` / `low` |
| `status` | enum | `pending` / `in_progress` / `answered` / `deferred` / `skipped` |
| `answer` | string | Captured response (required when `answered`) |
| `parent_id` | string | Optional: parent question for follow-ups |
| `code` | string | Auto-assigned: `Q1`, `Q2`... (per session) |

**Statuses:**
- `pending` — not yet asked
- `in_progress` — currently being discussed
- `answered` — answer recorded (non-empty `answer` required)
- `deferred` — postponed, still valid
- `skipped` — deliberately skipped, no longer relevant

**Key rules:**
- Nesting max 3 levels deep (parent → child → grandchild).
- One topic per question. Don't combine multiple asks.
- `area` enables filtering by section focus.
- Question text, answers, and rationale all support cross-references (`[[E3]]`).
- Answers support full markdown formatting.

---

### Task

Todo item for tracking work within a research.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | What needs to be done |
| `description` | string | Task details |
| `priority` | enum | `high` / `medium` / `low` |
| `status` | enum | `pending` / `in_progress` / `blocked` / `completed` / `failed` / `deferred` |
| `result` | string | Outcome (set when completing) |
| `code` | string | Auto-assigned: `T1`, `T2`... (per research) |

**Tasks vs Questions:** Questions capture knowledge through interview ("What is X?"). Tasks track work items ("Do X"). Use questions for knowledge extraction, tasks for action tracking.

**Key rules:**
- Always fill `result` when completing a task.
- Results support markdown and cross-references.
- Use `blocked` status with reason in description when waiting on something.

---

### Roadmap

Visual graph for learning paths, strategy maps, decision trees, or step-by-step guides. Unlike the auto-generated mindmap, roadmaps are deliberately designed by the AI or user.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Roadmap title |
| `description` | string | What this roadmap visualizes |
| `statuses` | string[] | Available node statuses in order (e.g. `["not_started", "in_progress", "completed"]`) |
| `status` | enum | `active` / `completed` / `archived` |
| `code` | string | Auto-assigned: `RM1`, `RM2`... (per research) |

**Key rules:**
- A research can have multiple roadmaps for different aspects.
- `statuses` defines the vocabulary for node progress. Choose statuses that fit the domain:
  - Learning path: `not_started`, `learning`, `practiced`, `mastered`
  - Marketing strategy: `planned`, `approved`, `launched`
  - Technical plan: `todo`, `in_progress`, `review`, `done`
- Leave `statuses` empty for purely structural graphs with no progress tracking.
- Create the entire graph in one `roadmap_create` call when possible — nodes and edges together.

**When to create a roadmap:**
- The research topic has a natural sequence or progression (learning path, onboarding flow, migration plan)
- You need to visualize dependencies or decision points (architecture decisions, strategy branches)
- The user wants a step-by-step guide with trackable progress
- Information needs spatial/hierarchical organization beyond flat sections

**When NOT to use a roadmap (use sections/entries instead):**
- Content is purely textual with no inherent sequence
- A simple list or table would suffice
- The information doesn't have meaningful relationships between items

---

### Roadmap Node

A node in a roadmap graph. Represents a step, milestone, decision point, or informational block.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Node title |
| `description` | string | Detailed text content (expandable in UI) |
| `node_type` | string | One of the nine types below (default `step`) |
| `status` | string | Current status (from parent roadmap's `statuses` list, or empty) |
| `position_x` | float | X position for layout |
| `position_y` | float | Y position for layout |
| `parent_id` | string | Optional: parent node for hierarchical nesting |
| `ref_type` | string | Optional: `entry` / `task` / `session` / `research` / `question` |
| `ref_id` | string | Optional: ID of the referenced entity |
| `ref_data` | object | Read-only: referenced entity resolved at read time, never stored |
| `metadata` | string | Optional: JSON blob for node-type-specific data (checklist items, URL, metric value) |
| `code` | string | Auto-assigned: `N1`, `N2`... (per roadmap) |

**Node types:**
- `step` — Regular action item or learning step (default)
- `milestone` — Key achievement or checkpoint
- `decision` — Fork in the path where a choice is needed
- `info` — Reference material or prerequisite note
- `group` — Container for related steps (visual grouping)
- `checklist` — Sub-items with checkboxes (`metadata` holds the items)
- `note` — Free-form sticky-note annotation
- `link` — External URL reference (`metadata` holds the URL)
- `metric` — KPI or numeric indicator (`metadata` holds the value)

**Key rules:**
- Use `temp_id` during creation to reference nodes in edges before they have real IDs.
- Status is free-form but should match the roadmap's `statuses` list for consistent UI rendering.
- Position coordinates are optional — the frontend can auto-layout, but explicit positions give more control.

---

### Roadmap Edge

A directed connection between two roadmap nodes.

| Field | Type | Description |
|-------|------|-------------|
| `source_node_id` | string | Source node UUID (the create tools take it as `source`, accepting a `temp_id`) |
| `target_node_id` | string | Target node UUID (the create tools take it as `target`, accepting a `temp_id`) |
| `label` | string | Edge label (e.g. "next", "if yes", "alternative") |
| `edge_type` | string | `default` / `success` / `warning` / `optional` |

**Edge types:**
- `default` — Normal progression
- `success` — Positive outcome path (e.g. "if passed")
- `warning` — Failure or risk path (e.g. "if blocked")
- `optional` — Non-required alternative path

**Key rules:**
- Edges are deleted automatically when their source or target node is removed (cascade).
- Use labels to clarify the relationship, especially for `decision` nodes with multiple outgoing edges.

---

### Revision

Every write that changes an entry appends a snapshot: content, title,
description, `entry_type`, status and tags as they stood after that write, plus
who wrote it (`agent`, `human`, `import`, `restore`), the session that was active
at the time, and a short summary of what changed.

| Field | Type | Description |
|-------|------|-------------|
| `revision` | int | 1-based per entry, never reused. A number, not a short code |
| `author_kind` | enum | `agent` / `human` / `import` / `restore` |
| `session_id` | string | Session **active when the write happened** — not the entry's own `session_id` |
| `summary` | string | "Updated content, tags", "Patched blocks: inserted 2", "Restored revision 1" |
| `content` | string | The entry's body at that point. Omitted from list responses, present when one revision is read |

**Not created by:** a write that changes nothing, or a checkbox tick — ticking a
box is not an edit to the document.

**Restoring** an earlier revision writes a new one holding its content, so
history is append-only and a restore is itself undoable. Checkbox state and
block ids survive a restore.

Read with `entry_history` and `entry_diff`; `GET /api/sessions/{id}/changes`
rolls it up per session. Not to be confused with a blocks entry's `rev`, a
content hash for optimistic concurrency. Full reference:
[Revisions](/llms/revisions.md).

---

### CrossRef

Links between documents, extracted automatically from `[[...]]` patterns.

| Syntax | Meaning |
|--------|---------|
| `[[E3]]` | Entry E3 in same research |
| `[[R2:E5]]` | Entry E5 in research R2 |
| `[[R2]]` | Research R2 itself |
| `[[RM1]]` | Roadmap RM1 in same research |
| `[[RM1:N3]]` | Node N3 in roadmap RM1 |

**Where they render:** Anywhere the web UI shows markdown — entry content, question text/rationale/answers, task titles/descriptions/results, session notes.

**Where they are indexed** (stored in the `crossrefs` table, and therefore visible to the graph views and the crossref API):

| Source | Indexed text | When |
|--------|--------------|------|
| Entry | `content` | create, update, rebuild |
| Question | `answer` | `question_update` |
| Task | `description` + `result` | `task_update` (not on create) |

**Resolution:** References are resolved when the target exists. Unresolved references are tracked and can be resolved later via `POST /api/researches/{id}/crossrefs/rebuild`, which re-scans entry content. Roadmap references (`[[RM1]]`, `[[RM1:N3]]`) are resolved against the roadmaps and roadmap_nodes tables.

**Visualization:** Shown on entry detail pages (outgoing/incoming), in the mindmap and the knowledge graph view (dashed / crossref edges), and preserved in export.

---

### Skill

A methodology document — how to run an interview, how to grade a source, how to build a roadmap — that a research *follows* and an agent opens at the moment it needs it. `instruction` says what **this research** is; a skill says how a **kind of work** is done, and the same skill can be followed by many researches without being copied.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | The address for management. A slug is not one: a built-in and a team's fork of it share a slug |
| `slug` | string | Derived from `name`. How a skill is addressed **inside a research**, where the resolution order is defined |
| `name` | string | Required |
| `tier` | enum | `private` / `team` / `builtin` — where it lives, and its precedence |
| `description` | string | The trigger line: **when** to load it, never what it contains. Max 200 characters (runes). The one field always in the agent's context |
| `body` | string | Markdown, max 16000 characters (runes), required. **Omitted from every list**; carried only by `skill_load` and `GET /api/researches/{id}/skills/{slug}` |
| `team_id` / `research_id` | string | The ownership pair: exactly one is set, or neither for a built-in |
| `user_id` | string | Who wrote it. A record, never a permission — as with a research |
| `ambient` | bool | A product skill: never counted against the cap, never detachable. Set from the shipped file at boot, not through the API |
| `forked_from` | string | The built-in slug this row copies, when it is a fork or a copy |
| `needs_trigger` | bool | The description was generated rather than written by a person; cleared on the first real edit |
| `body_tokens` | int | Estimate — characters over four. Derived at read time, never stored |
| `version` | int | |
| `attached` | bool | Library listings only: whether this research already follows it |
| `via_template` / `attached_at` | bool / string | Attached listings only. They describe the attachment, not the skill |

**Tiers, and what wins.** `private` (one research) over `team` over `builtin`. That order is both the sort of the index and the precedence: where two skills conflict, follow the higher one. Every `skill_load` response restates it in a `precedence` field.

**Attachment** is a row in `research_skills`, not a field on the skill. A research may follow **six** chosen skills — `ambient` ones are outside that budget, and writing a research-private skill spends it, because such a skill is attached on creation.

**No short code.** A skill is never referenced from content, so there is no `[[…]]` form for one.

**One MCP tool, `skill_load`.** The index — slug, name, tier, description, no bodies — rides in `research_get`, and only when the research follows at least one skill. Everything else (attach, detach, write, fork, copy, promote) is REST or the web UI: eleven routes under `/api/researches/{id}/skills…`, `/api/teams/{id}/skills` and `/api/skills/{skillId}`.

**A share link never exposes a skill** — not the index, not a body. Which methodology a team follows is working process, the same class as `instruction` and `memory`. There are no skills routes under `/api/shared/{token}/`, `SkillService.Load` refuses a share context before it resolves the slug, and the index is empty for one — so a route added later still fails closed.

Full reference, including the route table and the conflict codes: [Skills](/llms/skills.md).

---

## Workflow Summary

```
1. research_create → Research + Sections
2. research_update → Set instruction + seed memory
3. session_create → Session + initial Questions
4. question_update → Record answers
5. question_create → Follow-up questions (if needed)
6. entry_create → Synthesize answers into entries with [[refs]]
7. task_create → Track remaining work
8. roadmap_create → Build visual graphs for step-by-step guides, learning paths, or decision trees
9. research_update → Add to memory (insights for future sessions)
10. Repeat 3-9 for new sessions on uncovered areas
11. Mark sections completed → research completed
```

## Short Codes

| Entity | Pattern | Scope | Used in URLs |
|--------|---------|-------|-------------|
| Research | `R1`, `R2` | Global | `/research/R2` |
| Section | `S1`, `S2` | Per research | query param |
| Entry | `E1`, `E2` | Per research | `/research/R2/entry/E3` |
| Session | `SS1`, `SS2` | Per research | `/research/R2/session/SS1` |
| Question | `Q1`, `Q2` | Per session | `/research/R2/session/SS1/question/Q3` |
| Task | `T1`, `T2` | Per research | — |
| Roadmap | `RM1`, `RM2` | Per research | `/research/R2/roadmap/RM1` |
| Node | `N1`, `N2` | Per roadmap | — |

A revision has no short code: it is a plain number, 1-based per entry. A [share](#share) has none either — it is addressed by its token and never referenced from content. Nor does a [skill](#skill): inside a research it is addressed by slug, and everywhere else by id.

## Real-time Events

`GET /ws` on the REST port (`:8088` by default) is a WebSocket that pushes one JSON message per mutation, whatever caused it — a browser write, a REST write, or an MCP tool call over stdio, SSE or Streamable HTTP, since all of them run in one process. An event is a notification, not a payload: it names what changed, and the client re-reads it over REST. There is nothing to send in the other direction: the server discards whatever a client writes (512-byte read limit) and expects only pong and close frames.

### Connecting

- With `auth_enabled` the handshake needs the same credential as every other endpoint — a JWT, an API key or an OAuth token — as `Authorization: Bearer …`, or as `?token=…` because a browser cannot set headers on a WebSocket handshake. Missing or unresolvable: `401`, no upgrade. With auth off no credential is asked for.
- The credential is re-checked on the live connection roughly once a minute, on the keepalive. If it stops resolving to the same user — key deleted, token expired — the socket is closed with code **4401**, in the application range on purpose: `4401` means authenticate again, an ordinary drop means reconnect, and a client that cannot tell them apart retries the first forever.
- A [share link](#share) connects with `?share={token}` — plus `?unlock=…` for a password-protected one — and no user credential of any kind. It is checked before the user branch and in both modes, because with auth off falling through would register the visitor as an ordinary local client and hand them the whole server's stream. An unresolvable token is `401`, no upgrade. The link is re-resolved on the same keepalive timer, so revoking it or letting it expire closes that socket with **4401** within the minute.
- `Origin` is checked whether or not auth is on. It must equal the request `Host`, be loopback (`localhost`, `127.0.0.1`, `::1`), or match the configured `base_url`; anything else is refused at the handshake (`403`). A client that sends no `Origin` — curl, a script, an MCP client — is allowed through: the check exists to stop a page the user happened to visit from opening a socket to their server, which with auth off would hand over the whole stream.

### Envelope

| Field | Type | Present |
|-------|------|---------|
| `type` | string | always — `entity.verb`, listed below |
| `entity` | string | always — `research`, `section`, `entry`, `session`, `question`, `task`, `roadmap`, `team`, `crossref`, `share`, `skill` |
| `entity_id` | string | always — the id of the thing that changed, not of its parent |
| `research_id` | string | always present, empty for team-scoped events |
| `research_code` | string | when the id resolves — the same scope as a short code (`R7`), so a page routed as `/research/R7` can match an event without resolving a UUID first |
| `actor_user_id` | string | who caused it; absent with auth off and for anything an agent did over stdio |
| `actor_client_id` | string | which tab caused it, when the writer sent `X-Client-Id` |
| `reason` | string | `access.*` only |
| `name` | string | `access.revoked` only — the name of the team or research that was lost |
| `at` | int | server send time, milliseconds since the epoch. Use it instead of receipt time: a tab waking from sleep takes the whole queued backlog at once and would date every event to the moment it woke |

`target_user_id` is not on the wire. Directed events exist (below), but the addressee learns nothing from being named that the delivery did not already tell them.

### Types

| `type` | `entity` | `entity_id` | Notes |
|--------|----------|-------------|-------|
| `research.created`, `research.updated` | `research` | research | |
| `research.transferred` | `research` | research | the research now belongs to another team |
| `section.created`, `section.updated` | `section` | section | |
| `entry.created`, `entry.updated`, `entry.deleted` | `entry` | entry | `entry.updated` also covers `entry_patch` and a revision restore |
| `crossrefs.rebuilt` | `crossref` | research | `POST /api/researches/{id}/crossrefs/rebuild` finished. Every link in the research may have moved, so the graph and mindmap views re-read wholesale rather than patching |
| `session.created`, `session.updated` | `session` | session | |
| `question.created` | `question` | question | **one event per question**, so a batch of twelve sends twelve. It used to fire once per batch carrying the *session* id; a client written against that will now see the session id nowhere |
| `question.updated` | `question` | question | an answer or a status change |
| `task.created`, `task.updated`, `task.deleted` | `task` | task | |
| `roadmap.created`, `roadmap.updated`, `roadmap.deleted` | `roadmap` | roadmap | adding, changing or removing nodes and edges all report as `roadmap.updated` on the roadmap |
| `skill.attached`, `skill.detached` | `skill` | skill | this research started or stopped following a skill. The index `research_get` returns has changed |
| `skill.created` | `skill` | skill | a research-private skill was written, or a team/built-in one was copied into this research. A **team** skill being written emits nothing — it belongs to no research yet |
| `skill.updated` | `skill` | skill | the body or trigger line changed. A fork and a promotion report as this too, on the **new** row — its `entity_id` is not the one you were following. Editing a team skill directly emits nothing, because it names no single research |
| `team.created`, `team.updated`, `team.deleted`, `team.invited`, `team.invite_revoked`, `team.member_added`, `team.member_removed`, `team.member_role_changed` | `team` | team | no `research_id` |
| `access.changed` | `team` | team | directed. `reason: role_changed`. The new role is deliberately not in the event — read it back from the API rather than trusting a value pushed at you |
| `access.revoked` | `team` or `research` | team or research | directed. `reason: removed_from_team` (entity `team`) or `research_transferred` (entity `research`, with `research_id`) |
| `share.created`, `share.revoked` | `share` | share | a read-only link was handed out or taken back. Delivered by the ordinary research rule, so any member who may read the research learns a link exists — and never to a share visitor, who has no use for the news and, in the revoked case, is being disconnected by it |

There is no delete event for a research, section, session or question: none of them can be deleted. Only entries, tasks, roadmaps and teams can. A share is revoked rather than deleted, which is why `share.revoked` and not `share.deleted`. A skill can be deleted (`DELETE /api/skills/{skillId}`) but emits nothing when it is: the row is addressed by id and the researches that were following it are not known to that call — re-read the attached list rather than waiting for a message.

### Who receives what

Delivery is decided per event per connection, at send time — not once when the connection opens, which is what would let a revoked membership keep receiving updates until the tab was closed.

- With `auth_enabled: false` there are no users and one local team: every connection receives everything.
- With auth on, a research-scoped event goes to the users who may read that research — the same rule a REST read applies — and a team event to that team's members. An unidentified connection receives nothing.
- A **directed** event (`access.revoked`, `access.changed`) is addressed to one user and skips the research check, which is the whole point: by the time somebody is told they lost access, the ordinary rule already refuses to deliver it. Directed events therefore only occur with auth on.
- `access.revoked` carries `name` and `reason` because its recipient can no longer look either up — the moment it is sent, fetching what they lost answers 404.
- A **share** connection is scoped by the link, not by membership: events for its one research, filtered by the same `include` flags that gate its read routes — an event about a task on a link that excludes tasks is not harmless noise, it says a task exists and when somebody touched it. Team events, directed events and `share.*` events never reach it. `actor_user_id` and `actor_client_id` are stripped from what it receives: they name an account and a browser tab inside the owner's organisation, and there is no tab on the other side of a link to recognise its own writes.
- Membership verdicts are cached for at most a minute and dropped outright whenever anything touches a team, so the cache is never the reason someone keeps seeing what they lost.
- The broadcast queue is bounded. Under a burst the server drops events rather than stalling the write that produced them, so a client must be able to recover by re-reading, not by replaying.

### Recognising your own writes

A REST write may carry `X-Client-Id`: an opaque per-tab string, at most 64 characters (longer is truncated), invented by the client and never interpreted by the server. It comes back as `actor_client_id` on every event that write produced, so a tab can skip refetching what it already knows. `actor_user_id` cannot do this job — it is empty with auth off, and two tabs of one person share it, so one would suppress a change the other made and must display. MCP writes never carry a client id, so an agent's change always reads as somebody else's.
