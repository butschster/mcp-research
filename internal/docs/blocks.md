# Block documents

An entry with `entry_type: blocks` stores an article made of typed blocks instead
of markdown. Use it when the entry is a composed document — prose plus an alert,
a table, a chart, a custom visual — rather than a single stream of text.

Keep using `markdown` for ordinary notes. It is the default, it is cheaper to
write, and nothing about it changed.

## The document

`content` is a JSON document:

```json
{
  "version": 1,
  "blocks": [
    { "type": "heading",   "data": { "level": 2, "text": "Findings" } },
    { "type": "paragraph", "data": { "text": "Refers to [[E3]] and **matters**." } },
    { "type": "callout",   "data": { "variant": "warning", "text": "Numbers drift." } }
  ]
}
```

- Every block is `{ "type": string, "data": object }`. Array order is render order.
- Each stored block also carries an `id` — eight lowercase alphanumerics. The
  server fills it in when you omit it and **keeps yours when you send one**, which
  is what lets anything attached to a block (state, a comment, a deep link) survive
  you rewriting the document. Read the document, edit it, send it back with the ids
  intact; drop them and every block is treated as new. An id you send is kept only
  when it matches `^[a-z0-9]{4,16}$` (uppercase is folded down first) and is not
  already used in the document; anything else is replaced rather than trusted.
- A bare `[ ... ]` array is accepted too; the envelope is added for you.
- The document is **normalized server-side**: an unknown `type` and malformed
  `data` are **dropped, never an error**, so a bad payload degrades to fewer
  blocks instead of failing the call. Read back `content` to see what stored.
- `version` is stamped by the server; whatever you send is replaced.

**Validation is forgiving, so check your work.** If a block does not appear, the
field names did not match this catalog.

## Text fields

Text in `paragraph`, `heading`, `list`, `table`, `quote.text`, `callout.text`,
`checklist` item text, `image.caption`, `mermaid.caption` and `html.caption`
carries a restricted inline markdown subset: `**bold**`, `*italic*`, `` `code` ``,
`[label](url)`. Anything else is shown literally. The labels around them —
`quote.cite`, `callout.title`, `checklist.title`, `html.title`, `image.alt` — are
shown as plain text but are indexed like the rest.

`[[E3]]`, `[[R2:E5]]`, `[[RM1]]` work inside these fields and **are indexed** —
they land in the `crossrefs` table and become links in the web UI, exactly as in
a markdown entry. References inside `code`, `mermaid` and `html` bodies are **not**
indexed: a snippet mentioning a code is not a citation.

Literal `\n` in a prose field becomes a real newline. It does not in the one-line
fields above (`heading.text`, `quote.cite`, `callout.title`, `checklist.title`,
`html.title`, `image.alt`), where a real newline is collapsed to a space and a
backslash stays data, nor in `code`, `mermaid` and `html` bodies, which are stored
verbatim.

## Block catalog (12 types)

