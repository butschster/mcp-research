# Export

How to export research data as documents. Two scopes exist: a whole research, and a single session. Both come in a JSON form (structured data + pre-rendered markdown) and a raw `.md` download. A whole research has two further forms: an **Obsidian vault** (`?format=obsidian`, a zip of linked notes — a document to read) and the **portable JSON** (a research to move to another server). Everything here is a read endpoint; only the portable import writes.

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
  "roadmap_count": 2,
  "markdown": "# Full document as markdown string..."
}
```

`roadmap_count` is a number, not the roadmaps: this payload is already the whole
research, and the count exists so a client can tell whether the roadmap option of
the vault export applies at all. Use `roadmap_list` / `GET /api/researches/{id}/roadmaps`
to read them.

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
- `entries` — only entries whose `session_id` equals this session, with full content. `entry_create` links the entry to the research's active session automatically when `session_id` is not given, so entries created during an interview show up here without extra work. An entry the session **edited** without creating is not here; `GET /api/sessions/{id}/changes` covers those, with the revision range and a diff — see [Revisions](/llms/revisions.md).
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

## Obsidian Vault Export

```
GET /api/researches/{id}/export?format=obsidian
```

The same route as the JSON export, behind `format=obsidian`. Returns
`application/zip`: the research as a folder of cross-linked markdown notes, ready
to unzip into an [Obsidian](https://obsidian.md) vault. Any editor that reads
markdown can open it; only the links are Obsidian-flavoured. The archive is built
whole before the first byte is sent, so a failure is still a status code rather
than a truncated zip.

The archive has no wrapper folder — its entries are at the root, because every
extractor already creates a folder named after the zip. On the command line use
`unzip -d "R1 — Competitive Landscape" file.zip`, or the files land in the
current directory.

```
R1 — Competitive Landscape.zip
├── README.md                     goal, description, memory, instruction, and a
│                                 linked table of contents of every entry,
│                                 session, task and roadmap in the vault
├── 01 — Market Analysis/         one folder per section, numbered in section order
│   └── E1 — Pricing model.md
├── 02 — Competitors/
├── Unfiled/                      entries whose section is gone (usually absent)
├── Sessions/SS1 — Initial exploration.md
├── Tasks/T1 — Verify the claim.md
├── Roadmaps/RM1 — Migration plan.md      nodes and edges as a mermaid diagram
├── _html/E2 — Revenue chart (13b3d97a).html
├── _history/E1.md                only with revisions=true
└── _Unresolved/R2 E5.md          one note per reference pointing outside the export
```

**A cross-reference is retargeted at the note's filename and keeps its code as
the display text**: `[[E3]]` is exported as `[[E3 — Pricing model|E3]]`. The
reader sees exactly what the author wrote, and the link resolves — so backlinks,
unlinked mentions and the graph view work.

Obsidian's `aliases` do **not** resolve a hand-written link. An alias makes a
note findable in the quick switcher, in search and in link autocomplete, where
choosing a suggestion inserts a link to the *filename* with the alias as display
text; a `[[E3]]` typed into a file is matched against filenames only, and left
alone it renders as an unresolved link that offers to create an empty note.
Aliases are still written, for the switcher.

The retargeting happens on the way out and never in storage — the same kind of
rendering as a block document becoming markdown. Re-exporting rebuilds every link
from the current titles, so renaming an entry cannot leave a stale link behind.

- **Entries, sessions, tasks and roadmaps each get a note**, because each is a
  link target. A node link (`[[RM1:N3]]`) lands on its roadmap note, and
  `[[R1:E1]]` — the qualified form of a reference inside the same research —
  lands on the same note as `[[E1]]`.
- **Every note links home**, with a relative markdown link to `README.md` rather
  than a wikilink: two exports unzipped into one vault hold two `README.md`
  files, and a name could not tell them apart.
- **A reference to something outside the export** — another research, a deleted
  entry, a folder the options omitted (`[[SS1]]` with `sessions=false`, `[[T1]]`
  with `tasks=false`) — gets a stub under `_Unresolved/` explaining what it points
  at. The link is a real reference, so it stays visible as one rather than
  dangling silently. Only text shaped like a code (`E12`, `R2:E5`, `RM1:N3`)
  qualifies, and at most 200 stubs are written.
- **An entry's frontmatter** carries `code`, `title`, `aliases`, `research`,
  `section`, `type` (`markdown` or `blocks`), `status`, `tags`, `created`,
  `updated`, and the `session` that produced it. With `revisions=true` it also
  carries the current `revision` and its `author` (`agent`, `human`, `import`,
  `restore`); without that option neither key is present, whatever history the
  entry has. Empty values are dropped rather than written blank. A session note
  carries `code`, `title`, `aliases`, `research`, `focus`, `status`, `created`,
  `updated`; a task also `priority` and `completed`; a roadmap also `statuses`.
  Frontmatter is written with a YAML encoder, so a title containing `:` or a
  quote still parses.
- **Filenames** are `E1 — Title` and keep Cyrillic and other non-ASCII. What a
  filesystem refuses is removed (`: * ? " < > |`), a path separator becomes `-`,
  newlines and control characters become spaces, Windows reserved names get a
  trailing `_`, and a component longer than 100 bytes is cut on a rune boundary.
  The code leads the name, so two entries with the same title cannot collide.
