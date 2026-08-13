# Export

How to export research data as documents. Two scopes exist: a whole research, and a single session. Both come in a JSON form (structured data + pre-rendered markdown) and a raw `.md` download.

## Export Pages (Web UI)

| Scope | Page | Reached from |
|-------|------|--------------|
| Research | `/research/{code}/export` | Export button on the research page |
| Session | `/research/{code}/session/{sessionCode}/export` | Export button on the session page |

Both pages render a single printable document and expose the same two actions:

- **Download .md** — saves the markdown built by the server.
- **Print / PDF** — opens the browser print dialog (Ctrl+P / Cmd+P). Choose "Save as PDF". Navigation, footer, and interactive elements are hidden in print.

The research page renders: research header (title, goal, description, tags), table of contents with anchor links, all sections with full entry content, all sessions with questions and answers, all tasks with statuses and results.

The session page renders: parent research name, session title, focus, code, status, notes, all questions with answers, and the entries produced during that session.

### Print Individual Entries

Open any entry page and use Ctrl+P / Cmd+P. The entry page has print-optimized CSS that hides navigation, action buttons, cross-references, and related entries, showing only the breadcrumb path and document content.

## Research Export API

```
GET /api/researches/{id}/export
```

`{id}` accepts a UUID or a research short code (`/api/researches/R1/export`).

Returns all research data as structured JSON:

```json
{
  "research": { "name": "...", "goal": "...", ... },
  "sections": [
    {
      "name": "...",
      "display_name": "...",
      "entries": [
        { "title": "...", "entry_type": "markdown", "content": "full markdown...", "tags": [...] }
      ]
    }
  ],
  "sessions": [
    {
      "title": "...",
      "questions": [
        { "text": "...", "answer": "...", "status": "answered" }
      ]
    }
  ],
  "tasks": [
    { "title": "...", "status": "completed", "result": "..." }
  ],
  "markdown": "# Full document as markdown string..."
}
```

### Download Markdown File

```
GET /api/researches/{id}/export?format=md
```

Returns `text/markdown` with `Content-Disposition: attachment`, filename derived from the research name (spaces become `_`, `/ \ : " '` are stripped or replaced).

The markdown document structure:
1. Title and metadata (goal, description, tags)
2. Table of contents
3. Sections with full entry content
4. Sessions with all questions and answers
5. Tasks with statuses and results

## Session Export API

```
GET /api/researches/{id}/sessions/{sessionId}/export
```

Exports one session instead of the whole research. Both path segments accept a UUID or a short code, so `/api/researches/R1/sessions/SS1/export` works.

Returns JSON:

```json
{
  "research": { "name": "...", "code": "R1", ... },
  "session": { "title": "...", "code": "SS1", "focus": "...", "status": "active", "notes": "..." },
  "questions": [
    { "text": "...", "area": "...", "rationale": "...", "priority": "high", "status": "answered", "answer": "..." }
  ],
  "entries": [
    { "title": "...", "code": "E4", "section_id": "...", "entry_type": "markdown", "content": "full markdown...", "tags": [...] }
  ],
  "section_names": { "<section-id>": "Display name of the section" },
  "markdown": "# Session as markdown string..."
}
```

- `questions` — every question of the session, unfiltered, ordered by `position`.
- `entries` — only entries whose `session_id` equals this session, with full content. `entry_create` links the entry to the research's active session automatically when `session_id` is not given, so entries created during an interview show up here without extra work.
- `section_names` — map of section ID to display name (falls back to the slug `name`), so the client can label each entry without a second request.

### Download Markdown File

```
GET /api/researches/{id}/sessions/{sessionId}/export?format=md
```

Returns `text/markdown` with `Content-Disposition: attachment`, filename derived from the session title.

The markdown document structure:
1. Session title, parent research name, code, status
2. Focus as a blockquote, then `## Notes` if the session has notes
3. `## Questions (N answered of M)` — one `### Q:` block per question with area, priority, status, rationale, and answer
4. `## Entries produced in this session` — each entry with its code, section, tags, status, and full content (omitted when the session produced no entries)

## Portable Export / Import

The markdown/JSON exports above are for reading. To move a research to another server, use the portable format instead — it carries sections, entries, sessions, questions, tasks, and roadmaps in a versioned envelope (`version`, `exported_at`, `research`):

```
GET  /api/researches/{id}/export/portable   -> portable JSON (downloaded as <name>.json)
POST /api/researches/import                 -> body is that JSON, returns the new research_id and code
```

The `research_export` MCP tool returns the same portable payload for a research ID or short code.

Import re-creates entities from scratch: new UUIDs, new short codes, cross-references re-parsed from the imported content.

## Auth

Export endpoints are read endpoints: unauthenticated by default, but they require a bearer token (JWT or API key) when `auth_enabled` is set, and they only ever see the caller's own researches. `POST /api/researches/import` is a write endpoint and always requires the bearer token when `api_token` or `auth_enabled` is configured.

## Artifact Entries in Export

An entry with `entry_type: artifact` holds a whole HTML document instead of markdown, and each export treats it accordingly:

- **Markdown export** (`?format=md`, and the `markdown` field of the JSON responses) writes the document inside a fenced ```` ```html ```` block, preceded by the line `*HTML artifact — render the document below to view it.*`. Pasted inline it would leak its own `<style>` and `<script>` into whatever renders the export. The fence grows past any run of backticks inside the document, so an artifact containing fences cannot break out of the block.
- **JSON responses** (`GET .../export` and `GET .../sessions/{sessionId}/export` without `format=md`) carry `entry_type` next to `content` on every entry. Branch on it before rendering: an artifact's `content` is a raw HTML document, not markdown.
- **Export pages (Web UI)** render artifacts in the same sandboxed iframe the entry page uses, not as markdown.
- **Portable export** carries `entry_type` on each entry, so artifacts import as artifacts. Exports written before artifacts existed have no `entry_type`; those entries import as `markdown`, which is what they were.

## Cross-References in Export

Cross-references (`[[E3]]`, `[[R2:E5]]`, `[[RM1]]`) are preserved as-is in markdown export. In the web export pages, they are rendered as clickable links.
