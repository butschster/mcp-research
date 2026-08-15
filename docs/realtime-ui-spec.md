# Realtime UI specification — issue #68

Three surfaces: **connection state**, **access revoked while you look at it**,
**someone else changed what you are reading**. No new page, no settings screen.

Written against `frontend/components/**`, `frontend/assets/css/main.css`,
`frontend/composables/useRealtimeUpdates.ts`, `frontend/app.vue`,
`internal/api/ws/hub.go`, `internal/service/events.go` as of `feat/teams-and-roles`.

---

## 0. Read this first — what the UI needs that the backend does not yet give

Ordered by how much of this spec collapses without it.

### 0.1 `actor_user_id` is **not enough** to suppress your own write. Add `actor_client_id`. — BLOCKING

The whole of Surface 3 rests on "was this me?". `actor_user_id` answers it wrong twice:

- **Auth off** (the default, single local user): every write — from the browser
  and from the MCP agent — carries the same empty or same local identity. The UI
  can never tell its own checkbox tick from the agent rewriting the entry. It
  either flickers on every own-write or ignores every agent write. Both are the
  bug the 1200 ms timer exists to hide.
- **Two tabs, one user**: tab A ticks a checkbox, tab B is reading the same
  entry. Same `actor_user_id`, so tab B suppresses a change it must show.

**Ask:** the browser generates a per-tab id once (`crypto.randomUUID()`, held in
`sessionStorage`), sends it on the WS handshake (`?client_id=`) *and* on every
mutating REST request as `X-Client-Id`. The service copies it onto the event;
the envelope carries `actor_client_id`. The rule then becomes
`event.actor_client_id === myClientId → ignore`, which is exact in every mode.

Keep `actor_user_id` as well — it is what distinguishes "you, in another tab"
(copy: "You changed this in another tab") from "someone else" — but it cannot
carry the suppression decision alone.

### 0.2 The revocation event must carry the names, because after it lands nothing can be fetched — BLOCKING for Surface 2

Once access is gone, `GET /api/researches/{id}` returns 404. The research list
page never had the name in memory in the first place. So the directed event has
to be self-describing:

```
type:            "access.revoked"
scope:           "research" | "team"
reason:          "removed_from_team" | "research_transferred" | "team_deleted"
research_id:     string   // empty when scope=team
research_code:   string   // "R7" — used to match the open route
research_name:   string   // "Payment rails" — used in the copy
team_id:         string
team_name:       string   // "Northwind Ops" — used in the copy
ts:              RFC3339
```

Addressed to one user id, delivered only to that user's sockets. Emitted from
`TeamService.RemoveMember`, `TransferResearch`, and team deletion
(`team_service.go:279`, `:575`).

### 0.3 Role downgrade has no event at all — HIGH

`owner → viewer` is not revocation, but the open tab keeps rendering Edit,
Delete and the status picker, all of which now 403. Needs a sibling directed
event:

```
type: "access.changed", research_id, research_code, team_id, role: "viewer"|"editor"|"owner"
```

The UI feeds it straight into `useResearchRole().setFromResearch()` and the
controls disappear with no fetch. Without it, a demoted reader discovers the
downgrade by clicking Save and getting an error.

### 0.4 WebSocket close codes must distinguish "auth rejected" from "network" — HIGH

The footer has to choose between offering **Reconnect** and offering **Sign in
again**, and the client has to choose between retrying with backoff and stopping
dead. Ask for:

| Code | Meaning | Client behaviour |
|---|---|---|
| `4401` | token missing / invalid / expired / revoked (incl. the §2a kick) | stop retrying, state `offline(auth)` |
| `4403` | authenticated but not permitted | stop retrying, state `offline(auth)` |
| `1012` / `1001` / `1006` | restart, going away, abnormal | retry with backoff, state `reconnecting` |

`4401` on handshake also fixes the §7 reconnect loop on `/login`: the client
knows not to retry until a token exists.

### 0.5 `ts` on every envelope — MEDIUM

Needed for honest copy: "last synced 14:02", "changed 2 min ago". The client's
own receipt time is a lie after a tab wakes from sleep with a queued backlog.
Sequence numbers (§4b of the issue) are not needed by this spec; `ts` is.

### 0.6 Confirmed, and this spec assumes it

