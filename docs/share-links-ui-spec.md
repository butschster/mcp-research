# Read-only share links — UI specification

Issue: #50. Written against the catalog at `1baa500`, after teams (#51) shipped.

**Read the "Data-contract gaps" section first** (it is deliberately near the top).
Everything else can be built later; that section is backend work that has to
happen while the backend is being written.

---

## 1. The job

**Owner:** "let a client read this research without giving them an account, and
be able to take it back." Target: **4 interactions** from the research page —
`Share` → (defaults are already right) → `Create link` → `Copy`.

**Visitor:** "read what I was sent, and understand what I am looking at."
Target: **0 interactions** before content — the link opens onto the research,
unless a password was set.

The visitor is the more important of the two. They have no account, no product
knowledge, and no way to ask anyone but the person who sent them the link.

---

## 2. Data-contract gaps

Everything below is something the design needs and the contract in the brief
does not currently return. Ordered by how expensive they are to retrofit.

### 2.1 Leaks — these are correctness, not convenience

| # | Gap | Why it matters |
|---|-----|----------------|
| G1 | **Cross-reference payloads must be redacted, not just un-linked.** `components/entry/CrossReferencesBlock.vue` renders `ref.entry_title`, `ref.research_name` and `ref.research_code` straight from the API. For an outgoing ref to another research, the share crossref payload must return **`target_ref` only** — `entry_title`, `roadmap_title`, `research_name`, `research_code` and `resolved` absent. For **incoming** refs ("Referenced by"), rows whose source research is not the shared one must be **omitted entirely**: their existence alone tells the visitor that another research cites this one, and their title tells them what it is called. The UI cannot fix this — it can only render what it is handed. |
| G2 | **`/entries/{id}/related` must be restricted to the shared research.** The owner route returns tag-related entries across researches (`components/entry/RelatedEntriesBlock.vue` has an explicit `rel.research_id !== researchId` branch). Under a share it must return same-research rows only. |
| G3 | **The redacted `research` object must drop `role`.** `domain.Research` carries `Role`, `TeamID`, `TeamName`, `TeamPersonal`, `UserID`, `Instruction`, `Memory`. All must be absent. `Role` specifically: `useResearchRole()` computes `canWrite` from it, and with `auth_enabled: false` it returns `true` regardless — see §7.3. |
| G4 | **`/export` must honour the include flags.** The markdown export contains session Q&A and tasks. A share with `sessions: false` must not receive them in the export, and `instruction` / `memory` must be absent there too (issue acceptance criterion, and the export is the easiest place to forget). |
| G5 | **Events on a share socket must be scrubbed of `actor_user_id`.** It is an internal user id handed to an outsider, on every keystroke-sized event. `actor_client_id` is harmless (it is this visitor's own tab or somebody else's opaque uuid); `name` and `reason` are only sent where the recipient can no longer fetch the thing, which for a share means they should not be sent at all. |

### 2.2 Routes the design needs and the contract does not list

All under `/api/shared/{token}/researches/{id}`:

| # | Route | Screen that dies without it |
|---|-------|------------------------------|
| G6 | `GET /sections/{sectionId}/entries` | **The overview's default view.** The owner page opens on the first section, not on "all entries". Without this the sidebar is decoration. |
| G7 | `GET /tags` | The tag filter panel on the "All entries" view. |
| G8 | `GET /links` | The sidebar's "External links" item — a third of the sidebar. |
| G9 | `GET /roadmaps/{roadmapId}` | The roadmap detail page. `/roadmaps` lists them; the graph itself is a second call. |
| G10 | `GET /sessions/{sessionId}` | The session page (questions and answers). Only reachable when `include.sessions`. |
| G11 | `GET /tasks` | Nothing serves `include.tasks` otherwise — the flag exists in the create body with no route behind it. |
| G12 | `GET /entries/{entryId}/crossrefs`, `/links`, `/related` | The owner entry page reads these off **unscoped** `/api/entries/{id}/…` routes, which a share token must never reach. They have to be re-mounted under the share prefix. See G1 and G2 for what they may contain. |

Also: **the share entry route must resolve short codes**, not only UUIDs.
Every link in this product is built from a code (`/entry/E3`), and the owner
route `GET /api/researches/{id}/entries/{entryId}` already accepts either.

### 2.3 Signals the UI cannot infer