| `type` | `data` |
|---|---|
| `paragraph` | `{ text }` — required; empty ⇒ dropped |
| `heading` | `{ level: 2\|3\|4, text }` — `text` required; an out-of-range or missing level becomes 2. Level 1 is the entry title, which the page renders |
| `list` | `{ style: "unordered"\|"ordered", items: string[] }` — blank and non-string items are dropped, max 200; no items ⇒ block dropped. Unknown style becomes `unordered` |
| `table` | `{ header: bool (default true), rows: string[][] }` — non-array rows dropped, max 200 rows × 20 columns; no rows ⇒ dropped |
| `quote` | `{ text, cite? }` — `text` required |
| `code` | `{ code, language? }` — `code` required, stored verbatim. `language` is lowercased and reduced to `a-z0-9+#-_.`. `language: "mermaid"` is accepted as a spelling of the `mermaid` block below and draws the same way |
| `checklist` | `{ items: [{key?, text}], title? }` — items may also be plain strings; blank text ⇒ item dropped, max 200; no items ⇒ block dropped. `key` is what a tick hangs on: it is **kept when it matches `^[a-z0-9_-]{1,24}$`** (looser than a block `id` — `k1`, `backup`, `step_2` all survive; uppercase is folded down) **and is unused inside the block**, and minted otherwise. A replaced key is a lost tick. Ticks live in `data.state` (`{key: true}`) which is **server-owned: any `state` you send is dropped**. Tick with `entry_patch` `set_state`, or by clicking in the web UI |
| `mermaid` | `{ code, caption? }` — `code` is a mermaid source, required, stored verbatim, max 20000 chars. Drawn as a diagram: pan, zoom, fullscreen, and a link that reopens it in mermaid.live. A source mermaid cannot parse falls back to the source with a link to the editor, which reports the error |
| `callout` | `{ variant: "info"\|"warning"\|"success"\|"danger", text, title? }` — `text` required; an unknown variant becomes `info` |
| `divider` | `{}` — always kept |
| `image` | `{ url, alt?, caption? }` — `url` must be `http(s)://` or a domain-relative `/path`; anything else (including `javascript:`, `data:` and protocol-relative `//host`) ⇒ dropped |
| `html` | `{ html, title?, caption? }` — a complete self-contained HTML document rendered in a sandboxed iframe sized to its own height. `html` required |

Every block sits in the reading column. There is no width tier: a `width` field in
a payload is simply not stored, and a wide table scrolls inside its own container
rather than breaking the column.

## The html block

This is what the standalone `artifact` entry type became. Scripts run: the frame
is sandboxed with `allow-scripts` and **without** `allow-same-origin`, so the
document cannot read cookies, storage or the host page, and cannot call the API.

- Send one whole document — `<!doctype html>`, `<head>`, `<style>`, `<script>`,
  `<body>`. Inline everything; the frame cannot fetch external resources.
- **Close the script tag as `</script>`, never `<\/script>`.** The escaped form is
  only needed when HTML sits inside a JavaScript string literal. Here it is plain
  content in a JSON field, so the backslash makes the tag invalid: the browser
  never closes the script, swallows the rest of the document, and the frame renders
  blank with no error anywhere. This is the most common reason an otherwise correct
  artifact comes out empty.
- The frame reports its own height, so it is shown in full with no inner
  scrollbar. Do not write `body { height: 100% }` expecting a viewport — lay out
  for a document that grows. A document sized in `vh` grows on every report and
  the host stops it after twenty rounds.
- Read-only context arrives after load as `window.researchData` and a
  `research-data` event: the research, the entry, and the section list.
- `title` and `caption` are the author's framing around the frame, and they are
  indexed; the HTML body is not.

## Editing one block

`entry_update` replaces the whole document. `entry_patch` changes part of it,
addressing blocks by `id`:

```json
{ "entry_id": "…", "rev": "9f2c1a…", "ops": [
  { "op": "update", "id": "a1b2c3d4", "data": { "text": "Rewritten." } },
  { "op": "insert", "type": "paragraph", "data": { "text": "New." }, "after": "a1b2c3d4" },
  { "op": "delete", "id": "e5f6a7b8" },
  { "op": "move",   "id": "c3d4e5f6", "at": "start" },
  { "op": "set_state", "id": "b2c3d4e5", "item": "k1", "checked": true }
]}
```

- Ops apply **in order against one working copy**, so an op may target a block an
  earlier op inserted. Give the new block an `id` and the next op can point at it —
  but only an id the server would keep (`^[a-z0-9]{4,16}$`, unused in the document);
  an id it replaces leaves the next op pointing at nothing, which fails the patch.
- **A block id is the only handle.** There is no way to address a block by
  position, and no index or offset a patch will accept.
- Placement is `after` / `before` / `at: "start"|"end"`, at most one. None means
  append. Two anchors is an error, not a precedence rule.
- **A patch is strict.** An unknown block id, an unknown type, or data a type
  rejects fails the whole patch and writes nothing. This is the opposite of a
  whole-document write, and deliberately so: there you sent the article and can
  read back what stored, here you did not, so a silently dropped op would be
  indistinguishable from success. A patch that would empty the document, or push it
  past 400 blocks, is refused the same way.