`research_code` on the envelope (issue §1) — every guard and every route match
in this document uses it. `actor_user_id` — see 0.1. `entry.deleted` already
exists. `section.deleted` and `research.deleted` do **not** (issue §3), so the
"the thing you are reading no longer exists" state cannot fire for a deleted
section or research; if those service methods are added, they must emit.

### 0.7 Optional, cheap, saves a fetch — LOW

`entity_code` (`E3`, `SS1`, `Q2`) on the envelope. Lets a toast say
"E3 was updated" without resolving anything. Not required by any state below.

### 0.8 The composable contract this spec is written against

Not backend, but the builder needs it fixed before the components exist.
`useRealtimeUpdates()` must expose, at module scope, one socket only:

```ts
state:        Ref<'connected' | 'reconnecting' | 'offline'>
reason:       Ref<'auth' | 'gave_up' | null>   // meaningful only when offline
lastSyncedAt: Ref<string | null>               // RFC3339, set on open and on each event
clientId:     string                           // 0.1
isSelf(e):    boolean                          // actor_client_id === clientId
isSelfUser(e):boolean                          // actor_user_id === me, other tab
retryNow():   void                             // cancels backoff, connects immediately
onReconnect(cb): void                          // fires after a gap, so pages refetch
```

**Also delete a socket.** `ConnectionStatus.vue` currently opens its **own,
second, tokenless** WebSocket (`components/ConnectionStatus.vue:25-56`) parallel
to the one in `useRealtimeUpdates`. That is a third connection on top of the
double-connect in issue §7, and with auth on it is the one that gets 401-looped
forever. It becomes a presentational component with props (see the component
plan) and connects to nothing.

---

## 1. The job

> Know, without looking for it, whether what is on my screen is still true — and
> when it stops being true, be told what changed and what to do, without losing
> my place or my draft.

Target cost: **zero steps** in the healthy case (ambient only), **one click** to
recover in every unhealthy case (Retry / Reload / Back to list / Sign in again).
No state in this spec is allowed to exist without exactly one obvious next click.

---

## 2. Screens and entry points

| # | Region | Lives in | Reached from | Replaces |
|---|---|---|---|---|
| A | Connection indicator | `app.vue` footer (`.footer-right`) | always present on chromed pages | today's self-connecting `ConnectionStatus` |
| B | Connection escalation toast | `ToastHost` (already mounted in `app.vue`) | raised by A's state machine | nothing — new, but no new host |
| C | Activity blip | `app.vue` nav (`.nav-right`) | already there | today's 5 s blip on *any* event, incl. your own |
| D | Access-revoked page notice | research-scoped page body | directed `access.revoked` | the page body |
| E | Access-revoked editor banner | above `MdEditor` in `entry/[entryId].vue` | directed `access.revoked` while `editing === true` | nothing; D is suppressed in favour of it |
| F | Remote-change bar (view mode) | under the entry header, above the view toggle | `entry.updated` from another actor | today's silent `refresh()` |
| G | Remote-change banner (edit mode) | above `MdEditor` | `entry.updated` from another actor while editing | nothing |
| H | Revocation / transfer toast | `ToastHost` | directed `access.revoked` while **not** on that research | the list silently losing a card |

Nothing here is a page. Nothing here is reachable by navigation.

With `auth_enabled: false`, D, E and H never fire — the events do not exist —
and no code path may require `useTeams()` or a team name to render.

---

## 3. Layout

### A — the footer indicator

It stays in the footer, on the right, beside nothing else. It is the quietest
element in the product and that is the point: a health indicator that competes
with content is a health indicator that gets ignored when it finally matters.

Structure: a `<button type="button">` (always a button, always one tab stop —
see 6.1), containing a 6px dot and a label at `--type-xs`,
`color: var(--color-text-muted)`, `gap: var(--space-2)`. Existing `.ws-dot` /
`.ws-*` styles carry over unchanged. Padding `var(--space-1) var(--space-2)`,
`border-radius: var(--radius-sm)`, transparent border that becomes
`var(--color-border)` on hover — the same treatment `.user-menu-trigger` uses,
so it reads as the same class of control.

```
DESKTOP  (footer, container max-width 1200px, padding 0 var(--space-6))
┌──────────────────────────────────────────────────────────────────────────┐
│  Research                                                 ● Live         │
└──────────────────────────────────────────────────────────────────────────┘
                                                            ▲
                                              6px dot + --type-xs label,
                                              --color-text-muted, one tab stop

  degraded:                                                 ◐ Reconnecting…
  offline:                                                  ○ Offline
```