| # | Gap | What the UI does with it |
|---|-----|--------------------------|
| G13 | **Password-required must be distinguishable from dead.** `GET /api/shared/{token}` must answer `401 {"error":"password_required"}` for a valid, locked token with no unlock cookie. This *is* a deliberate exception to "revoked, expired and nonexistent all return the same 404" — a UI that cannot tell them apart cannot ask for a password. Make the exception knowingly; do not let it widen. |
| G14 | **Wrong password vs rate-limited.** `POST /unlock` → `401 {"error":"invalid_password"}` and `429` with `Retry-After` when throttled. Without the distinction the UI says "wrong password" to someone who has been locked out, and they keep trying. |
| G15 | **`share.expires_at` in the visitor payload.** So the banner can say when the link stops working, and the visitor can act before it does. Nullable = never. |
| G16 | **`owner_name` must be the share's creator**, the same person `created_by_name` names in the owner list. The banner says "shared by X"; if it is the research's creator and someone else made the link, the banner lies. Specify the fallback for an empty name: the UI drops the byline entirely (§5.2). |
| G17 | **`share.research_code` is load-bearing, not decoration.** It is the only way the client can decide whether `[[R7:E5]]` is an internal reference or a foreign one. It is in the contract — this is a note not to drop it during redaction, since it is the *only* research identifier that survives. |
| G18 | **`active_share_count` on `GET /api/researches/{id}`.** Drives the count badge on the `Share` button, which is how an owner sees at a glance that a research is exposed without opening anything. The UI degrades to no badge if absent — but "no badge" and "zero shares" then look identical, which is the wrong failure for a security signal. |
| G19 | **Who may share — `owner` or `owner`-and-`editor` — has to be decided and reflected.** The UI gates on `canWrite` (editor+) by default; if the answer is owner-only it gates on `canAdmin`. Either is a one-line change, but the button must never be shown to somebody the API will refuse. If the rule ever becomes more subtle than the role, send `research.can_share`. |
| G20 | **`/ws?share={token}` close semantics.** On revoke or expiry the socket must close with `4401` (the existing "credential stopped being valid" code), and must not use `4401` for ordinary churn. The UI's response is in §7.5: on `4401` it probes `GET /api/shared/{token}` once and shows the dead screen on 404, the offline state otherwise. That works with what exists today — this is a note that the close code carries meaning now. |

---

## 3. Screens and entry points

### Visitor

| Route | File | Reached from | Condition |
|-------|------|--------------|-----------|
| `/s/{token}` | `pages/s/[token]/index.vue` | the link itself | always |
| `/s/{token}?section={id}` | same | sidebar | always |
| `/s/{token}/entry/{code}` | `pages/s/[token]/entry/[entryId].vue` | entry cards, cross-refs, prev/next | always |
| `/s/{token}/roadmaps` | `pages/s/[token]/roadmaps.vue` | header icon | `include.roadmaps` |
| `/s/{token}/roadmap/{code}` | `pages/s/[token]/roadmap/[roadmapId].vue` | roadmap list, `[[RM1]]` refs | `include.roadmaps` |
| `/s/{token}/session/{code}` | `pages/s/[token]/session/[sessionId].vue` | session cards on the overview | `include.sessions` |
| `/s/{token}/tasks` | `pages/s/[token]/tasks.vue` | header icon | `include.tasks` |
| `/s/{token}/export` | `pages/s/[token]/export.vue` | header menu | `include.export` |

`pages/s/[token].vue` is the **parent route** (Nuxt renders it with a
`<NuxtPage/>` inside). It owns: the one fetch of `GET /api/shared/{token}`, the
banner, the password gate, the dead-link screen, the realtime subscription and
the read-only lock. Children render content and nothing else. A child never
decides whether the visitor is allowed in.

**Not built (and why), see §10:** `/s/{token}/mindmap`, `/s/{token}/graph`,
`/s/{token}/session/{code}/question/{code}`, entry revision history.

Sub-routes that are switched off by an `include` flag are **not rendered as
disabled** — the entry point is absent and the route, if guessed, shows the same
"not part of this link" empty state (§6.7). The API 404s them anyway.

### Owner

| Surface | Where | Replaces |
|---------|-------|----------|
| `Share` button + count | `pages/research/[id]/index.vue` header, immediately left of the icon-button group | nothing |
| Share dialog (list / create / reveal) | modal from that button | nothing |
| Revoke confirmation | `ConfirmModal` over the dialog | nothing |

**Manage-shares does not live in `DetailsPanel.vue`.** That panel is the
research's content — goal, description, instruction, memory — and it is open to
viewers. Shares are an access-control surface and belong with the action that
creates them, in one place, so nobody has to know there are two.

### Entry-point decision (recorded disagreement)

The visual designer's position: the header already carries a status badge, a
team chip, four icon links and a kebab. A fifth control is one too many; put
`Share` in the menu next to Export.

The interaction designer's position: this is a p0 feature whose entire value is
that people find it. Buried in a kebab, beside `Download JSON`, it will be found
by the people who read the changelog.

**Resolved in favour of the header, with a condition.** `Share` goes in as a
*labelled* `btn btn-sm` — a different kind of object from the four icon links,
so it reads as one item, not a fifth icon. It carries a `.btn-count` badge of
active shares, which earns its place independently: **an owner should be able to
see that a research is exposed without clicking anything.** Export, Members,
Move to team and Archive all stay in the kebab. The header does not grow again.

---

## 4. Layout

Tokens only. Every number below is a token from `frontend/assets/css/main.css`;
the design invents none. Note the product has **one theme** (dark) — there is no
light-mode contrast pass to do.

### 4.1 Visitor — desktop (≥769px)

