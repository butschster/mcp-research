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
- `instruction` governs all future sessions — set it during initialization.
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
- The WebSocket at `/ws` needs the same credential as everything else when auth is on (`?token=…`, because a browser cannot set headers on a handshake), and the hub decides delivery **per event**: a research event reaches only those who may read it, a team event only its members. Membership is re-checked on each event, so removing someone stops their updates on the socket they already have open.
- That local team survives if auth is later turned on. It has no members, so its researches are **readable by every signed-in user and writable by none** until the first registration claims them — a deliberate compromise between stranding them behind a team nobody can join and letting the first caller take them for themselves.

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

A revision has no short code: it is a plain number, 1-based per entry.
