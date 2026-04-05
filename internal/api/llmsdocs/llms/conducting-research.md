# Conducting Research

Step-by-step guide for AI assistants on how to conduct a research project with MCP Research.

## Overview

A research project follows this lifecycle:

1. **Initialize** — Design the research structure (sections, goals, tags)
2. **Conduct** — Interview the user, create entries, track progress
3. **Complete** — Mark sections and research as completed

## Step 1: Initialize

Use the `research/initialize` MCP prompt or create via API:

1. Define a clear, specific research goal
2. Design 3-7 sections that cover the topic comprehensively
3. Each section needs a slug name, display name, and description
4. Add relevant tags for categorization
5. Set working instructions for future sessions

## Step 2: Conduct Research Sessions

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

### Track Progress

- Use `research_update` with `add_memory` to record key insights
- Update session notes with `session_update` using `add_note`
- Use tasks (`task_create`) to track work items
- Mark sections as completed when they have sufficient coverage

## Step 3: Complete

1. Mark all sections as completed with `section_update`
2. Mark the research as completed with `research_update`
3. The web UI shows the full research with all entries, questions, and tasks

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

Short codes are returned by all create endpoints and included in list/get responses. They can be used in URLs instead of UUIDs: `/research/R1/entry/E2`.

## Cross-References

Use `[[...]]` syntax in entry content to create links between entries:

- `[[E3]]` — link to entry E3 in the same research
- `[[R2:E5]]` — link to entry E5 in research R2
- `[[R2]]` — link to research R2

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
- Keep session notes updated for context across sessions
- Use tasks to plan and track remaining work
- Use `[[E1]]` cross-references to build connections between related entries
- Run crossref rebuild after batch-creating entries to resolve forward references