- An `update` op replaces the block's `data` entirely — send every field you want
  to keep. It carries the block's ticks forward as long as the block keeps its
  type; changing `type` drops them.
- `rev` is the document revision. Every `entry_patch` result carries the one it
  just wrote, so a chain of patches can pass it along and each is rejected if
  someone else wrote in between. `entry_read` does not return a `rev` today, so a
  first patch leaves it null and relies on strictness. `set_state` never needs one:
  it names one item and cannot clobber prose.
- `set_state` sets an absolute value, never a toggle, and only on a `checklist`
  block whose item key exists.
- The tool refuses an entry that is not `entry_type: blocks`, and `entry_update`'s
  `text_replace` refuses a blocks entry — running a string replacement over the
  stored JSON is how a document stops parsing. Patch it, or send the whole
  document in `content`.
- Over REST the same ops are `POST /api/entries/{id}/patch` with `{rev, ops}`,
  which answers `409` when `rev` does not match and returns the new `rev` beside
  the entry. The checkbox in the web UI posts there too.

## Keeping a human's ticks

A checklist is ticked by a person and rewritten by an agent, so the server keeps
the two apart:

- Ticks live beside the block, never in what an author sends. A `state` field in
  your payload is dropped.
- A whole-document `entry_update` **carries ticks forward** by block `id` and item
  `key`. Read the entry, edit it, send it back with both intact and nothing is
  lost. A block whose `type` changed loses its ticks even when the id survives.
- If you drop the ids, the server falls back to matching **item text**, compared
  case-insensitively and with runs of whitespace collapsed. Reword an item and its
  tick is gone; keep the wording and it survives a regenerated document.
- What the write tells you: `entry_patch` returns `blocks`, `state_preserved` and
  `state_lost`; the REST write responses carry a `block_report` with
  `blocks_reidentified` beside those. **MCP `entry_update` returns none of them**,
  so on a whole-document rewrite a lost tick is silent — which is the reason to
  echo the ids and keys back rather than to check afterwards.

## Entry title and description

Omit `title` and it is taken from the first `heading`, else the first sentence of
the first `paragraph`, else a lone `html` block's `title`. If none of those exist
the call fails — a document with nothing to name itself by is a mistake, not a
default. `description` falls back the same way to the first paragraph that is not
already the title, or an `html` block's `caption`.

## The artifact alias

`entry_type: artifact` still works: pass a bare HTML document as `content` and it
is stored as a blocks document with one `html` block. The document's
`<title>` and `<meta name="description">` are lifted onto the block, so it names
itself as before. Nothing stores the type `artifact` any more — reading such an
entry back returns `entry_type: blocks`.

## Export and import

- **Markdown export** (`?format=md`, research and session): blocks are serialized
  to markdown. A `callout` becomes a labelled blockquote, a `checklist` becomes a
  GitHub task list with the ticks as they stand, a `mermaid` block becomes a
  ```mermaid fence — which GitHub and this app both draw — and an `html` block
  becomes a named note saying to view it in the web UI, its document being the one
  thing that cannot be markdown.
- **JSON export**: a blocks entry also carries `content_markdown` beside
  `content`, so a reader does not need to know the block format.
- **Portable export/import**: `entry_type` travels in the file. A file written
  before block documents existed has no `entry_type` and imports as markdown.

## Changing type

Switching an entry to `blocks` requires `content` in block form in the same call
— wrapping a markdown document in one paragraph would silently discard its
structure, so it is refused instead. The other direction, `blocks` → `markdown`,
converts what is stored.

## Caps

400 blocks per document, 200 list or checklist items, 200×20 table cells. 20000
characters per text field — but 300 for a `heading`, 5000 for a `callout`, 1000 for
a caption or an `alt`, 2000 for an `image.url`, 40 for a `language` — and 100000
for `code`, 20000 for `mermaid`, 200000 for `html`. Text over a cap is clamped on a
character boundary, not cut mid-character.
