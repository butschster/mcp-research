# Roadmap Timeline — Range bars (Gantt) + Axis zoom: build-ready UI spec

Surface: the **Timeline** view inside
`frontend/pages/research/[id]/roadmap/[roadmapId].vue` and its read-only twin
`frontend/pages/s/[token]/roadmap/[roadmapId].vue`.
Components touched: `frontend/components/roadmap/RoadmapTimeline.vue`,
`RoadmapTimeAxis.vue`, `RoadmapNodePopover.vue`, and `frontend/utils/roadmap.ts`.

Two additions, both to the Timeline only. Graph and Stages are untouched.
Everything here is additive: a roadmap that has no `node_end_date` on any node,
and never touches the zoom control, looks exactly as it does today.

---

## 0. DATA-CONTRACT GAPS — read this first

The only new field is `node.node_end_date` (ISO `YYYY-MM-DD` or `""`). Lane
assignment and unit bucketing are 100% client-side from `node_date` /
`node_end_date`. But the field is only **half-wired** in the backend right now,
and the human-editable path is the part that is missing:

1. **READ — already fine, just confirm.** `RoadmapHandler.Get` /
   `GetByResearch` serialize the `domain.Roadmap` directly
   (`writeJSON(..., rm)`), and `domain.RoadmapNode.NodeEndDate` has
   `json:"node_end_date,omitempty"`, and `roadmap_repo.go` already SELECTs
   `node_end_date`. So the GET payload emits it automatically once a row has it.
   **Frontend action:** add `node_end_date?: string` to `RawRoadmapNode` in
   `utils/roadmap.ts`. No backend read work.

2. **WRITE — GAP, and it blocks human-created bars.** `PUT
   /api/roadmap-nodes/{nodeId}` does **not** accept `node_end_date` today:
   - `handlers/roadmap.go` `Update` input struct has no `NodeEndDate *string`
     and does not pass it into `service.UpdateRoadmapNodeRequest`.
   - `service.UpdateNode` (roadmap_service.go ~line 725) applies `req.NodeDate`
     but never reads `req.NodeEndDate` — the field exists on the struct
     (`UpdateRoadmapNodeRequest.NodeEndDate *string`, line 75) but is dropped on
     the floor.
   Until this is closed, the **"End date" input added to the node popover
   (§4, Addition 1) cannot save**, and a human can only ever produce point
   nodes. Agents writing over MCP that populate `node_end_date` on create are
   unaffected — bars from agent data render fine read-only.

3. **VALIDATION — GAP, UI depends on it.** The create path already calls
   `normalizeNodeRange(nr.NodeDate, nr.NodeEndDate)` (roadmap_service.go line
   173, 621) but **that function is not defined yet** (only `normalizeNodeDate`
   exists). It must: reject a malformed end like `normalizeNodeDate` does, and
   **reject `end < start` with a distinct error** (e.g. `ErrNodeEndBeforeStart`)
   that `writeRoadmapError` maps to a 400 field error on `node_end_date` (mirror
   the existing `ErrInvalidNodeDate → writeFieldError(..., "node_date")` case at
   handler line 20). The popover surfaces that field error inline. The client
   also guards optimistically (§6), but the server 400 is the source of truth.
   `UpdateNode` must call the same `normalizeNodeRange` over the resulting
   `{NodeDate, NodeEndDate}` pair, not `normalizeNodeDate` alone — otherwise a
   PUT that sets only the end, or clears the start, can leave `end < start`
   persisted.

Nothing else is needed. No new endpoint, no new query param, no per-cell server
computation. **If item 2/3 are deferred, say so and ship bars read-only** — the
render path (Addition 1 display) and zoom (Addition 2) need only the read field.

---

## 1. The job

> Read a multi-year plan as durations, not just dots — see what runs long, what
> overlaps, and which quarter a thing lands in — and shrink a 3-year axis to fit
> one screen.

