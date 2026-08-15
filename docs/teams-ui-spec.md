# Teams & membership — UI specification

A design review of the teams surface and the specification to build from. It
came out of one complaint: the lists read badly, managing invited people on a
single team's page reads worse, and the buttons look imported from another
product. All three turned out to be true, and the third is measurable.

Written against the state of `fix/ui-wave-0`. It assumes the decisions in
`ui-improvement-plan.md` §4 — the CSS split, the token contract, and the rule
that a class earns a place in `system.css` at three unrelated consumers.

> **State: built.** Everything below shipped on `fix/ui-wave-0` except where
> this block says otherwise. Four deviations, each a decision rather than an
> omission:
>
> - **`.section-heading` was not promoted** (§3). Its three consumers turned
>   out to be three *designs* under one name: the two export views are a
>   document heading with a rule under it, the team page is a section label.
>   Promoting one would have forced the other two to override it — which is the
>   exact `.danger-zone` bug this work removed. The team page's version stays
>   scoped, next to a count and whatever that section lets you do.
> - **No avatar hue from the user id** (§3). The `--hue-*` tokens are semantic
>   — red means blocked, amber means session. Spending them on identity would
>   make a red avatar read as a state.
> - **The header `+ Invite` stays on a one-member team** (§4). What the review
>   measured was a *filled* button duplicated by a neutral one; the header is
>   neutral now and the inline prompt is a text link, so the two are ranked
>   rather than competing. Hiding a header action at one member and revealing it
>   at two would move the control as the team grows. The duplicate that did go
>   is "Invite someone" in the empty-Researches state — three invitations on one
>   screen was the original complaint wearing different clothes.
> - **Bulk transfer validates first, then moves** (§5). `POST
>   /api/teams/{id}/researches` checks every research before moving any, so a
>   refusal moves nothing; it is not a database transaction, because these
>   repositories share one connection and have no `Tx` plumbing. Per-research
>   `research.transferred` and `access.revoked` events are kept rather than
>   collapsed into a count — those are per-research facts, and dropping them
>   would leave open tabs showing a research their reader can no longer load.
>
> One defect was found during the build that is not in this document and
> affected every page in the product: `.page-bar > .title-with-code` carried
> `flex: 1 1 20rem`, and `.page-bar` turns into a column below 768px — so the
> basis was measured vertically and every page opened on a phone with 320px of
> nothing between the breadcrumbs and its own title.

---

## 1. What is wrong

### 1.1 The three lists do not share a frame, let alone a rhythm

`.data-rows`/`.data-row` has three consumers and `TeamRowList` — the list you
see **first** — is not one of them. It kept its own subgrid
(`TeamRowList.vue:52-56`) through the split that named it as a consumer.

| List | Padding | Columns | Row height |
|---|---|---|---|
| Teams (`TeamRowList.vue:63`) | `--space-4` | 5, subgrid | **56px** |
| Members (`TeamMemberList.vue:91`) | `--space-3` | 5 | **72px** |
| Invites (`TeamInviteList.vue:85`) | `--space-3` | 4 | **48px** |

72px against 48px, under two byte-identical `<h2>`s (`teams/[id].vue:32, :47`).
One page, two lists, half again the height, and nothing saying which one you
came for.

The affordances diverge further than the geometry. A member has a face, a
bordered role picker and a 28px icon button; an invitation has none of those —
an email, grey text for a role, and two bare text links. Two rows about the
same thing: a person and their access.

### 1.2 Three defects nobody had filed

- **`teams/[id].vue:478-503` redeclares `.danger-zone`/`.danger-row`**, which
  were extracted *from this page* into `system.css:862-865` during the split,
  and wins on specificity. The primitive is a red-bordered box; the page renders
  a borderless column. Two designs, one class name, and the discipline check
  reports the global pair as having one consumer — the one overriding it.
- **`teams/[id].vue:58` — `v-if="team.research_count > 0"`.** A team with no
  researches says nothing about researches. The single moment the product must
  say *"there is no work in here yet"* is the moment it goes silent. This one
  `v-if` is where the membership loop breaks.
- **The loading state does not resemble the loaded one.** `.member-skeleton` is
  52px against a 72px row, in gapped cards against flush rules. The page
  visibly reassembles itself.

### 1.3 The buttons: weight, then placement, then style

`/teams/[id]` puts controls at **21.7, 26.6, 28, 32 and 47px in five visual
families** into one viewport. `/research/[id]` puts six controls at **30px in
one family**.