```
┌───────────────────────────────────────────────────────────────────────────┐
│ ◉ Read-only shared view — shared by Elena Marsh · Entries, roadmaps    ▾ │  sticky, --z-elevated
└───────────────────────────────────────────────────────────────────────────┘  h ≈ 36px, --type-xs
     │  .container (max-w 1200px, padding 0 var(--space-6))
     ▼
   Research / Pricing benchmark, Q3                        ← Breadcrumbs, non-linking root
   ┌──────────────────────────────────────────────────────────────────┐
   │  R7  Pricing benchmark, Q3                [active] [Roadmaps 3]  │  .page-header
   │  How three competitors price seat-based tiers.                   │  .page-title / .card-meta
   └──────────────────────────────────────────────────────────────────┘
                                                            gap var(--space-5)
   ┌───────────┬──────────────────────────────────────────────────────┐
   │ Sidebar   │  Section title                          [completed]  │  .layout-sidebar
   │ (reused,  │  ────────────────────────────────────────────────    │
   │  sticky)  │  ┌────────────────────────────────────────────────┐  │
   │           │  │ E3  Seat pricing at Kestrel        [answered]  │  │  EntryCard
   │ All 24    │  │ Their tiers cluster around 12 seats…           │  │
   │ Method  6 │  └────────────────────────────────────────────────┘  │
   │ Findings 9│  ┌────────────────────────────────────────────────┐  │
   │ ───────   │  │ E4  …                                         │  │
   │ Links   3 │  └────────────────────────────────────────────────┘  │
   └───────────┴──────────────────────────────────────────────────────┘

   ─────────────────────────────────────────────────────────────────────
   Research                                            ● Live            .app-footer, ConnectionStatus
```

Header actions on the visitor overview are **only** the view-switching icon
links that the include flags allow — Roadmaps, Tasks — plus a kebab holding
`Export` when `include.export`. No status dropdown, no Details, no Members, no
Move, no Archive. The `StatusBadge` stays: a reader benefits from knowing the
research is still active.

Spacing: identical to the owner page, because it is the same page. The banner
adds `var(--space-4)` of top padding to `.main-content` equivalent, matching
`app.vue`.

### 4.2 Visitor — ≤768px

```
┌─────────────────────────────────────────┐
│ ◉ Read-only · Elena Marsh            ▾ │  sticky, wraps to 1 line, --type-xs
└─────────────────────────────────────────┘
  ┌───────────────────────────────────────┐
  │ R7                                    │  .title-with-code wraps (existing rule)
  │ Pricing benchmark, Q3                 │  --type-xl at this width
  │ [active]  [Roadmaps]  [⋯]             │  .research-actions flex-wrap
  └───────────────────────────────────────┘
  ┌───────────────────────────────────────┐
  │ All 24 │ Method 6 │ Findings 9 │ Lin… │  .sidebar becomes a horizontal scroller
  └───────────────────────────────────────┘     (existing @media rule, free)
  ┌───────────────────────────────────────┐
  │ E3  Seat pricing at Kestrel           │
  │ Their tiers cluster around 12 seats…  │
  └───────────────────────────────────────┘
```

Expanded banner disclosure at this width is a full-width panel pushing content
down (not an overlay): it is read once, and an overlay over content the visitor
has not seen yet is the wrong trade.

### 4.3 Share dialog — desktop

`ModalOverlay size="lg"` (max-width 720px, max-height 85vh, scrolls).

```
  ┌── Share "Pricing benchmark, Q3" ───────────────────────── ✕ ─┐
  │  Anyone with the link can read this research.               │  .modal-help
  │  They don't need an account.                                │
  │                                              [ + New link ] │  .btn-sm .btn-primary
  │  ─────────────────────────────────────────────────────────  │
  │  Client review, March                    Active · 47 views  │  ShareRowList row
  │  3 Mar by Elena Marsh · Entries, roadmaps  [Password]       │  --type-xs muted
  │  Last opened 2 hours ago         Show link · Revoke         │
  │  ─────────────────────────────────────────────────────────  │
  │  Board pack                             Expires in 6 days   │
  │  1 Mar by Elena Marsh · Entries         Never opened        │
  │                                                    Revoke   │
  │  ─────────────────────────────────────────────────────────  │
  │  Q1 handoff                                       Revoked   │  dimmed, no actions
  └─────────────────────────────────────────────────────────────┘
```

Row grid: `minmax(0, 1.6fr) minmax(0, 1fr) auto`, `gap: var(--space-3)`,
`padding: var(--space-3) var(--space-1)`, `border-bottom: 1px solid
var(--color-border)` — the same construction as
`components/team/TeamMemberList.vue`, so the two lists look related, because
they are.

At ≤768px the row collapses to two columns, actions on their own row, exactly
the pattern `TeamMemberList` already uses.

### 4.4 Share dialog — create and reveal

```
  ┌── New share link ──────────────────────────────────── ✕ ─┐
  │  Label                                                   │  .field-label
  │  [ Client review, March                              ]   │  .text-input
  │  Only you see this. It's how you'll recognise the link.   │  --type-xs muted
  │                                                          │
  │  What the link shows                                     │  fieldset legend
  │  Sections, entries and cross-references · always         │  static line, not a checkbox
  │  ☐ Roadmaps                                              │
  │  ☐ Interview sessions, with questions and answers        │
  │  ☐ Tasks                                                 │
  │  ☑ Downloading and printing                              │
  │                                                          │
  │  Stops working                                           │
  │  [ 30 days                                          ▾ ]  │
  │                                                          │
  │  ☐ Require a password                                    │
  │                                                          │
  │                                [ Cancel ] [ Create link ]│  .modal-actions
  └──────────────────────────────────────────────────────────┘

  ┌── Link ready ──────────────────────────────────────── ✕ ─┐
  │ ⚠ This is the only time this link is shown. Copy it now  │  .warning-banner
  │   — it cannot be shown again after you reload.           │
  │                                                          │
  │  ┌────────────────────────────────────────────────────┐  │
  │  │ https://…/s/9f2c…c41a                    [ Copy ]  │  │  .empty-command + .copy-btn
  │  └────────────────────────────────────────────────────┘  │
  │  Send the password separately — not in the same message. │  only when a password was set
  │                                                          │
  │                                              [ Done ]    │
  └──────────────────────────────────────────────────────────┘
```