- **Block documents** render through the same markdown projection as the `.md`
  export, so checklists keep their ticks and mermaid stays a fence — which
  Obsidian draws. The live-editor link under a diagram is omitted here for that
  reason.
- **An `html` block becomes a real file** under `_html/` with a callout linking
  to it. Obsidian sanitizes inline HTML and runs no scripts, so an inlined
  artifact would be a broken shell of itself; a separate file opens in a browser
  and is still the artifact.
- **A tag Obsidian cannot index** (one containing a space, a tab or `#`) stays
  exact in the frontmatter, and a line at the end of the note names the tags that
  will not appear in the tag pane. The vocabulary is the author's; it is not
  mangled to fit a reader.

Query options — everything is included by default except revisions:

| Param | Default | Effect |
|---|---|---|
| `sessions` | `true` | `false` omits `Sessions/` and the session link in each entry's footer. The `session` frontmatter key stays: it is provenance, not a link |
| `tasks` | `true` | `false` omits `Tasks/` |
| `roadmaps` | `true` | `false` omits `Roadmaps/` |
| `html` | `true` | `false` omits `_html/`; the callout stays and says the file was left out |
| `revisions` | `false` | `true` adds `_history/{code}.md` for every entry that has history — a table of revision, date, author, session and summary — plus a link to it in the entry's footer, and the `revision` / `author` frontmatter keys |

`false`, `0`, `no`, `off` and `n` turn an option off; `true`, `1`, `yes`, `on`
and `y` turn it on (case-insensitive). Anything else keeps the default rather
than guessing, and a mistyped flag never turns the export into an error.

### From the MCP tool

`research_export` takes an optional `format`:

| Value | Result |
|---|---|
| absent, `null`, `""`, `portable`, `json` | the portable JSON payload below — unchanged, and the input for `research_import` |
| `obsidian`, `vault`, `zip` | a JSON object describing the vault download |
| anything else | a validation error: `format must be portable or obsidian` |

A tool result cannot carry a zip, so `format: "obsidian"` returns the link
instead:

```json
{
  "format": "obsidian",
  "url": "https://host/api/researches/R1/export?format=obsidian",
  "research": "Competitive Landscape",
  "contains": "README.md with a linked table of contents, a folder per section, ...",
  "links": "Cross-references are retargeted at the notes they name and keep their codes as display text ...",
  "auth": "The link needs the same credentials as the REST API ...",
  "description": "Download the zip and unzip it into a folder in an Obsidian vault. Options: ..."
}
```

- The research is resolved before the link is built, so a bad `research_id` is an
  error now rather than a link that 404s later.