- **Weight.** Outside modals, form submits and error recovery, a filled
  `.btn-primary` appears in a page header exactly three times in the whole
  product — and two of them are these pages. The research page, the product's
  centre of gravity, has no filled button in its header at all. So the teams
  pages are the only ordinary read surface wearing one, and the eye lands on
  the button before the team.
- **Placement.** `+ Invite` appears twice on a one-member team — filled in the
  header (`:24`) and neutral in the body (`:43`) — which is *every* newly
  created team.
- **Style.** `+ Invite` is 26.6px beside a 30px `⋯` in the same row: `.btn-sm`
  overshoots its own `min-height: 26px`, so the stated height never binds.
  That is the exact defect `tokens.css` records as fixed. `RoleSelect` ignores
  the `.select` primitive and is the only select in the product without a
  chevron. `TeamMemberList` shadows the global `.btn-icon`. `.link-btn` at
  `padding: 0` gives a 21.7px target — under the 24px floor — twice per row,
  12px apart, distinguished by colour alone.

**Cross-cutting, found here, affects every page:** at ≤768px `.btn` gets
`min-height: 36px` later in source than `.btn-icon`'s `30px`, so every `⋯`
trigger in the product renders **36×30 — a vertical rectangle** on mobile.

---

## 2. The membership loop, which is the real work

`create team → invite → colleague sees nothing`.

The data layer is fine and needs no work: `team.*` events already reach every
member, `research.transferred` reaches anyone who can now read it, and the
access-verdict cache is flushed on both. The colleague's empty list *does*
repaint the instant a research moves in.

**Every failure in this loop is a missing sentence, and there are four.**

| Where | What is not said |
|---|---|
| Create toast (`teams/index.vue:95`) | That the researches they already have stayed in their personal team |
| `teams/[id].vue:58` | Anything at all — the zero case renders no element |
| `research/[id]/index.vue:71` | That "Move to team…" in a `⋯` menu is the only transfer control there is |
| `pages/index.vue:56` | The invitee reads *"An agent connected to your account can create one"* — founder advice, given to a viewer on somebody else's team, which reads as the invitation having failed |

`GettingStartedBanner` compounds the last one: it fires on `researches.length`
unfiltered, so a fresh invitee is also told to install an MCP server.

### The way out: one structural move, four sentences

**The team page leads with its researches.** The most consequential fact about
a team is what work is in it, and it is currently the smallest thing on the
page and absent when it is zero.

```
Researches (N)      ← what the team is for
Members (N)         ← who can see it
Pending invites (N) ← who is on their way   (one line when empty)
Danger zone
```

This is also the hierarchy fix: three flat lists become one lead and two
supporting sections.

**Sentence 1 — empty Researches, owner.** Replaces the silent `v-if`:

> **Nothing to see in here yet**
> Members of {team} can only read researches that live in this team. Your other
> researches are still in your personal team — move one across and everyone
> here gets it.
> `[ Move researches here ]` `[ Invite someone ]`

This is the one place a filled button survives on this surface: conditional
emphasis on the action that unblocks the flow.

**Sentence 2 — empty Researches, viewer or editor.** Names a human:

> **{team} has no researches yet**
> Researches added to this team will appear here for everyone in it.
> {ownerName} can move one across.
> `[ Copy {ownerEmail} ]` `[ Your researches ]`

Buildable today — the members endpoint needs only viewer rights.

**Sentence 3 — the invitee's landing (`pages/index.vue:52`)**, split on role, and
`GettingStartedBanner` gated on owning a team.

**Sentence 4 — the create toast**: *"{name} is ready. Move a research into it,
then invite someone."* Order matters: the current toast recommends inviting
first, which is exactly how the colleague arrives to an empty list.

**Both directions of transfer keep a home.** "Move to team…" on a research page
is the right control when you are looking at one research. What is added is the
inverse — from a team, pull work in.

---

## 3. Components

**3 new, 5 extended, 1 class promoted, 2 deleted from `system.css`.**