Amber is correct **here** and nowhere else in this feature: this is the one
moment where something is about to be lost.

---

## 5. Component plan

Reuse is the default. `extend` means an existing file gains a prop, a slot or a
path helper; the owner-side behaviour must not change.

### 5.1 Visitor

| Element | Verdict | Detail |
|---|---|---|
| Share shell / route parent | **new** — `pages/s/[token].vue` | Fetches the share payload once, renders `ShareBanner` + `<NuxtPage/>` + footer, owns password gate and dead-link screens, sets the read-only lock, opens the socket. No existing page is a token-addressed parent route; `pages/invite/[token].vue` is the closest precedent and is a leaf. |
| Banner | **new** — `components/share/ShareBanner.vue` | Props: `ownerName?: string`, `include: ShareInclude`, `expiresAt?: string \| null`, `live: boolean`. Emits: none. Nothing existing fits: `WarningBanner.vue` is a fixed in-memory-mode message with no props, `team/ViewerNotice.vue` is a five-word badge. |
| Password gate | **new** — `components/share/SharePasswordGate.vue` | Props: `busy: boolean`, `error: '' \| 'wrong' \| 'throttled'`, `retryAfter?: number`. Emits: `submit: [password: string]`. No existing single-field auth card exists outside `pages/login.vue`, which is bound to `useAuth`. |
| Dead link / expired / bad token | **reuse** — `components/EmptyState.vue` | `icon`, `title`, `description`, default slot left empty. |
| Transport failure | **reuse** — `EmptyState` + a `Try again` button in its default slot | Mirrors `pages/invite/[token].vue`'s `loadFailed` branch exactly. |
| Live indicator | **reuse** — `components/ActivityIndicator.vue` | `:active` `label="Updating"`, placed in the banner's right end. |
| Connection state | **reuse** — `components/ConnectionStatus.vue` | Fed from `useConnectionBanner()` as `app.vue` does. |
| Breadcrumbs | **reuse** — `components/Breadcrumbs.vue` | Root crumb is the research name and is **not** a link to `/`. |
| Section sidebar | **reuse as-is** — `components/research/Sidebar.vue` | Emits only; no links inside it. Zero change. |
| Entry list | **extend** — `components/research/EntriesView.vue` | Replace the two inline `` `/research/${researchSlug}/entry/…` `` literals with `entryPath()` (§5.3). Recommended but not required: this component duplicates `EntryCard.vue` verbatim twice — collapsing it to `<EntryCard>` removes the duplication and the path fix lands in one place instead of three. Do not make the share feature depend on that refactor. |
| Entry card | **extend** — `components/EntryCard.vue` | `entryPath()`. |
| Entry body | **reuse as-is** — `components/blocks/BlockRenderer.vue` | Already gates its checkboxes on `canWrite`; the read-only lock (§7.3) turns them off. |
| Cross-references | **extend** — `components/entry/CrossReferencesBlock.vue` | Its `refLink()` becomes path-aware, and a row whose target is outside the shared research renders as **static text, not a `NuxtLink`**. Depends on G1 for the data. |
| Related entries | **extend** — `components/entry/RelatedEntriesBlock.vue` | `entryPath()`; its cross-research branch is unreachable under a share once G2 lands, and must render inert if it ever is reached. |
| External links | **reuse as-is** — `components/entry/ExternalLinksBlock.vue`, `components/research/ExternalLinksView.vue` | Outbound `http(s)` links only. |
| Prev/next | **extend** — `components/entry/EntryNavigation.vue` | `entryPath()`. |
| Sessions on the overview | **extend** — `components/research/ActiveSessionsGrid.vue`, `PastSessionsList.vue` | `sessionPath()`. Rendered only when `include.sessions`. |
| Questions | **extend** — `components/QuestionList.vue` | `questionPath()` — and under a share the question text is **not** a link (no question detail route in v1); the component gets a `:linkQuestions="false"` prop rather than a path that goes nowhere. |
| Roadmap | **reuse as-is** — `components/roadmap/*` | `RoadmapNodePopover.vue` already gates status chips on `canWrite`; the lock handles it. Node dragging is off because the page passes `:nodes-draggable="false"`. |
| Tasks | **reuse as-is** — `components/tasks/KanbanBoard.vue`, `KanbanCard.vue`, `TaskDetailModal.vue` | All three gate on `canWrite`. Only mounted when `include.tasks`. |
| Badges, codes, tags, progress | **reuse as-is** — `StatusBadge`, `ShortCode`, `TagList`, `ProgressBar` | |
| Export document | **reuse** — the markup of `pages/research/[id]/export.vue`'s `<article class="export-document">` | The share export page is a new thin page reading the share export endpoint; the print CSS in `main.css` already hides `.btn` and `.no-print`, and must also hide `.share-banner` — one selector added to the existing `@media print` block. |

### 5.2 Owner