```
≤768px  (main.css already sets .ws-label { display: none })
┌────────────────────────────────────┐
│  Research                    ●     │   dot only — see 6.4, the accessible
└────────────────────────────────────┘   name must survive the hidden label
```

### B — the escalation toast

`useToasts().push()`. No layout of its own; `ToastHost` owns the corner
(`--space-5` inset, `--z-toast`, `min(380px, …)`).

```
                          ┌──────────────────────────────────────────┐
                          │ ⓘ  Live updates are paused               │
                          │    Reconnecting since 14:02. What you    │
                          │    see may have changed since then.      │
                          │    Retry now                          ×  │
                          └──────────────────────────────────────────┘
```

### D — access-revoked page notice

Replaces the page body — everything below `Breadcrumbs` — with `EmptyState`.
`.empty-state` already gives `padding: var(--space-16) var(--space-12)` and
centres a title, a `--type-sm` description capped at `40ch`, and an action row.
Breadcrumbs stay so the URL still explains itself.

```
DESKTOP
┌──────────────────────────────────────────────────────────────────────────┐
│ Research / Payment rails                                                 │
│                                                                          │
│                                                                          │
│                              ⃠   (glyph, --type-2xl, opacity .7)          │
│                                                                          │
│                     You no longer have access to R7                      │
│                                                                          │
│         Your access to “Payment rails” ended when you were removed       │
│         from the team Northwind Ops. Ask an owner of that team to         │
│         invite you again.                          (--type-sm, 40ch)     │
│                                                                          │
│                    [ Back to research list ]                             │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘

≤768px — .empty-state drops to padding var(--space-8) var(--space-4); the
button goes full width via the existing .btn min-height 36px rule. No other
change; this region has no columns to collapse.
```

### E / G — the editor banners

Full-width inline strip directly above the editor, inside the normal document
flow. **Not sticky, not floating.** A floating notice over a text area covers
the text it is talking about, and on a 70vh editor it would sit over the
paragraph the writer is looking at.

Material: the existing `.warning-banner` recipe —
`background: rgba(240,184,73,0.06)`, `border: 1px solid rgba(240,184,73,0.2)`,
`border-radius: var(--radius)`, `padding: var(--space-3) var(--space-4)`,
`margin-bottom: var(--space-6)`, `--type-sm`. E uses the `danger` variant
(`--color-error` at the same alphas) because it is terminal; G uses `warning`.

```
DESKTOP — E, revoked mid-edit (the case that must not lose work)
┌──────────────────────────────────────────────────────────────────────────┐
│ ⚠  Your access to this research ended while you were editing.            │
│    Saving is no longer possible. Copy your text before you leave.        │
│    [ Copy my draft ]  [ Download as .md ]  [ Leave editor ]              │
└──────────────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────────────┐
│  MdEditor — untouched, still editable, still holds every character       │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
   [ Save (disabled, title: “You no longer have access…”) ] [ Cancel ]

≤768px — actions stack to a column at gap var(--space-2), full-width buttons.
```

### F — remote-change bar, view mode

Between the entry header block and `.view-toggle`. It pushes the content down
once, on arrival, and never again — the content itself does not move while
being read.

```
DESKTOP
┌──────────────────────────────────────────────────────────────────────────┐
│ E12  Reconciliation rules                    [Edit] [History] [Copy] [🗑] │
│ Short description…                                                       │
│ ◇ agent · edited 3m ago · r7 · View history →                            │
├──────────────────────────────────────────────────────────────────────────┤
│ ◇  An agent updated this entry 2 minutes ago.        [ Reload ]  [ × ]   │  ← F
├──────────────────────────────────────────────────────────────────────────┤
│ [Rendered] [Source]                                                      │
│ ┌──────────────────────────────────────────────────────────────────────┐ │
│ │  entry content — unchanged until Reload is pressed                   │ │
│ │                                                                      │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘

≤768px
┌────────────────────────────────────┐
│ ◇ An agent updated this entry      │
│   2 minutes ago.                   │
│ [ Reload ]                     [×] │
└────────────────────────────────────┘
   label wraps, actions drop to their own row, gap var(--space-2)
```

