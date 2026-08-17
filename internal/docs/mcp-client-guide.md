# MCP Client Guide

Practical guide for AI assistants and MCP clients interacting with the Research server via MCP tools. Covers tool conventions, team roles and what they let you write, content formatting, nullable fields, common pitfalls, and the difference between MCP and REST API access.

## Two Ways to Interact

The Research server exposes two interfaces. Use the one that matches your integration:

| Interface | When to use | Reference |
|-----------|-------------|-----------|
| **MCP tools** | Claude Desktop, Claude Code, Cursor, ChatGPT (via Streamable HTTP), any MCP client | This guide + tool descriptions returned by the server |
| **REST API** | Custom integrations, scripts, webhooks, non-MCP clients | [OpenAPI Specification](/api/openapi.yaml) |

Both interfaces operate on the same data and produce the same results. MCP tools are thin wrappers around the same service layer that the REST API uses.

**MCP prompts** (`research/initialize`, `research/conduct`) return workflow instructions that tell you which tools to call in which order. `research/initialize` takes an optional `topic` argument; `research/conduct` requires `research_id`. They are the recommended starting point for new research projects, but every action they describe can also be done with individual tool calls.

## Available MCP Tools

### Research

| Tool | Purpose |
|------|---------|
| `research_create` | Create research with sections, tags, and goal. Optional `team_id` picks the team it lands in; omitted, it goes to your personal team |
| `research_get` | Load full research context (sections, entry counts, active session) |
| `research_list` | List every research you can reach, with optional status filter. Marks a shared one with `team` and a read-only one with `access: "read-only"` |
| `research_update` | Update name, goal, status, instruction, memory, tags |
| `research_add_section` | Add a new section to an existing research |
| `research_export` | Export a full research (sections, entries, sessions, questions, tasks, roadmaps). `format` defaults to `portable` (JSON for `research_import`); `format: "obsidian"` returns a link to download it as an Obsidian vault |
| `research_import` | Re-create a research from a portable export payload. Optional `team_id` picks the team it lands in; omitted, it goes to your personal team |

### Entries

| Tool | Purpose |
|------|---------|
| `entry_create` | Create an entry in a section — markdown by default, or a block document via `entry_type` |
| `entry_read` | Read full entry content |
| `entry_list` | List entries in a section (metadata only, no content) |
| `entry_update` | Update title, content, description, status, tags, session link, entry type, or replace text in a **markdown** entry (`text_replace`) |
| `entry_patch` | Edit blocks of a `blocks` entry by id — update, insert, delete, move, tick a checklist item — as one atomic, strict change. The surgical edit path for block documents; `text_replace` is refused on them |
| `entry_delete` | Delete an entry (also removes its cross-references and external links) |
| `entry_history` | List an entry's revisions: who wrote each one (agent, human, import, restore), in which session, and what it changed. Read it before rewriting an entry another session wrote |
| `entry_diff` | Unified diff between two revisions; with no numbers, the most recent change. Block documents are compared as markdown, not JSON |

### Sessions & Questions

| Tool | Purpose |
|------|---------|
| `session_create` | Create a Q&A session with initial questions |
| `session_get` | Load session with all questions and progress |
| `session_update` | Update title, focus, status, or notes |
| `question_create` | Add questions to an existing session |
| `question_update` | Record answer, change question status |
| `question_list` | List questions with filtering |

### Tasks

| Tool | Purpose |
|------|---------|
| `task_create` | Create a todo item in a research |
| `task_list` | List tasks (sorted by priority) |
| `task_update` | Update status, priority, result |
| `task_delete` | Remove a task |

### Sections

| Tool | Purpose |
|------|---------|
| `section_list` | List sections for a research |
| `section_update` | Update section display name, description, status, position (the slug `name` is immutable) |

### Roadmaps

| Tool | Purpose |
|------|---------|
| `roadmap_create` | Create a full roadmap with nodes and edges |
| `roadmap_get` | Load roadmap with all nodes, edges, and reference data |
| `roadmap_list` | List roadmaps for a research |
| `roadmap_update` | Update title, description, statuses, status |
| `roadmap_delete` | Delete a roadmap and all its nodes/edges |
| `roadmap_add_nodes` | Add nodes and edges to an existing roadmap |
| `roadmap_update_node` | Update a single node |
| `roadmap_remove_nodes` | Remove nodes (edges auto-cascade) |

## Access: You Can See More Than You Can Write

