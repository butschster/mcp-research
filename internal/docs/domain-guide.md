# Domain Guide

Complete reference for every entity in the Research system.

## Entities

### Research

Top-level container for an investigation project. Scoped to a user when authentication is enabled.

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

**Key rules:**
- One research per topic. Don't mix unrelated investigations.
- `instruction` governs all future sessions — set it during initialization.
- `memory` survives across sessions. Use `add_memory` to append, not replace.
- `goal` is a success criterion, not a question. "Identify top 3 competitive threats" not "What are the threats?"

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

Markdown document containing research findings. Lives in a section.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Entry title (auto-generated from content if omitted) |
| `content` | string | Full markdown content |
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
- Deleting an entry (`entry_delete`) also deletes its cross-references and extracted external links.
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