Type: message at `--type-sm`, `color: var(--color-text)` (not muted — it is the
one thing on the page asking for a decision); timestamp inside the same
sentence, `relativeTime()` from `useRelativeTime`. `Reload` is a `.btn.btn-sm`;
the dismiss is the `.copy-btn`-weight ghost `×` already used by
`GettingStartedBanner`.

---

## 4. Component plan

| Element | Verdict | Component / props |
|---|---|---|
| A Footer indicator | **extend** `components/ConnectionStatus.vue` | Rip out its private socket. New props: `state: 'connected'\|'reconnecting'\|'offline'`, `reason?: 'auth'\|'gave_up'`, `lastSyncedAt?: string`. Emits `retry`. Root becomes `<button>`. Keep `.ws-dot` / `.ws-connected` / `.ws-reconnecting` classes; add `.ws-offline`. Update `ConnectionStatus.stories.ts` — it already mocks the three states, add `offline(auth)` and a `lastSyncedAt` story. |
| A state machine (grace/escalate timers) | **new** `composables/useConnectionBanner.ts` | Not a component. Wraps `useRealtimeUpdates()` state with the 5 s grace and 20 s escalation timers, the `visibilitychange` re-evaluation, and the toast raise/dismiss. Exists because `app.vue` must not grow a second timer block, and because the timers are the part with the subtle bugs (wake-from-sleep) and should be testable on their own. |
| B Escalation toast | **reuse** `useToasts().push()` | `{ variant: 'info' \| 'error', title, message, action: { label, onClick }, timeout: 0 }`. `ToastHost` is already mounted in `app.vue`. Nothing new. |
| C Activity blip | **reuse** `components/ActivityIndicator.vue` | Unchanged props (`active`, `label`). Only the trigger in `app.vue` changes: fire on events where `!isSelf(event)`, and never while `state !== 'connected'`. |
| D Revoked page notice | **reuse** `components/EmptyState.vue` | `icon`, `title`, `description`, default slot for the `[Back to research list]` button. Add `tabindex="-1"` on `.empty-title` for focus handoff (6.3) — one attribute, not a new prop. |
| D/E/H page-level plumbing | **new** `composables/useAccessRevoked.ts` | Subscribes to `access.revoked` / `access.changed`, matches against the current route's `research_code`/`research_id`, and returns `{ revoked, notice }` plus raising toast H when the event is for a research you are not on. New because every research-scoped page needs identical matching logic and six copies of it is exactly the class of bug issue §1 documents. |
| E / G Inline banners | **new** `components/InlineNotice.vue` | Props: `variant: 'warning' \| 'danger' \| 'info'`, `title?`, `message`, `dismissible?: boolean`; slot `actions`; emits `dismiss`. Lives at `frontend/components/`. Nothing existing fits: `WarningBanner.vue` takes no props at all (it hard-codes the in-memory-mode text and fetches `/api/health` itself), `EmptyState` replaces a whole region rather than sitting in flow, and a toast is transient and corner-anchored — staleness that persists must not be told by something that dismisses itself. Build it on the existing `.warning-banner` token recipe so it is visually the same object, and refactor `WarningBanner` onto it in the same pass. |
| F Remote-change bar | **reuse** `InlineNotice` (variant `info`) | With `EntryAuthorBadge` in its message slot for the `◇ agent` / `●  person` glyph. No separate component — a second inline-notice shape is how this catalog grows a fourth kind of card. |
| F/G entry logic | **new** `composables/useRemoteChanges.ts` | Returns `{ pending, count, lastAt, authorKind, reload(), dismiss() }` for one entity id. Coalesces a burst of events into one bar, holds them while a modal or drag is open (6.5), and applies the `isSelf` rule. New because the suppression rule is the thing being fixed and it must exist in exactly one place — it replaces the ad-hoc 1200 ms window in `entry/[entryId].vue:281-285` and `shouldSuppressRefresh()` in `roadmap/[roadmapId].vue:297`. |
| "Copy my draft" / "Download as .md" (E) | **reuse** `navigator.clipboard` + `useDownload`'s blob-link pattern | The draft is already in memory (`editForm.content`); no request, no endpoint. |
| Role downgrade application (0.3) | **reuse** `composables/useResearchRole.ts` | `setFromResearch({ role })` on `access.changed`. The controls already key off `canWrite` / `isViewer`, and `TeamViewerNotice` already explains the absence. |
| Revoked-team page (`/teams/{id}`) | **reuse** `EmptyState` | Same shape as D, different copy. |