| Element | Verdict |
|---|---|
| `PageHeader`, `ConfirmModal`, `EmptyState`, `ModalOverlay`, `ModalHeader`, `InviteDialog` | reuse unchanged |
| `TeamRowList` | adopt `.data-rows`/`.data-row`; add a research count; make the personal team a link |
| `TeamMemberList` | delete the scoped `.btn-icon`; collapse name and email into one identity block (72px → 56px); `ActionMenu` instead of the `×`; avatar hue from the user id |
| `TeamInviteList` | one `ActionMenu` instead of two `.link-btn`s; add the `invited by` cell (already on the wire, unused); expired as a badge, not colour alone |
| `RoleSelect` | put `.select` on the `<select>` and delete the scoped copy — **third consumer**, which is what the primitive needed |
| `team/TeamResearchList.vue` | **new.** `ResearchCard` is a 400px grid card for the home page; this is a rule list with its own columns |
| `team/AddResearchDialog.vue` | **new.** `TransferModal` moves *one* research *out* from a research page; this pulls *N* *in* from a team page. Same sentence, inverted subject |
| `DangerZone.vue` + `DangerRow.vue` | **new.** Owns the label/note/disabled-reason triple written by hand twice here and zero times in settings. Deletes the class pair from `system.css` and the shadowing override |
| `.section-heading` | **promote** — three unrelated consumers today with three different rules |

---

## 4. States worth naming

- **A one-member team** loses the duplicate `+ Invite`. The header goes neutral;
  a single inline prompt under the one member row carries the emphasis and
  disappears when a second member exists.
- **Fifty members** gets a filter above the list at >12, with the filtered count
  in a live region.
- **A failed member load** currently renders as `members = []`, which reads as
  "you are alone". It needs its own state: *"Nobody has lost access — the list
  just didn't arrive."*
- **An emptied invite list** collapses to one muted line rather than vanishing:
  an owner who revokes the last invitation currently watches a heading
  evaporate under the cursor.
- **The last owner who cannot leave** gets a way out, not just a reason: the
  note carries "Choose a new owner", which scrolls to Members and focuses the
  first non-owner's role select. Same for "delete a team that holds researches".
- **`.data-row--busy`** changes from `opacity: 0.6; pointer-events: none` to a
  faint colour plus `aria-busy` and `inert`. Opacity on text is forbidden by the
  token contract, and `pointer-events: none` does not stop Tab — so today a
  keyboard user can focus and press a control that silently does nothing.

---

## 5. Backend gaps

> **No bulk transfer endpoint.** Only `POST /api/researches/{id}/transfer`
> exists. Moving twelve researches is twelve requests, twelve partial-failure
> states, and twelve events repainting every member's list one row at a time.
> Wanted: `POST /api/teams/{id}/researches` taking a list, transactional, one
> event carrying the count.

> **`GET /api/researches` defaults to active.** The candidate list must ask for
> all statuses explicitly, or an owner holding eight archived researches is told
> there is nothing to move.

> **Team events carry no scope.** `team.*` has no name and no count, so the
> invitee's list repaints silently instead of saying *"3 researches were added
> to Acme"*. Nice to have; the flow is not blocked without it.

---

## 6. Where the panel disagreed

- **The filled primary.** Keeping it: "+ Invite" is genuinely the primary action
  on a one-member team. Removing it: nothing outside a modal is filled anywhere
  else, so these pages read as imported. **Removing won, with the emphasis moved
  into the body as a conditional prompt** — the product has exactly one other
  filled header button, and two exceptions is drift, not a pattern. A permanent
  filled button decays into noise; one that appears only when the team is empty
  is emphasis that carries information.
- **Merging the lists into one component.** It would end the rhythm problem and
  produce four booleans wearing one name. **Four components stay, with an
  explicit contract**: every row list uses the primitive, every row is two zones,
  every row's controls use the same affordance.
- **A fourth list on a page already suffering from three.** Fair, and answered
  by making Researches the *first* section rather than a fourth — the complaint
  was about hierarchy, and hierarchy means something outranks the rest.
- **Burying "Remove member" in `⋯`.** Costs a click on a large team. **It wins
  anyway**: the current control is a 28px unlabelled glyph in one list and a
  21.7px colour-only text link in the other, and destruction should be reachable
  rather than tempting. This multiplies `ActionMenu` from one per page to one
  per row, so the outstanding fix to its roles must land with it.

---

## 7. Deliberately out of scope

`invite/[token].vue` — six outcomes, right copy and right button on each; the
strongest file in the feature. `InviteDialog`'s focus choreography and clipboard
fallback. A nav team-switcher, which would duplicate the `?team=` filter.
Per-research permissions — access is a role in a team, full stop. Email delivery
of invitations: the model is that you pass the link along yourself, and the copy
says so. Pagination on any of these lists.