- `url` is absolute when the server has a base URL configured; otherwise it is the
  bare path `/api/researches/R1/export?format=obsidian` and `description` says so.
  Do not invent an origin for it.
- The URL is a normal read endpoint: it needs the API token or JWT as an
  `Authorization: Bearer` header. Pasting it into a browser tab only works on a
  server running without auth. Say that when you hand the link to a user.
- The query options above apply to that URL; the tool has no parameters for them.

**The vault does not travel back.** It is a document format; use the portable
JSON below to move a research between servers.

## Portable Export / Import

The markdown/JSON exports above are for reading. To move a research to another server, use the portable format instead — it carries sections, entries, sessions, questions, tasks, and roadmaps in a versioned envelope (`version`, `exported_at`, `research`):

```
GET  /api/researches/{id}/export/portable   -> portable JSON (downloaded as <name>.json)
POST /api/researches/import[?team={id}]     -> body is that JSON, returns the new research_id and code
```

The `research_export` MCP tool returns the same portable payload for a research ID or short code — that is what `format` defaults to.

Import re-creates entities from scratch: new UUIDs, new short codes, cross-references re-parsed from the imported content.

**Where the import lands.** The new research goes into the caller's personal team unless another one is named: `?team={id}` on the REST route, `team_id` on the `research_import` tool. Naming a team you are not in is `not found`; naming one where you are only a `viewer` is refused with `your role in this team does not allow this`. Ownership does not travel with the payload — the export carries no team and no user.

**No skills travel with an export either.** Which [skills](/llms/skills.md) a research follows is not in the portable payload and not in any of the reading formats, so an imported research follows nothing and its team library is not reconstructed. Re-attach them on the destination server.

**No template provenance either.** `template_slug` and `template_version` are on the research record but not in `ExportResearch`, so an import lands with no methodology recorded — and could not honour one anyway: a slug names a row in *this* server's template library, and the destination may have a different set, a fork under the same slug, or none. The stamp says which methodology this research was started from on this server; it is not a portable reference. If it matters to the reader, write it into `instruction` or an entry, which do travel. See [Templates](/llms/templates.md).

**No history travels with an export.** Revisions are not in the portable payload, and every entry an import creates starts at revision 1 attributed to `import` rather than to an agent that never wrote it. Export a research, import it elsewhere, and who wrote what before the export is only in the original server. The vault's `_history/` tables (`revisions=true`) are a readable record, not a transferable one — nothing imports them back.

## Export Through a Share Link

