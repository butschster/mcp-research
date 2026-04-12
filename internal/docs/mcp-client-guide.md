# MCP Client Guide

Practical guide for AI assistants and MCP clients interacting with the Research server via MCP tools. Covers tool conventions, content formatting, nullable fields, common pitfalls, and the difference between MCP and REST API access.

## Two Ways to Interact

The Research server exposes two interfaces. Use the one that matches your integration:

| Interface | When to use | Reference |
|-----------|-------------|-----------|
| **MCP tools** | Claude Desktop, Claude Code, Cursor, ChatGPT (via Streamable HTTP), any MCP client | This guide + tool descriptions returned by the server |
| **REST API** | Custom integrations, scripts, webhooks, non-MCP clients | [OpenAPI Specification](/api/openapi.yaml) |

Both interfaces operate on the same data and produce the same results. MCP tools are thin wrappers around the same service layer that the REST API uses.

**MCP prompts** (`research/initialize`, `research/conduct`) are interactive workflows that call MCP tools internally. They are the recommended starting point for new research projects, but every action they perform can also be done with individual tool calls.

## Available MCP Tools

### Research

| Tool | Purpose |
|------|---------|
| `research_create` | Create research with sections, tags, and goal |
| `research_get` | Load full research context (sections, entry counts, active session) |
| `research_list` | List all researches with optional status filter |
| `research_update` | Update name, goal, status, instruction, memory, tags |
| `research_add_section` | Add a new section to an existing research |
| `research_export` | Export full research as JSON or markdown |
| `research_import` | Import a research from exported JSON |

### Entries

| Tool | Purpose |
|------|---------|
| `entry_create` | Create a markdown entry in a section |
| `entry_read` | Read full entry content |
| `entry_list` | List entries in a section (metadata only, no content) |
| `entry_update` | Update title, content, status, tags, or do text replacement |

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
| `section_list` | List sections for a research (includes allowed_entry_statuses per section) |
| `section_update` | Update section name, description, status, position, or allowed_entry_statuses |

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

## Nullable and Optional Fields

Many MCP tool parameters are optional. You can omit them entirely or pass `null` — both are equivalent and result in the default value being used.

### Fields that accept null

**`session_create`**: `focus`, and on each question: `area`, `rationale`, `priority`, `parent_id`
**`entry_create`**: `title`, `description`, `status`, `session_id`
**`task_create`**: `description`, `priority`
**`question_create`**: same question fields as in `session_create`
**`roadmap_create` / `roadmap_add_nodes`**: on nodes — `description`, `node_type`, `status`, `position_x`, `position_y`, `parent_id`, `ref_type`, `ref_id`, `metadata`; on edges — `label`, `edge_type`

### Default values when null or omitted

| Field | Default |
|-------|---------|
| `priority` | `medium` |
| `status` (entry) | First status from section's `allowed_entry_statuses` (usually `draft`) |
| `node_type` | `step` |
| `edge_type` | `default` |
| `title` (entry) | Auto-generated from first line of content |
| `description` (entry) | Auto-generated from lines 2-5 of content |
| `position_x`, `position_y` | `0` (frontend auto-layouts) |

### Example: creating a session with minimal fields

Only `research_id`, `title`, and question `text` are required:

```json
{
  "research_id": "uuid-here",
  "title": "Initial exploration",
  "questions": [
    { "text": "What are the main components?" },
    { "text": "How do they interact?" }
  ]
}
```

Fields like `focus`, `area`, `rationale`, `priority`, `parent_id` can all be omitted or set to `null`.

## Content Formatting

All content fields that support markdown (`content` in entries, `answer` in questions, `notes` in sessions, `description` and `result` in tasks) follow these rules.

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

### Cross-references

Use `[[...]]` syntax in content to create links between documents:

| Syntax | Target |
|--------|--------|
| `[[E3]]` | Entry E3 in same research |
| `[[R2:E5]]` | Entry E5 in research R2 |
| `[[R2]]` | Research R2 |
| `[[RM1]]` | Roadmap RM1 in same research |
| `[[RM1:N3]]` | Node N3 in roadmap RM1 |

