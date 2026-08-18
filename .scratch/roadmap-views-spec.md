# Roadmap view modes — build-ready UI specification

Surface: `frontend/pages/research/[id]/roadmap/[roadmapId].vue` (editable) and its
read-only twin `frontend/pages/s/[token]/roadmap/[roadmapId].vue`.
Composable: `frontend/composables/useRoadmap.ts`.

Three views over one roadmap: **Graph** (existing, unchanged), **Stages** (kanban
columns by `node.stage`), **Timeline** (nodes on a date axis by `node.node_date`).
A view toggle switches between them. Additive — edges never disappear silently.

---

## 0. DATA-CONTRACT GAPS — action these in the API now

The contract you gave lists the *read* fields. The design also needs them written,
and needs them on the GET payload. Concretely:

1. **`GET .../roadmaps/{roadmapId}` must return the new fields.** Add to the
   `Roadmap` payload: `view` ("graph"|"stages"|"timeline"), `stages` ([]string).
   Add to each `RoadmapNode` in the payload: `stage` (string), `node_date`
   (string "YYYY-MM-DD" or ""). Today `internal/domain/roadmap.go` has none of
   these — the frontend cannot open in the right view or place a single card
   until they ship on the read path. **This is the one that blocks everything.**

2. **`PUT /api/roadmap-nodes/{nodeId}` must accept `stage` and `node_date`.**
   The v1 edit path (set stage / date from the node popover) writes through this
   handler. `handlers/roadmap.go UpdateNode` input struct has neither field, and
   `service.UpdateRoadmapNodeRequest` will need `Stage *string` and
   `NodeDate *string`. Without these the board modes are read-only *even for the
   owner* and stages can only be set by the agent over MCP.

3. **`PUT /api/roadmaps/{id}` must accept `stages []string` and `view`** (for
   the owner to define the columns and set the default opening view). Today
   `Update` accepts title/description/statuses/status only. `stages` mirrors how
   `statuses` is already handled — copy that path exactly.

4. **NOT needed — confirmed derivable, do not add:**
   - *Dependency chips* ("depends on R1:E3"). Derived on the client from the
     existing `edges[]` (source/target node ids) joined to each node's `code`.
     No `predecessors` field required.
   - *Per-stage progress counts.* Derived from `nodes[]` + `statuses[]`, same as
     the existing header progress.

