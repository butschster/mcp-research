# Skills

A skill is a methodology document the agent opens when it decides it needs it.

> **`instruction` says what *this research* is. A skill says how a *kind of work*
> is done. A [template](/llms/templates.md) says how a kind of research is
> *started*.**

`research_get` lists the skills a research follows — name, tier and one line
saying when to use it. The bodies are not there. When you are about to do the
work a line names, call `skill_load` and read it then.

## Why it works this way

`instruction` is returned in full on every `research_get`. At 200–400 words that
is right. A methodology — how to run an interview, how to grade a source, how to
build a trade-off table — is 800–2500 tokens, a research needs one or two of
them *at the moment of use*, and there are three to six over its life. Putting
them all in an always-loaded field costs about 5k tokens on every call to save
one tool call.

So the index is small and arrives with the research, and the bodies are one call
away.

## Reading them

```
research_get(research_id) -> {
  research: {...},
  sections: [...],
  active_session: {...},
  skills: [
    {slug: "evidence-grading", name: "Grading evidence", tier: "team",
     description: "Use when recording a claim that rests on a source..."},
    ...
  ],
  skills_hint: "Each skill says when to use it. Call skill_load ..."
}

skill_load(research_id, slug) -> {slug, name, tier, description, body,
                                  version, updated_at, precedence}
```

Four fields per entry, never a body. Because the product skills are always in it,
the index is never empty in a working install — so a missing `skills` key means
the built-ins failed to load, not that this research has none.

`skill_load` takes the research **UUID or the `R1` short code** — whichever you
have. `version` and `updated_at` come back with the body, so an agent holding an
older copy can tell it has moved on.

Two rules the tool enforces rather than suggests:

- **One slug per call.** There is no batch form. Loading everything up front is
  the thing this design exists to prevent.
- **Load at the point of use**, not while orienting. A skill read three steps
  before the work it describes has usually been forgotten by the time it matters.

`skill_load` is the **only** MCP tool for skills. Attaching, writing, editing and
detaching are REST or web-UI acts — see below — so nothing an agent does over MCP
changes which skills a research follows.

## Tiers, and what wins

| Tier | Owner | Notes |
|---|---|---|
| `private` | one research | Where a rule that applies only here lives. Never offered to another research. |
| `team` | a team | Reusable across that team's researches. |
| `builtin` | nobody | Ships with the binary, refreshed on upgrade. Editing one forks a team copy; the original is never changed. |

**The index is ordered by tier, and that order is the precedence.** Where two
skills conflict, the higher one wins: private over team, team over built-in.
`skill_load` restates this in every response, because it is the only rule
governing which of two skills the agent follows. It says nothing about
`instruction`, which answers a different question — what *this research* is.

**A slug resolves against the attachment first.** A team's fork keeps the slug of
the built-in it copies, so a research that follows the built-in and never took
the fork still gets the built-in body: what the research actually follows is
looked up before what it merely could attach. Only when a slug is attached to
nothing here does the wider scope (this research's private skills, its team's
library, the built-ins) answer, highest tier first.

## Product skills are not counted

Some built-ins describe the product rather than a domain — how to manage a
research, how to write entries, what an artifact is, how roadmaps work. Four of
the six shipped skills are marked `ambient`: once attached they are **never
counted against the cap** and **cannot be detached** (`not_allowed`, 403). The
flag is structural, set from the shipped file at boot, and never editable through
the API.

**They need no attaching.** The always-on set is unioned into every index and
every attached list, so a research nobody has curated still gets them — and the
agent still learns from the first `research_get` that skills exist. Their bodies
load without an attachment too. Everything else does have to be attached, and
there is still no MCP tool for attaching one. There is exactly one way an agent
attaches a skill: `research_create` with a `template_slug` attaches the skills
that [template](/llms/templates.md) names, at creation and once. Those rows carry
`via_template: true` in the attached listing, so a reader can tell a methodology's
choice from a person's. Everything after that — attach, detach, fork, copy,
promote — is REST or the web UI.

Everything non-ambient counts. A research may follow **six** chosen skills; the
seventh is refused with `skill_cap_reached`, and so is writing a seventh private
one, because a private skill is attached the moment it is created. The limit is a
budget over what somebody decided to follow, not over what the product needs to
function.

### What ships

Six built-ins, refreshed from the binary on every boot — an upgrade rewrites the
shipped rows and never touches a team's fork of one. A built-in that disappears
from a later build is left in the database rather than deleted: a research may
still be following it.

| Slug | Ambient |
|---|---|
| `managing-a-research` | yes |
| `writing-entries` | yes |
| `writing-artifacts` | yes |
| `building-roadmaps` | yes |
| `evidence-grading` | no |
| `structured-interviewing` | no |

## Writing one

Three fields matter, and one of them decides whether the skill is ever used.

**`description` is a trigger, not a summary.** It is the only line always in the
agent's context, capped at 200 characters — counted in runes, so a Cyrillic
trigger line is not worth half a Latin one — and the agent matches situations
rather than topics. Write *"Use when comparing two or more suppliers against a
fixed rubric"*, not *"A guide to vendor scoring"*. A description that describes
gets a skill nobody opens. Over the cap is a field error, never a silent
truncation: a truncated trigger line is a skill that stops working.