`team_list` names the teams you belong to and your role in each. Pass a `team_id` to `research_create` (or `research_import`) to put new work where your colleagues can see it — without it every research lands in your personal team and someone has to move it by hand.


A research is owned by a **team**, and your role in that team decides what you may do to it. You may therefore be able to read a research, list its entries and export it, and still be refused when you try to write to it. The creator of a research has no standing privileges over it: `user_id` records who made it and is never consulted for permission.

| Role | May do |
|------|--------|
| `viewer` | Every read tool (`*_get`, `*_list`, `*_read`, `entry_history`, `entry_diff`, `research_export`) |
| `editor` | The above, plus every create/update/delete tool |
| `owner` | The above, plus team management and moving a research to another team (REST only) |

**How you find out, before you fail:**

- `research_list` marks each item it needs to. A research owned by a team that is not your own personal one carries `team` with the team's name; one you may only read carries `access: "read-only"`. **No `access` key means you may write to it.**
- `research_get` returns the research record itself, which carries `team_id`, `team_name`, `team_is_personal` and `role` (`viewer` / `editor` / `owner`).
- With `auth_enabled: false` there is no caller and no check: `access` never appears, and every tool is permitted.

**How a refusal arrives** — as an ordinary tool result with `isError: true`, never as a protocol error:

| Text | Means | Do |
|------|-------|----|
| `your role in this team does not allow this` | You are in the team but only a `viewer` | Do not retry. Tell the user which research it was and that they need editor rights, or pick another research |
| `not found` | Either no such id, **or** it belongs to a team you are not in — the two are deliberately indistinguishable | Re-run `research_list` and use an id from it |

`research_create` and `research_import` both take an optional `team_id`. Omit it and the research lands in your own personal team, so a research you created this session is always writable. Name a team and it goes there instead — you must be an `editor` or `owner` of it: a team you are not in is `not found`, one where you are only a `viewer` is refused. Use `team_list` to get the id; `research_create` echoes back `team` with the team's name when the research did not land in your personal one.