| Element | Verdict | Detail |
|---|---|---|
| `Share` header button | **extend** — `pages/research/[id]/index.vue` | `v-if="canWrite"` (see G19), `.btn .btn-sm`, `.btn-count` badge from `active_share_count`. |
| Share dialog | **new** — `components/research/ShareDialog.vue` | Props: `visible`, `researchName`, `shares: Share[]`, `loading`, `creating`, `error`, `issuedUrl`, `busyId`, `recoverableLinks: Record<string,string>`. Emits: `create: [payload]`, `revoke: [share]`, `showLink: [share]`, `close`. Three internal views (`list` / `create` / `reveal`). Nothing existing fits: `team/InviteDialog.vue` is the closest — same compose→reveal shape — but its fields, its lifecycle and its list semantics are all different, and generalising it would leave one component serving two features badly. |
| Share rows | **new** — `components/research/ShareRowList.vue` | Props: `shares`, `busyId?`, `recoverableLinks`. Emits: `revoke`, `showLink`. Deliberately modelled on `team/TeamInviteList.vue` and `team/TeamMemberList.vue` — same grid, same dimmed-dead-row rule, same `link-btn` actions — but its columns (views, last opened, expiry, password) have no overlap with either. |
| Copy-once block | **new** — `components/CopyableSecret.vue` | Props: `value: string`, `hint?: string`. Emits: `copied`. This markup — `.empty-command` + `.copy-btn` + the clipboard-absent fallback to a selected readonly input — exists in `team/InviteDialog.vue` and a variant in `EmptyState.vue`, and this would be the third copy. The fallback matters (plain-HTTP LAN deployments have no `navigator.clipboard`) and is exactly the kind of thing that drifts between copies. Refactoring `InviteDialog` onto it is recommended and **not** a blocker. |
| Revoke confirmation | **reuse** — `components/ConfirmModal.vue` | `variant="danger"`, `:loading="!!busyId"`. |
| Dialog shell | **reuse** — `components/ModalOverlay.vue` | Owns Escape, focus trap, focus restore. Do not reimplement. |
| Toasts | **reuse** — `useToasts()` | |

### 5.3 Composables and plumbing

| Element | Verdict | Detail |
|---|---|---|
| Share state | **new** — `composables/useShare.ts` | Module state, same argument as `useResearchRole`: `{ token, include, ownerName, researchCode, researchId, expiresAt, active }`, plus `shareFetch()` which prefixes `/api/shared/{token}` and sends **no** `Authorization` header. |
| Path building | **new** — `composables/useResearchPaths.ts` | `researchPath(slug)`, `entryPath(slug, code)`, `sessionPath`, `questionPath`, `roadmapPath`, `roadmapsPath`, `tasksPath`, `exportPath`. Each returns `/s/{token}/…` when a share is active and `/research/{slug}/…` otherwise. **This is the load-bearing piece of the whole feature**: 20 files across `components/` and `pages/` build `/research/…` by hand (`grep -rn "/research/" frontend/components frontend/pages`). Only the ones the share view renders need converting; the rest can follow later, and the helper makes that safe. |
| Cross-ref rendering | **extend** — `composables/useCrossRefs.ts` | See §8. |
| Read-only lock | **extend** — `composables/useResearchRole.ts` | See §7.3. |
| WebSocket | **extend** — `composables/useRealtimeUpdates.ts` | A module-level share token; `socketUrl()` emits `?share={token}` instead of `?token=`; `start()`'s watcher must not take the "nobody is signed in → disconnect" branch when a share token is set. |
| Route guard | **extend** — `middleware/auth.global.ts` | `if (to.path.startsWith('/s/')) return`, beside the existing `/invite/` clause and for the same reason. |
| Chrome | **extend** — `app.vue` | `isChromeless` gains `route.path.startsWith('/s/')`. |
| Keyboard nav | **extend** — `composables/useKeyboardNav.ts` | It binds `Shift+G` → `router.push('/')`. On a share page that is a one-keystroke trip to the login wall. Guard it: return early when a share is active. (It has the same latent bug on `/login` and `/invite/` today.) |
| Revocation watcher | **extend** — `composables/useAccessRevoked.ts` | Return early when a share is active: it fires unauthenticated `authFetch` calls and `useTeams().load()` on every `access.*` event, none of which mean anything to a visitor. |

---

## 6. States

### 6.1 Visitor shell — loading

Skeleton, never a spinner, and never a partially-rendered banner: the banner
names a person, and a banner that says "shared by" with a blank is worse than no
banner. `pages/s/[token].vue` renders `.skeleton-card` blocks in the shape of
the page (header bar, sidebar, three entry cards) — the exact markup
`pages/research/[id]/index.vue` already uses for `pending`.

### 6.2 Visitor shell — dead link (invalid / revoked / expired)

**One screen for all three.** The backend returns an identical 404 and the UI
must not speculate about which.

```
icon:        &#x1F517;
title:       This link isn't available
description: It may have been turned off, it may have expired, or the address
             may be incomplete — these links are long and easy to cut short.
             Ask the person who sent it for a new one.
```

No button. The next action is "ask the sender", which the copy names, and there
is nowhere in this product a signed-out visitor can usefully be sent. A `Sign
in` button here would be exactly the wall this feature exists to avoid.

### 6.3 Visitor shell — the server did not answer

Distinct screen, because "check how you copied the link" is the wrong
instruction for a server that is down. Straight from `pages/invite/[token].vue`:

```
title:       Couldn't open this link
description: The server didn't answer. The link is probably fine — try again in
             a moment.
button:      Try again          → refetches
```

### 6.4 Visitor shell — password required

```
title:       This link is password-protected
description: Ask the person who sent you the link for the password.
label:       Password
button:      Open
```

Wrong password (`role="alert"`, `.inline-error`, below the field):

```
That password doesn't match. Check it with whoever sent you the link.
```