Cross-references work in: entry content, question text, question answers, question rationale, task results, task descriptions, and session notes. They are stored in the `crossrefs` table and rendered as clickable links in the web UI.

## Entry Types and Statuses

Entries are the primary knowledge artifacts. They live in sections and represent synthesized research findings.

### Statuses

Entry statuses are **defined per section** via the `allowed_entry_statuses` field. Each section specifies which statuses are valid for its entries. The first status in the list is the default for new entries.

**Default statuses** (used when a section doesn't specify custom ones):

| Status | Meaning | When to use |
|--------|---------|-------------|
| `draft` | Work in progress (default) | Initial creation, incomplete content |
| `active` | Published and current | Entry is complete and represents current understanding |
| `completed` | Finalized, no further changes expected | Section is being closed out |
| `archived` | Historical, superseded by newer entries | Content is outdated but preserved for reference |

**Custom statuses** — sections can define domain-specific statuses. Examples:
- Sources section: `["found", "reading", "reviewed", "rejected"]`
- Experiments: `["planned", "running", "analyzed", "discarded"]`
- Recommendations: `["proposed", "approved", "implemented"]`

**How to use:**
1. Check `section_list` or `research_get` — each section includes `allowed_entry_statuses`
2. When creating/updating an entry, use a status from that list
3. If status is omitted on `entry_create`, the first status in the list is used as default
4. Setting an invalid status returns an error listing the allowed values

### Content patterns

Entries should be self-contained markdown documents. Common patterns:

- **Finding**: synthesized answer to a research question with supporting evidence
- **Comparison**: table-based analysis of alternatives (tools, frameworks, approaches)
- **Guide**: step-by-step instructions derived from research
- **Summary**: high-level overview of a topic area with links to detailed entries via `[[E1]]`
- **Reference**: collected facts, specs, or configuration details

### Auto-generated fields

When `title` is omitted or null, the server extracts it from the first non-empty line of content (stripping markdown heading markers, max 100 chars).

When `description` is omitted or null, the server generates it from lines 2-5 of content (stripping markdown, max 200 chars).

This means you can create entries with just `research_id`, `section_id`, and `content` — everything else is inferred.

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

### 1. Forgetting required fields

Every create tool has a small set of required fields. The server returns a validation error listing all missing fields — read the error message, it tells you exactly what to fix.

**Minimum required fields by tool:**

| Tool | Required |
|------|----------|
| `research_create` | `name` |
| `entry_create` | `research_id`, `section_id`, `content` |
| `session_create` | `research_id`, `title` |
| `question_create` | `session_id`, at least one question with `text` |
| `task_create` | `research_id`, `title` |
| `roadmap_create` | `research_id`, `title` |

### 2. Confusing tool errors with protocol errors

MCP tools in this server **never return Go errors** (protocol-level errors). All failures are returned as `CallToolResult` with `isError: true` and a descriptive text message. If you get a protocol error (like `-32602: invalid params`), it means the JSON schema validation failed before the tool code ran — check your parameter types.

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

## Short Codes

Every entity gets an auto-assigned short code on creation. These codes can be used in URLs and cross-references instead of UUIDs:

| Entity | Pattern | Scope | URL example |
|--------|---------|-------|-------------|
| Research | `R1`, `R2` | Global | `/research/R1` |
| Entry | `E1`, `E2` | Per research | `/research/R1/entry/E2` |
| Session | `SS1`, `SS2` | Per research | `/research/R1/session/SS1` |
| Question | `Q1`, `Q2` | Per session | — |
| Task | `T1`, `T2` | Per research | — |
| Roadmap | `RM1`, `RM2` | Per research | `/research/R1/roadmap/RM1` |
| Node | `N1`, `N2` | Per roadmap | — |

Codes are returned by all create and get endpoints.
