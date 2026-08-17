---
slug: writing-artifacts
name: Writing an artifact
description: Use when about to write an html block or an artifact entry — a chart, a dashboard, an interactive layout the block catalog cannot express.
ambient: true
---

# Writing an artifact

An artifact is a self-contained HTML document rendered in a sandboxed iframe
inside an entry. It is for the finding that is genuinely visual — a chart, a
comparison a table cannot carry, a small interactive model. It is not for
decorating prose that was fine as prose.

## The one rule people get wrong

**Write one theme and hard-code it.**

Do not ship both a light and a dark palette. An artifact that switches on
`prefers-color-scheme` flips to its light branch when the page is printed —
because print does not match the dark query — and then prints dark text on the
dark surface the frame sits on. The result is unreadable and the author never
sees it, because it only happens on paper.

Pick the palette that suits the content, set an explicit background and an
explicit text colour on `body`, and stop there.

## What the sandbox gives you

- Inline `<style>` and `<script>` run.
- **No network.** No CDN, no external stylesheet, no font, no remote image, no
  fetch. Everything is inlined or it is not there. A chart library you cannot
  embed is a chart you draw in SVG yourself.
- No access to the parent page, no cookies, no storage.

## Sizing

**The document sizes itself; the frame follows.** Never lay out against the
viewport — `100vh` inside the iframe means the frame's height, which is being
decided by your content, and the two chase each other. Let the content set the
height and let the host resize to it.

Wide content — a table, a diagram, a code block — scrolls inside its own
container with `overflow-x: auto`. The page body must never scroll sideways.

## Data

A read-only `researchData` bridge may be delivered to the document after load.
**It may also never arrive** — a share link, an export, a stale render. So:
render something correct with no data, then enrich if it appears. An artifact
that shows an empty frame until a message lands is broken most of the times
anyone looks at it.

## Style

Match the product it lives in: restrained, typographic, no gradients or heavy
shadows. Label axes. Give every number a unit. If the chart needs a paragraph to
explain what it shows, write the paragraph as a normal block above it — the
artifact should carry the picture, not the argument.

Keep it under a few hundred lines. An artifact nobody can read the source of is
an artifact nobody can fix.