Throttled:

```
Too many tries. Wait a minute, then try again.
```

The field is not cleared on a wrong answer — retyping a long password from a
chat message is the failure mode, not a security risk here.

### 6.5 Overview — empty research

```
icon:        &#x1F4C4;
title:       Nothing here yet
description: {ownerName} hasn't added anything to this research yet. This page
             updates by itself when they do — you can leave it open.
```

The next action for a read-only visitor is genuinely "leave it open", and it is
true: the socket is live. With no `ownerName`: `Nothing has been added to this
research yet.`

### 6.6 Overview — empty section

Reuses the existing `EntriesView` empty state, with the copy changed under a
share: the owner's version says "Claude will populate this section with research
entries", which means nothing to a client.

```
title:       No entries in this section
description: There's nothing here yet.
```

### 6.7 A sub-route that is not part of this link

Reached only by typing a URL, since the entry point is absent.

```
title:       Not part of this link
description: The person who shared this research didn't include this. Ask them
             if you need it.
button:      Back to the research   → /s/{token}
```

### 6.8 Entry removed while being read

The entry fetch 404s after a `entry.deleted` event.

```
title:       This entry was removed
description: It was deleted while you were reading. The rest of the research is
             still here.
button:      Back to the research
```

### 6.9 Overloaded

- **40+ sections:** the sidebar scrolls (desktop, `.sidebar` is sticky with its
  own overflow) and becomes a horizontal scroller at ≤768px. Existing behaviour,
  no change.
- **500 entries in "All entries":** renders as-is, as on the owner page. The
  `.grid > *:nth-child(n)` stagger caps at the 9th child already. Not paginating
  is a conscious carry-over: the owner page does not, and a share that behaves
  differently from the real page is a worse bug than a long page.
- **A 300-node roadmap:** Vue Flow, unchanged.
- **A research name of 200 characters with no spaces:** `.page-title` has
  `overflow-wrap: anywhere`; the banner's own strings must set it too.
- **A share label of 200 characters:** the row's label cell is `minmax(0, 1.6fr)`
  with `overflow-wrap: anywhere` — it wraps, it does not truncate. A truncated
  label is not a label, and this list exists to be recognised by it.
- **60 shares on one research:** the dialog body scrolls (`modal-lg` is
  `max-height: 85vh; overflow-y: auto`). No pagination; revoked rows stay,
  dimmed, as lapsed invites do.

### 6.10 Share dialog — empty

There is no empty state, by construction: **with no shares, the dialog opens
straight onto the create form.** The lead sentence does the explaining. This is
the only empty state in the feature that was worth designing away rather than
writing copy for.

After revoking the last one, the list stays and shows:

```
All links revoked. Nobody outside the team can open this research now.
                                                        [ + New link ]
```

### 6.11 Share dialog — loading and error

- Loading the list: three `.skeleton-card` rows at row height. The `+ New link`
  button is live immediately — creating does not depend on the list.
- Create failed: `.inline-error` with `role="alert"` above the actions, inline
  and not a toast, because the form is still on screen and still filled in
  (`InviteDialog` precedent). Copy: the server's message, falling back to
  `Couldn't create the link. Try again.`
- Revoke failed: the row un-busies and a toast fires — `Couldn't revoke that
  link. It is still active.` The row must **not** show "Revoked" until the
  server says so (§7.6).

---

## 7. Interaction rules

### 7.1 Keyboard — visitor

Tab order: banner disclosure → header view links → sidebar items → content.
Nothing traps. No skip-link is added, because no page in this product has one
and inventing one here would be the only one.

The sidebar's `.sidebar-item` elements are `<div>`s with `@click` on the owner
page — they are not focusable and not announced. **That is a pre-existing
defect and this spec does not inherit it silently:** if it is not fixed, a
keyboard-only visitor cannot change section, and on a share view there is no
alternative route to the other sections. Minimum fix: `role="tab"`/`tabindex="0"`
plus Enter/Space, in `components/research/Sidebar.vue`, which fixes the owner
page at the same time. Flagged rather than designed around.

### 7.2 Keyboard — share dialog

`ModalOverlay` already gives Escape-to-close, a Tab cycle inside the dialog and
focus restore to the trigger. Do not reimplement any of it.

- Open with no shares → focus the `Label` input.
- Open with shares → focus `+ New link`.
- `list` → `create` → focus `Label`.
- `create` → `reveal` → focus **`Copy`**, and announce in a live region that is
  already in the DOM before the text lands (the `InviteDialog` note about a live
  region that mounts with its text already in it is a real bug, not a style
  point): `Link ready. Copy it now — it is shown once.`
- Escape from `reveal` closes the dialog (see 7.4). Escape from `create` returns
  to `list` if a list exists, otherwise closes.
- The revoke `ConfirmModal` stacks over the dialog; both are `--z-overlay` and
  the later one in the DOM paints above. Focus restores back into the row's
  `Revoke` button. This already happens on `pages/teams/[id].vue`.

### 7.3 The read-only lock

This is the part that will be got wrong if it is not spelled out.

`useResearchRole().canWrite` returns **`true` when `authEnabled` is false**. On a
server running without auth, a `/s/{token}` page would therefore render every
edit control — Edit, Delete, the status dropdown, the roadmap status chips, the
block checkboxes — for an anonymous visitor.

Two independent defences, and both are required:

1. **The share pages never import the edit machinery.** They are new thin pages
   composing read-only components; `MdEditor`, `HistoryPanel`, the delete
   confirm and the status dropdown are not in the tree. A page that does not
   import an editor cannot render one, whatever a flag says.
2. **`useResearchRole` gains a share lock.** Add a module-level
   `shareLocked` ref; `canWrite` and `canAdmin` become
   `!shareLocked.value && (…)`; `isViewer` stays false (the team read-only badge
   is the wrong explanation here — the banner is the right one).
   `pages/s/[token].vue` sets it in setup; `setFromResearch()` clears it, so an
   owner navigating from a share view back into the app publishes their real
   role and the lock lifts. Do **not** clear it on unmount — the composable's own
   comment explains why unmount-clearing races the incoming page.

Defence 2 is what protects the components defence 1 does reuse:
`BlockRenderer`'s checkboxes, `RoadmapNodePopover`'s status chips,
`KanbanBoard`'s drag.

### 7.4 The link is shown once

Follows the invite-link precedent in `pages/teams/[id].vue` exactly, because it
is the same problem and a second answer to it would be worse than either.

- The URL is held in a `recoverableLinks: Record<shareId, url>` ref for the life
  of the tab. Never persisted.
- A row whose link is still in memory offers `Show link`. After a reload, it does
  not, because it cannot — the server has only the hash.
- **Dismissing without copying is allowed.** Escape and the close button always
  work; a dialog that refuses to close is a trap, and the a11y reviewer's
  objection wins over the interaction designer's proposed blocking confirm.
  Instead: the `Done` button is labelled **`Copy and finish`** until a copy has
  happened, and closing uncopied fires a warning toast —

  ```
  Link not copied. You can still show it from the Share dialog until you
  reload this page.
  ```

  — and the row keeps its `Show link` action. Nothing is lost that can be
  recovered, and the one thing that cannot be recovered is said plainly in the
  amber strip before it happens.
- Clipboard absent (plain HTTP on a LAN): fall back to a readonly input,
  focused and selected, with `Copying isn't available on this connection —
  select the link above and copy it manually.` `CopyableSecret` owns this.

### 7.5 Realtime

The visitor's socket is `/ws?share={token}`. The page subscribes with
`useResearchRealtime()` exactly as the owner page does — same 80ms burst
coalescing, same per-entity refetch map, same `onResync` superset refetch after
a dropped connection.

Rules when an update repaints under the visitor's hands:

- **Never move focus and never scroll.** A repaint is data changing, not
  navigation.
- The `ActivityIndicator` in the banner blips for 5 seconds (`isSelf()` is
  irrelevant — a visitor never writes).
- The overview's section list and counts refetch together, for the same reason
  the owner page does it: the counts live on the research payload and the list
  does not, so refetching one leaves the other lying.
- **On close code `4401`:** probe `GET /api/shared/{token}` once. 404 → replace
  the entire page with the dead-link screen (§6.2). Anything else → the ordinary
  offline state. See G20.
- A revocation that arrives while the visitor is reading **replaces the page
  immediately.** Recorded disagreement: the interaction designer argued for
  letting them finish the paragraph behind a "this link has been turned off"
  banner. Rejected — revocation is not advisory, and the owner pressing Revoke
  has to mean the content is off the screen. There is no unsaved work to
  protect, which is the one thing that makes `useAccessRevoked` hesitate on the
  owner side.

### 7.6 Optimistic vs. server

Nothing on the visitor side is optimistic: there are no writes.

On the owner side, **nothing is optimistic either**, and both cases are
deliberate:

- **Create** cannot be — the URL comes from the server.
- **Revoke** must not be. A row that flips to "Revoked" before the server agrees
  tells an owner that access is closed when it may not be. The row goes busy
  (`opacity: 0.6`, actions disabled — the `TeamInviteList` busy treatment), then
  flips on the response, or un-busies and toasts on failure.

---

## 8. Cross-references under a share

`renderRefs(text, researchSlug)` in `composables/useCrossRefs.ts` currently turns
every `[[…]]` into an `<a class="crossref-link">`. It is called from 14 places
and funnels through `useInlineMarkdown.ts` for most content.

Extend it to take the share context from module state (safe: it is a plain
function reading a module ref, the same shape `useResearchRole` uses, and it
defaults to today's behaviour with no share active — Storybook's stub is
unaffected).

**The rule, stated once:** a reference renders as a link only when its target is
inside the shared research **and** the `include` flag for that target's type is
on. Otherwise it renders as **bare text with no markup at all**.

| Ref | Share of R7, roadmaps on | Share of R7, roadmaps off |
|---|---|---|
| `[[E3]]` | `<a href="/s/{t}/entry/E3" class="crossref-link">E3</a>` | same |
| `[[R7]]` | `<a href="/s/{t}">R7</a>` | same |
| `[[R7:E5]]` | `<a href="/s/{t}/entry/E5">R7:E5</a>` | same |
| `[[RM1]]`, `[[RM1:N3]]` | `<a href="/s/{t}/roadmap/RM1">RM1</a>` | `RM1` — plain text |
| `[[R2]]`, `[[R2:E5]]` | `R2:E5` — plain text | same |

"Bare text with no markup" means exactly that: **no `<span>`, no class, no
`title`, no cursor change, no colour.** Not `.crossref-link` with
`pointer-events: none` — a distinct visual treatment is itself a statement that
something is there. The brackets are stripped, so `as covered in [[R2:E5]]`
reads as `as covered in R2:E5`, which is how a document code reads in prose. The
string `R2:E5` was already in the author's own text; the UI adds nothing to it.