**Share links are not yours to issue.** A research can be published as a revocable, read-only link that someone opens without an account, but there is no MCP tool that creates, lists or revokes one — it is done in the web UI or over REST (`POST /api/researches/{id}/shares`), by a person deciding to give something away. If a user asks for one, point them at the share dialog on the research page rather than looking for a tool. Nothing you do over MCP is affected by a link existing, and a share token never authenticates an MCP call. [Domain Guide → Share](/llms/domain-guide.md#share).

Full model — roles, invitations, transfer and the team REST routes: [Domain Guide](/llms/domain-guide.md#team).

## Nullable and Optional Fields

Read this before composing any tool call. Input schemas are generated from Go structs, and the generated schema lists **every** property of a tool (and of every nested object) in `required`, with `additionalProperties: false`. Optionality is expressed by nullability, not by absence:

| Parameter kind in the schema | How to skip it | Effect of `null` |
|------------------------------|----------------|------------------|
| `"type": ["null", "string"]` / `["null","number"]` (pointer in Go) | send `null` | default value is used |
| `"type": ["null", "array"]` (any list: `tags`, `statuses`, `questions`, `nodes`, `edges`, `node_ids`) | send `null` or `[]` | treated as empty |
| `"type": "string"` / `"integer"` (plain scalar) | send `""` or `0` — **not** `null` | rejected: `null` is not a valid string/integer |

Consequences:

- **Send every property.** Omitting one currently fails schema validation with `-32602 invalid params: required: missing properties: [...]` before the tool code runs. Use `null` (or `""` / `0`) rather than leaving a property out.
- **Five exceptions**, the only tools whose schema does not require everything: `entry_history` requires `entry_id` alone (`limit` may be omitted), `entry_diff` requires `entry_id` alone (`from` and `to` may be omitted), `research_export` requires `research_id` alone (`format` may be omitted), `research_import` requires `data` alone (`team_id` may be omitted), and `research_create` requires everything except `team_id`. Sending `null` for those six properties works as well, so "send every property as `null`" is still a correct strategy everywhere.
- **Never send `null` into a plain scalar.** The optional-but-not-nullable parameters are: `research_create` → `description`, `goal`; each `sections[]` item → `display_name`, `description`, `position`; `research_add_section` → `display_name`, `description`, `position`; each question item → `position`.
- **List filters are nullable**: `research_list.status`, `entry_list.status`, `question_list.status` / `area` / `priority`, `task_list.status` / `priority`. `null` or `""` means "no filter".
- Unknown property names are rejected outright (`additionalProperties: false`).

### Default values when a nullable field is null

| Field | Default |
|-------|---------|
| `priority` (question, task) | `medium` |
| `status` (entry) | `draft` |
| `node_type` | `step` |
| `edge_type` | `default` |
| `title` (entry) | Auto-generated from first non-empty line of content |
| `description` (entry) | Auto-generated from lines 2-5 of content |
| `position_x`, `position_y` | `0` (frontend auto-layouts) |
| `session_id` (entry) | The research's currently active session, if there is one |
| `limit` (`entry_history`) | `20` newest revisions; the result says `truncated: true` when more exist |
| `format` (`research_export`) | `portable` — the JSON `research_import` takes. `obsidian` returns a vault download link instead; `json` / `vault` / `zip` are accepted aliases, anything else is a validation error |
| `team_id` (`research_create`, `research_import`) | Your personal team |
| `to` (`entry_diff`) | The newest revision |
| `from` (`entry_diff`) | The revision before `to` — so a call with neither shows the most recent change |

### Example: creating a session with the fewest meaningful values

`research_id`, `title`, and each question's `text` are the only fields carrying information; everything else is still present, as `null` or `0`:

```json
{
  "research_id": "uuid-here",
  "title": "Initial exploration",
  "focus": null,
  "questions": [
    { "text": "What are the main components?", "area": null, "rationale": null, "priority": null, "parent_id": null, "position": 0 },
    { "text": "How do they interact?", "area": null, "rationale": null, "priority": null, "parent_id": null, "position": 1 }
  ]
}
```

`research_id` here must be the UUID — `session_create` does not resolve `R1`-style codes (see Short Codes).

## Content Formatting

All content fields that support markdown (`content` in a markdown entry, `answer` in questions, `notes` in sessions, `description` and `result` in tasks) follow these rules.

They do **not** apply to an entry with `entry_type: blocks`: there `content` is a JSON block document, markdown is not parsed, and each block field has its own rules — see [Block Documents](/llms/blocks.md).

### Newlines

**Use actual newline characters**, not literal `\n` sequences. The server normalizes literal `\n` to real newlines as a safety net, but it is better to send properly formatted content from the start.

```
Good:  "# Title\n\nParagraph one.\n\nParagraph two."
       (where \n is an actual newline character in the JSON string)

Bad:   "# Title\\nParagraph one.\\nParagraph two."
       (literal backslash-n — will be auto-corrected but indicates a client bug)
```

When composing multi-line markdown, ensure your JSON serializer encodes newlines as `\n` (JSON escape for real newline), not as `\\n` (escaped backslash + n).

### Markdown support

Content is rendered using GitHub Flavored Markdown with `breaks: true` (single newline = line break). Supported formatting:

- Headings (`# H1` through `###### H6`)
- Bold, italic, strikethrough (`**bold**`, `*italic*`, `~~strike~~`)
- Lists (ordered and unordered, nested)
- Code blocks (fenced with triple backticks, language hint supported)
- Tables (GFM pipe syntax)
- Links and images
- Blockquotes
- Mermaid diagrams (fenced code blocks with `mermaid` language — see below)

### Mermaid diagrams

Use fenced code blocks with the `mermaid` language identifier to embed diagrams. They are rendered as interactive SVG in the web UI. All Mermaid diagram types are supported: flowchart, sequence, class, state, ER, Gantt, pie, mindmap, timeline, etc.

Example in entry content:

```
\`\`\`mermaid
graph TD
    A[Start] --> B{Decision}
    B -->|Yes| C[Action 1]
    B -->|No| D[Action 2]
    C --> E[End]
    D --> E
\`\`\`
```

Tips:
- Keep diagrams focused — large diagrams become hard to read
- Combine diagrams with markdown text for context
- Use descriptive node labels
- Mermaid blocks work in entry content, question answers, session notes, and task descriptions/results
- Diagrams are interactive: drag to pan, ctrl/⌘ and scroll to zoom, a button for fullscreen,
  and a link that reopens the diagram in mermaid.live
- A diagram that fails to parse keeps its source and links to the editor, which reports the
  syntax error
- In a `blocks` document diagrams get their own block type: `{"type": "mermaid", "data":
  {"code": "flowchart TD\n  A --> B", "caption": "optional"}}`. A `code` block with
  `language: "mermaid"` is accepted as the same thing

### Cross-references

Use `[[...]]` syntax in content to create links between documents:

| Syntax | Target |
|--------|--------|
| `[[E3]]` | Entry E3 in same research |
| `[[R2:E5]]` | Entry E5 in research R2 |
| `[[R2]]` | Research R2 |
| `[[RM1]]` | Roadmap RM1 in same research |
| `[[RM1:N3]]` | Node N3 in roadmap RM1 |

`[[...]]` is rendered as a clickable link anywhere the web UI displays markdown: entry content, question text, rationale and answers, task titles, descriptions and results, session notes.

Only three sources are additionally **indexed** into the `crossrefs` table, and therefore feed the graph views and `GET /api/researches/{id}/crossrefs`:

| Source | Indexed text |
|--------|--------------|
| Entry | `content` (on create, update, and crossref rebuild) |
| Question | `answer` (on `question_update`) |
| Task | `description` + `result` (on `task_update`) |

Put references you want in the knowledge graph into entry content: the graph view draws an edge only for a resolved reference whose target is an entry, so `[[R2]]` and `[[RM1]]` are stored and clickable but never become graph edges. A reference to a target that does not exist yet is stored unresolved and can be fixed later with `POST /api/researches/{id}/crossrefs/rebuild`, which re-scans entry content only.

## Entry Types and Statuses

Entries live in sections and represent synthesized research findings.

### Types

`entry_type` picks how `content` is interpreted. It is optional on `entry_create` and defaults to `markdown`.

| Type | `content` holds | Rendered as |
|------|-----------------|-------------|
| `markdown` (default) | Markdown, with `[[E3]]` cross-references and ```mermaid blocks | Markdown in the page |
| `blocks` | A block document `{version:1,blocks:[{type,data}]}` | Each block by its own renderer |
| `artifact` | A complete, self-contained HTML document | Sugar: stored as a `blocks` document holding one `html` block |

Use `blocks` when the entry is a composed document — prose plus an alert, a table, a chart, a checklist, a diagram, a hand-built visual — rather than one stream of text. Twelve block types are available: `paragraph`, `heading`, `list`, `table`, `quote`, `code`, `callout`, `divider`, `image`, `checklist`, `mermaid`, `html`. Text fields carry the inline markdown subset and `[[E3]]` references, and they are indexed; `code`, `mermaid` and `html` bodies are not. **Read [Block Documents](/llms/blocks.md) for the field-by-field catalog before writing one** — an unknown type or a mistyped field is dropped silently, so a guessed field name means a missing block.

Editing one: send the whole document again in `content` (forgiving — bad blocks are dropped) or change part of it with `entry_patch`, which addresses blocks by their `id` and is strict and atomic. `text_replace` does **not** work on a blocks entry and is rejected: it would run a string replacement over the stored JSON. Keep the block `id`s and the `checklist` item `key`s you read back — they are what a human's ticks hang on, and `entry_update` reports nothing when they are lost.

`artifact` still works and needs no document: pass the HTML as `content` and it is wrapped in one `html` block, taking its title from `<title>`. Reading the entry back returns `entry_type: blocks` — nothing stores `artifact` any more.

Everything else stays `markdown`.

Writing HTML — as an `html` block inside a document, or via the `artifact` alias:

- Send one whole document: `<!doctype html>`, `<head>`, `<style>`, `<script>`, `<body>`. Inline everything — external requests are not available to the frame.
- Give it a `<title>`. Using the alias, the entry title comes from there when `title` is omitted, and `<meta name="description">` fills `description` the same way. Inside a document, prefer the block's own `title` field.
- Scripts run. The frame is sandboxed with `allow-scripts` and without `allow-same-origin`, so the document cannot read cookies, storage or the host page — and cannot fetch from the API.
- The frame reports its own height, so it is shown in full with no inner scrollbar. Do not set `height: 100%` on `body` expecting a viewport; lay out for a document that grows.
- Read-only host context arrives after load as `window.researchData` and a `research-data` event: the research (id, code, name, goal), the entry (id, code, title, tags) and the section list. Render from it rather than hardcoding names.
- **Markdown exports do not contain the HTML.** A markdown export names the block and points at the web UI; a wall of markup in a `.md` file is neither readable nor markdown. The Obsidian vault export (`research_export` with `format: "obsidian"`) is the exception: there each `html` block is written as a real `.html` file under `_html/` and linked from the note. `research_export` / `research_import` carry `entry_type`, so the document survives a round trip. See [Export](/llms/export.md).

`[[E3]]` inside HTML is stored as literal text: cross-references are extracted from markdown content and from block text fields, never from a `code`, `mermaid` or `html` body.

### Statuses

| Status | Meaning | When to use |
|--------|---------|-------------|
| `draft` | Work in progress (default) | Initial creation, incomplete content |
| `active` | Published and current | Entry is complete and represents current understanding |
| `completed` | Finalized, no further changes expected | Section is being closed out |
| `archived` | Historical, superseded by newer entries | Content is outdated but preserved for reference |

### Content patterns

Entries should be self-contained markdown documents. Common patterns:

- **Finding**: synthesized answer to a research question with supporting evidence
- **Comparison**: table-based analysis of alternatives (tools, frameworks, approaches)
- **Guide**: step-by-step instructions derived from research
- **Summary**: high-level overview of a topic area with links to detailed entries via `[[E1]]`
- **Reference**: collected facts, specs, or configuration details

### Auto-generated fields

When `title` is null or empty, the server extracts it from the first non-empty line of content (stripping markdown heading markers, max 100 chars; `Untitled` if the content is blank).

When `description` is null, the server generates it from lines 2-5 of content (stripping markdown, max 200 chars).

For a `blocks` entry both come from the document instead: the title from the first `heading`, else the first sentence of the first `paragraph`, else a lone `html` block's `title` — and the call **fails** when the document offers none of those. The description falls back to the first paragraph that is not already the title, or an `html` block's `caption`.

When `session_id` is null, the server links the entry to the research's currently active session if one exists. Pass an explicit session ID to override that, and use `entry_update` with `session_id: ""` to unlink.

This means only `research_id`, `section_id`, and `content` need real values — everything else can be `null` and is inferred.

## Roadmap Node Types

Roadmap nodes have a `node_type` field that determines their visual appearance and semantic meaning:

| Type | Purpose | When to use |
|------|---------|-------------|
| `step` | Regular action item (default) | Most nodes — learning steps, tasks, stages |
| `milestone` | Key achievement or checkpoint | Mark significant completions or phase transitions |
| `decision` | Fork in the path | When the path branches based on a choice |
| `info` | Reference material or prerequisite | Context that isn't an action step |
| `group` | Container for related steps | Visual grouping of related nodes |
| `checklist` | Sub-items with checkboxes | Use `metadata` JSON for items list |
| `note` | Free-form annotation | Sticky-note style commentary on the graph |
| `link` | External URL reference | Use `metadata` JSON for the URL |
| `metric` | KPI or numeric indicator | Use `metadata` JSON for the value |

### Entity references on nodes

Nodes can reference existing research entities via `ref_type` and `ref_id`. Referenced data is resolved at read time (always shows current state):

| ref_type | References | Synced data |
|----------|-----------|-------------|
| `entry` | An entry | Title, status, content preview, section |
| `task` | A task | Title, status, priority, result |
| `session` | A session | Title, status, question progress |
| `research` | Another research | Name, status, section/entry counts |
| `question` | A question | Text, status, answer |

## Common Pitfalls

### 1. Sending empty values for fields that need content

Beyond schema validation, each create tool checks a few fields for non-empty values and returns `isError: true` with a list of everything missing — read the message, it names each field.

**Fields that must carry a real value:**

| Tool | Must be non-empty |
|------|-------------------|
| `research_create` | `name` |
| `entry_create` | `research_id`, `section_id`, `content` |
| `session_create` | `research_id`, `title` |
| `question_create` | `session_id`, at least one question with `text` |
| `task_create` | `research_id`, `title` |
| `roadmap_create` | `research_id`, `title` |

### 2. Confusing tool errors with protocol errors

MCP tools in this server **never return Go errors** (protocol-level errors). All failures are returned as `CallToolResult` with `isError: true` and a descriptive text message. If you get a protocol error (like `-32602: invalid params`), the input schema rejected your arguments before the tool code ran. The two usual causes: a property was left out (all properties are required — see Nullable and Optional Fields), or `null` was sent into a plain string/integer parameter.

### 3. Forgetting temp_id in roadmap creation

When creating roadmaps, edges reference nodes by `temp_id`. If you forget to set `temp_id` on a node that an edge references, the edge will have a broken connection.

```json
{
  "nodes": [
    { "temp_id": "n1", "title": "Step A" },
    { "temp_id": "n2", "title": "Step B" }
  ],
  "edges": [
    { "source": "n1", "target": "n2" }
  ]
}
```

### 4. Not setting result when completing tasks

When changing a task status to `completed` or `failed`, always set the `result` field in the same call. The result is the permanent record of what happened.

### 5. Sending answers without setting status

When answering a question, set both `answer` and `status: "answered"` in the same `question_update` call. Setting status to `answered` without an answer will fail validation.

### 6. Combining replace and append parameters

`research_update` rejects `memory` together with `add_memory`, and `session_update` rejects `notes` together with `add_note`. Pick one per call: the plain field replaces the whole value, the `add_*` field appends a single item. Set the other to `null`.

### 7. Editing a blocks entry as if it were text

`text_replace` is rejected on an entry with `entry_type: blocks` — it would rewrite the stored JSON as a string. Use `entry_patch` for part of the document, or `entry_update` with the whole document in `content`. Switching an entry **to** `blocks` also needs the block-form `content` in the same call; without it the call fails rather than wrapping the markdown in one paragraph.

### 8. Rewriting an entry without reading what happened to it

An entry may have been written by another session, corrected by a person, or
already fixed by you. `entry_history` costs one call and tells you who last
touched it and what they changed; `entry_diff` shows the change itself. Nothing
is lost if you overwrite it — every write that changes something appends a
revision, and an earlier one can be restored — but undoing someone's correction
and not noticing is a real failure mode, and the history is how you avoid it.

Three more things worth knowing about revisions:

- A write that changes nothing appends nothing, so re-sending identical content
  leaves the history as it was.
- `entry_history` returns no content — it is metadata per revision. Use
  `entry_diff` for what changed and `entry_read` for the current text.
- A `rev` (the 12-character content hash a `blocks` entry carries for optimistic
  concurrency, returned by `entry_read` and `entry_patch`) is **not** a revision
  number. Never pass one into `from` / `to`.

See [Revisions](/llms/revisions.md).

### 9. Assuming every research you can see is one you can write to

Reading a research is not permission to change it. Check `research_list` for
`access: "read-only"`, or `role` on the record `research_get` returns, before you
plan a session, an entry or a task against a research you did not create. A
`viewer` gets `your role in this team does not allow this` on the first write —
after the user has already answered your questions. See
[Access](#access-you-can-see-more-than-you-can-write).

## Short Codes

Every entity gets an auto-assigned short code on creation. These codes can be used in URLs and cross-references instead of UUIDs:

| Entity | Pattern | Scope | URL example |
|--------|---------|-------|-------------|
| Research | `R1`, `R2` | Global | `/research/R1` |
| Section | `S1`, `S2` | Per research | — |
| Entry | `E1`, `E2` | Per research | `/research/R1/entry/E2` |
| Session | `SS1`, `SS2` | Per research | `/research/R1/session/SS1` |
| Question | `Q1`, `Q2` | Per session | `/research/R1/session/SS1/question/Q1` |
| Task | `T1`, `T2` | Per research | — |
| Roadmap | `RM1`, `RM2` | Per research | `/research/R1/roadmap/RM1` |
| Node | `N1`, `N2` | Per roadmap | — |

Which tools hand codes back:

- `research_create`, `research_import`, `research_get` — research code
- `entry_create`, `entry_list`, `entry_read`, `entry_patch` — entry codes (`entry_read` returns the whole entry record, so `code` is on it)
- `roadmap_create`, `roadmap_list`, `roadmap_get`, `roadmap_update_node` — roadmap and node codes
- `session_get` — session code (inside the `session` object)
- Section, question, and task codes are **not** returned by any tool right now: `research_get`, `question_list`, `session_get` questions, `task_list` and `task_create` return UUIDs only. Use those UUIDs in subsequent tool calls, and the REST API (`GET /api/researches/{id}`, `/tasks`, `/sessions/{sessionId}`) if you need the codes themselves.

Where codes are accepted as tool input:

| Accepts UUID **or** code | Accepts UUID only |
|--------------------------|-------------------|
| `research_get`, `research_update`, `research_export` (`research_id`), `session_get` (`session_id`), `roadmap_get` (`roadmap_id`) | every other tool — `research_id` in `entry_create` / `entry_list` / `section_list` / `session_create` / `task_create` / `task_list` / `roadmap_create` / `roadmap_list`, `entry_id`, `question_id`, `task_id`, `session_id` in `session_update`, `roadmap_id` in `roadmap_update` / `roadmap_delete` / `roadmap_add_nodes` / `roadmap_remove_nodes`, `node_id` |

So keep the UUIDs returned by create calls. Short codes are for humans, URLs, and `[[...]]` cross-references — REST routes resolve them in `{id}` / `{sessionId}` / `{entryId}` / `{roadmapId}` path segments, MCP tools mostly do not.
