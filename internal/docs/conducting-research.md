# Conducting Research

Step-by-step guide for AI assistants on how to conduct a research project.

> **MCP or REST API?** This guide shows MCP tool names and REST endpoints side by side. Use whichever matches your integration. For MCP-specific details (nullable fields, content formatting, common pitfalls), see the [MCP Client Guide](/llms/mcp-client-guide.md). For REST API details, see the [OpenAPI spec](/api/openapi.yaml).

## Overview

A research project follows this lifecycle:

1. **Initialize** — Design the research structure (sections, goals, tags)
2. **Conduct** — Interview the user, create entries, track progress
3. **Complete** — Mark sections and research as completed

### Check first that you may write

A research belongs to a team, not to whoever created it, and your role there can
be read-only. Before continuing someone else's research — a session, an entry, a
task, a status change — confirm you may write to it:

- `research_list` marks a shared research with `team` and a read-only one with
  `access: "read-only"`. No `access` key means you may write.
- `research_get` returns `role` on the research: `viewer`, `editor` or `owner`.

A `viewer` gets `your role in this team does not allow this` on the first write.
Discovering that after an interview wastes the user's answers, so check before
you ask. A research you created yourself is always writable. See
[Access](/llms/mcp-client-guide.md#access-you-can-see-more-than-you-can-write).

## Step 1: Initialize

Use the `research/initialize` MCP prompt, which runs this in three turns, or do
it call by call:

1. Find out what decision is waiting on the research before you propose any
   structure. A section list shown first anchors the person to it.
2. **Check whether a methodology already covers this.** `template_list` returns
   the kickoff methodologies available — matching criteria only, no bodies — and
   `template_get` returns the one that fits in full: what to ask before proposing
   anything, what structure to suggest, what a good entry looks like here, and
   when the research is finished. Follow it; it is instructions, not a form. An
   empty list, or nothing that fits, is a normal answer — design the research
   yourself and say so.
3. Define a clear, specific research goal
4. Design the sections *from that conversation*. Fewer than you think: create one
   when you have something to put in it, since the conductor aims at the
   least-covered sections and an empty one is a standing instruction to invent
   content for it. Each needs a slug name, display name, and description
5. Add relevant tags for categorization
6. Create it with `research_create`, passing `template_slug` when you followed a
   methodology. That stamps `template_slug` / `template_version` on the research
   and attaches the skills the methodology names — read `skills_attached` and
   `skills_unavailable` in the reply. The slug is an MCP argument only; `POST
   /api/researches` has no field for it
7. Set working instructions for future sessions. `instruction` is what **this**
   research is — its scope, its constraints, what counts as done here. How the
   work is done belongs to the template and the skills; do not copy methodology
   into it

See [Templates](/llms/templates.md).

## Step 2: Conduct Research Sessions

### Note the Skills, Load Them Late

`research_get` may return a `skills` array — each entry a name, a tier and one
line saying **when** to use it. The bodies are not there, and the key is absent
entirely when the research follows nothing, which is normal.

Read the lines when you load the research; do not load the bodies then. When you
are about to do the work one of them names — start an interview, grade a source
that two people disagree about, build a roadmap — call `skill_load` with that
slug and read it at that moment. One slug per call. A skill read three steps
before the work it describes has usually been forgotten by the time it matters.

Where two skills conflict, the higher tier wins: research-private over team over
built-in, which is why the index arrives in that order. `instruction` is not part
of that ordering — it answers a different question, what this research is, and
still governs tone and depth. See [Skills](/llms/skills.md).

### Create a Session

Start a Q&A session focused on specific sections:

1. Create a session with `session_create` (MCP) or `POST /api/sessions` (API)
2. Include initial questions covering gaps in the research
3. Set a focus area to keep the interview directed

### Interview Loop

For each question:

1. Present the question clearly to the user
2. Record their answer with `question_update`
3. If the answer raises follow-ups, add them with `question_create`
4. Defer or skip questions that are out of scope

### Create Entries

As information accumulates:

1. Write entries with well-structured markdown in `entry_create`
2. Place each entry in the appropriate section
3. Use tags for cross-cutting concerns
4. Title and description are auto-generated from content if not provided
5. Use `[[E1]]` syntax to cross-reference other entries
6. Each entry gets an auto-assigned short code (E1, E2, ...)
7. Entries created while a session is active are linked to it automatically, which is what the session export lists as "entries produced in this session"

### Build on What Exists, Don't Overwrite It

Before rewriting an entry a previous session produced:

1. `entry_history` — who wrote it last, in which session, and what they changed
2. `entry_diff` — the change itself, if the history suggests someone corrected something

A revision whose author is `human` is a person's edit in the web UI: treat it as
the strongest reason to extend rather than replace what is there.

Every write that changes something leaves a revision, so nothing is permanently
lost and an earlier version can be restored. But undoing another session's
correction without noticing is exactly the failure this history exists to catch.
See [Revisions](/llms/revisions.md).

### Track Progress

- Use `research_update` with `add_memory` to record key insights
- Update session notes with `session_update` using `add_note`
- Use tasks (`task_create`) to track work items
- Mark sections as completed when they have sufficient coverage

### Build Roadmaps

When the research has a natural progression, sequence, or decision tree, create a visual roadmap:

1. Identify whether the topic suits a roadmap (learning path, strategy, migration plan, onboarding flow)
2. Create a roadmap with `roadmap_create` (MCP) or `POST /api/roadmaps` (API)
3. Include all nodes and edges in one call using `temp_id` for node references in edges
4. Define custom `statuses` that fit the domain (e.g. `["not_started", "learning", "mastered"]`)
5. Update node statuses with `roadmap_update_node` as the user progresses
6. Extend the graph with `roadmap_add_nodes` as new steps emerge

**When to create a roadmap during research:**

| Situation | Action |
|-----------|--------|
| Topic has a clear learning sequence | Create a roadmap after initial Q&A reveals the learning path |
| Research uncovers a multi-step process | Build a roadmap showing the steps and dependencies |
| User asks "how do I get from A to B?" | Create a roadmap with the progression |
| Multiple alternatives exist | Use decision nodes to show branching paths |
| Research maps a system architecture | Create a roadmap showing component dependencies |

**Example: After a session on "learning Vue 3", create a roadmap:**
- Nodes: HTML/CSS basics → JavaScript ES6+ → Vue 3 Fundamentals → Composition API → State Management → Testing → Deployment
- Edge types: `default` for the main path, `optional` for alternatives
- Statuses: `not_started`, `in_progress`, `completed`

**Tips:**
- Create roadmaps when enough information has been gathered (typically after 1-2 sessions)
- Use `milestone` nodes to mark key achievements or checkpoints
- Use `decision` nodes when the path branches based on choices
- Use `info` nodes for prerequisites or reference material that isn't a step
- Keep node descriptions concise — detailed content belongs in entries, link conceptually

## Step 3: Complete

1. Mark all sections as completed with `section_update`
2. Mark the research as completed with `research_update`
3. The web UI shows the full research with all entries, questions, and tasks
4. Hand the user a document if they want one — see [Export](/llms/export.md):
   - the research and per-session export pages produce markdown or PDF
   - `research_export` with `format: "obsidian"` returns a link to a zip shaped like an Obsidian vault (a folder per section, a note per entry, `[[E3]]` resolving as a link) — offer this when the user keeps notes in Obsidian or wants the research as files. The link needs their bearer token
   - `research_export` with no `format` returns the portable JSON, which is for moving the research to another server, not for reading

## Short Codes

Every record gets an auto-assigned short code on creation:

| Entity | Prefix | Scope | Example |
|--------|--------|-------|---------|
| Research | `R` | global | `R1`, `R2` |
| Section | `S` | per research | `S1`, `S2` |
| Entry | `E` | per research | `E1`, `E2` |
| Session | `SS` | per research | `SS1`, `SS2` |
| Question | `Q` | per session | `Q1`, `Q2` |
| Task | `T` | per research | `T1`, `T2` |
| Roadmap | `RM` | per research | `RM1`, `RM2` |
| Node | `N` | per roadmap | `N1`, `N2` |

REST responses carry the `code` field on every entity. MCP tools are less complete: `research_create`, `research_import`, `research_get`, `entry_create`, `entry_list`, `entry_read`, `entry_patch`, `session_get` and the `roadmap_*` tools return codes, while section, question and task codes are only reachable through the REST API. Codes can be used in URLs instead of UUIDs (`/research/R1/entry/E2`), but as tool arguments only `research_get`, `research_update`, `research_export`, `session_get` and `roadmap_get` resolve them — see the [MCP Client Guide](/llms/mcp-client-guide.md).

## Cross-References

Use `[[...]]` syntax in entry content to create links between documents:

- `[[E3]]` — link to entry E3 in the same research
- `[[R2:E5]]` — link to entry E5 in research R2
- `[[R2]]` — link to research R2
- `[[RM1]]` — link to roadmap RM1 in the same research
- `[[RM1:N3]]` — link to node N3 in roadmap RM1

### How it works

1. When an entry is created or updated, the server parses all `[[...]]` patterns from the content
2. Each reference is resolved to a target entry/research UUID and stored in the `crossrefs` table
3. If the target doesn't exist yet (e.g. `[[E5]]` before E5 is created), the reference is stored as unresolved
4. Use `POST /api/researches/{id}/crossrefs/rebuild` to re-scan all entries and resolve stale references
5. On server startup, codes are automatically backfilled for any records missing them

### Viewing cross-references

- **In entry view**: `[[E3]]` renders as a clickable badge-style link navigating to the target entry
- **In mindmap**: cross-references appear as purple dashed edges between entry nodes. Hover an edge to see which entries are connected and highlight the source/target nodes
- **Via API**: `GET /api/researches/{id}/crossrefs` returns all resolved and unresolved references

### Best practices for cross-referencing

- Reference foundational entries from higher-level ones: "See [[E1]] for goroutine basics"
- Use cross-research references when topics span projects: "Compare with [[R2:E3]]"
- After creating entries that are referenced by earlier entries, run rebuild to resolve forward references
- Keep entries self-contained — cross-references add context but each entry should be readable alone

## Best Practices

- Ask one question at a time for clarity
- Prioritize high-priority questions first
- Write entries that are self-contained and useful on their own
- Use the research's instruction field as your guide for tone and depth
- Load a skill when you reach the work it names, not while orienting — and one at a time
- Keep session notes updated for context across sessions
- Use tasks to plan and track remaining work
- Use `[[E1]]` cross-references to build connections between related entries
- Run crossref rebuild after batch-creating entries to resolve forward references
- Create roadmaps when the research reveals step-by-step processes, learning paths, or decision trees
- Build the full roadmap graph in one `roadmap_create` call rather than adding nodes one at a time
- Choose roadmap statuses that match the domain vocabulary (learning, marketing, engineering, etc.)
