# UI improvement plan

> **State, as of the last commit on `fix/ui-wave-0`.** Waves 0-4 are done. Wave 5
> is partial — the task board works without a mouse and off-column statuses are
> visible; the answer editor, the dead skeletons and the tab-state-in-URL are
> not. Waves 6 and 7 are untouched. Three review passes ran over the result and
> their findings are folded in; where the code went a different way from what is
> written below, the text has been corrected rather than left standing.

A working document, not a report to file. It is the result of a four-way review
of the whole frontend — a UX/UI panel over every surface, a walk of twelve user
journeys, a state/keyboard/contrast/responsive audit, and a component-library
and CSS-architecture audit — merged into one ordered plan.

Where the reviewers disagreed, this document decides. Section 4 holds those
decisions and they are binding: they are the contract every later change is
written against, and re-opening one is a deliberate act, not a detail to
reinterpret at the keyboard.

Every claim carries a `file:line`. Numbers were measured — contrast ratios from
the tokens and the box model, control heights from padding plus border plus
line-height, token usage from the tree — not estimated.

---

## 1. What the product is

MCP Research is a workspace for structured research that an AI agent conducts
together with a person. The agent connects over MCP from Claude Code, Cursor,
Claude Desktop or ChatGPT and **writes**: researches, sections, entries,
interview sessions of questions and answers, its own task list, roadmaps. The
person **reads, corrects, shares and exports**. Content arrives over WebSocket
in real time while they watch.

So the web UI is, in order of weight:

1. **A reading and review surface** over long-form content an agent produced
   while the person was elsewhere.
2. **A collaboration surface** — teams, roles, invitations, and share links that
   let someone with no account read exactly one research.
3. **A light editing surface** — fix an entry, answer a question, move a task.
4. **An export surface** — Markdown, print, Obsidian vault.

Four audiences:

| Who | What they need from the UI |
|---|---|
| **Researcher-owner** | Starts the research from their editor, comes to the web to read what came out, correct it, and decide what to do next |
| **Teammate (editor)** | Joins an existing research mid-flight; needs to find their bearings in someone else's material |
| **Share-link reader** | No account. Sees one research. Must never see anything else, and must understand what they are looking at |
| **Team owner** | Invites people, changes roles, moves researches between teams |

**The product's characteristic failure mode** — it recurs in five distinct forms
in section 6 — is the UI assuming the reader knows what only the agent knows:
short codes, `[[E3]]` syntax, what a "section" is, and MCP tool names printed as
instructions to a human.

---

## 2. What exists today

Nuxt 4 SPA (`ssr: false`), embedded into the Go binary. **No Tailwind and no
component framework** — a hand-written design system in
`frontend/assets/css/main.css` plus scoped `<style>` per component. Dark theme
only.

| | Count |
|---|---|
| Components | 74 |
| Pages | 26 (8,273 lines) |
| Composables | 26 |
| Storybook stories | 78 (every component but `entry/Foldable.vue`) |
| `main.css` | 1,189 lines · 52 custom properties · 211 class selectors |

Measured facts that drive most of section 4:

| Finding | Number |
|---|---|
| `--type-xs` + `--type-sm` share of all explicit type declarations | **84%** (268 of 320) — and they are 1px apart |
| `--type-3xl` and `--type-4xl` uses | **0** |
| Raw `border-radius` values below `--radius-sm` | **84** (`4px` ×37, `3px` ×34, `2px` ×13) |
| Distinct `box-shadow` values | **18** |
| Numeric `font-weight` declarations | **176**, including five `650` and two `800` |
| Raw hex literals in `.vue` files | **87** |
| Files honouring `prefers-reduced-motion` | **5 of 100** — and `main.css` is not one of them |
| Copies of the `.short-code` rule | **7 identical + 9 renamed clones**, while `ShortCode.vue` has 3 call sites |
| Names for one page-header layout | **7** |
| Identical CSS lines shared by `login.vue` and `register.vue` | **363**, plus a third copy in a story |

---

## 3. Fix these regardless of any design work

These are defects. They are not waiting on a decision in section 4 and they are
not styling. Ordered by consequence.

### 3.1 `renderRefs` does not escape, and its output goes into `v-html` — 15 sites, one of them public

`composables/useCrossRefs.ts:53` builds

```ts
return `<a href="${href}" class="crossref-link">${ref}</a>`
```

with neither `href` nor `ref` escaped, and returns the surrounding text
untouched. Every call site below passes a **raw agent-authored field** and feeds
the result to `v-html`:

```
components/EntryCard.vue:14                        entry.description
components/research/EntriesView.vue:43, :100       entry.description
components/QuestionList.vue:39                     q.text
components/tasks/KanbanCard.vue:17                 task.title
components/tasks/TaskDetailModal.vue:26,101,139    task.title, description, result
components/mindmap/QuestionNode.vue:6              data.text
pages/research/[id]/entry/[entryId].vue:67         entry.description
pages/research/[id]/session/[sessionId]/index.vue:37, :119   session.notes, entry.description
pages/research/[id]/session/[sessionId]/question/[questionId].vue:17, :31   question.text → <h1>, rationale
pages/s/[token]/entry/[entryId].vue:41             entry.description   ← public
pages/s/[token]/session/[sessionId].vue:34         session.focus       ← public
```

A title containing `<` truncates silently on screen; a title containing markup
executes. Content written by one team member renders in another's browser, and
on the share surface it renders for a stranger with no account.

One page already got this right and says so —
`pages/research/[id]/session/[sessionId]/export.vue:171` carries the comment
*"renderRefs does not escape, so escape first"*. `composables/useInlineMarkdown.ts:35`
is the correct pipeline: escape, then markup, then refs. Route all fifteen sites
through it, and escape the interpolations inside `renderRefs` itself.

`components/SearchModal.vue:144-148` has the same shape: `highlight()`
regex-injects into unescaped text, consumed by `v-html` at `:39, :59, :60`. That
one is on every page in the product.

### 3.2 Prev/Next on the entry page is always wrong

`pages/research/[id]/entry/[entryId].vue:681`:

```js
const currIndex = computed(() => siblings.value.findIndex((e: any) => e.id === entryId))
```

`entryId` is the route param — a **short code** (`E3`), because
`composables/useResearchPaths.ts:26` builds `/entry/${entry.code || entry.id}`.
`siblings[].id` is a UUID. So `currIndex` is `-1` for every reader who arrived by
clicking a link, which is everyone. Consequences:

- `prevEntry` requires `currIndex > 0` → **"Prev" never renders**.
- `nextEntry` requires `currIndex < length - 1`, and `-1 < length - 1` is true →
  **"Next →" points at the first entry of the section from every entry**,
  including from that entry itself, where it links to the page you are on.

The share twin is correct and is the patch:
`pages/s/[token]/entry/[entryId].vue:162` compares `e.id === entry.value?.id`
and guards `currentIndex >= 0`.

### 3.3 Three CSS custom properties are used and never defined

Verified: zero definitions in `main.css`.

| Token | Used at | Effect |
|---|---|---|
| `--color-border-hover` | `main.css:572` — `.btn-icon:hover { border-color: … }` | Invalid at computed-value time, so the border **inherits** instead of changing. The hover state of every icon button in every page header does nothing. |
| `--color-danger` | `app.vue:248` | Falls back to `#dc2626`, an off-palette red, on the Sign-out item |
| `--radius-md` | `components/team/TeamChip.stories.ts:43,56` | Storybook only |

Fixed by the token contract in 4.1; listed here because two of them are live
bugs, not cosmetics.

### 3.4 Timestamps are parsed with `new Date()` in four places, and can be hours wrong

`composables/useRelativeTime.ts:16` exists and its docblock explains why: SQLite
emits `2026-08-14 15:59:40` and the JSON encoder emits `2026-08-14T15:59:40Z`.
A bare `new Date()` reads the first as **local** and the second as **UTC**.
Bypassed at `components/tasks/TaskDetailModal.vue:74,80`,
`components/research/ShareRowList.vue:101-102`,
`components/research/ExportDocument.vue:16` and
`pages/research/[id]/session/[sessionId]/export.vue:39` — so a task's date and a
share's creation date can disagree by hours with the same instant shown
elsewhere.

Related: `pages/settings.vue:137` prints `last_used_at` as a raw ISO string on
the one screen that answers "is this key still in use?".

### 3.5 Writes that fail silently

| Where | What happens |
|---|---|
| `pages/settings.vue:63-72, :139` | **Revoking an API key has no confirmation** and `catch { /* ignore */ }`. An irreversible credential revocation, one click, silent whether it worked or not — in a product that wraps *archiving a research* in a `ConfirmModal`. |
| `pages/research/[id]/session/[sessionId]/index.vue:218` | `catch { /* ignore */ }` on **adding a question**. The form sits there with the text still in it and nothing happens. |
| `pages/research/[id]/tasks.vue:114-189` | `saveField`, `updatePriority`, `confirmStatusChange`, `createTask` have **no `try/catch`** — an unhandled rejection and nothing on screen. |
| `pages/research/[id]/index.vue:424-433, :456-463` | `handleDetailsSave` and `toggleArchive`: same. |
| `components/ResearchCard.vue:80-87` | Archive has no confirm, no error path, no rollback. A failed archive looks exactly like a successful one. |
| `pages/research/[id]/entry/[entryId].vue:583-587` | A failed status change rolls back and `console.error`s — the badge flicks back with no explanation. |
| 4 × `alert()` | `pages/index.vue:123`, `pages/research/[id]/index.vue:449`, `pages/research/[id]/entry/[entryId].vue:549` and `:647` — a native modal dialog over an editor full of unsaved work, in a product with a working `ToastHost` that supports retry actions. |

