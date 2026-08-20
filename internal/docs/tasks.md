# Tasks

Tasks are AI-managed todo items within a research. Use them to plan work, track action items, and record outcomes.

## MCP Tools

| Tool | Description |
|------|-------------|
| `task_create` | Create a new task in a research |
| `task_list` | List tasks for a research (sorted by priority, then creation time) |
| `task_update` | Update task status, priority, title, description, or result |
| `task_delete` | Remove a task from the list |

## When to Create Tasks

- Breaking down a complex research into concrete work items
- Tracking follow-up actions discovered during Q&A sessions
- Planning next steps after completing a session
- Recording action items from research findings (e.g. "benchmark X", "verify claim Y")
- Documenting work that needs to happen outside the research tool (e.g. "deploy to staging", "schedule meeting")

## Tasks vs Questions

Questions and tasks serve different purposes:

| | Questions | Tasks |
|---|-----------|-------|
| **Purpose** | Capture knowledge through interview | Track work items and actions |
| **Example** | "What are the main competitors?" | "Create competitor comparison table" |
| **Outcome** | `answer` field with findings | `result` field with outcome |
| **Lifecycle** | pending → answered / skipped / deferred | pending → in_progress → completed / failed |
| **Context** | Lives in a session | Lives in a research |

Use questions for knowledge extraction, tasks for action tracking.

## Statuses

```
pending → in_progress → completed
                      → failed
                      → blocked (waiting on something)
                      → deferred (postponed)
```

- `pending` — not yet started (default)
- `in_progress` — actively being worked on
- `blocked` — waiting on external input or dependency. Describe the blocker in `description`.
- `completed` — done. Always set `result` with the outcome.
- `failed` — attempted but unsuccessful. Set `result` with what went wrong.
- `deferred` — postponed for later. Still valid but not prioritized.

## Priority

`high`, `medium` (default), `low`

Tasks are listed sorted by priority (high first), then by creation time within the same priority level.

## A task inside a document

A `blocks` entry can show tasks with a `task_ref` block — a plan, a handover note
or a meeting summary that carries the work it names instead of describing it:

```json
{ "type": "task_ref",
  "data": { "tasks": ["3f1c0b6a-6a2e-4b1c-9a77-1f2e3d4c5b6a"], "note": "Before the next call" } }
```

- **It references, never copies.** The block stores ids; the titles, statuses and
  priorities are read from the tasks each time the document is rendered, so a
  renamed or completed task is right everywhere at once.
- **A tick in the document is a status change on the task** — the board,
  `task_list` and the document cannot disagree, because there is one place the
  state lives. `entry_patch`'s `set_state` is refused here; use `task_update`.
- **Reference the uuid.** `task_create` returns `task_id` and `task_list` returns
  `id`; the short code `T4` is accepted too but no MCP tool hands it to you.
- Create the tasks first: the block does not create them, and a reference to
  nothing simply draws no row.

Field-by-field rules are in [Block Documents](/llms/blocks.md).

## Best Practices

- **Always set `result` when completing or failing** — it captures the outcome and becomes part of the research record
- Use `blocked` with a clear reason in description when waiting on something
- Task results support markdown and `[[...]]` cross-references — link to entries that were created as a result
- Keep task titles actionable and specific: "Benchmark SQLite WAL modes" not "Database stuff"
- Create tasks as you discover work during sessions — don't wait until the end
- Review and update task statuses regularly during research to keep the list accurate
- Use high priority sparingly — if everything is high priority, nothing is
