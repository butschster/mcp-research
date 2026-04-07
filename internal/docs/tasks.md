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

## Best Practices

- **Always set `result` when completing or failing** — it captures the outcome and becomes part of the research record
- Use `blocked` with a clear reason in description when waiting on something
- Task results support markdown and `[[...]]` cross-references — link to entries that were created as a result
- Keep task titles actionable and specific: "Benchmark SQLite WAL modes" not "Database stuff"
- Create tasks as you discover work during sessions — don't wait until the end
- Review and update task statuses regularly during research to keep the list accurate
- Use high priority sparingly — if everything is high priority, nothing is