### 3.6 A failed fetch renders as "you have nothing"

`composables/useApi.ts` returns `useFetch`'s `error` and **no page destructures
it**. The universal shape is `data ?? []` → the empty state. So a server that is
down tells a new user "No research projects yet" and hands them the onboarding
command. Sites: `pages/index.vue:65`, `pages/research/[id]/index.vue:161`,
`pages/research/[id]/entry/[entryId].vue:225`, `pages/research/[id]/tasks.vue:69`,
`pages/research/[id]/roadmaps.vue:38`, `pages/settings.vue:143`.

`pages/s/[token].vue:19-26` and `components/entry/HistoryPanel.vue:24-38` already
distinguish error from empty. They are the model.

### 3.7 Every skeleton on the owner-side pages is unreachable

`pages/research/[id]/index.vue` makes four sequential top-level
`await useApi(...)` calls (`:205, :306, :309, :313`). The component is therefore
a Suspense boundary: nothing renders until all four resolve, so `pending` is
already `false` when the template first runs and `v-if="pending"` at `:2-10`
never fires. There is no `<NuxtLoadingIndicator>` in `app.vue` and no page
transition configured. Clicking a research freezes the previous page for the sum
of four round trips with no feedback.

Same pattern and same dead skeleton in `entry/[entryId].vue` (3 awaits),
`tasks.vue`, `session/[sessionId]/index.vue`, `question/[questionId].vue`,
`roadmaps.vue` and `export.vue`. `pages/s/[token].vue:142-149` documents this
exact trap and avoids it by not awaiting.

### 3.8 Two things cannot be done at all with a keyboard or on a touch device

- **Change a task's status.** `pages/research/[id]/tasks.vue:141-171` is the only
  path and it is HTML5 drag-and-drop, which does not fire on touch.
  `components/tasks/TaskDetailModal.vue:52-57` renders status as a read-only
  badge. Meanwhile `components/tasks/KanbanBoard.vue:196-206` collapses the board
  to one column at mobile widths — implying it is meant to be used there.
- **Use the knowledge graph.** `pages/research/[id]/graph/index.vue:41-53` is a
  bare `<canvas>` with mouse handlers only: no `tabindex`, no `role`, no touch
  events, no text fallback. `error` is destructured at `:76` and never rendered,
  so a failed fetch is a permanently black screen beside a sidebar reading
  "0 nodes".

### 3.9 A person cannot answer a question

`PUT /api/questions/{id}` is registered at `internal/api/server.go:253` and
works — the reviewer exercised it against a live server. **No component in the
frontend calls it.** `question/[questionId].vue` renders text, rationale, answer
and cross-refs, and no control of any kind; `QuestionList.vue` is read-only.

The asymmetry is the indefensible part: `session/[sessionId]/index.vue:80-98`
gives the human a form to **add** a question — the half only the agent needs —
while withholding the half a reviewer needs on every visit. A session also cannot
be completed, paused or renamed, so a run that ends without the agent marking it
`completed` leaves a research showing a perpetually active session with no way to
close it.

---

## 4. Decisions

These resolve the places where the four reviewers disagreed. They are binding.

### 4.1 The token contract

After the CSS split (4.2), the tokens are the only thing holding the look
together, so this list is the contract. Names and values, decided.

```css
:root {
  /* ── Surface & ink ──────────────────────────────────────────────── */
  --color-bg:             #0c1220;
  --color-bg-deep:        #080e1e;   /* NEW. auth hero, graph canvas, code wells */
  --color-surface:        #151d2e;
  --color-surface-hover:  #1e2940;
  --color-surface-raised: #1b2438;   /* NEW. popovers, dropdowns, dialogs — so a
                                        floating layer separates from a .card */
  --color-border:         rgba(148, 163, 184, 0.12);
  --color-border-strong:  rgba(148, 163, 184, 0.22);
  --color-border-hover:   var(--color-border-strong);  /* NEW. fixes main.css:572 */
  --color-text:           #e2e8f0;
  --color-text-muted:     #7f8ea3;
  --color-text-faint:     #5b6b80;   /* NEW. replaces `opacity` on muted text —
                                        see 4.5, this is a contrast fix */

  /* ── Accent & status ────────────────────────────────────────────── */
  --color-primary:        #6cc5e0;
  --color-primary-hover:  #7cd1ea;   /* NEW. literal at main.css:307 */
  --color-primary-muted:  rgba(108, 197, 224, 0.10);
  --color-on-primary:     var(--color-bg);  /* NEW. fixes settings.vue:210 */
  --color-success:        #34d399;
  --color-warning:        #f0b849;
  --color-error:          #ef6b6b;
  --color-info:           #6b9df0;
  --color-danger:         var(--color-error);  /* NEW. one red for destruction */

  /* ── Categorical: tags, graph, mindmap, roadmap ─────────────────── */
  --hue-1: #6cc5e0;   /* cyan   — entries          */
  --hue-2: #34d399;   /* green  — done             */
  --hue-3: #f0b849;   /* amber  — sessions         */
  --hue-4: #ef6b6b;   /* red    — blocked/failed   */
  --hue-5: #a78bfa;   /* violet — sections, refs   */
  --hue-6: #fb923c;   /* orange — questions        */

  /* ── Elevation ──────────────────────────────────────────────────── */
  --shadow-1:     0 2px  8px rgba(0, 0, 0, 0.18);   /* NEW. cards, flow nodes */
  --shadow-2:     0 8px 24px rgba(0, 0, 0, 0.30);   /* NEW. menus, popovers   */
  --shadow-3:     0 16px 48px rgba(0, 0, 0, 0.40);  /* NEW. modals, toasts    */
  --shadow-glow:  0 4px 20px rgba(108, 197, 224, 0.20);  /* NEW */
  --shadow-focus: 0 0 0 3px rgba(108, 197, 224, 0.15);   /* NEW */

  /* ── Type ───────────────────────────────────────────────────────── */
  --type-3xs:  0.6875rem;  /* 11px — NEW. canvas/flow chrome ONLY, see 4.4 */
  --type-2xs:  0.75rem;    /* 12px — NEW. canvas/flow chrome ONLY          */
  --type-xs:   0.875rem;   /* 14px — badges, codes, meta. NEVER prose      */
  --type-sm:   0.9375rem;  /* 15px — the UI default                        */
  --type-base: 1.0625rem;  /* 17px — body copy, reading                    */
  --type-md:   1.1875rem;  /* 19px — NEW. card titles, sub-headings        */
  --type-lg:   1.25rem;
  --type-xl:   1.5rem;
  --type-2xl:  2rem;
  --type-3xl:  2.5rem;     /* RETUNED from 2.75 — login.vue:219 wrote 3.25rem
                              rather than use either dead top step          */
  /* --type-4xl: DELETED. Zero uses. */

  /* ── Weight ─────────────────────────────────────────────────────── */
  --weight-normal:   400;
  --weight-medium:   500;
  --weight-semibold: 600;
  --weight-bold:     700;
  /* 650 and 800 fold into these. Only 300–800 static weights of Outfit are
     loaded, so 650 already rounds — it was never a decision. */

  /* ── Radius ─────────────────────────────────────────────────────── */
  --radius-hair: 2px;      /* NEW. bar-shaped things only: progress, marks */
  --radius-xs:   4px;      /* NEW. absorbs all 2/3/4px chip radii          */
  --radius-sm:   6px;
  --radius:      10px;
  --radius-lg:   16px;

  /* ── Measure ────────────────────────────────────────────────────── */
  /* Stated in rem, not ch — corrected during implementation. `ch` is the
     advance of `0`, which in Outfit is 0.656em against an average lowercase
     of about 0.5em, so 72ch is roughly 94 characters. Worse, `ch` resolves
     against each element's own font-size, so one token became a different
     width on every heading level and an h2 hung its underline past the text
     beneath it. */
  --measure:       38rem;  /* NEW. long-form reading — see 4.3            */
  --measure-wide:  52rem;  /* NEW. diffs, tables, code                    */
  --measure-prose: 34rem;  /* NEW. page leads, empty-state copy           */

  /* ── Chrome ─────────────────────────────────────────────────────── */
  --nav-h: 66px;           /* NEW. main.css:417 guesses 56px for this     */
}
```

Delete, as unused: `--transition-spring`, `--type-4xl`, `--z-base`.

**Decisions inside this list, and why:**

- **`--type-xs` stays at 14px.** The panel wanted it pulled to 13px so xs and sm
  are a real distinction. Rejected: the contrast audit measured
  `--color-text-muted` at 14px on `--color-surface-hover` at **4.36:1**, already
  under AA. Shrinking a size that is already failing makes a measured failure
  worse. The distinction is bought instead by adding `--type-md` at the top and
  by the rule that **`--type-xs` never carries a sentence** — today it does, at
  `ToastHost.vue:113`, `share/Banner.vue:113`, `QuestionList.vue:286` and the
  entry provenance line; those move to `--type-sm`.