Net: **one new presentational component** (`InlineNotice`), one extended
(`ConnectionStatus`), three new composables, zero new hosts, zero new pages.

---

## 5. States

### 5.1 A — connection indicator

Two clocks turn three composable states into five display states. Both clocks
are re-evaluated on `document.visibilitychange`, because a `setTimeout` armed
before a laptop lid closed fires late and would paint "Reconnecting…" over a
connection that is already back.

| Display state | Entered when | Dot | Label | Toast | Announced? |
|---|---|---|---|---|---|
| **Live** | `connected` | `--color-success`, existing glow | `Live` | — | no |
| **Grace** | `reconnecting`, < **5 s** | unchanged from Live | `Live` | — | no |
| **Reconnecting** | `reconnecting`, ≥ 5 s | `--color-warning`, `pulse-dot` | `Reconnecting…` | — | no |
| **Stale** | `reconnecting`, ≥ **20 s** | as above | `Reconnecting…` | B raised once, sticky | yes, once, via toast |
| **Offline** | `offline` | `--color-text-muted`, opacity .5, no pulse | `Offline` | B raised immediately, sticky | yes, once, via toast |

The grace window is the whole design. A closed lid, a wifi handover and a
`make build` server restart all produce a 1–4 s gap; a banner for each of those
is a banner nobody reads by Thursday.

**Loading** (first connect, before the socket opens): treated as Grace — the
footer shows `Live` for up to 5 s, then `Reconnecting…`. It must never render
`Offline` on page load, which is what today's component does for the whole
handshake.

**Overloaded**: a flapping connection (open/close every 2 s) must not raise a
toast per cycle. The escalation toast is raised at most once per *degradation
episode*; an episode ends only after 10 s of continuous `connected`. The toast
id is held in the composable and dismissed on recovery.

**Recovery copy.** If the escalation toast was raised and the connection then
returns, dismiss it and push one auto-dismissing success toast:

> **Back in sync** — Reconnected and refreshed what was on screen.

If it was never raised (Grace or Reconnecting only), recovery is **silent**. The
refetch happens; nobody is told; that is the reward for a healthy connection.

**Toast copy, stale (info, sticky):**

> **Live updates are paused**
> Reconnecting since 14:02. Anything changed since then is not on your screen
> yet.
> `Retry now`

**Toast copy, offline / `reason: 'gave_up'` (error, sticky):**

> **Not receiving live updates**
> The connection to the server keeps failing. Your page will not update on its
> own until it is back.
> `Try again`

**Toast copy, offline / `reason: 'auth'` (error, sticky):**

> **Your session ended**
> You were signed out, or your access key was revoked. Sign in again to keep
> getting live updates.
> `Sign in again` → `logout('/login?next=' + route.fullPath)`

**Error state of the indicator itself:** there is none. It has no request to
fail. If `state` is somehow undefined it renders Grace, i.e. Live — a health
indicator that itself errors and shouts is worse than one that stays quiet.

### 5.2 C — activity blip

| State | Behaviour |
|---|---|
| Event from another actor | blip on, `Updating`, 5 s (unchanged) |
| Event from me (`isSelf`) | nothing — this is the fix |
| Event from me, another tab | blip on, same as another actor |
| `state !== 'connected'` | never shown; A already carries the story |
| Burst of 30 events | one continuous blip, timer refreshed, never a queue |

### 5.3 D / E / H — access revoked

| Situation | What happens |
|---|---|
| On the research, view mode, no dialog open | Page body → D. Breadcrumbs stay. URL unchanged. Focus moves to the notice title. |
| On the research, a modal open | Close the modal first (see 6.3), then D. |
| On the research, **editing, no changes** (`editForm` deep-equals the loaded entry) | D. There is nothing to protect. |
| On the research, **editing, dirty** | **E, not D.** Editor stays mounted and editable, Save disabled with a title, the draft is reachable by Copy and Download. |
| On a *different* research, or the list, or `/teams` | H toast + refetch of the current page's list. No modal, no redirect. |
| The list page itself | Refetch `GET /api/researches` (authoritative; do not splice the card out client-side) + toast H. |
| `/teams/{id}` for the team you left | D-shaped `EmptyState`: "You are no longer a member of Northwind Ops" / `[ Back to teams ]`. |
| Auth off | Impossible. No event, no code path. |
| Revoked twice / racing events | Idempotent: `revoked` is a one-way latch per route. A second event changes nothing. |
| Revocation arrives during page load | The in-flight fetch will 404 anyway; show D with the event's copy rather than the generic "Research not found", because it names the reason. |