Target step count: **zero** to read (bars and the auto-picked zoom render on
load); **one** to change zoom (click a segment); **two** to give a node a
duration (open node → set End date).

---

## 2. Screens and entry points

One screen: the Timeline view body, reached from the roadmap page's view toggle
(`Graph | Stages | Timeline`). No new route, no new page. Two additions land
inside the existing `RoadmapTimeline.vue`:

- **A1 Range bars** — replaces nothing; a node with a valid `node_date` +
  `node_end_date` that today renders as a point/card now renders as a bar in a
  new "ranges" band above the existing per-cell point row. Point nodes render
  exactly as today.
- **A2 Zoom control** — a new timeline-local segmented control
  (`Month | Quarter | Year`) added to a thin toolbar strip at the top of the
  timeline body. It changes the axis unit only; it adds no data and persists
  nothing.

The node popover (`RoadmapNodePopover.vue`) gains one input ("End date") next to
the existing "Date" input, so a human can create a bar. This is the only entry
point that writes `node_end_date`.

---

## 3. Layout

The timeline body is a vertical stack of three fixed-height chrome/scroll
regions, all sharing one column grid so columns line up:

```
.rm-timeline-wrap  (flex column, height:100%)
├─ .rm-timeline-bar     ← NEW local toolbar (flex-shrink:0, does NOT scroll)
│     [ Month | Quarter | Year ]        No date (3) ▸
├─ .rm-timeline-scroll  (flex:1, overflow:auto — the one scroll container)
│   ├─ RoadmapTimeAxis   (position:sticky top:0)   ← band row + unit-label row
│   ├─ .rm-ranges-band   ← NEW: greedy-laned bars, grid columns = units
│   └─ .rm-timeline-lane ← EXISTING per-cell point/card/diamond row
└─ .rm-tray  (details, flex-shrink:0)   ← EXISTING undated tray, unchanged
```

The three grid regions (`RoadmapTimeAxis`, `.rm-ranges-band`,
`.rm-timeline-lane`) all use `grid-template-columns: repeat(N, {cellWidth}px)`
with the same `N` and `cellWidth`, so a bar in the ranges band lines up with the
axis cell and the point cell beneath it.

**Column width is a JS constant in px, not a spacing token** — the same
established exception the current file already makes (`const cellWidth = 200`).
Grid track sizing for a horizontally-scrolled axis is a layout dimension, not
padding; keeping it in JS is deliberate and matches today's code. Per unit:

| Unit    | cellWidth | Rationale                                             |
|---------|-----------|-------------------------------------------------------|
| month   | 200px     | unchanged from today; a compact card fits a 200 col   |
| quarter | 200px     | 12 quarters ≈ 2400px for a 3-yr plan                  |
| year    | 240px     | 3 years ≈ 720px, fits most screens without scroll     |

Everything else uses tokens:

- Local toolbar strip: padding `var(--space-2) var(--space-5)`, bottom border
  `1px solid var(--color-border)` — matches `.rm-tray-head`.
- Zoom control: reuse the `.segmented` primitive (`system.css`), buttons at
  `--control-h-sm`. Identical box to `RoadmapViewToggle`.
- Ranges band: `grid-auto-rows: var(--space-8)` (32px lane pitch); each bar
  fills the lane minus `var(--space-1)` top/bottom.