- **Two micro sizes, not four.** The library audit asked for 12/11/10/9px for
  flow-node chrome. 10px and 9px are rejected outright — nothing at that size
  meets contrast on this palette, and the graph's 9–11px greys are exactly what
  measured 2.69–3.83:1. The sites using them move up to 11–12px.
- **One new radius, not three.** `2px` vs `3px` vs `4px` was never a decision
  anybody made; 84 sites picked one at random. They collapse to `--radius-xs`.
  `--radius-hair` survives only for genuinely bar-shaped elements.
- **Five shadows, not eighteen and not seven.** The 18 live values cluster on
  three altitudes; the pairs that differ by 4px of blur and 0.03 of alpha are
  noise. `main.css:232`'s `.card:hover` inset ring and
  `KanbanBoard.vue:135`'s drop-target ring stay bespoke — they are signatures,
  not elevation.
- **`--color-danger` is an alias of `--color-error`, not a new red.** There are
  four reds in the tree (`#ef6b6b`, `#ef4444`, `#dc2626`, `currentColor`). One
  survives.

### 4.2 CSS architecture — the split

**Target files.**

```
assets/css/
  tokens.css     :root only. No selectors. The file a designer opens.  (~90 lines)
  base.css       Reset, body, links, focus-visible, ::selection, the noise and
                 glow layers, @keyframes fade-up and skeleton-shimmer,
                 the global prefers-reduced-motion block.               (~120)
  system.css     The primitives and only the primitives:
                 .container .card .grid .btn* .badge* .tag* .form-* .modal-*
                 .link-btn .skeleton-card .user-avatar .layout-sidebar
                 .page-header .page-title, the surviving utilities, and the
                 @media print / responsive blocks that govern them.     (~500)
  markdown.css   .markdown-content and descendants, .crossref-link,
                 .search-mark, .block-doc.                              (~70)
  mermaid.css    .mermaid-* and its @media print block.                 (~155)
```

`main.css` at 1,189 lines becomes roughly 710 after ~480 lines leave, and then
splits so that **no file exceeds 500 lines** and the file answering "what is the
design system here?" is 90 lines of `:root`.

**`markdown.css` and `mermaid.css` must be global files, not `:deep()` blocks.**
The HTML is a *string* — `marked` output for one, `useMermaidViewer.ts`'s
hand-built DOM for the other — so no element in it carries a scope attribute.
And eight components render markdown independently, so a `:deep()` block in any
one of them leaves the other seven unstyled. Import them **after** `system.css`:
`main.css:945` `.markdown-content pre.mermaid-error` depends on
`.markdown-content pre` at `:788` being declared first.

**The rule for tomorrow.** Two questions, in order:

1. **Is the element rendered by Vue, in a template you can point at?**
   *No* — it comes from `v-html`, a string, or a composable that builds DOM →
   it goes in `markdown.css` / `mermaid.css` / `vue-flow.css`, with a comment
   naming the file that produces the HTML.
   *Yes* → question 2.
2. **Do three or more *unrelated* components use this exact class?**
   *Yes* → `system.css`, and its name comes from the design vocabulary
   (`.btn-*`, `.card-*`), never from a feature (`.tasks-header`).
   *No* → `<style scoped>` in the component that renders it. If exactly two
   components need it, the answer is usually "one should be using the other",
   not "promote the class".

Three is the threshold on purpose. Two occurrences is the moment to extract a
component, not to promote a class — promoting at two is how `.research-header`,
`.tasks-header` and `.roadmaps-header` became three names for one rule.

**Guard it.** Add a `scripts/` check that fails when a selector in `system.css`
has fewer than three referencing files, and when a `.vue` scoped block contains a
hex literal, a `box-shadow`, or a `font-size` that is not a `var()`. Without it
the split silently undoes itself.

**What must stay global as a design decision, not by oversight.** Moving these
into one component would license two components to drift, and several already
have:

| Class family | Why |
|---|---|
| `.short-code` | `R1`/`E3`/`SS2` must be pixel-identical everywhere. Currently copied into 7 scoped blocks and renamed into 9 more, with weight drifting 600→700 and radius 4px→3px. This is the canonical proof of the rule. |
| `.badge-*` | A status means the same on an entry, a session, a question and a task. **Built as strings** by `StatusBadge.vue:2`, so a static grep reports zero references — annotate, never purge. |
| `.tag`, `.tag-hue-*` | The hue is derived from the tag string by `useTagHue`; the same tag must be the same colour everywhere or the derivation is pointless. Same string-built caveat. |
| `.crossref-link`, `.markdown-content`, `.mermaid-*`, `.search-mark` | Cannot be scoped — `v-html` output carries no scope attribute. |
| `.btn` family | A control must look like a control regardless of which component drew it. |
| `.card`, `.card-title`, `.card-meta` | The product's one container. |
| Form and dialog vocabulary | `main.css:616-621` already records why: nine `.modal-title`s in two disagreeing designs, three names for one text field. |
| `.research-header`, `.entry-header`, `.title-with-code` | **Do not move these back into pages.** `main.css:507-519` documents that the move was reverted twice, the second time breaking `/s/{token}` — share pages deliberately render the owner's markup. Resolve by promoting to `.page-bar`, not by re-scoping. |
| `.empty-state`, `.sr-only`, `.user-avatar`, `.skeleton-card`, the focus ring, `@media print` | Accessibility and identity affordances that must not vary. |

**Scoping hazards, per move:**

- `:slotted()` is required for anything a parent passes into a slot.
  `ActionMenu.vue:88-120` gets this right; every new slot-driven component must.
- A scoped rule *does* reach a child component's root element. That is why
  `CrossReferencesBlock.vue:95` `.crossrefs-block` works — it is `Foldable.vue`'s
  root. But the three `.crossrefs-title` copies (`:100`, `ExternalLinksBlock:32`,
  `RelatedEntriesBlock:65`) target a non-root element inside `Foldable` and are
  **dead**.
- Stories count as consumers. `.badge-active` is referenced only from
  `team/ViewerNotice.stories.ts`; `.sidebar-item` only from
  `entry/RevisionRow.stories.ts` outside its owner.
- `main.css:545-548` documents a live specificity collision: `.btn-icon` exists
  globally *and* as unrelated scoped rules in `ResearchCard.vue:99` and
  `TeamMemberList.vue:107`, which win — so one glyph is a bordered 30px chip in a
  page header and a borderless 28px ghost in a card.

**Dead, verified across components, pages, `app.vue`, composables and all 78
stories — safe to delete (~20 lines):** `.grid-3` · `.flex-between` ·
`.flex-center` · `.gap-1` · `.gap-3` · `.gap-4` · `.gap-6` · `.mb-6` ·
`.ml-auto` · `.w-full` · `.text-muted` · `.text-sm` · `.text-xs` ·
`.font-medium` · `.font-bold` · `.task-widget` · `.progress-card` · the empty
`.markdown-content { }` at `:769`.

### 4.3 The reading surface gets a measure

`main.css:588` gives `.entry-content` padding and no `max-width`. The entry page
has no sidebar, so a paragraph runs about 1,090px at 17px — roughly **130
characters per line**, against a comfortable 45–75. The one surface whose entire
purpose is reading a long document is the only one with no reading width, while
`DiffView.vue:198` caps itself at `90ch` and `teams/index.vue:110` caps a
one-line lead at `65ch`.

```css
.entry-content .markdown-content,
.entry-content .block-doc { max-width: var(--measure); }
.entry-content table,
.entry-content pre,
.entry-content .mermaid-diagram { max-width: var(--measure-wide); }
```

One rule, and it lands on both the owner's entry page and the shared one.

**As implemented it goes no further than that.** The selector is scoped to
`.entry-content`, and session notes and the printable export document render
`.markdown-content` outside any `.entry-content`, so they get no measure. That is
a gap, not a decision: both are long-form reading surfaces and both should have
one. Extending it means either giving those two containers the cap directly or
promoting the rule off `.entry-content` — the second is cleaner and should wait
until the export document stops being a fork of a component.

### 4.4 Dark-only stands; the graph joins the token system

**No light theme now.** The premise — long-form dark reading is fatiguing — is
right, but the cause is the missing measure (4.3), not the palette. And the cost
is understated: a light theme must survive 87 raw hexes in `.vue` files, 18
hardcoded `rgba(0,0,0,x)` shadows, the `rgba(255,255,255,0.02)` table hover, the
noise overlay, the ambient glow, `#fff` in the login hero, 14 One-Dark syntax
colours and the `md-editor-v3` `:deep()` overrides. That is a ~40-file project.

The decisive argument: **the product already prints to white.**
`main.css:1060-1079` is a complete print stylesheet and
`pages/research/[id]/export.vue` is a printable document surface. The
light-background need is served where it actually arises.

Every token in 4.1 is nonetheless written so a `[data-theme="light"]` block
becomes a ~45-line addition later: semantic names, derived tints via
`color-mix`, no raw `rgba(0,0,0,…)` outside the shadow tokens.

**The graph is not an exception.** `graph/index.vue` and `GraphSidebar.vue` use
zero colour tokens — every text colour is a bare `rgba(255,255,255,α)` across
**ten distinct alphas**, every font size is `px` across six off-scale values,
background is `#111` against `--color-bg: #0c1220`. Close enough that the
difference reads as sloppiness rather than intent, and the contrast audit
measured four failures there. Resolve with a scoped token block, so one decision
replaces fifty:

```css
.graph-page {
  --color-bg:         var(--color-bg-deep);
  --color-text:       rgba(255, 255, 255, 0.85);
  --color-text-muted: rgba(255, 255, 255, 0.55);  /* raised from 0.45 — 4.49:1 */
  --color-border:     rgba(255, 255, 255, 0.06);
}
```

then use the ordinary token names inside.

### 4.5 One destructive path, one red, and no `opacity` on text

**`.btn-danger` is a fill for confirmations and a ghost everywhere else** — and
the fill is `--color-danger` with `--color-bg` ink, not `#ef4444` with white.
`ConfirmModal.vue:79-84` currently redefines the global class, so the same class
name means "quiet outlined red" in a page and "loud filled red" inside a dialog,
and the filled version measures **3.76:1** — a failure. Delete the local
override; add `.btn-danger--solid` to `system.css`. A destructive confirm
*should* be the loudest thing on screen; it just has to be readable.

Also `ConfirmModal.vue:59-62`: `.confirm-icon.info` pairs an **amber** background
with a **cyan** foreground — an unfinished copy-paste from the danger variant.

**`opacity` is forbidden on text.** It is the systematic cause of most contrast
failures: `--color-text-muted` has only 0.55 of headroom over 4.5:1 on
`--color-surface`, so any opacity below ~0.9 pushes it under. Three of the five
worst measurements are opacity, not colour choice. Use `--color-text-faint` at
full opacity instead.

### 4.6 The missing primitives, introduced *during* the split

This is the cheap moment and it will not come again. Each of these dissolves work
that is currently written out by hand many times over.

| Primitive | Replaces |
|---|---|
| `.cluster` / `.stack` | ~11 identical flex rows/columns: `.research-actions`, `.entry-actions`, `.title-with-code`, `.page-header-actions`, `.tasks-actions`, `.entry-tags`, `.q-badges`, `.skeleton-list`, `.group-body`, `.password-form`, `.rail-skeletons` |
| `.page-bar` | **7 names for one layout** — but see below: it was written, shipped unadopted and withdrawn. Adopting it across fourteen pages, four of them owner/share twins, is an unreviewable diff to buy a renamed class. It arrives as the internals of `PageHeader.vue` instead. |
| `.data-rows` / `.data-row` / `--busy` / `--dead` | The row chrome shared by `TeamMemberList`, `TeamInviteList`, `ShareRowList`, `TeamRowList` — and `.share-rows` has **no rule at all**, so that list is missing its top border |
| `.select` | Three byte-identical copies of a styled `<select>`, each re-encoding the same chevron SVG with `stroke='%237f8ea3'` hardcoded — which breaks the moment `--color-text-muted` changes |
| `.tabs` / `.tab` | With `role="tablist"`/`aria-selected` baked in, so `session/[sessionId]/index.vue:41-65` stops shipping tabs a screen reader cannot see |
| `.segmented` | `entry/[entryId].vue:841-865` builds one by overriding `.btn`'s hover, transform and shadow |
| `.check-row` | Local to `ShareDialog.vue:338`; needed by `ObsidianExportDialog` and `GraphSidebar` |
| `.danger-zone` / `.danger-row` | Local to `teams/[id].vue:484-509`, which is why `settings.vue` could not use it and invented a bare text button with no confirm |
| `.centered-gate` | `share/PasswordGate.vue:60`, `invite/[token].vue:214`, `s/[token].vue:228` |

**Note the four lists stay four components.** Merging `TeamMemberList`,
`TeamInviteList`, `ShareRowList` and `TeamRowList` into one slot-driven component
would produce a boolean-props component wearing four names — they have genuinely
different columns. Only the row *chrome* is shared. Pick one dead-row rendering
while you are there: an expired invite is `--color-text-muted`, a dead share is
`opacity: 0.55`; same concept, two renderings, and the opacity one fails
contrast.

### 4.7 Density: rows, not cards, for lists — but verify in a browser first

`EntriesView.vue` renders every entry as a `.card` with 24px padding, a border, a
hover lift and a staggered fade. Forty entries is roughly **4,800px of scroll to
read 40 titles**. The product already has a good compact row idiom four times
over (`.member-row`, `.share-row`, `.past-session-item`, `.result-item`), each
written from scratch.

Decision: promote the row chrome now (4.6, mechanical), and stage the
entries-and-questions-as-rows change separately, after looking at it in a
browser. It changes what the product feels like, which is not a change to make
from a diff. A compact/comfortable toggle is rejected — a preference nobody finds
is not a fix for a bad default.

### 4.8 Missing components

| Component | Replaces | Call sites collapsed |
|---|---|---|
| `PageHeader.vue` | The `.page-header > .research-header > .title-with-code` block | **11 pages**. `PageHeaders.stories.ts` already documents this pattern with stub components — the codebase has admitted it exists |
| `auth/AuthShell.vue` | login/register hero + form panel | 2 pages + 1 story; **~700 lines deleted** (363 identical CSS lines ×2, plus a third transcription in `AuthPages.stories.ts` that uses raw values instead of tokens) |
| `ModalHeader.vue` | `.modal-header` + `.modal-close` + a 16-line close SVG | **5 components** |
| `EditableField.vue` | The pencil/value/edit-form triple | **8 blocks** across `DetailsPanel` and `TaskDetailModal`, which share ten byte-identical rule bodies |
| `flow/FlowNode.vue` | The Vue Flow card shell | **10 node components**. Must be the root element, not a wrapper — Vue Flow measures the node box |
| `flow/FlowToolbar.vue` | mindmap/roadmap toolbars | 3 pages |
| `CommandLine.vue` | `.empty-command` + `.command-text` + `.copy-btn` + the copied-flag timer | **4 components**; the copy-to-clipboard-with-feedback logic is written out three times |
| `AppNav.vue` | `app.vue:85-118` + `:150-250` | 1 file — but it stops `AppNav.stories.ts` from being fiction |
| `ExportToolbar.vue`, `entry/RefRow.vue`, `session/NotesCard.vue`, `session/ProgressSummary.vue` | see the extraction list | 2–5 sites each |
| `utils/truncate.ts` | **10 local `truncate` definitions** in two subtly different versions — four guard `if (!text) return ''` and six do not, so a null title throws in half of them | 11 files |

Adopt what already exists and is unused or under-used: `ShortCode.vue` (3 of ~20
sites), `EntryCard.vue` (1 site while 3 places inline it), `PrintButton.vue`
(**zero** sites, and its label says "Export" while all three hand-rolled copies
say "Print / PDF"), `CopyableSecret.vue` (not used on the one screen that shows a
secret exactly once), `EmptyState.vue` (14 of 15 sites).

---

## 5. Findings by surface

Severity: **critical** the surface fails at its job · **high** visible recurring
friction · **medium** polish that compounds · **low** nice to have.
Cost: **S** a class or token change · **M** rework one component · **L** a new
component or a restructured screen.

### 5.1 Research list / home

| Sev | Cost | Where | What |
|---|---|---|---|
| high | S | `internal/storage/research_repo.go:138` vs `ResearchCard.vue:36` | Ordered by `created_at DESC`, prints `relativeTime(updated_at)`. A research rewritten ten minutes ago sits below one untouched since yesterday while displaying "10 minutes ago". Order by `updated_at`. |
| high | S | `ResearchCard.vue:99-113` | Archive is `opacity: 0` until hover, with no `:focus-within` rule — a keyboard user focuses an invisible destructive control, and on touch it is unreachable. `TeamMemberList.vue:30-39` is the same control done right. |
| high | S | `ResearchCard.vue:31-36`, `EntriesView.vue:11-17`, `TagList.vue:3-8` | Tag filters are `<span @click>`: not focusable, not announced, no role. The primary filtering mechanism on two surfaces is mouse-only. |
| high | S | `ResearchCard.vue:2 + :30` | A `NuxtLink` (`<a>`) containing a `<button>` — invalid nesting. |
| medium | S | `pages/index.vue:84` | `statusFilter` defaults to `'active'`, so a research the agent marks `completed` — which the documented workflow tells it to do — vanishes from home with no explanation. |
| medium | S | `GettingStartedBanner.vue:7` | *"A read-only view of your AI-driven research sessions"* — the first sentence a new user reads, and it has been false since entry editing, tasks, teams and share links shipped. |
| medium | S | `pages/index.vue:171-187` | `.page-header-row`, `.page-header-actions`, `.skeleton-list` re-declared identically to `main.css:599-614`. |
| medium | M | `pages/index.vue:44-52` | No sort, no pagination. 60 active researches is 60 animated cards. |
| low | S | `pages/index.vue:7-12` | The only action on the landing page is **Import JSON** — the least likely thing a first-time user wants, in the primary slot. |

`pages/index.vue:65-71` is the best empty state in the product: it hands the user
the exact sentence to type into Claude, with a copy button. Use it as the
template for the others.

### 5.2 Research overview