**D copy — removed from team:**

> **You no longer have access to R7**
> Your access to “Payment rails” ended when you were removed from the team
> Northwind Ops. Ask an owner of that team to invite you again.
> `[ Back to research list ]`

**D copy — research transferred away:**

> **R7 moved to another team**
> “Payment rails” was moved to a team you are not a member of. Ask whoever moved
> it to add you, or to move it back.
> `[ Back to research list ]`

**E copy — dirty editor (danger):**

> **Your access ended while you were editing**
> This research was moved out of your reach, so this draft can no longer be
> saved. Copy it somewhere safe before you leave the page.
> `[ Copy my draft ] [ Download as .md ] [ Leave editor ]`

`Leave editor` → `EmptyState` D, i.e. the same terminal state, reached by choice.
`Copy my draft` flips to `✓ Copied` for 2 s, the `EmptyState`/`copy-btn` pattern.

**H copy (info, self-dismissing):**

> “Payment rails” is no longer shared with you. It has been removed from your
> list.

**Loading / empty / error for D:** D *is* the terminal state — it has no fetch,
no spinner, no failure mode. The one hazard is the empty name: if
`research_name` is missing (0.2 not delivered), fall back to the code alone —
"You no longer have access to R7" — and drop the second sentence's quoted title
rather than rendering `“”`.

### 5.4 F / G — remote change

| Situation | Behaviour |
|---|---|
| `isSelf(event)` | Nothing. No refetch, no bar, no blip. |
| Same user, other tab | Bar, copy: "You changed this entry in another tab." |
| Another actor, view mode | Bar F. Content untouched until `Reload`. |
| Another actor, edit mode | Banner G. No reload offered. |
| N events in a burst | One bar. Copy pluralises: "An agent made 4 changes to this entry." Timestamp = newest. |
| `entry.deleted`, view mode | Body → `EmptyState`: "This entry was deleted" / "Someone deleted this entry while you had it open." / `[ Back to Payment rails ]` |
| `entry.deleted`, edit mode dirty | G escalates to danger: "This entry was deleted while you were editing. Saving it will fail." + `[ Copy my draft ] [ Download as .md ]` |
| Reload fails (404/network) | The bar switches to danger in place: "Could not reload — the entry may have been deleted. `Try again` / `Back to Payment rails`". The old content stays on screen; it is stale but it is *something*, and it may be the only copy of a paragraph the reader was about to quote. |
| A modal is open (History, Delete confirm, status dropdown) | Bar is held, not rendered, until the dialog closes. See 6.5. |
| List/aggregate pages (research index, tasks, sessions, roadmaps, mindmap, graph) | **Silent refetch**, as today — a list has no cursor to disturb and a card appearing explains itself. Blip C is the ambient signal. |
| Kanban card mid-drag | Refetch deferred until `dragend`. |
| Roadmap detail / question detail body | Same bar as F. Retire `shouldSuppressRefresh()`. |

**F copy** (author kind from the entry's existing `author_kind` provenance, via
`EntryAuthorBadge` — no new backend field needed):

> ◇ An agent updated this entry 2 minutes ago. `Reload` ×
> ● Someone else updated this entry 2 minutes ago. `Reload` ×
> You changed this entry in another tab. `Reload` ×

**G copy (warning):**

> **Someone else saved a new version**
> Your draft is untouched, but saving it will overwrite their changes.
> `[ See what changed ] [ Keep editing ]`

`See what changed` opens the existing `EntryHistoryPanel` (`showHistory = true`),
which already renders `DiffView` against the newest revision. `Keep editing`
dismisses the banner; it reappears if another change lands.

**Empty state:** none of F/G has one — no bar is the healthy state.
**Loading:** `Reload` disables itself and reads `Reloading…` while the refetch is
in flight; the bar stays until the new content is painted, then removes itself.

---

## 6. Interaction rules

### 6.1 The footer indicator is always a button

Not "a button when broken". An element that appears in the tab order only when
something is wrong is an element nobody knows exists.