`components/entry/CrossReferencesBlock.vue` needs the same rule applied to its
`refLink()`, and must render a non-internal row as a plain element rather than a
`NuxtLink` — but its real fix is server-side (G1): it displays
`ref.entry_title` and `ref.research_name`, and no amount of link suppression
helps if the title of another research's entry is in the payload.

---

## 9. Data contract, per screen

`{t}` = token, `{r}` = `share.research_id`. Base for every visitor call:
`/api/shared/{t}/researches/{r}`. No `Authorization` header, ever — `useApi()`
attaches one whenever a token exists in `useAuth`, so the share pages must use
`shareFetch()`, not `useApi()`. An owner who is signed in and opens their own
share link must be treated as a visitor.

| Screen | Reads | Fields needed | Missing |
|---|---|---|---|
| Shell | `GET /api/shared/{t}` | `share.label`, `share.include`, `share.research_id`, `share.research_code`, `share.owner_name`, `research.{code,name,goal,description,tags,status}`, `sections[].{id,code,name,display_name,description,status,entries_count}` | `share.expires_at` (G15); `research` must be free of `role`, `team_*`, `user_id`, `instruction`, `memory` (G3) |
| Overview, one section | `GET …/sections/{sectionId}/entries` | `id, code, title, description, status, tags, entry_type, section_id` | **whole route** (G6) |
| Overview, all entries | `GET …/entries` | as above | — |
| Overview, tag filter | `GET …/tags` | `tag, count` | **whole route** (G7) |
| Overview, external links | `GET …/links` | grouped links + `total` | **whole route** (G8) |
| Overview, sessions | `GET …/sessions` | `id, code, title, status, focus, started_at` | — (gated on `include.sessions`) |
| Entry | `GET …/entries/{code}` | `id, code, title, description, content, blocks, status, tags, section_id, entry_type, session_id` | must resolve short codes |
| Entry cross-refs | `GET …/entries/{id}/crossrefs` | `outgoing[]`, `incoming[]` | **whole route** (G12) **and redaction** (G1) |
| Entry links | `GET …/entries/{id}/links` | `url, title` | **whole route** (G12) |
| Entry related | `GET …/entries/{id}/related` | `id, code, title, tags` | **whole route** (G12), same-research only (G2) |
| Entry siblings (prev/next) | `GET …/sections/{sectionId}/entries` | `id, code, title` | same as G6 |
| Roadmap list | `GET …/roadmaps` | `id, code, title, description, status, statuses` | — |
| Roadmap detail | `GET …/roadmaps/{code}` | nodes, edges, positions | **whole route** (G9) |
| Session | `GET …/sessions/{code}` | session + `questions[]` + `entries[]` | **whole route** (G10) |
| Tasks | `GET …/tasks` | `id, title, description, status, priority, result` | **whole route** (G11) |
| Export | `GET …/export` | the export document | must honour `include` (G4) |
| Unlock | `POST /api/shared/{t}/unlock` | — | distinct 401 / 429 (G13, G14) |
| Live | `WS /ws?share={t}` | `type, entity, entity_id, research_id, research_code` | scrub `actor_user_id` (G5); `4401` on revoke (G20) |

Owner side:

| Screen | Reads / writes | Missing |
|---|---|---|
| `Share` button badge | `GET /api/researches/{id}` | `active_share_count` (G18) |
| Share list | `GET /api/researches/{id}/shares` | — |
| Create | `POST /api/researches/{id}/shares` | response must carry the **absolute** URL, not a token to be assembled client-side — the server knows `base_url` and the client does not when it sits behind a proxy |
| Revoke | `DELETE /api/shares/{id}` | — |

---

## 10. Out of scope

Not forgotten — deliberately not designed:

- **`/s/{token}/mindmap` and `/s/{token}/graph`.** Both aggregate questions and
  tasks into a single payload (`useResearchMindmap`, `GET …/graph`), so both
  would need the `include` filter applied *inside* the graph builder, server
  side. That is a second, subtler leak surface than the read routes, and it
  should not ride along on the first release. The `/graph` route in the brief's
  contract can exist; the UI will not call it yet.
- **The question detail page** (`/session/{s}/question/{q}`). Questions render
  in the session page's list; a per-question route is one more surface for one
  more payload. `QuestionList` gets `:linkQuestions="false"` under a share.
- **Entry revision history.** `HistoryPanel` exposes who edited what and when,
  including author kind. That is internal working process, of a piece with
  `instruction` and `memory`, and it should stay out.
- **Search inside a shared view.** Held for #54, as the issue says. `SearchModal`
  is not mounted (the share pages are chromeless).
- **Comments from visitors.** Its own issue, its own security posture.
- **A "shared" marker on `ResearchCard`** in the research list. Worth having —
  it is how an owner notices an old link on a research they forgot about — but
  it needs `active_share_count` in the list payload as well as the detail one.
  Do it next, not now.
- **Narrower share scopes** (`scope: session | entry | roadmap` in the
  migration). The UI designed here is research-scoped only. The column can
  exist; nothing in this spec sets it to anything but `research`.
- **Analytics beyond `view_count` / `last_seen_at`.** Per the issue: anything
  finer is surveillance of the recipient.
- **A light theme.** There is not one.
