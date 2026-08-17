# Writing an Artifact

An artifact is a complete HTML document you author and this app renders inside a
sandboxed iframe: your own layout, your own CSS, your own scripts and charts,
sized to its content and printable as part of the entry.

Use one when the finding **is** a picture — a component map, a comparison table
that wants real columns, a chart, a decision grid. Prose belongs in markdown or
in `paragraph` blocks; an artifact that is a styled essay is a worse essay.

The block format is in [Block Documents](/llms/blocks.md). This is how to write
the document that goes inside it.

---

## The rule that catches everyone: one theme

**Never write both a light and a dark palette. Pick one and hard-code it.**

Do not do this:

```css
/* WRONG — two themes in one artifact */
:root       { --bg:#fff;     --fg:#15171c; }
@media (prefers-color-scheme: dark) {
  :root     { --bg:#131519;  --fg:#e9ebef; }
}
```

The reason is not taste. The iframe is a separate document with its own media
queries, and `prefers-color-scheme` is not guaranteed to resolve the same way it
did on screen — printing is the case where it reliably does not. A two-theme
artifact then renders its **light** palette while the host still paints a dark
surface behind the frame, and the result is dark text on a dark page: headings
vanish, body text goes muddy. That exact bug shipped, and it is invisible until
someone exports a PDF.

The host reads dark. Write dark, unconditionally:

```css
/* RIGHT — one palette, no query */
:root {
  --bg:#0c1220; --surface:#151d2e; --fg:#e2e8f0; --muted:#7f8ea3;
  --line:rgba(148,163,184,.18); --accent:#6cc5e0; --warn:#f0b849;
}
body { background: var(--bg); color: var(--fg); }
```

Those values are the host's own tokens, so an artifact using them sits in the
page instead of on top of it. You may depart from them deliberately — a chart
needs its own series colours — but depart in one direction only.

If you genuinely want a light artifact, make it light in every context: light
background, dark text, no query. It will look like a sheet of paper laid on the
page, which is a legitimate choice. What must never happen is the palette
depending on where it is rendered.

---

## What the frame gives you, and what it withholds

The document is loaded through `srcdoc` with `sandbox="allow-scripts"` and
**without** `allow-same-origin`, at an opaque origin.

**You have:** JavaScript, `<canvas>`, SVG, CSS animation, `requestAnimationFrame`,
inline data URIs, and the full height of the entry.

**You do not have:** cookies, `localStorage`, `sessionStorage`, the host page's
DOM, the API, or any credential. Reading any of those throws or returns nothing.
That is deliberate — an artifact is untrusted content authored by a model.

**Send one self-contained document.** `<!doctype html>`, `<head>`, `<style>`,
`<script>`, `<body>`, everything inlined. Nothing blocks an external `<script
src>` at the browser level, but do not use one: the same document is written to
a real `.html` file by the Obsidian export and is read where that host is not
reachable. A CDN link turns the artifact into a blank box the first time someone
opens it offline.

### Close the script tag as `</script>`

Not `<\/script>`. The escaped form belongs inside a JavaScript string literal;
here the HTML is plain content in a JSON field, so the backslash makes the tag
invalid. The browser never closes the script, swallows the rest of the document,
and the frame renders blank with no error anywhere. This is the single most
common reason an otherwise correct artifact comes out empty.

---

## Size: lay out for a document, not a screen

The frame has no fixed height. An injected reporter measures the document and
tells the host how tall to be, on load, on font load, on resize, and on every
mutation — so the artifact appears in full with no inner scrollbar.

That inverts one habit:

- **Do not size anything to the viewport.** `height:100vh`, `min-height:100vh`,
  `100dvh` — the frame *is* the viewport, so the document grows, reports a
  larger height, gets a taller frame, and grows again. The host stops the loop
  after twenty rounds and freezes the height wherever it happened to be.
- **Do not centre with `position:fixed`** for the same reason.
- Let the content set the height. Flow down the page as a document would.

A hard ceiling of 60000px exists as a backstop; nothing real should approach it.

---

## Horizontal scrolling clips on paper

`overflow-x: auto` with a wide child works on screen — the reader drags it. In a
PDF there is nothing to drag, so whatever is off to the right is simply gone,
and the scrollbar prints as a grey bar.

If a diagram is wider than the page, either scale it to fit
(`svg { width:100%; height:auto }` with a `viewBox`, no `min-width`), or split it
into two stacked diagrams. Reserve a horizontal scroller for material where
losing the right-hand side is acceptable.

---

## Printing

The entry page prints, and the artifact is the one part of it that keeps its own
look — it is a document someone authored, not part of our chrome. The host
handles the mechanics: colour adjustment is forced for every element, the frame
loses its border, and the block avoids being split across pages.

What is left to you:

- One theme (above). This is the whole game.
- No horizontal scrollers you cannot afford to lose.
- Nothing that only exists on hover or after a click. A print shows the initial
  state; a chart that draws itself on `mouseover` prints empty.
- Animations are fine — they print at whatever frame they reached, so make the
  first frame the meaningful one rather than opacity zero.

---

## Reading the research

After load the frame receives read-only context, both as `window.researchData`
and as a `research-data` event:

```js
window.addEventListener('research-data', (e) => render(e.detail))
if (window.researchData) render(window.researchData)   // already arrived
```

The payload:

```json
{
  "research": { "id": "…", "code": "R1", "name": "…", "goal": "…" },
  "entry":    { "id": "…", "code": "E7", "title": "…", "tags": ["…"] },
  "sections": [ { "id": "…", "name": "Findings" } ]
}
```

Both forms exist because the timing is not yours to control — bind the listener
*and* check the global, as above.

**It can be absent.** On a shared research the bridge is `null`, and the artifact
still has to render. Treat the context as decoration: never make the document's
main content depend on it. Data the artifact is *about* belongs inlined in the
document, not fetched from the bridge.

---

## Style

The artifact sits inside an entry, under a title the reader has already read.

- **One idea per artifact.** Two unrelated diagrams are two artifacts, and each
  gets its own caption and its own place in the document.
- **Lead with the answer.** The first thing on screen is the finding, not a
  legend or a title repeating the entry's own.
- **Label with numbers.** "~5 min", "2 × 2", "5–6 consecutive" — an artifact
  earns its space by carrying quantities prose would bury.
- **Use the host's type scale loosely:** 13–15px body, 21–23px for the one
  heading, uppercase 12px with letter-spacing for section labels. Do not ship
  28px headings; the entry page already has the big type.
- **Borders over fills.** A 1px `rgba(148,163,184,.18)` outline reads on screen
  and on paper; a large saturated fill fights both.
- **Say what a colour means.** If green means "exists" and amber means "needs
  work", put that legend in the artifact — the reader will not infer it, and the
  PDF has no tooltip.
- **No fixed pixel widths on the outer container.** The entry column is narrower
  on a phone and wider in fullscreen; use `max-width` and let it breathe.

---

## Checklist before sending

- [ ] One palette, no `prefers-color-scheme` anywhere
- [ ] One self-contained document, nothing loaded from the network
- [ ] `</script>` closed plainly
- [ ] No `100vh`, no `position: fixed`
- [ ] Wide content scales rather than scrolls sideways
- [ ] Renders correctly with `window.researchData` undefined
- [ ] The first painted frame is the meaningful one