- `<button type="button">`, one tab stop, at the end of the footer.
- Click / Enter / Space:
  - `connected` → `retryNow()` acts as **Refresh now**: refetch the current page
    and stamp `lastSyncedAt`. A visible, honest action for someone who suspects
    staleness.
  - `reconnecting` / `offline(gave_up)` → cancel backoff, reconnect immediately.
  - `offline(auth)` → same as the toast's `Sign in again`.
- No dropdown, no popover, no panel. There is nothing to configure and one thing
  to do.
- Hover/focus reveals the full timestamp via `title` **and** `sr-only` text —
  `title` alone is invisible to touch and to a screen reader.

### 6.2 What is announced, and what is not

- **The indicator is not in a live region.** A flaky café connection would
  otherwise narrate itself every four seconds.
- Escalation goes through `ToastHost`, which is already
  `role="region" aria-label="Notifications"` with per-toast
  `role="status"` / `role="alert"`. One announcement per episode.
- The revocation notice D is announced by receiving focus (6.3), not by a live
  region — a message you have been moved to does not also need to be read at you.
- F's arrival is announced once, politely: the bar carries `role="status"`.
  Suppressed under `prefers-reduced-motion`? No — reduced motion suppresses the
  slide-in animation, not the announcement.

### 6.3 Focus, on the two destructive transitions

Replacing a page body under a focused element drops focus to `<body>`, which
puts a keyboard user back at the top of the document with no idea why.

On revocation (D):
1. If a dialog is open, close it. `ModalOverlay` restores focus to its trigger on
   close — and that trigger is about to be unmounted, so **guard**:
   `restoreTo?.isConnected` before `focus()`. (This is a live bug in
   `ModalOverlay.vue` today for any modal whose trigger unmounts; it silently
   focuses a detached node.)
2. `await nextTick()`, render D.
3. `focus()` the `EmptyState` title, which carries `tabindex="-1"`.

On reload (F):
1. Refetch, paint.
2. The bar unmounts — so move focus to the entry `<h1>` (`tabindex="-1"`) if
   focus was inside the bar; otherwise leave focus exactly where it was, because
   a reader who never touched the bar should not be moved by it.
3. `sr-only` `role="status"`: "Entry reloaded."

Escape: closes the F bar (same as `×`). Escape does **not** close D or E — there
is nothing behind them to return to. Escape inside the editor keeps its existing
meaning (`MdEditor` owns it); E must not steal it.

### 6.4 Colour is never the only signal

- The dot's meaning is duplicated in the label — except at ≤768px, where
  `main.css:1020` hides `.ws-label`. The button therefore carries a computed
  `aria-label` ("Live updates connected. Last synced 14:02.") on every viewport,
  plus `sr-only` text so the state survives the hidden label. **This is an
  existing defect in the shipped component**, not a new requirement.
- `InlineNotice` variants differ by border colour *and* by an icon *and* by the
  first word of the title, matching how `ToastHost` already encodes variant on
  the left border plus icon.
- The product is dark-only — `main.css` defines no light theme and no
  `prefers-color-scheme` block. Contrast checked against `--color-bg #0c1220`
  and `--color-surface #151d2e`: `--color-text-muted` ≈ 6.2:1, `--color-warning`
  ≈ 8.9:1, `--color-error` ≈ 5.1:1, `--color-success` ≈ 9.4:1. All pass AA at
  `--type-xs` (14px). If a light theme is ever added, the muted dot on Live is
  the value to re-check first.
- Long strings: research and team names are user input. `.toast-message` already
  has `overflow-wrap: anywhere`; `InlineNotice`'s message needs the same, and
  `EmptyState`'s title needs it too (`.empty-title` currently has none, so a
  200-character research name in D would blow the layout out sideways).

### 6.5 What waits, what does not

- **Optimistic:** nothing in this spec. Every state here reports something the
  server already decided.
- **Deferred repaint.** A refetch that would repaint under an open dialog, an
  open dropdown, or a drag in progress is queued and flushed on close/drop. The
  rule lives in `useRemoteChanges`; pages pass a `blocked: Ref<boolean>`.
- **The entry body never repaints without a click.** This is the load-bearing
  rule of Surface 3 (see the recorded disagreement, §9).
- **Lists repaint freely.** They have no cursor, no selection, no scroll anchor
  worth defending.
- **Reconnect refetches.** `onReconnect` fires after any gap longer than the
  grace window; each page passes its existing `refresh()`. On the entry page it
  does **not** auto-apply — it goes through the same bar F, because a reconnect
  is exactly when the content is most likely to have moved under the reader.