5. **OPTIONAL, fast-follow only (flag, don't build for v1):** an intra-column
   **order / rank** field on `RoadmapNode` (e.g. `stage_order int`). v1 orders
   cards within a column by existing array order (code order). Drag-to-reorder
   and drag-between-columns both need a persisted rank; until that field exists,
   dragging cannot be truthful, so v1 does not offer it (see §6).

Everything below is buildable the moment items 1–3 land. Nothing in the design
needs a field beyond those.

---

## 1. The job

> Look at the same roadmap three ways — as a dependency graph, as ordered phase
> columns, or on a date axis — and switch between them in one click.

Target: **1 step** to change view (single control). Setting a node's stage/date
is a secondary, owner-only, 2-step path (open node → pick value) reusing the
existing detail popover.

---

## 2. Screens and entry points

No new routes. The two existing roadmap pages gain a **view toggle** in the
toolbar and swap their canvas body between three renderers.

| Where | Entry | Replaces |
|---|---|---|
| Owner roadmap page | Toolbar view toggle | Nothing — Graph stays the default body |
| Shared roadmap page | Same toggle (read-only) | Nothing |

Opening view = `roadmap.view` from the payload (default `"graph"` if absent).
The toggle overrides **locally only** in v1 (ephemeral; not persisted per user).
Persisting a new default is the owner editing `roadmap.view` — deferred unless
gap item 3 ships; if it does, an owner-only "Set as default view" is a trivial
add, out of scope for v1 wiring.

---

## 3. Layout

### Toolbar (both pages, desktop)

The view toggle is the primary new control. It sits at the **left of
`toolbar-right`**, ahead of progress, so it reads before the mode-specific
controls it governs. The graph-only controls (LR/TB, Auto layout, Fit view) are
grouped after it and **hide entirely in Stages/Timeline** — a computed layout is
not dragged, auto-laid-out, or fit-to-viewport.

```
DESKTOP (graph mode)
┌───────────────────────────────────────────────────────────────────────────┐
│ [< Back]  Title  R3        [ Graph | Stages | Timeline ]  · 4/9 ▓▓▓░░  │ LR TB │ Auto layout │ Fit │
└───────────────────────────────────────────────────────────────────────────┘

DESKTOP (stages / timeline mode) — graph-only cluster gone, separator collapses
┌───────────────────────────────────────────────────────────────────────────┐
│ [< Back]  Title  R3        [ Graph | Stages | Timeline ]  · 4/9 ▓▓▓░░       │
└───────────────────────────────────────────────────────────────────────────┘
```

Toolbar tokens are unchanged from today: `padding: var(--space-3) var(--space-5)`,
`gap: var(--space-2)` in `toolbar-right`, `--toolbar-sep` divider. The view
toggle uses the existing `.segmented` control (system.css) at `--control-h-sm`,
`--type-xs`.

### Stages board

Full-bleed inside `.roadmap-canvas` (which is already `flex:1; min-height:0`).
A horizontal flex row of columns, `gap: var(--space-4)`, `padding: var(--space-5)`,
`overflow-x: auto` (the board scrolls sideways, the page does not). Each column
is fixed width so 12+ columns scroll rather than crush:
`width: 300px` (min 260 / max 320 via `--space` multiples; column min-width
matches the card min-width of 280px + padding). Column body scrolls vertically
independently (`overflow-y: auto`), header is sticky within the column.

```
DESKTOP — Stages
┌─ Discovery ── 3 ─┐ ┌─ Design ── 2 ─┐ ┌─ Build ─ 0 ─┐ ┌─ Launch ─ 1 ─┐ ┌─ Без этапа ─ 2 ─┐
│ ┌──────────────┐ │ │ ┌───────────┐ │ │             │ │ ┌──────────┐ │ │ ┌────────────┐  │
│ │ E1 card…     │ │ │ │ E4 card…  │ │ │  (empty)    │ │ │ M1 ◆ …   │ │ │ │ S2 card…   │  │
│ │ → depends R1 │ │ │ └───────────┘ │ │  Nothing    │ │ └──────────┘ │ │ └────────────┘  │
│ └──────────────┘ │ │ ┌───────────┐ │ │  here yet   │ │              │ │ ┌────────────┐  │
│ ┌──────────────┐ │ │ │ E5 card…  │ │ │             │ │              │ │ │ E9 card…   │  │
│ │ E2 card…     │ │ │ └───────────┘ │ │             │ │              │ │ └────────────┘  │
│ └──────────────┘ │ └───────────────┘ └─────────────┘ └──────────────┘ └────────────────┘
        →→→ scrolls horizontally when columns exceed viewport →→→
```

### Timeline

A month-granularity axis. Header row = month cells (with quarter band labels
above), body = one horizontal lane band; cards drop into the month cell of their
`node_date`, stacking vertically when several share a month. Milestones render as
a diamond marker pinned to the axis line, not as a full card (see §3-timeline).
Undated nodes live in a persistent, collapsible **tray** below the axis.

```
DESKTOP — Timeline
        Q1 2026                    Q2 2026
┌ Jan ─┬ Feb ─┬ Mar ─┬ Apr ─┬ May ─┬ Jun ─┐
│      │ E1   │      │ E4   │      │ ◆ M1 │   ← milestone diamond on axis
│      │ card │      │ card │      │Launch│
│      │      │      │ E5   │      │      │
│      │      │      │ card │      │      │
└──────┴──────┴──────┴──────┴──────┴──────┘
▸ No date (3)   ← tray, collapsed by default when axis has content
```

### ≤768px

The toolbar already wraps (`flex-wrap`) and hides the title. The view toggle
stays visible (it is the point of the screen); it drops below the back link on
the wrap line. Graph-only controls still hide in board modes, which *reduces*
wrap pressure on small screens.

- **Stages on mobile:** columns keep `width: 300px` and scroll horizontally
  (kanban is inherently horizontal; a single-column collapse would lose the
  phase axis, which is the feature). A `scroll-snap-type: x proximity` on the
  board and `scroll-snap-align: start` on columns makes one-column-at-a-time
  swiping comfortable.
- **Timeline on mobile:** month cells keep a fixed `width: 120px` and scroll
  horizontally. The undated tray goes full-width below.

```
MOBILE ≤768px — Stages
┌───────────────────────────┐
│ [<]  [Graph|Stages|Time]  │  ← toggle wraps under
│ · 4/9                     │
├───────────────────────────┤
│ ┌─ Discovery ─ 3 ─┐  ┌─ De…│  ← snap-scroll columns
│ │ E1 card…        │  │ E4 …│
│ │ E2 card…        │  │     │
│ └─────────────────┘  └─────│
└───────────────────────────┘
```

---

## 3-stages. Stages layout details

- **Column header:** stage name (`--type-xs`, `--weight-semibold`, truncated
  with `title` attr on overflow) + count badge reusing the `.tab-count` visual
  (`--color-surface-hover`, `--radius-xs`, tabular-nums). Sticky within the
  column at `top:0`, `background: var(--color-surface)`.
- **Empty column:** a stage in `roadmap.stages[]` with zero nodes still renders
  (columns are declared, not derived — same as `statuses[]`). Body shows a muted
  centered line, `--color-text-faint`, `--type-xs`: **"Nothing here yet"** /
  RU **"Здесь пока пусто"**. No action button inside the column (adding a node
  to a stage is an agent/API action in v1; the board is not a node creator).
- **"Unassigned" / leftover column:** collects any node whose `stage` is `""`
  **or** a value not present in `roadmap.stages[]`. Rendered as the **last**
  column, header **"Unassigned"** / RU **"Без этапа"**, with a subtly distinct
  header tint (`--color-surface-hover`) so it reads as not-a-real-phase. It only
  appears when it has ≥1 node. When every node is assigned, it is absent.
- **Card inside a column:** the extracted `RoadmapNodeCard` (see §5), full width
  of the column, `margin-bottom: var(--space-3)`. Same content the graph node
  shows today (code, title, status chip, type badge, ref preview/result/session
  bar). No Vue Flow `<Handle>`.
- **Dependency EDGES across columns — RECOMMENDATION: do not draw them.**
  Instead, each card gains an optional footer line listing predecessor codes it
  depends on, e.g. **"↳ depends on E1, R2:E5"** (rendered from `edges[]`, links
  via the code so click scrolls/opens the target). On card hover/focus, all its
  predecessors and successors across every column get a highlighted border
  (`--color-border-strong`) and non-related cards dim to `opacity: 0.5`.
  **Justification:** columns scroll vertically and independently and the board
  scrolls horizontally; persistent SVG edges would need an absolutely-positioned
  overlay recomputed on every scroll and reflow, and across four+ columns it
  reads as spaghetti, not dependency. The chip is *truthful, cheaper, keyboard-
  reachable, and screen-reader legible* where a drawn line conveys nothing to
  assistive tech. This satisfies "gracefully handled" without pretending the
  kanban is a graph. (Interaction + a11y chairs agreed; the visual chair wanted
  faint drawn lines for at-a-glance flow — overruled because scroll makes them
  lie about position. The hover-highlight is the compromise that keeps the
  at-a-glance relationship.)

## 3-timeline. Timeline layout details

- **Granularity: month, with quarter band labels above. Justification:** research
  roadmaps are sparse and coarse (a handful of dated milestones over weeks to
  months); week columns produce a mostly-empty axis, quarter columns collapse
  distinct dates into the same cell and lose the ordering that is the whole
  point. Month is the granularity where dated nodes separate without a desert of
  empty cells; the quarter band gives the wider frame for free. The axis spans
  from the earliest to the latest `node_date` present (no dates → see states),
  rounded out to whole months.
- **Same-date stacking:** nodes in the same month cell stack vertically in the
  lane, code order, `gap: var(--space-2)`. The cell grows; the lane band scrolls
  vertically if a single month is very tall.
- **Undated nodes:** a persistent tray under the axis, header
  **"No date (N)"** / RU **"Без даты (N)"**, collapsible (`<details>`-style
  disclosure). Collapsed by default when the axis has ≥1 dated node; expanded by
  default when nothing is dated (so the roadmap is never a blank axis — see
  states). Cards in the tray are the same `RoadmapNodeCard`, laid out as a
  horizontal wrap.
- **Edges in timeline:** same rule as stages — not drawn. Same dependency chip +
  hover-highlight. A drawn Gantt dependency arrow across months is even less
  tractable than in kanban because the vertical position is arbitrary.
- **Milestones (`node_type: milestone`) — special treatment: YES.** A milestone
  renders as a compact **diamond marker** (◆, using `--hue-5` violet already
  assigned to milestones in the step-node tint) sitting *on the axis line* with
  its code + short title beneath, rather than as a full stacking card. This is
  the Gantt convention (a milestone is a point in time, not a bar) and it makes
  the dated checkpoints legible among regular cards. Undated milestones fall into
  the tray as ordinary cards. Non-milestone nodes always render as full cards.

---

## 4. States

Written for **every** view and every new component.

### Loading
Reuse the existing `.skeleton-card` treatment already in the page. Board modes
show 3 ghost columns each with 2 skeleton cards; timeline shows a ghost axis
(6 month cells) with 2 skeleton cards. No spinners.

### Error
Unchanged from today: centered `card-meta` line + Retry. Copy owner:
**"Couldn't load this roadmap."** + **[Retry]**. Shared page keeps its existing
**"This roadmap could not be loaded."** + **[Try again]**.

### Empty & edge states — with copy and next action

| Situation | Where | Copy (EN / RU) | Next action |
|---|---|---|---|
| View set to stages/timeline but `stages[]` empty / no dates AND roadmap has nodes | Board body | title **"No stages defined yet"** / **"Этапы ещё не заданы"** · desc **"This roadmap's nodes aren't organised into phases. Ask the assistant to set stages, or switch to Graph."** / **"Узлы этой карты не разбиты на этапы. Попросите ассистента задать этапы или переключитесь на Граф."** | Button **[Switch to Graph]** (local toggle). Owner also: inline hint that stages are set via the agent/API. |
| Stages view, `stages[]` defined but **every** node unassigned | Board | Columns render **empty**; the "Unassigned" column holds everything. A one-line banner above the board: **"No nodes assigned to a stage yet."** / **"Ни один узел ещё не привязан к этапу."** | Banner links **[Switch to Graph]**; cards remain usable in the Unassigned column. |
| Timeline, **nothing dated** | Timeline | No axis. EmptyState: title **"No dates on this roadmap"** / **"На этой карте нет дат"** · desc **"Add a date to nodes to place them on a timeline, or switch to Graph. All nodes are listed below."** / **"Добавьте узлам даты, чтобы разместить их на шкале, или переключитесь на Граф. Все узлы — ниже."** | The undated tray is shown **expanded** with all nodes, so the screen is never empty. Button **[Switch to Graph]**. |
| Roadmap has **zero nodes** | Any view | Existing roadmap-level empty applies (no nodes at all) — out of scope, unchanged. | — |
| Long stage names | Column header | Truncate to one line, ellipsis, `title` attr carries full text. Column width fixed. | — |
| 12+ columns | Board | Horizontal scroll (already specced). A right-edge scroll affordance (fade mask `--color-bg` gradient) hints more columns. | — |
| Cyrillic content | Cards / headers | `overflow-wrap: anywhere` already on card titles; headers truncate. Cyrillic needs no special casing — verify the count badges stay tabular-nums. | — |
| Overloaded column (50+ cards) | Column | Column body scrolls vertically; header stays sticky with live count. | — |

---

## 5. Component plan (reuse / extend / new)

Default is reuse. Every `new` row justifies itself.

| Element | Disposition | Detail |
|---|---|---|
| View toggle | **new** — `RoadmapViewToggle.vue` | 3-segment control over the existing `.segmented` class. Props: `modelValue: 'graph'\|'stages'\|'timeline'`. Emits `update:modelValue`. **Why new:** `TabBar.vue` is underline tabs bound to tab *panels* (roving tabindex, `role=tab`); a view switch is a pressed-button segmented control (`aria-pressed`), which `.segmented` already styles but has no component wrapper. Both roadmap pages need it identically — a shared component stops the two pages diverging (the exact lesson `RoadmapCard.vue`'s docblock records). |
| Node card body | **extend** (extract) — new `RoadmapNodeCard.vue` carved out of `RoadmapStepNode.vue` + `RoadmapRefNode.vue` | The presentational markup + CSS (code, title, status chip, type badge, ref preview/result/session bar, tint-by-type) moves into `RoadmapNodeCard.vue`. The two Vue Flow node components become thin wrappers: `<Handle>` + `<RoadmapNodeCard :data>`. **Load-bearing:** `<Handle>` from `@vue-flow/core` only works inside a Vue Flow provider, so the graph node components cannot be reused as-is inside a plain CSS grid. Extracting the body is the single move that lets all three modes render the *same* card and keeps the catalog from growing a fourth. Add optional prop `deps?: {code,label}[]` (dependency chips, board modes only) and `compact?: boolean` (timeline stacking). |
| Stages board | **new** — `RoadmapStagesBoard.vue` | Props: `stages: string[]`, `nodes`, `edges`, `canWrite`. Emits `node-click`. Owns column layout, unassigned bucketing, horizontal scroll, hover-highlight of related cards. No equivalent exists. |
| Stage column | **new** — `RoadmapStageColumn.vue` | Props: `name`, `count`, `nodes`, `leftover?: boolean`. Emits `node-click`. Sticky header + count badge + empty-column copy. Split out so the empty-column and leftover states live in one place; used only by the board. |
| Timeline | **new** — `RoadmapTimeline.vue` | Props: `nodes`, `edges`, `canWrite`. Emits `node-click`. Owns month/quarter axis computation, cell stacking, milestone diamonds, undated tray. No existing time axis. |
| Time axis header | **new** — `RoadmapTimeAxis.vue` | Props: `months: {key,label,quarterLabel}[]`. Pure presentational header row (month cells + quarter bands). Split from `RoadmapTimeline` for testability; nothing existing renders a time axis. |
| Dependency chip | **reuse pattern** (no new component) | Rendered inline in `RoadmapNodeCard` footer from the `deps` prop, styled like the existing `.ref-entity-code`/`.rm-code` chip. Codes link via the app's existing cross-ref/navigation path. |
| Count badge | **reuse** | The `.tab-count` visual from `TabBar.vue`'s CSS (or lift to a shared class) for column/tray counts. |
| Empty / no-stages / no-dates states | **reuse** — `EmptyState.vue` | Props `title`, `description`, default slot for the **[Switch to Graph]** button. |
| Progress in toolbar | **reuse** — unchanged | Existing `.toolbar-progress`. Works in all views (node-based, view-independent). |
| Node detail popover | **extend** — `RoadmapNodePopover.vue` | Add, gated on `canWrite`, a **Stage** select (options = `roadmap.stages[]` + "Unassigned") and a **Date** field (native `<input type=date>` → `node_date`). This is the v1 write path for board modes. Emits `update-stage(nodeId, stage)` and `update-date(nodeId, date)`; the page wires them to `PUT /api/roadmap-nodes/{id}`. On the shared page these controls do not render (read-only). |
| Canvas host | **reuse** — the page's `.roadmap-canvas` wrapper | Swap child by `v-if` on active view: `<VueFlow>` (graph) / `<RoadmapStagesBoard>` / `<RoadmapTimeline>`. |

**Note on duplication already flagged in the repo:** the shared page's own
comment (`s/[token]/roadmap/[roadmapId].vue`, ~L200) says the Vue Flow chrome
setup should become a `RoadmapCanvas` component. This change adds a second body
swap to *both* pages — strong moment to lift the shared canvas host so the toggle
+ three-way body swap live in one component both pages import. Recommended, and
noted as the tidiest path, but the spec is buildable without it (duplicate the
`v-if` swap across both pages as they duplicate the Vue Flow block today).

---

## 6. Interaction rules

### View toggle
- Keyboard: three `<button aria-pressed>` inside a `role="group"
  aria-label="Roadmap view"`. Plain Tab reaches each; Enter/Space activates.
  (Not a roving-tabindex tablist — these switch a view, they don't own panels.)
- Switching view is **instant and local**, no server round-trip in v1. State
  lives in the page (`ref`), initialised from `roadmap.view`.
- Focus stays on the toggle after switching; the newly rendered body is not
  auto-focused (would steal focus mid-scan). First card in the body is the next
  Tab stop.

### Board / timeline cards
- Each card is a `<button>` (or `role=button, tabindex=0`), Enter/Space opens the
  same `RoadmapNodePopover` used in graph mode. Focus order = column order, then
  next column (DOM order left-to-right, top-to-bottom).
- Hover **and focus** both trigger the related-card highlight (keyboard parity).
- Dependency chip codes are links: Enter navigates/scrolls to the target card
  (same column-scroll-into-view behaviour as click).

### Popover (v1 edit)
- Opening a node from any view opens the popover; focus moves to the popover
  (existing behaviour). On close, focus returns to the card that opened it.
- Escape closes the popover (existing).
- **Optimistic:** status change is already optimistic today; **stage change and
  date change are optimistic too** — the card moves column / axis cell
  immediately, then the `PUT` confirms. On failure, revert and surface the
  existing inline error. Rationale: reflow is cheap and the agent-driven realtime
  path will re-sync anyway.

### Realtime repaint under the user's hands
- The page already defers a remote refresh while `isInteracting()` (drag
  settling). Board/timeline have **no drag in v1**, so the only interaction to
  protect is an **open popover**: if a `roadmap` event lands while the popover is
  open, defer the repaint until the popover closes (extend the existing
  `refreshWhenIdle` guard to also check "popover open"), so a card doesn't
  restack out from under the reader mid-edit.
- Column membership and axis position are recomputed from fresh data on every
  (non-deferred) refresh — a remote stage/date change animates the card to its
  new column/cell via a simple CSS transition, it does not require a full remount.

---

## 7. Data contract (per screen)

All three views read the **same single endpoint** already in use:

- Owner: `GET /api/researches/{id}/roadmaps/{roadmapId}` (via `useRoadmap`).
- Shared: `GET /api/shared/{token}/researches/{id}/roadmaps/{roadmapId}`.

Fields consumed:

| View | Fields used | Missing today (see §0) |
|---|---|---|
| Graph | (unchanged) nodes, edges, positions, ref_data | — |
| Stages | `roadmap.stages[]`, `node.stage`, node code/title/status/type/ref_data, `edges[]` (for chips) | `roadmap.stages`, `node.stage` on GET (gap 1); write via gap 2/3 |
| Timeline | `node.node_date`, `node_type` (milestone), node card fields, `edges[]` | `node.node_date` on GET (gap 1); write via gap 2 |
| Toggle default | `roadmap.view` | `roadmap.view` on GET (gap 1) |

The shared read path must return the same new fields (they are not sensitive —
stage names and dates are roadmap structure, like node titles). No redaction
needed; confirm `clampForShare`/`redactForShare` do not strip them.

---

## 8. Out of scope (deliberately not designed)

- **Drag-to-reorder / drag-between-columns** — needs a persisted rank field
  (gap 5); v1 is read + popover-set. Fast follow.
- **Creating nodes from the board** (no "+ add card"). Nodes are created by the
  agent/API; the board organises existing nodes.
- **Editing the stage list from the UI** (`stages[]` is owner/agent-set via the
  API; no in-board "add column" affordance in v1).
- **Persisting a user's chosen view** beyond the roadmap default — toggle is
  local/ephemeral in v1.
- **Drawn dependency edges in board/timeline** — explicitly rejected (§3); chips
  + highlight instead.
- **Week/day timeline granularity and zoom** — month only for v1.
- **Light theme** — the product is dark-theme-only (tokens define one theme);
  contrast was checked against the single dark surface set.
- **Print/export of the board and timeline** — existing roadmap export is graph;
  not extended here.
