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
  intact; drop them and every block is treated as new. An unsafe or duplicate id is
  replaced rather than trusted.
- A bare `[ ... ]` array is accepted too; the envelope is added for you.
- The document is **normalized server-side**: an unknown `type` and malformed
  `data` are **dropped, never an error**, so a bad payload degrades to fewer
  blocks instead of failing the call. Read back `content` to see what stored.
- `version` is stamped by the server; whatever you send is replaced.

**Validation is forgiving, so check your work.** If a block does not appear, the
field names did not match this catalog.

## Text fields

Text in `paragraph`, `list`, `table`, `quote`, `callout`, `image.caption`,
`mermaid.caption` carries
a restricted inline markdown subset: `**bold**`, `*italic*`, `` `code` ``,
`[label](url)`. Anything else is shown literally.

`[[E3]]`, `[[R2:E5]]`, `[[RM1]]` work inside these fields and **are indexed** —
they land in the `crossrefs` table and become links in the web UI, exactly as in
a markdown entry. References inside `code`, `mermaid` and `html` bodies are **not**
indexed: a snippet mentioning a code is not a citation.

Literal `\n` in a text field becomes a real newline. In `code`, `mermaid` and
`html` it does not — a backslash there is data.

## Block catalog (11 types)

| `type` | `data` |
|---|---|
| `paragraph` | `{ text }` — required; empty ⇒ dropped |
| `heading` | `{ level: 2\|3\|4, text }` — `text` required; an out-of-range or missing level becomes 2. Level 1 is the entry title, which the page renders |
| `list` | `{ style: "unordered"\|"ordered", items: string[] }` — blank and non-string items are dropped, max 200; no items ⇒ block dropped. Unknown style becomes `unordered` |
| `table` | `{ header: bool (default true), rows: string[][] }` — non-array rows dropped, max 200 rows × 20 columns; no rows ⇒ dropped |
| `quote` | `{ text, cite? }` — `text` required |
| `code` | `{ code, language? }` — `code` required, stored verbatim. `language` is lowercased and reduced to `a-z0-9+#-_.`. `language: "mermaid"` is accepted as a spelling of the `mermaid` block below and draws the same way |
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
  to markdown. A `callout` becomes a labelled blockquote, a `mermaid` block becomes
  a ```mermaid fence — which GitHub and this app both draw — and an `html` block
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

400 blocks per document, 20000 characters per text field, 100000 for `code`,
20000 for `mermaid`, 200000 for `html`, 200 list items, 200×20 table cells. Text over a cap is clamped
on a character boundary, not cut mid-character.