| Sev | Cost | Where | What |
|---|---|---|---|
| high | M | `research/Sidebar.vue:86-91` | `sectionProgressWidth()` maps a **status enum** to a fake percentage — `completed → 100%`, `active → 50%`, `draft → 10%`. Every active section shows a half-full bar whether it holds 1 entry or 40. Delete the bar (the count beside the name already answers "how much is here") or compute it from `entries_count`, both fields being on the payload. |
| high | S | `research/Sidebar.vue:14-70` | `role="tablist"` with every tab at `tabindex="0"`, no arrow keys, no `aria-controls`, no `tabpanel`. With 8 sections that is 10 tab stops before the content. `HistoryPanel.vue:229-241` already implements roving tabindex — lift it. |
| high | M | `research/[id]/index.vue:21-89` | Ten controls in one flex row with no hierarchy and no primary. Reduce to identity (status + team) on the left, one primary **Share**, one `⋯` menu; move Tasks/Roadmaps/Mindmap/Graph into a `.tabs` strip — they are *views of this research*, which is what a tab means. |
| high | S | `research/ActiveSessionsGrid.vue:1-17` | The active session is the single most valuable object on the page — what the agent is doing *right now* — and the card shows a code, a badge and a title. No progress, no last activity. (The sessions payload carries no progress fields, so this needs an API field or a second request.) |
| medium | S | `research/[id]/index.vue:91` | The research **goal** — the sentence saying what all this is for — is `--type-sm` muted below a ten-control toolbar. It is the page's subtitle, not metadata. |
| medium | S | `research/PastSessionsList.vue:4, :35` | Labelled "Completed sessions" but the parent passes everything `!== 'active'`, including archived; and `align-items: flex-end` orphans the toggle against the page edge. |
| medium | S | `research/[id]/index.vue:263-274` | Landing opens the *first section*, not "All entries" — for 8 sections, an arbitrary slice. |
| low | S | `EntriesView.vue:141-146` | *"Claude will populate this section"* names one vendor; the product supports Cursor, Claude Desktop and ChatGPT. The share path two lines above already has corrected wording; the viewer path does not. |

### 5.3 Entry read and edit — the centre of gravity, and the weakest surface

| Sev | Cost | Where | What |
|---|---|---|---|
| critical | S | `main.css:588` | No measure. See 4.3. The highest-value single line in this document. |
| critical | S | `entry/[entryId].vue:62-64` | The **delete** button is icon-only with no `aria-label` and no `title`. A screen reader announces "button". |
| critical | S | `entry/[entryId].vue:681` | Prev/Next always wrong. See 3.2. |
| high | S | `entry/[entryId].vue:917` | `.status-dropdown-overlay { z-index: 9999 }` — above `--z-toast: 300`. A status picker paints over a notification. |
| high | M | `entry/[entryId].vue:556-570` | The status dropdown computes `position: fixed` coordinates **once** on open and never recalculates, so it detaches on scroll. Teleported to `<body>`, so Tab from the trigger jumps to the end of the page. No `aria-haspopup`, no `aria-expanded`, no Escape, no arrow keys, no focus return. |
| high | M | `entry/[entryId].vue:24-65` | Five controls at equal weight — status, Edit, History, Copy, Delete. `Edit` is the action the page exists for and looks exactly like `Copy`. **Four different heights**: 24.00 / 28.40 / 26.60 / 26.20px. This is the disease the `--control-h` comment at `main.css:33-38` was written to end; it was cured on the research page and left here. |
| high | S | `entry/[entryId].vue:299` | A realtime `refresh()` repaints the article under a reader mid-page, with no scroll anchoring and no "updated" affordance. The edit-mode version at `:91-94` is exactly right; read mode has nothing. |
| medium | S | `entry/HistoryPanel.vue:2` | `<ModalOverlay size="xl">` with **no `labelledby`** — the app's largest dialog announces itself as "dialog". The prop exists and `ConfirmModal.vue:2` uses it. |
| medium | S | `entry/[entryId].vue:882-895` | Fourteen One-Dark literals, one of which (`:890`) is `--color-primary` written out. The palette is legitimate; the duplicate is not. |
| medium | M | `entry/[entryId].vue:174-188` | `md-editor-v3` with 25 lines of `:deep()` pushing tokens into a third-party widget. The edit mode is visibly a different application, and the override block is a standing upgrade liability. Out of scope for a token pass — flagged, not scheduled. |
| medium | S | `main.css:788` | `.markdown-content pre` has `overflow-x: auto` and no `max-height`. A 400-line code block grows the card to thousands of pixels. `entry/Foldable.vue` exists for this and is not used here. |
| medium | S | `main.css:1172` | Markdown tables get `display: block; overflow-x: auto` **only at ≤768px**. Between 769px and the card width a wide table overflows with no scroll container. |
| low | S | `entry/[entryId].vue:137` | The Rendered/Source toggle is neither persisted nor in the URL. |

What is excellent here and must survive the rework: the
`remoteChangedWhileEditing` banner (`:91-94`), `useUnsavedWork` (`:517-524`), the
`isSelf(event)` discrimination (`:284-300`), and all of `HistoryPanel` — real
error states with retry, an `sr-only role="status"` announcement, a `.pane-stale`
dim instead of a blank, and rail arrow-key navigation.

### 5.4 Sessions and questions

| Sev | Cost | Where | What |
|---|---|---|---|
| critical | M | `question/[questionId].vue`, `QuestionList.vue` | No way to answer a question, correct an answer, defer or skip. See 3.9. |
| critical | S | `session/[sessionId]/index.vue:218` | `catch { /* ignore */ }` on adding a question. |
| high | S | `session/[sessionId]/index.vue:41-65` | Tabs are bare `<button>`s — no `role="tab"`, no `aria-selected`, no `tabpanel` — while the research sidebar 200 lines away does use `role="tab"`. Two tab implementations, one accessible. |
| high | S | `session/[sessionId]/index.vue:189` | Active tab is component-local, **not in the URL**. A teammate cannot be sent to the Changes view; a refresh drops back to Questions. `research/[id]/index.vue:268-270` already does this correctly for `activeSection`. |
| high | S | `QuestionList.vue:41` | `<StatusBadge :status="q.priority" />` — **priority rendered in the same badge shape and palette as status**. "High" is red exactly like "Failed"; "Low" is grey exactly like "Archived". Split `StatusBadge` and `PriorityMark`. |
| high | S | `session/[sessionId]/index.vue` | A session cannot be completed, paused or renamed. |
| high | M | `session/[sessionId]/index.vue:105-124` | The Entries tab re-implements the entry card instead of using `EntryCard` — third copy, and it has drifted: a hardcoded path instead of `entryPath()`, and no artifact badge. The **share twin uses the component correctly**. |
| medium | S | `question/[questionId].vue:14` | The breadcrumb prints a **fake question code**: `` `Q${questionIndex + 1}` `` from a status-ordered flat list. A question whose real code is `Q7` shows as `Q3` — a code-shaped label with the wrong number, in a product where codes are the addressing scheme. Same flattening makes Next jump from the last pending question to the first answered one. |
| medium | S | `session/[sessionId]/index.vue:37, :242` | Session notes get `white-space: pre-wrap` **and** `.markdown-content` **and** `marked.parse()` — pre-wrap on parsed HTML doubles every blank line. And the agent's own notes render at `--type-sm` muted. |
| medium | S | `QuestionList.vue:51` | A full-page `EmptyState` — 64px padding, icon, title — rendered *inside a collapsed accordion group* when filters exclude everything, offering no way to clear the filters that caused it. |
| medium | S | `QuestionList.vue:279, :289` | `margin-left: calc(0.625rem + 0.6rem + var(--space-2))` — hand-computed alignment that assumes the code chip's exact font size, and is simply wrong for a question with no code. Use a grid. |
| low | S | `session/[sessionId]/index.vue:58-64` | The Changes tab has no count, unlike the two beside it. |

### 5.5 Tasks

| Sev | Cost | Where | What |
|---|---|---|---|
| critical | M | `tasks.vue:141-171`, `TaskDetailModal.vue:52-57` | Status changes by mouse drag only. See 3.8. |
| high | S | `KanbanBoard.vue:67`, `KanbanCard.vue:13-17` | `blocked` and `deferred` are folded into "Todo" **with no marker**, and the card renders only code, title and a high-priority badge. The domain guide explicitly tells the agent to use `blocked` with a reason — so the one status meaning "I need you" is the one the board erases. There is also no way to *set* either from the UI. |
| high | S | `tasks.vue:97` | The `failed` column is titled **"Rejected"**. `StatusChangeModal` then says "Rejected", the API records `failed`, and badges elsewhere say `failed`. |
| medium | S | `KanbanBoard.vue:181` | `max-height: calc(100vh - 200px)` against ~213px of real chrome, and `vh` not `dvh` — so a full column produces an inner scrollbar *and* a page scrollbar, and on mobile Safari its bottom sits under the URL bar. |
| medium | S | `KanbanBoard.vue:30-32` | "No tasks" at `opacity: 0.5` on muted text — **2.25:1**, the worst measurement in the product, ×4 columns on an empty board. |
| medium | S | `tasks.vue:2-7` vs `KanbanBoard.vue:197,202` | The skeleton breaks at 768px; the board breaks at 1024px and 640px. Between 641–1024px the skeleton shows 4 or 1 columns and the board then shows 2 — a guaranteed jump. |
| medium | S | `task_repo.go:92` | Columns order by priority then `created_at ASC`, with no drag-to-reorder and no explanation — so changing a priority makes the card jump within its column. |
| low | S | `KanbanBoard.vue:78, :84, :101` | Drag state managed by `classList.add/remove` and `document.querySelectorAll` — imperative DOM mutation that a realtime re-render wipes. |

### 5.6 Roadmaps, graph and mindmap