### 6.6 Session lifecycle

`logout()` must close the socket (issue §2b) and `login()` reopen it; the
footer therefore shows Grace → Live across a sign-in rather than Offline. On
`/login` and `/register` (chromeless) there is no footer and no indicator, and
the composable must not connect at all while `authEnabled && !token` — which is
also what stops the 401 reconnect loop.

---

## 7. Data contract, per surface

| Surface | Reads | Fields needed | Gap |
|---|---|---|---|
| A indicator | WS `/ws` only | socket lifecycle, close code | **0.4** close codes 4401/4403 |
| A "last synced" | envelope | `ts` | **0.5** |
| B toast | — | — | — |
| C blip | envelope | `entity`, `actor_client_id`, `actor_user_id` | **0.1** |
| D / H revocation | directed WS event | `type`, `scope`, `reason`, `research_id`, `research_code`, `research_name`, `team_id`, `team_name` | **0.2** — the whole surface |
| D list refetch | `GET /api/researches` | existing | — |
| Role downgrade | directed WS event | `role` | **0.3** |
| F bar | envelope + `GET /api/researches/{id}/entries/{entryId}` | `entity`, `entity_id`, `research_code`, `actor_client_id`, `ts`; `author_kind` already in the entry payload (`handlers/entry.go:209`) | **0.1**, **0.5** |
| F match by code | envelope | `research_code` | issue §1, already planned |
| G banner | same as F | same | same |
| "This entry was deleted" | `entry.deleted` | already emitted | — |
| "This section/research was deleted" | — | — | **no such event exists** (issue §3); this state cannot be built until `section.deleted` / `research.deleted` are emitted |
| E draft rescue | nothing | in-memory `editForm.content` | — |

---

## 8. Out of scope — deliberately not designed

- A connection or activity **settings screen**. Nothing here is configurable.
- A **history / log of events** ("what changed while I was away"). It needs
  payloads and sequence numbers (issue §4b/§4c) and it is a page.
- **Presence** — who else is reading this entry. No backend for it, and it is a
  different feature.
- **Live collaborative editing / conflict merge.** G warns and points at the
  diff; it does not merge. Last write wins, as today.
- **Offline queueing of writes.** Out of reach without a write buffer.
- **Optimistic list updates from event payloads.** Every refetch here is a full
  refetch, as today; the payload work in issue §4c would change that and is a
  separate design.
- **Onboarding for any of this.** No coach marks, no first-run tooltip on the
  dot.
- **A per-research mute for update notifications.** If bar F turns out to be
  noisy on a busy research, that is a signal to revisit the coalescing window,
  not to add a preference.

---

## 9. Where the panel disagreed

**1. Should the entry body auto-reload when the reader has not scrolled?**
The interaction designer argued for it: a reader who just opened the entry and
has not moved has nothing to lose, and a bar demanding a click for a change they
would accept anyway is friction. The visual designer and the accessibility
reviewer argued against: "has not scrolled" is not the same as "is not reading",
a heuristic that is right 90% of the time still repaints a document under
someone's cursor once a day, and a rule with an exception cannot be explained in
the UI.

**Chosen: never auto-reload the entry body.** The entry is a document people
quote from and copy out of; a repaint mid-read costs more than a stale byte. The
bar is one click and it says exactly what it will do. Lists get the opposite
rule for the opposite reason.

**2. One indicator or two?**
The interaction designer wanted the connection state escalated into the sticky
nav when degraded, because a footer on a long research page is below the fold —
an offline state nobody can see is not a design. The visual designer refused two
homes for one signal: an indicator that teleports is an indicator you cannot
learn the position of, and the nav already holds search, the blip and the user
menu.

**Chosen: one home (footer), escalation via the toast.** The toast is corner-
anchored, always in view, already the app's one notification surface, and it
carries the action. This is why the escalation threshold matters so much — 20 s
is long enough that a toast is proportionate.

**3. Redirect on revocation?**
The interaction designer proposed redirecting to the research list, on the
grounds that the page is dead and the list is where they can act. The
accessibility reviewer objected: an unrequested navigation is disorienting,
destroys the back button's meaning, and gives a screen-reader user no
explanation of why they are somewhere else. The visual designer noted the URL
still has value — it is what you paste when asking to be re-invited.

**Chosen: in place, never redirect.** The notice names the reason and offers the
list as the one next click.