A [read-only share link](/llms/domain-guide.md#share) may carry the research export, and only if the link was created with `include.export`:

```
GET /api/shared/{token}/researches/{id}/export
GET /api/shared/{token}/researches/{id}/export?format=md
GET /api/shared/{token}/researches/{id}/export?format=obsidian
```

Same handler, same payload shape as the authenticated route above, with three differences that are the point of the feature:

- `instruction` and `memory` are absent, along with `user_id` and every team field. They are stripped on the read path every export goes through, so there is no format that carries them out.
- `sessions` is empty unless the link includes sessions, and `tasks` unless it includes tasks. An export that carried the interview transcript would hand over in one file exactly what the creator chose to leave out of the pages.
- `/export/portable` is not mounted under the prefix at all: it is a re-importable copy of the record rather than a reading of it. **Session export is not shared either** — only the research-scoped route is mounted.

### The vault through a share

`?format=obsidian` works on a link that includes downloading, and the archive is narrowed to what that link publishes. The narrowing happens in `service.clampForShare`, next to the code that reads the options rather than in the handler — the vault's parts come from a query string the visitor can type, and the include flags gate *routes*, which is a check the vault never passes through.

| Requested | What a visitor gets |
|---|---|
| `sessions=true` | only if the link includes sessions |
| `tasks=true` | only if the link includes tasks |
| `roadmaps=true` | only if the link includes roadmaps |
| `revisions=true` | never — no flag publishes an entry's history |
| — | the `session:` key in an entry's frontmatter is dropped, and the `Session:` footer link with it |

Revisions and provenance are refused outright rather than gated, for the same reason the shared entry pages omit them: who edited what, when, and from which session is working process, like `instruction`.

The research itself needs no special handling — the vault builds from `ResearchService.Get`, which is where `instruction` and `memory` are redacted, so it starts from the published research. This route was a `404` for a while on the belief that it did not; it is `internal/api/share_routes_test.go` that now settles the question, by unzipping the response and asserting against filenames.

Without `include.export` the route answers the same `404 this link is no longer available` as a revoked token: a link that does not offer a download should look like a server that has none.

## Auth

Export endpoints are read endpoints: unauthenticated by default, but they require a bearer token (JWT or API key) when `auth_enabled` is set, and they only ever see researches owned by a team the caller belongs to — a research in someone else's team is `404`, indistinguishable from one that does not exist. **Exporting needs no more than read access**: a `viewer` may export a whole research, a session, or the Obsidian vault, exactly as an `editor` can. `POST /api/researches/import` is a write endpoint: it always requires the bearer token when `api_token` or `auth_enabled` is configured, and it needs editor or owner rights in whichever team it imports into.

A share token is not a bearer token. It authenticates nothing on the routes above — it only opens the mirrored, redacted export under `/api/shared/{token}/…`, and only when the link includes it.

## Block Documents in Export

An entry with `entry_type: blocks` stores a JSON document of typed blocks (see
[Block Documents](/llms/blocks.md)), so every export has to render it rather than write
`content` out:

- **Markdown export** (`?format=md`, and the `markdown` field of the JSON
  responses) serializes the blocks: headings, lists, tables, quotes and code
  become their markdown equivalents, a `callout` becomes a labelled blockquote, a
  `mermaid` block becomes a ```mermaid fence with a link to mermaid.live below it, and a `checklist` becomes a GitHub
  task list carrying the ticks as they stand.
- **An `html` block is named, not emitted.** The export gets
  `*<title> — interactive HTML, view in the web UI.*` and its caption. A whole
  HTML document in a markdown file is not readable, is not markdown, and inlined
  it would leak its `<style>` and `<script>` into whatever renders the file. The
  same applies to legacy `artifact` rows written before the type was folded into
  the `html` block.
- **The Obsidian export is the exception on both counts.** One serializer renders
  a block document everywhere, with two overrides for the vault: the mermaid.live
  link under a diagram is dropped (Obsidian draws the fence itself), and an `html`
  block is written as a real file under `_html/` — wrapped in `<!doctype html>`
  when the block is a fragment — with a callout in the note linking to it. It is
  the one export where the HTML survives.
- **JSON responses** (`GET .../export` and `GET .../sessions/{sessionId}/export`
  without `format=md`) carry `entry_type` next to `content`, and for a blocks
  entry also `content_markdown` — the same rendering the `.md` file uses. Read
  that field and you need no knowledge of the block format.
- **A document that cannot be parsed** yields
  `*This entry holds a block document that could not be read.*` rather than the
  raw JSON or a silent gap.
- **Portable export** carries `entry_type` on each entry, so a blocks entry
  imports as a blocks entry, and the import validates the document before writing
  anything. Exports written before entry types existed have no `entry_type`; those
  entries import as `markdown`, which is what they were.

## Cross-References in Export

Cross-references (`[[E3]]`, `[[R2:E5]]`, `[[RM1]]`) are preserved as-is in markdown export. In the web export pages, they are rendered as clickable links.

The Obsidian vault is the one export that rewrites them: each reference is retargeted at the note's filename and keeps its code as the display text (`[[E3]]` becomes `[[E3 — Pricing model|E3]]`), because Obsidian matches a wikilink against filenames and an `aliases` entry does not resolve a hand-written one. See [Obsidian Vault Export](#obsidian-vault-export) above for the full rule and what happens to a reference pointing outside the export.