- Bar: padding `var(--space-1) var(--space-2)`, radius `var(--radius-xs)`,
  code chip `0.5625rem` mono (mirrors the card's existing code chip value),
  title `var(--type-3xs)`.
- Point row: unchanged (`.rm-timeline-cell` padding `var(--space-3)`, gap
  `var(--space-2)`).

### Desktop (≥769px), Month zoom, mixed bars + points

```
┌──────────────────────────────────────────────────────────────────────────┐
│ [ Month | Quarter | Year ]                              No date (3) ▸      │ ← toolbar (fixed)
├──────────────────────────────────────────────────────────────────────────┤ ┐
│ Q1 2026              │ Q2 2026              │ Q3 2026            │ …        │ │ sticky axis
│ Jan │ Feb │ Mar  ⬤2  │ Apr │ May │ Jun      │ Jul │ Aug │ Sep    │ …        │ │ (band + units)
├──────────────────────────────────────────────────────────────────────────┤ │
│ RANGES                                                                     │ │
│ ▐ E3 Build parser ───────────▌         ▐ E9 Migrate store ──────▌          │ │ lane 0
│        ▐ E7 Draft spec ▌   ▐ E4 Pilot ─────────────▌                       │ │ lane 1
├──────────────────────────────────────────────────────────────────────────┤ │ scrolls
│ POINTS (existing)                                                          │ │ (x + y)
│ Jan │ Feb        │ Mar      │ Apr │ May    │ Jun   │ …                      │ │
│     │ [card E1]  │ ◆ M1     │     │[card E2]│       │                       │ │
│     │            │ Kickoff  │     │        │        │                       │ │
└──────────────────────────────────────────────────────────────────────────┘ ┘
▸ No date (3)   ← tray, collapsed
```

- `⬤2` in the axis = the existing per-unit node-count pill (kept). It now counts
  point nodes **and** bars that start in that unit, so the count still means
  "things anchored here."
- A bar too narrow to hold its label (a short single-cell bar) prints its label
  clipped to the right, overflowing into the next lane cell, with a `title`
  attribute for the full text.

### Desktop, Year zoom, sparse data (two nodes 3 yrs apart, one long bar)

```
┌──────────────────────────────────────────────────────────┐
│ [ Month | Quarter | *Year* ]              No date (1) ▸    │
├──────────────────────────────────────────────────────────┤
│  2026            │  2027            │  2028               │  ← no band row at year zoom
├──────────────────────────────────────────────────────────┤
│ RANGES                                                    │
│ ▐ E3 Platform build ─────────────────────────────────▌    │  ← 5-yr bar, one lane
├──────────────────────────────────────────────────────────┤
│ POINTS                                                     │
│  2026            │  2027            │  2028               │
│  [card E1]       │                  │  ◆ M9 Launch        │
└──────────────────────────────────────────────────────────┘
```

At year zoom the band row (the coarser-unit caption) disappears — year is the
top unit, there is nothing coarser to caption with.

### ≤768px, Month zoom

The timeline already scrolls horizontally on mobile; that is unchanged. The
toolbar strip stays; the segmented control is small (3 short labels) and does
not wrap. The "No date (n)" affordance drops to icon-density if space is tight
but stays tappable.

```
┌───────────────────────────────┐
│ [Mo|Qtr|Yr]        No date(3)▸ │ ← fixed toolbar
├───────────────────────────────┤
│ Q1 26      │ Q2 26     │ …     │  (sticky, horiz-scroll →)
│ Jan│Feb│Mar│Apr│May│Jun│ …     │
├───────────────────────────────┤
│ ▐E3 Build ─────▌               │  ← bars scroll with axis
│    ▐E7 Spec▌                   │
├───────────────────────────────┤
│ Jan│Feb      │Mar   │ …        │
│    │[card E1]│◆ M1  │          │
└───────────────────────────────┘
```

On ≤768px, bias the **auto-default one unit coarser** (see §5) so the first
paint is legible without immediately scrolling — the smaller viewport is the
whole reason zoom exists.

---

## 4. Component plan

| Element | Disposition | Detail |
|---|---|---|
| Zoom segmented control | **reuse** | `.segmented` CSS primitive from `system.css`, wrapped in a trivial `RoadmapGranularityToggle.vue` mirroring `RoadmapViewToggle.vue` (props `modelValue: TimeUnit`, emit `update:modelValue`; `role="group"`, `aria-pressed`). New file, but it is the same 20-line wrapper the house already uses for the view toggle — not a new pattern. |
| Timeline local toolbar strip | **new (markup only)** | A `div.rm-timeline-bar` added inside `RoadmapTimeline.vue`. Not a component — it is layout owned by the timeline, holding the granularity toggle (left) and the "No date (n)" jump (right). |
| Axis (band + unit labels) | **extend** | `RoadmapTimeAxis.vue`: add prop `unit: TimeUnit`. Rename the "quarters" band to a generic **coarser-unit band** — month→quarter caption, quarter→year caption, year→(band hidden). Cell label comes from the unit (`Jan` / `Q1` / `2026`). Keep the per-cell count pill. |
| Range bar | **new** | `RoadmapTimelineBar.vue` (in `components/roadmap/`). Props: `data: RoadmapCardData`, `deps?: {code}[]`, `startLabel: string`, `endLabel: string`, `highlighted?: boolean`, `dimmed?: boolean`, `invalid?: boolean`. Emits: `click`, `hover: [boolean]`. Renders a compact horizontal bar: code chip + clipped title, status-driven fill/left-accent, dep chips inline when they fit (else a `+n` affordance), an inline `⚠` when `invalid`. **Why new:** `RoadmapNodeCard` is a full vertical card (title row, preview, footer, progress, deps block) that cannot compress to a ~24px spanning element with in-bar label clipping; and no existing component renders a horizontally-spanning grid item. Positioning (`grid-column`, `grid-row`) is applied by the parent, not the component. |
| Point card / milestone diamond | **reuse** | `RoadmapNodeCard` (compact) and the existing `.rm-milestone` markup, unchanged. Only bars are new; points are today's rendering. |
| Bucketing / lane assignment | **extend** | `utils/roadmap.ts`: generalize `buildMonthAxis` → `buildTimeAxis(nodes, unit)` returning `{ units, bars, points, undated }`; add greedy `assignLanes(bars)`; add `TimeUnit = 'month'|'quarter'|'year'` and a `TimelineUnit` type; add `node_end_date` to `RawRoadmapNode`. Keep `buildMonthAxis` as a thin `unit='month'` shim if any caller still needs it, else replace its callers. |
| Timeline shell | **extend** | `RoadmapTimeline.vue`: hold `unit` state (+ auto-default), render the toolbar + ranges band + existing point row, wire hover-highlight (see below). |
| End-date input | **extend** | `RoadmapNodePopover.vue`: add an "End date" `<input type="date">` beside the existing "Date" input; emit `update-end-date: [nodeId, date]`. Read-only summary line gains "End DATE". Page (`[roadmapId].vue`) adds `onUpdateEndDate` calling a new `updateNodeEndDate` in `useRoadmap.ts` (mirrors `updateNodeDate`, optimistic, PUTs `node_end_date`). |

### Hover-highlight — a recorded decision

The Stages board has hover-highlight (`highlightIds`, `relatedIds`, dim the
rest); the **current Timeline does not** — its point cards pass no
`highlighted`/`dimmed`. The brief says "keep hover-highlight working on bars
too," which means it must be *added* to the timeline, not merely preserved.

**Decision: add it to the timeline for both bars and points**, using the
existing `relatedIds(id, edges)` helper and the card's existing
`highlighted`/`dimmed` props — identical mechanics to the board, so the two
board-like views finally behave the same. The bar exposes the same
`highlighted`/`dimmed` styling. Dep chips already work on point cards via
`:deps`; they carry to bars via the new bar's `deps` prop. This is a small,
justified scope addition and is flagged so the builder knows it is net-new
wiring, not a copy of existing timeline behavior.

---

## 5. Auto-default zoom

On first render, pick `unit` from the total dated span (earliest start →
latest end, counting `node_end_date`):

| Span | Default unit |
|---|---|
| ≤ 12 months | month |
| > 12 and ≤ 36 months | quarter |
| > 36 months | year |

On ≤768px, shift one step coarser (month→quarter, quarter→year, year→year).

**Recommend auto-default over always-month.** A 3-year plan opening at month
zoom is 36 columns of horizontal scroll on first paint — the exact problem the
feature exists to solve; opening it already-fitted is the point. The user can
always zoom back in; that costs one click, whereas an unusable first paint costs
confidence. The chosen unit is **local, ephemeral state** (like the view
toggle): not persisted per user, re-derived on load, but a manual override
survives a realtime refresh within the session.

---

## 6. States

Applies per region. "Overloaded" is called out because it is where this design
earns its keep.

### Timeline as a whole
- **Loading:** owned by the page (`roadmap-loading` skeleton) — unchanged.
- **Empty (nothing dated):** `months.length === 0` → the existing `EmptyState`
  ("No dates on this roadmap"), unchanged. **The zoom control is hidden** —
  there is no axis to zoom. The undated tray shows expanded, as today.
- **Error:** owned by the page (`error` + Retry) — unchanged.

### Ranges band (bars)
- **No bars, all points:** the band renders zero lanes → collapses to zero
  height. Layout is today's exactly. No empty-state copy — an absent band is
  not a failure, and a "no ranges yet" strip would be noise on every
  point-only roadmap.
- **All bars, no points:** the point row collapses to zero height; the ranges
  band fills. Both are valid; neither shows placeholder copy.
- **Overloaded — many overlapping bars:** greedy lanes grow downward; the band
  height = `lanes × var(--space-8)`. It scrolls vertically with the rest of the
  scroll container. No lane cap — capping would hide a bar; a tall band that
  scrolls is honest.
- **Overloaded — one 5-year bar:** auto-default picks year; the bar spans all
  year cells in one lane. At month zoom (if the user drills in) the same bar is
  60 columns wide and the axis scrolls — expected, and the reason year is the
  default here.
- **Bad data (`end < start`):** the backend 400s new writes (§0.3), but MCP or
  pre-validation data can carry it. The bar component receives `invalid={true}`;
  the client **renders it as a point at `node_date` (ignores the end)** with an
  inline `⚠` and `title="End date is before start date"`. Never render a
  negative-width or backwards bar. If a client optimistic edit produces
  `end < start`, block the local patch and surface the field error from the
  popover before the PUT.

### Points row / milestone
- Unchanged. **Milestone with a `node_end_date`:** *ignore the end.*
  **Recommend point-not-bar for milestones.** A milestone is semantically an
  instant ("v1.0 ships"); giving it a duration muddies the one node type whose
  job is to mark a moment. Render the diamond at `node_date` as today. If a
  milestone carries an end, surface it only as text in the popover
  ("End DATE"), never as a bar. Alternative (treat as a bar) rejected: it would
  make milestones indistinguishable from range steps and break the diamond
  affordance the graph and timeline share.

### Axis
- **Sparse data at coarse zoom:** e.g. two nodes 3 years apart at year zoom →
  3 cells, two populated, one empty. Empty cells still render (continuous time),
  as `buildMonthAxis` already does for gap months.
- **Band at year zoom:** hidden (no coarser unit). At month/quarter it shows the
  next-coarser caption.

### Copy (the states that need words)
- Empty timeline (unchanged, from today): title **"No dates on this roadmap"**,
  body **"Add a date to nodes to place them on a timeline, or switch to Graph.
  All nodes are listed below."**, action **"Switch to Graph"**.
- Bad-range tooltip: **"End date is before start date"** (on the `⚠`).
- Popover End-date field error (from the 400): **"End date must be on or after
  the start date."**
- Popover End-date help, when the node has no start date:
  **"Set a date first — an end date needs a start."** (disable the End input
  until `node_date` is set.)

---

## 7. Interaction rules

- **Keyboard path.** Toolbar first: the granularity segmented control is three
  plain buttons (Tab to each, Enter/Space to select, `aria-pressed` on the
  active one — same as `RoadmapViewToggle`, deliberately not a roving tablist).
  Then, in DOM order: every **bar** is a `<button>` (lane order top→bottom,
  left→right by start), then every **point** card/diamond in cell order. Enter/
  Space on any opens the node popover — identical to today's card click.
- **Focus visible.** Bars reuse the card's focus ring
  (`outline: 2px solid var(--color-primary); outline-offset: 2px`).
- **Popover open/return.** Opening the node popover is the existing flow
  (`openNode`). On close, focus returns to the bar/card that opened it
  (`ModalOverlay` already restores focus to the trigger — verify the bar is the
  `document.activeElement` at open time).
- **Escape** closes the popover (existing `ModalOverlay` behavior). Escape does
  nothing to zoom.
- **Optimistic vs server.** Zoom change: **pure client state, instant**, no
  request. Setting/clearing a date or end date: **optimistic** (patch local
  node, re-lane immediately), then PUT, revert on failure — exactly the existing
  `updateNodeField` pattern extended to `node_end_date`. The one guard: if the
  optimistic value would make `end < start`, do not patch; show the field error
  and skip the PUT.
- **Realtime repaint.** The page already defers a realtime refresh while the
  popover is open or the user is interacting (`refreshWhenIdle`) — that carries
  over. Zoom is local state and is preserved across a refresh (it is not derived
  from the payload). Lane assignment is deterministic from the data, so a
  refresh that did not change any date re-lanes identically — no visible jump. A
  refresh that *did* change a date (an agent moved a node) re-lanes; if the
  popover is open it waits, as today. A manual zoom override set this session is
  **not** reset by a refresh.

---

## 8. Data contract (per screen)

Single screen (Timeline). Reads from `GET
/api/researches/{id}/roadmaps/{roadmapId}` via `useRoadmap` (already fetched;
no new read).

Fields consumed per node: `id`, `code`, `title`, `description`, `node_type`,
`status`, `node_date`, **`node_end_date` (new — add to `RawRoadmapNode`)**,
`ref_type`, `ref_id`, `ref_data`; edges for deps/highlight (already present).

Writes (only from the popover End-date input): `PUT
/api/roadmap-nodes/{nodeId}` with `{ node_end_date }`.

**Gaps repeated from §0 so they are not missed:**
- Read path emits `node_end_date` automatically (domain serialized directly) —
  confirm, no work.
- **Write path does not yet accept/persist `node_end_date`** (handler
  `Update` struct + `service.UpdateNode` both drop it) — required for the
  End-date input to function.
- **`normalizeNodeRange` is called but undefined**; it must exist and reject
  `end < start` as a 400 field error on `node_end_date`, and `UpdateNode` must
  call it over the resulting pair.

No other backend field is needed. Bucketing into month/quarter/year cells and
greedy lane assignment are computed client-side in `utils/roadmap.ts`.

---

## 9. Out of scope (deliberately not designed)

- **Drag to resize/move a bar on the axis.** Editing stays through the popover,
  matching how the board sets stage/date today. No drag on any board-like view.
- **Dependency lines drawn between bars.** Deps stay as chips (`depends on E3`),
  the established board/timeline convention; drawing edges across an
  independently-scrolling axis is the reason chips exist.
- **A "today" marker / date cursor line.** Reasonable follow-up; not specified
  here.
- **Persisting the chosen zoom** per user or on the roadmap. Ephemeral, like the
  view toggle; a stored default `view` already exists but no stored zoom.
- **Week/day granularity.** Month is the finest unit; the data is
  month-meaningful (plans, not sprints).
- **Collapsing/grouping bars by stage or ref_type** in the ranges band. One flat
  laned band; grouping is a later call.
- **Colouring bars by a progress %** (partial-fill Gantt). Bars are
  status-tinted, not progress-filled; session progress bars stay a card feature.
