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
- Set `session_id` to link the entry to the session that produced it (helps track provenance).

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

### CrossRef

Links between documents, extracted automatically from `[[...]]` patterns.

| Syntax | Meaning |
|--------|---------|
| `[[E3]]` | Entry E3 in same research |
| `[[R2:E5]]` | Entry E5 in research R2 |
| `[[R2]]` | Research R2 itself |

**Where they work:** Entry content, question text/answers/rationale, task titles/results, session notes. All rendered as clickable links in the web UI.

**Resolution:** References are resolved when the target exists. Unresolved references are tracked and can be resolved later via rebuild.

**Visualization:** Shown on entry detail pages (outgoing/incoming), in the mindmap (dashed edges), and preserved in export.

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
8. research_update → Add to memory (insights for future sessions)
9. Repeat 3-8 for new sessions on uncovered areas
10. Mark sections completed → research completed
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