**`body`** is markdown, 600–2500 tokens, and required. Over 16000 characters
(runes again) is refused: at that size it is two skills.

**`name`** is required and derives the slug. Uniqueness is per tier *scope* —
globally for built-ins, per team, per research — so a built-in and a team's fork
of it deliberately share a slug, while a second fork of the same built-in in the
same team is `slug_taken`. A name with no Latin characters gets a generated
`skill-xxxxxxxx` slug rather than failing to save.

## What a skill is not

- **Not a starting kit.** A skill is read at the moment of the work it describes,
  not once before a research exists to say how to *start* one. That is a
  [template](/llms/templates.md): read once at kickoff, in full, by a model that
  has nothing yet. A template may **name** skills, and `research_create` with a
  `template_slug` attaches them — which is the one place the two meet. Everything
  else about them is separate: a template is chosen and then done, a skill is
  opened again and again for as long as the research runs.
- **Not a place for rows.** A skill creates no sections, no questions, no tasks.
  It changes how you do the work; it never stands in for doing it.
- **Not tone or format policy for one research.** That belongs in the research
  itself; a skill that legislates entry length or house voice is misfiled and
  will conflict with the research that borrows it next.

## Sharing

**A share link never exposes a skill** — not the index, not a body. Which
methodology a team follows is working process, the same class as `instruction`
and `memory`, and none of it is on the public surface. There are no skills routes
under `/api/shared/{token}/`, and `skill_load` refuses a share context on its own
so a route added later still fails closed.

## REST

Slugs address a skill only inside a research, where the resolution order is
defined. Management is by id, because a built-in and a fork of it share a slug.
Thirteen routes, all of them under the ordinary authenticated API:

| Method | Path | Does |
|---|---|---|
| `GET` | `/api/researches/{id}/skills` | what it follows, in precedence order, always including the product skills. `{data, cap, chosen}` — `cap` is the server's number so no client carries its own, `chosen` counts only what somebody picked, so it is legitimately smaller than the list |
| `GET` | `/api/researches/{id}/skills/library?q=` | what it may attach: the non-ambient built-ins and its team's library, each with `attached`. The always-on product skills are absent — they are already on, so offering them is a control with nothing behind it. `q` matches name or description. Research-private skills are never listed, its own included |
| `GET` | `/api/researches/{id}/skills/{slug}` | one skill **with its body** — the same resolution `skill_load` uses |
| `POST` | `/api/researches/{id}/skills` | `{slug}` attaches an existing skill; `{name, description, body}` writes a research-private one and attaches it. `201` |
| `DELETE` | `/api/researches/{id}/skills/{slug}` | detach. `204`. Ambient is refused |
| `PUT` | `/api/researches/{id}/skills/{slug}` | **built-in only** — copies it into the team with your edit and moves the attachment to the copy in one step. Answers `{data, forked: true}`; follow the slug to the new row. A team or private skill here is `not found`, not an update |
| `POST` | `/api/researches/{id}/skills/{slug}/copy` | team or built-in → a research-private copy, attachment swapped. Already private: returns it unchanged |
| `POST` | `/api/researches/{id}/skills/{slug}/promote` | research-private → the team library, attachment follows, **the private original is deleted**. Anything else is `not_allowed` |
| `GET` | `/api/teams/{id}/skills` | a team's library on its own terms, for a team with no research yet |
| `GET` | `/api/skills/{skillId}` | one skill with its body, addressed the way `PUT` and `DELETE` are |
| `POST` | `/api/teams/{id}/skills` | write a team skill. Created unattached — attach it to a research separately |
| `PUT` | `/api/skills/{skillId}` | edit a team or private skill in place. A built-in is `not_allowed`: editing one is the fork route above |
| `DELETE` | `/api/skills/{skillId}` | delete a team or private skill. A built-in is `not_allowed`; a team skill other researches follow is `skill_in_use` |

Every write needs `editor` or better, asked of whoever owns the thing being
changed: the research for the `/api/researches/{id}/skills…` routes, the team for
`/api/teams/{id}/skills` and for forking or promoting into a library, and — for
the by-id routes — whichever of the two the skill belongs to. A `viewer` gets
`403`, a non-member `404`. With `auth_enabled: false` there is nobody to check
and every route is permitted.

Conflicts carry a `code` so a client can tell them apart without matching on
prose:

| `code` | Status | Means |
|---|---|---|
| `skill_cap_reached` | 409 | six chosen skills already, on attach or on writing a private one |
| `already_attached` | 409 | this research follows that skill already |
| `slug_taken` | 409 | a skill of that name exists in the same tier scope |
| `skill_in_use` | 409 | a team skill other researches follow; the message carries the count |
| `not_allowed` | 403 | detaching an ambient skill, writing to a built-in, promoting something that is not private |

Validation failures are `400` and carry a `field`: `name` (empty),
`description` (over 200), `body` (empty or over 16000).