| Sev | Cost | Where | What |
|---|---|---|---|
| critical | L | `graph/index.vue:41-53, :76` | Canvas with mouse handlers only; `error` never rendered. See 3.8. The cheap 20% — a `role="img"` with a summarising `aria-label` plus a visible "View as list" falling back to the entries view — is **S**. |
| critical | S | `mindmap/index.vue:11-58` | `.toolbar-right` is `display: flex` with no `flex-wrap` and no `overflow-x`, holding ~11 `nowrap` controls ≈1100px wide. Below ~1150px the layout toggle, Expand all, Collapse and Fit view run off-screen and are unreachable; at 375px two thirds of the toolbar is gone. |
| high | S | `graph/index.vue:90-95` **and** `:124-133` | The node palette is defined **twice in the same file** — once for the legend, once for the renderer. They agree today; nothing keeps them agreeing. |
| high | S | `graph/index.vue:93` vs `:92` | `question: '#fbbf24'` beside `session: '#f0b849'` — two ambers 4% apart as the sole distinction between two node types, on a canvas, unlabelled at low zoom. Assign `--hue-6`, and let `BASE_RADIUS` (`:135-141`) carry meaning too. |
| high | S | `roadmaps.vue:39-46, :108-121` | A hand-rolled empty state that **overrides the global `.empty-state`**, telling a person to *"use the `roadmap_create` MCP tool"* — an instruction to call an API, given to someone looking at a web page. `EmptyState` has a `command` prop built for this, and the **share twin uses it correctly**. |
| high | S | `roadmaps.vue:94`, `s/[token]/roadmaps.vue:74` | `minmax(340px, 1fr)` overflows below a 364px viewport — 4px of horizontal page scroll on a Galaxy S8, 44px at 320px. Use `minmax(min(340px, 100%), 1fr)`. |
| medium | S | `useResearchGraph.ts:37` | Session nodes are **off by default**, while the sidebar offers them as a counted toggle. The one node type answering "which session produced this?" is hidden on arrival. |
| medium | S | `graph/index.vue:586`, `GraphSidebar.vue:96` | Double-click opens a node; nothing says so. The sidebar documents right-click only, and single click does nothing. A 786-line view is a picture unless you guess the gesture. |
| medium | S | `research/[id]/index.vue:35-40` | `/graph` and `/mindmap` are adjacent icon buttons with `title` tooltips and nothing saying that one is the *hierarchy* and the other the *reference network*. |
| medium | S | mindmap nodes, `:2` | `<div @click>` with no `role`, no `tabindex`, no key handler. Vue Flow focuses the node wrapper, but Enter does not reach the inner handler. |
| medium | S | roadmap/mindmap node files | Node max-widths of 360/380/400/420px and a private palette per node type — four independent node-styling systems for three views of the same data. |

### 5.7 Sharing — the strongest surface in the product

`s/[token].vue` has five named screens (`loading / ready / locked / dead /
unreachable`), each with real copy and a next action. `ShareDialog` moves focus
deliberately on every view change and announces it. `ShareRowList` keeps dead
rows dimmed because "who did I give this to" is the question the list answers.
`PasswordGate` deliberately does not clear the field on a wrong answer, with the
reason written down. Treat this surface as the house standard.

| Sev | Cost | Where | What |
|---|---|---|---|
| medium | M | `app.vue:43-46` | The share shell is chromeless — correct, since the nav's search reaches past the one shared research — but the consequence is that a client sent a 60-entry research has only the sidebar. A **share-scoped** search box inside the shell closes it; the API already scopes by token. |
| medium | M | `ShareDialog.vue:58-66` | A share's password cannot be changed or recovered. Forgetting it costs a revoke-and-recreate, invalidating a URL already sent. |
| medium | S | `ShareRowList.vue:25-40` | "Show link" and "Revoke" are adjacent bare `.link-btn`s at `--type-xs`, distinguished only by a hover colour — colour-only differentiation of an irreversible action. |
| medium | S | `ShareRowList.vue:132` | The comment admits the gap: *"there is no `--color-danger`"*. Fixed by 4.1. |
| low | S | `share/Banner.vue:113`, `:27-37` | Banner prose at `--type-xs`; `.share-contents` is `display: none` below 768px, so a mobile visitor loses the "what this link includes" summary — the one fact that scopes the page. |
| low | S | `ShareDialog.vue:51-56` | The expiry `<select>` wears `.text-input`, so it has no chevron and does not look like a picker. |

### 5.8 Teams and roles — second strongest

`teams/[id].vue` gets the hard things right: optimistic role change with rollback
on refusal (`:283-305`), focus moved to a real heading after a row is destroyed
(`:221-224`), disabled-control reasons exposed via `aria-describedby` because a
`title` never reaches a keyboard (`:80-82`), and reinvite that revokes the old
link only *after* the new one exists (`:381-392`). `invite/[token].vue` handles
six outcomes, each with the right sentence and the right button.

| Sev | Cost | Where | What |
|---|---|---|---|
| high | M | `teams/index.vue:10`, `teams/[id].vue:62-66`, `pages/index.vue:55-62` | **The create-team → invite → colleague-sees-nothing loop has no signpost.** The model is stated once ("Teams own researches"); the consequence is never stated — *your existing researches are still in your personal team*. A brand-new team shows no prompt to move one; the only transfer control is buried in a `⋯` menu as "Move to team…"; and the invited colleague lands on a page reading *"An agent connected to your account can create one"* — advice for a founder, given to someone invited to read somebody else's work. They will conclude the invitation was broken. |
| medium | S | `teams/index.vue:47` | `placeholder="Отдел интеграций"` — a Russian placeholder in an otherwise entirely English UI, on the first field a new team owner sees. |
| medium | S | `teams/[id].vue:467`, `teams/index.vue:110` | `.page-lead` declared twice with different rules; `.section-heading` declared once and needed in three places. |
| medium | S | `TeamMemberList.vue:107-121` | Declares its own `.btn-icon`, shadowing `main.css:549` — 28px and borderless against 30px and bordered. |
| medium | S | `roadmaps.vue`, `roadmap/[roadmapId].vue`, `export.vue` | `ViewerNotice` is missing from these three research-scoped pages, so a viewer meets suppressed-but-unexplained controls. |
| low | S | `TeamMemberList.vue:9` | Ten members are ten identical cyan circles; derive a hue from the user id the way `useTagHue` already does. |

### 5.9 Export

`export.vue:129-133` maps each failure status to the one act that can answer it
(404 → reload, 401/403 → sign in, else → retry) and `:137-139` explains slowness
only when it is slow. Good surface.

| Sev | Cost | Where | What |
|---|---|---|---|
| medium | M | `research/[id]/index.vue:61-64` vs `export.vue` | **The two halves of one feature are on different screens.** "Download .md", "Obsidian .zip" and "Print / PDF" live on `/export`; "Download JSON" — the portable format that round-trips through the home page's Import button — is in the research `⋯` menu, and neither mentions the other. |
| medium | S | `research/[id]/index.vue` | Export is two clicks behind a hidden `⋯` menu, for the action that ends most reading sessions, while Share beside it earned a labelled button. |
| medium | M | `export.vue:53`, `ExportDocument.vue` | The printable document is the second long-form reading surface and inherits the missing measure. |
| medium | L | `session/[sessionId]/export.vue` vs `ExportDocument.vue` | A 368-line fork of a 403-line component: **26 shared class names**, ~230 near-identical CSS lines, the same render pipeline written twice, and measurable drift (`.entry-code` at `3px`/`700` vs `--radius-sm`/**`650`**; `.entry-heading` at `--type-lg` vs `--type-base`). Give `ExportDocument` a `scope` prop. |
| low | S | `export.vue:3-4, :39-42` | Inline `style="height: 60px; margin-bottom: 2rem"`; and at 768px the row wraps so the deliberate Print-vs-download grouping evaporates. |

### 5.10 Auth and settings — the least-maintained files

| Sev | Cost | Where | What |
|---|---|---|---|
| critical | S | `settings.vue:207-217` | `.auth-button { background: var(--color-primary); color: white }` — **#ffffff on #6cc5e0 is 1.9:1**, a WCAG failure by a factor of two, on the button that creates an API key. `main.css:300-305` gets it right with `color: var(--color-bg)`. Delete the local class; use `.btn .btn-primary`. |
| critical | S | `settings.vue:63-72, :139` | API-key revoke: no confirm, silent failure. See 3.5. |
| critical | S | `GettingStartedBanner.vue:14-16` | **The onboarding dead-ends at its first step.** *"Add to Claude — configure mcp-research as an MCP server"* with no URL, no config snippet, no mention that an API key is needed, no link to `/settings`. The strings `sse`, `mcp-remote`, `claude_desktop_config` and the server's own base URL appear in **no** frontend component; `internal/docs/mcp-client-guide.md` is served only over MCP — to the agent that is already connected. Step 2 (a copyable prompt) is beautifully done and useless until step 1 succeeds. Fix: show `{base_url}/mcp` and `{base_url}/sse` with a copy button, and a "Create an API key →" link; put a ready-to-paste JSON block beside the new key. |
| high | S | `settings.vue:112-115` | The once-only API key is a bare `<code>` with **no copy button**, under the words "Copy it now — it won't be shown again". `CopyableSecret.vue` exists and is used by both `ShareDialog` and `InviteDialog` for exactly this. |
| high | M | `login.vue:400-434`, `settings.vue:207` | Two different `.auth-button`s in one product: a gradient with `animation: btnShift 4s infinite` on the auth pages, and the inaccessible cyan one in settings. Rename the auth one `.auth-cta`; delete settings'. |
| medium | S | `settings.vue:150` | `.page-title { font-weight: 600 }` overrides the global `700`, so Settings' H1 is visibly lighter than every other page's. |
| medium | S | `settings.vue:121` | The key-name input has **no label at all**. Also `QuestionList.vue:5,9` (two unlabelled selects) and `pages/index.vue:18` (status select unlabelled while the team select beside it has `aria-label`). |
| medium | S | `settings.vue:80, :142, :199` | No breadcrumbs, unlike `/teams` reached from the same menu; the only empty state in the app that is not `EmptyState`; and `.auth-error` duplicating `.inline-error`. |
| medium | S | `login.vue:80` | The auth error `<div>` has no `role="alert"`, no `aria-describedby`, no `aria-invalid` — a failed sign-in is silent to a screen reader. `share/PasswordGate.vue` gets all of this right. |
| medium | S | `settings.vue:100-106` | `.keys-table` is a 4-column `width: 100%` table with no `overflow-x` and no responsive rule — horizontal page scroll below ~500px. |
| low | S | `login.vue:219, :222` | `font-size: 3.25rem` off-scale, and the only pure-white ink in the product. Keep the hero but demote it: drop the orb opacity and remove `btnShift` from the button — the hero may be atmospheric, the primary control should not shimmer. |

### 5.11 Cross-cutting chrome

| Sev | Cost | Where | What |
|---|---|---|---|
| high | S | `main.css` (whole file) | **No `prefers-reduced-motion` block anywhere**, while `main.css` owns `fade-up` on every card (`:1039-1046`), a 0.05→0.45s stagger (`:1048-1056`), `.card:hover` and `.btn:hover` translates, and two `infinite` animations (`skeleton-shimmer`, `pulse-dot`) — the latter being WCAG 2.2.2 territory when an API is slow. `html { scroll-behavior: smooth }` is unguarded too. One block, lands on every screen. |
| high | S | `SearchModal.vue:9-10` | The command palette does **not** use `ModalOverlay`: no focus trap, no focus restore, no `role="dialog"`, no `aria-modal`. Escape works only while focus is in the input, and since it is teleported to the end of `<body>`, Tab walks the entire page behind it. `ModalOverlay.vue:21-29` exists to prevent exactly this. |
| high | S | `SearchModal.vue:122-124` | The search URL is a reactive computed — **a request per keystroke**, no debounce, no loading state. Typing "authentication" fires 14 requests, and during each one the modal reads "No results for …". |
| high | S | `SearchModal.vue:34, :53, :150` | Arrow keys move a purely visual `cursor` class: no combobox/listbox/option roles, no `aria-activedescendant`, and the active item is never scrolled into view, so past ~8 results the highlight leaves the screen. |
| high | S | `ConfirmModal.vue:79-84` | Redefines the global `.btn-danger`. See 4.5. |
| high | S | `ModalOverlay.vue:125-133` | `.modal-card` sizes `sm`/`md` have **no `max-height` and no `overflow-y`** (only `lg`/`xl` do), and the overlay is `align-items: center` — so a dialog taller than the viewport is clipped at the top with no way to scroll to it. At 375×667 with a soft keyboard up, `ConfirmModal` with a long message and `TransferModal` are both at risk. Also no background scroll lock. |
| high | S | `ToastHost.vue:17-18` | `role="alert"` **combined with an explicit `aria-live="polite"` on the same element**, so the explicit value wins and error toasts queue politely. And the live region *is* the toast, inserted with its content — a live region must exist in the DOM beforehand. Put the `aria-live` containers on the persistent `.toast-host` at `:4`. |
| medium | S | `ActionMenu.vue:16-22` | `role="menu"` whose children carry no `role="menuitem"`, and no arrow keys — a screen reader announces a menu with zero items. Decision: **drop `role="menu"`**. These are one-shot commands, not a menu bar; an honest popover of buttons is cheaper and correct. |
| medium | S | `app.vue:63-68, :92-113, :220, :247` | The user menu's click-outside listener is registered on mount and never removed; no Escape, no `aria-expanded`, no `aria-haspopup`, no focus return; `z-index: 100` as a literal; `var(--color-danger, #dc2626)` where the fallback is load-bearing. `ActionMenu` does all of this correctly — render the nav menu through it with a `#trigger` slot. |
| medium | S | `ActivityIndicator.vue:2` | A plain `<div>` with no `role` — the "somebody else is changing this page" signal does not exist for a screen reader. `ConnectionStatus.vue:21` is the model and is the only component in the app that gets this right. |
| medium | S | `StatusBadge.vue:11-26` | One map serves both status (11 values) and priority (3), which is why priority renders in status clothing. Split it. |
| medium | S | `main.css:94` + everywhere | **No skip link.** On a research page a keyboard user passes the logo, search, the user menu, 10 sidebar tabs and 6 header controls before reaching content. |
| medium | S | `useKeyboardNav.ts` | Binds exactly one shortcut — `Shift+G` → `/` — documented nowhere, and it guards `INPUT`/`TEXTAREA`/`SELECT` but not `[contenteditable]`. `SearchModal.vue:89-96` registers `Cmd/Ctrl+K` on `window` with an anonymous handler that is **never removed**, and it fires while the markdown editor has focus, where Ctrl+K means "insert link". |
| medium | S | `ModalOverlay.vue:131-151` | Four hardcoded dialog widths — the dialog scale living inside one component. Promote to `--dialog-sm/md/lg/xl`. |
| low | S | `main.css:106-116` | The noise overlay sits at `--z-noise: 200`, above `--z-overlay: 100`, so the grain lies over every modal and text input. Deliberate per the `--z-toast` comment, but worth confirming. |
| low | S | `SearchModal.vue:5` | `⌘K` shown on every platform while the handler accepts `metaKey \|\| ctrlKey`. |

**Modal label audit** — `ConfirmModal.vue:2`, `team/InviteDialog.vue:2` and
`research/TransferModal.vue:2` pass `labelledby` and associate their labels
correctly. `entry/HistoryPanel.vue:2`, `tasks/CreateTaskModal.vue:2` and
`tasks/StatusChangeModal.vue:2` pass none; the two task modals also have three
`<label class="form-label">` with no `for`, inputs with no `id`, and icon-only
close buttons with no `aria-label`.

---

## 6. Where the UI assumes the reader knows what only the agent knows

Five forms, ordered by frequency. This is the product's characteristic failure
and none of it is expensive to fix.

1. **`[[E3]]` renders as the bare string "E3".** `useCrossRefs.ts:55` emits
   `<a href="…">E3</a>` — no `title`, no hover card, no resolved entry name. Every
   entry body is dense with these, so learning what E3 is costs two navigations,
   times however many the agent wrote. `/api/entries/{id}/crossrefs` already
   returns the titles.
2. **Short codes are never explained.** `R1`, `E3`, `SS1`, `Q3`, `T2` appear as
   coloured chips on every screen with no tooltip and no legend anywhere.
3. **MCP tool names are given as instructions to humans.** `roadmaps.vue:45` —
   *"Use the `roadmap_create` MCP tool"*. A person cannot invoke an MCP tool; they
   can only ask an agent to. `pages/index.vue:69-70` gets it exactly right by
   handing over **a sentence to type into Claude**. Every agent-facing empty state
   should follow that pattern.
4. **"Section" is never defined**, though it is the primary navigation axis of
   every research — and the agent names them in snake_case (`open_questions`),
   with `display_name` optional and often absent.
5. **The teams model is stated but its consequence is not.** See 5.8.

A related dead end: a dangling `[[E9]]` is linked unconditionally
(`useCrossRefs.ts:51`), and the reader lands on `entry/[entryId].vue:225` —
`<EmptyState icon="🔍" title="Entry not found" />` with no description, no action
and **no breadcrumbs**, because the breadcrumbs live inside the `v-else-if="entry"`
branch at `:7`. Browser Back is the only way out. `EmptyState.vue:12-13`'s own
comment says *"An empty state that names no action is a dead end"* — and six
owner-side not-founds pass no slot.

---

## 7. What the interface cannot do at all

Ranked by cost × frequency. Each is a missing screen or control, not a polish
item.

1. **Answer a question, or correct an answer.** `PUT /api/questions/{id}` exists
   and nothing calls it. Needs: an answer editor plus status controls on
   `question/[questionId].vue`.
2. **Tell what changed since the last visit.** No activity feed, no unread state,
   no per-entity "new" marker, and **no timestamp on entry cards at all**. The
   only change signal is a 5-second "Updating" blip in the nav with no entity, no
   link and no history. The WebSocket already carries every event needed.
3. **Search within one research.** `SearchModal` is global only; tags are the
   only in-research filter and they exist only if the agent applied them. Search
   results also show no snippet, so a body match is a title with nothing
   highlighted.
4. **See who referenced this entry when the referrer is a question, task or
   session note.** The data is stored and returned — `entry_code: null`,
   `source_type: "question"` — but `CrossReferencesBlock.vue:37, :66-89` renders it
   as a blank inert row, because `refLink()` returns `''` for any source type
   other than `entry`. **The interview answer that cites an entry is unreachable
   from that entry**, in a product whose thesis is provenance.
5. **Change a task's status without a mouse; set a task to `blocked` or
   `deferred`.**
6. **Close, pause or rename a session.**
7. **Recover from a research with zero sections.** The page renders *"Select a
   section — choose one from the sidebar"* beside a sidebar containing none.
8. **Change or recover a share link's password.**
9. **Print a single entry.** `PrintButton.vue` exists, unused.

**Built and buried** — paths that exist and nobody would find: double-click to
open a graph node; the graph's session nodes (a filter off by default); **External
links** (a genuinely useful roll-up of every URL the agent cited, grouped by
domain, sitting as a nameless icon below a sidebar divider); shift-click to
compare revisions; the entry provenance chip (the best answer in the product to
"did a human or the agent write this", at 11px grey); the `?team=` URL filter;
"Move to team…"; the **Details panel**, which holds `instruction` and `memory` —
the fields governing every future agent run — behind `⋯ → Details`.

---

## 8. Order of work

Each wave makes the next cheaper. Wave 0 and Wave 1 are prerequisites for
everything after them.

### Wave 0 — defects, ship independently

Nothing here waits on a decision.

1. **Escape `renderRefs` and its 15 call sites**, plus `SearchModal.highlight()`. (3.1)
2. **Fix Prev/Next** on the entry page — the patch is in the share twin. (3.2)
3. **Define `--color-border-hover` and `--color-danger`.** Two lines, two live bugs. (3.3)
4. **Route the four `parseTimestamp` bypasses through the composable.** (3.4)
5. **`.entry-content { max-width: var(--measure) }`.** One rule, the product's central job. (4.3)
6. **`settings.vue`'s 1.9:1 button** → `.btn .btn-primary`. (5.10)

### Wave 1 — the token contract (no visual change intended)

Land the whole `:root` from 4.1 **before** the CSS split starts, since the split's
entire point is that scoped styles are written against tokens.

7. Add every new token; retune `--type-3xl`; delete `--type-4xl`, `--transition-spring`, `--z-base`.
8. Sweep-replace the 84 raw radii, the 18 shadows, the 176 numeric weights and the hexes that duplicate token values. Mechanical, greppable, reviewable.
9. Delete the ~20 lines of dead CSS (4.2).

Falls out for free: `ShareRowList`'s admitted gap, `app.vue:248`'s `#dc2626`, `entry:741`'s fourth red.

### Wave 2 — safety, and the destructive-action contract

10. Delete `ConfirmModal`'s `.btn-danger` override; add `.btn-danger--solid`; fix `.confirm-icon.info`. (4.5)
11. Replace all four `alert()` calls with toasts carrying retry actions.
12. Add confirms and error handling to every silent write in 3.5.
13. Destructure `error` in the six pages of 3.6 and give each an error branch with **Try again**.
14. Promote `.danger-zone` / `.danger-row`; adopt in `settings.vue`.
15. `SearchModal` and the entry status dropdown adopt `ModalOverlay` / `ActionMenu`; delete `z-index: 9999`; add `max-height` + `overflow-y` + scroll lock to `ModalOverlay`.

### Wave 3 — accessibility, highest ratio in this document

16. **One global `prefers-reduced-motion` block.** Lands on 100% of screens. (5.11)
17. `aria-label` on every icon-only control, starting with entry delete.
18. `labelledby` on `HistoryPanel`, `CreateTaskModal`, `StatusChangeModal`; `for`/`id` on their inputs; `role="alert"` on the auth error.
19. Convert `<span @click>` tag filters to `<button aria-pressed>`.
20. Fix the `ToastHost` live-region conflict; give `ActivityIndicator` a `role`.
21. Remove `opacity` from text: `.readonly-badge` (3.17:1), `.kanban-empty` (2.25:1), `.rail-hint` (3.44:1), the graph greys. Use `--color-text-faint`.
22. Add a skip link.

### Wave 4 — the CSS split, with the primitives introduced during it

The expensive-but-cheap-now wave. Each component gets **one** visit, in which its
tokens, its shadow, its radius and its duplicated class are all fixed together.

23. Split the files (4.2), preserving import order.
24. Introduce `.cluster` / `.stack` / `.page-bar` / `.data-row` / `.select` / `.tabs` / `.segmented` / `.check-row` / `.centered-gate`. (4.6)
25. Mechanical component moves: `.entry-nav-btn`, `.ws-label`, `.filter-bar`, `.empty-state`, the dead global `.progress-bar` pair.
26. Per-component token adoption, in payoff order: the 10 flow nodes → the 5 modals → the graph's scoped token block.
27. Retire the 16 `.short-code` clones for `ShortCode.vue`; resolve the three-way `.btn-icon` collision; add `utils/truncate.ts` and delete 10 local copies; add the two missing `tagHue` imports.
28. Add the `scripts/` lint guard so the split does not undo itself.

### Wave 5 — the surfaces that are functionally broken

29. **Task status without a mouse** — a `.segmented` control in `TaskDetailModal`, feeding the existing `StatusChangeModal` path. (3.8)
30. **An answer editor on the question page**, plus session status controls. (3.9)
31. Show `blocked`/`deferred` on kanban cards; split `StatusBadge` / `PriorityMark`; rename the "Rejected" column.
32. Fix the four sequential awaits so the existing skeletons work; add `<NuxtLoadingIndicator>`. (3.7)
33. Session tab state into the URL; real `role="tab"` via the Wave 4 primitive.
34. Research header triage: views become tabs, actions collapse to one primary plus one menu.
35. Roving tabindex on the research sidebar; delete or fix the fake section progress bar.
36. Mindmap toolbar wrapping; the roadmap grid `minmax` overflow; the kanban `max-height`.

### Wave 6 — the component extractions

In dependency order, so each makes the next cheaper:
`ModalHeader` → `CommandLine` → `PageHeader` → `EditableField` → `AuthShell` →
`AppNav` → `FlowNode` → `ExportDocument` session mode. Then the risky global
moves: the `.sidebar` family (5 referencing files, one of which is a *different*
sidebar), `.warning-banner`, `.readonly-badge`, and the page-header family — none
of which moves before `PageHeader.vue` exists.

### Wave 7 — the product gaps

37. Onboarding: connection details and an API key on the banner and in settings. (5.10)
38. Inline `[[E3]]` resolves to a title on hover; the crossref block renders question/task/session sources. (6, 7)
39. A "what changed" surface: timestamps on entry cards first, then a per-research activity list.
40. Search within a research, and inside the share shell; debounce the global search; combobox roles.
41. The team → invite → move-a-research signpost. (5.8)
42. Graph: one palette, `question` off amber, the `role="img"` label and a list fallback.
43. Entries and questions as rows by default — **after** looking at it in a browser. (4.7)

---

## 9. Deliberately not decided here

- **A light theme.** Deferred, with the reasoning in 4.4. Waves 1 and 4 are
  exactly the work it would need first, so deferring costs nothing.
- **`md-editor-v3` theming.** The edit mode is visibly a different application
  and the 25-line `:deep()` block is an upgrade liability. Either scope it
  properly or accept it as a documented foreign surface — but not during a token
  pass.
- **Whether the noise overlay should sit above dialogs.** Currently deliberate;
  confirm with eyes.
- **Storybook's auto-import allow-list.** `.storybook/main.ts:38-42` hand-maintains
  what Nuxt auto-imports, and **11 components currently use an auto-imported
  composable with no explicit `import`** — they work in the app and crash in
  Storybook the moment the list falls behind. The fix is the explicit import, not
  a local copy of the function; `EntryCard.vue:20-21` and `TagList.vue:13` show
  the pattern. Worth its own pass.

## 10. What needs a browser before it is called done

Static reading took this as far as it goes.

1. **`animation-delay` retiming.** `main.css:1048-1056` sets delays by
   `:nth-child`. When a realtime event prepends a card, every later card's delay
   changes — which per spec retimes the animation and, with `fill-mode: both`, can
   push a finished one back into its delay phase at `opacity: 0`. Engine-dependent.
2. **Control heights.** The arithmetic in 5.3 assumes `button`/`select` inherit
   `line-height: 1.55` from `body`. The *mismatches* hold either way; the absolute
   numbers do not.
3. **The sidebar's horizontal scroller at 768px on a trackpad** — 10 items,
   ~1200px of content, `scrollbar-width: none` and no fade or arrow.
4. **Modal height at 375×667 with a soft keyboard up.**
5. **The teleported entry status dropdown during scroll.**
6. **Graph canvas on touch, and whether its backing store goes stale when the
   sidebar collapses** — there is no `ResizeObserver`, only a `window.resize`
   listener.
7. **Outfit's Cyrillic/CJK fallback.** The Google Fonts subset has no Cyrillic, so
   Cyrillic and CJK fall through to the system stack — a different face at a
   different optical size, mid-list. Either request `&subset=cyrillic` or state an
   explicit fallback stack; the stack also names no CJK face.
8. **`md-editor-v3` keyboard conflicts** — `Cmd/Ctrl+K` and `Shift+G` while the
   editor has focus.
9. **Screen-reader pass on `ToastHost`** — NVDA and VoiceOver differ on the
   `role="alert"` + `aria-live="polite"` conflict.
10. **Print output** — page breaks inside code blocks and mermaid diagrams,
    whether toolbars are suppressed, and whether `[[E3]]` prints as a readable code.
