# Templates

A template is a **kickoff methodology the model reads** — not a skeleton the code
clones.

It carries no sections, no questions and no rows. It carries a name, the criteria
an agent matches on, and a markdown body saying what to ask the person before
proposing anything, what structure to suggest, what a good entry looks like, and
when the research is finished. The model then designs the research itself.

> A template says how a kind of research is **started**. A skill says how a kind
> of work is **done**.

## Using one

```
template_list()     -> {templates: [{slug, name, tier, description, when_to_use, when_not_to_use}], usage_hint}
template_get(slug)  -> {slug, name, tier, when_to_use, when_not_to_use, body, skills, version, usage_hint}
research_create(..., template_slug: "technology-comparison")
```

`template_list` takes **no arguments** — send `{}`. It never carries a body:
four methodologies arriving in a kickoff that needs one is the cost this feature
exists to avoid. `template_get` takes the slug you read there (an id works too,
but you will not have one) and is the only call that returns a body.

Three rules the kickoff prompt states and this one repeats because they are the
whole point:

- **Ask what decision is waiting before you offer a template.** A structure shown
  first anchors the person to it — they recognise most of it, accept, and the
  parts that do not fit quietly redefine the work.
- **The body is instructions, not a form.** Follow it. Its section list is a
  starting point to adapt to the conversation, never a list to recite.
- **Create a section when you have something to put in it.** The conductor aims
  at the least-covered sections, so an empty one is a standing instruction to
  invent content for it.

Passing `template_slug` to `research_create` does the only structural thing in
the feature: it stamps `template_slug` and `template_version` on the research,
and attaches the [skills](/llms/skills.md) that methodology names. The version is
stamped because the built-ins are refreshed from the binary at every boot —
without it, an upgrade would silently change the text behind a research already
in flight.

The tool answers with `skills_attached` and, when the methodology named something
this research could not get, `skills_unavailable`. Neither key appears when it is
empty. A skill that could not be attached never fails the creation: a research
that exists with five of six skills beats a create call rolled back because a
template named something a later build removed. Read `skills_unavailable` and
tell the user which part of the methodology will not be backed by a skill.

Two things `template_slug` does **not** do. It does not create sections,
questions or tasks — you design those from the conversation. And it is an **MCP
argument only**: `POST /api/researches` has no `template_slug` field, so a
research created over REST carries no provenance and no attached skills.

## What ships

Four, refreshed from the binary at every boot.

| Slug | Ends in | Skills it attaches |
|---|---|---|
| `technology-comparison` | a pick you can defend, closing on an ADR-shaped entry | `evidence-grading` |
| `competitive-landscape` | a recommendation with the two facts that would reverse it | `evidence-grading` |
| `user-interview-study` | findings traceable to what real people said | `structured-interviewing`, `evidence-grading` |
| `literature-review` | conclusions a reader could check, including what was excluded | `evidence-grading` |

Each carries a correction to the obvious version of itself. A few worth knowing,
because they are the difference between the template and a section list:

- **Comparison:** criteria are fixed and weighted *before* candidates are named,
  or they get retrofitted around a favourite. "Keep what we have" is candidate
  zero.
- **Landscape:** a company is a direct competitor only if a real buyer considered
  it in the same purchase. Everything else — the spreadsheet, doing nothing — is
  a substitute, and half of real scans find the threat there.
- **Interview study:** two people make a finding, one makes a signal. Every
  finding carries a verbatim quote and a participant id.
- **Literature review:** what was excluded and why is the load-bearing artifact.
  A list of what you kept is unfalsifiable.

## Two tiers

| Tier | Owner | Lifecycle |
|---|---|---|
| `global` | nobody | ships with the binary, refreshed at boot; editing one forks a team copy |
| `team` | a team | written by that team, theirs alone |

There is no third. A team cannot publish its own template server-wide: a template
body steers a model, and lending one team's instructions to another team's
kickoff needs a trust story this product does not have.

A team's fork keeps its parent's slug — it is the same methodology, edited — and
shadows the global one in every list that team sees, so the same slug is never
offered twice. Resolution follows the same order: a slug is looked up in your
teams' libraries first and falls back to the global set, which is why forking has
any effect at all. What you can see is taken from your own memberships, never
from anything you send, so a slug can never resolve into a team you are not in.
With `auth_enabled: false` there is no caller to scope to and every template on
the instance is visible. Writing one in that mode needs a team id, and
`GET /api/teams` refuses without a signed-in user — so use the id the local
instance actually uses: **`team-local`**. `POST /api/teams/team-local/templates`
works and the template lands in the team tier, not the global one.

## Writing one

**`when_to_use` is required and is what an agent matches on**, before it has read
anything else. Name the situation and the outcome: *"Use when a decision is being
made between named tools and the reasoning will have to survive review."* Both
criteria are capped at 240 characters, counted in runes.

**`when_not_to_use` is worth as much.** Knowing when a methodology is wrong is
what stops it being applied to everything.

**The body** is markdown, up to 24000 characters. It is read once, in full, by a
model that has just started — so unlike a skill description it can be a real
document. Give it four parts: what to ask before proposing anything, the
structure to propose, the working rules, and when it is done.

**`skills`** names the methodology [skills](/llms/skills.md) a research started
this way should follow. Slugs only, resolved as the owning team's copy if there
is one and the built-in otherwise. A *built-in* template naming a skill that does
not exist fails the boot rather than quietly shipping broken; a team template is
not checked that way, so a slug that resolves to nothing comes back on a single
read as `skills_resolved[].missing` rather than being dropped from the list.

**`name` is required and derives the slug**; `description` is one line for a
picker and is optional. An edit that restates only some fields inherits the rest
from the row it is editing rather than blanking them — restating a body must not
silently drop the line an agent matches on.

## Saving one from a research

`GET /api/researches/{id}/templates/draft` returns a **skeleton, not a capture**,
and creates nothing. A template carries no rows, so there is nothing to copy:
what comes back is `{data: {name, when_to_use, body, skills}, hint}` — the
sections that research actually grew, the non-ambient skills it follows, and
headings with the judgement left blank (*Before you propose anything*,
*Structure to propose*, *What good looks like here*, *When it is done*). The
`when_to_use` it returns is a placeholder that says to rewrite it. Somebody
writes the methodology into it and posts it to `/api/teams/{id}/templates`.

Anything else would be dishonest about what a template is.

## Sharing

A share link exposes no template list and no body. `TemplateService` refuses a
share context before it resolves anything, and there are no template routes under
`/api/shared/{token}/`, so a route added later still fails closed. Which
methodology a team follows is working process, the same class as `instruction`,
`memory` and skills.

The stamp goes too. `redactForShare` blanks `template_slug` and
`template_version` alongside `instruction`, `memory` and the team fields, on the
one read path every reader of a research goes through — a slug is a name a team
chose, and "acme-q4-layoff-diligence" is not something a read-only link should
teach.

## REST

| Method | Path | Does |
|---|---|---|
| `GET` | `/api/templates` | the global set plus your teams', forks shadowing their parents, **no bodies**. `{data: [...]}`, each with derived `research_count` and `body_words` |
| `GET` | `/api/templates/{slug}` | one, **with its body**. The path value is tried as an **id first, then as a slug** — an id is not a way past team scoping, so a template of a team you are not in is `404` either way. Also carries `skills_resolved` |
| `GET` | `/api/teams/{id}/templates` | that team's own library, without the global set mixed in |
| `POST` | `/api/teams/{id}/templates` | write one for a team. `201 {data}` |
| `POST` | `/api/templates/{slug}/fork` | copy a **global** into a team and apply the edit; `team_id` goes in the body. `200 {data, forked: true}` — follow the slug to the new row. A slug that is not a global one is `404` |
| `PUT` | `/api/templates/{templateId}` | edit a team template in place; bumps `version`. A global one is `not_allowed` — editing what ships is the fork route |
| `DELETE` | `/api/templates/{templateId}` | delete a team template. `204`. A global one is `not_allowed` |
| `GET` | `/api/researches/{id}/templates/draft` | the skeleton above. `{id}` accepts a research short code |

Read by slug, write by id — the same split skills use, because a fork keeps its
parent's slug and a slug is therefore an address only within one caller's scope.

Every write asks the owning **team** for `editor` or better, since there is no
research to ask about: a non-member gets `404`, a member who is only a `viewer`
gets `403`. With `auth_enabled: false` there is nobody to check and every route
is permitted.

| `code` | Status | Means |
|---|---|---|
| `slug_taken` | 409 | a template of that name already exists in the same tier — including forking a global twice into one team |
| `not_allowed` | 403 | writing to or deleting a template that ships with the app |

Validation failures are `400` and carry a `field`: `name` (empty), `body` (empty,
or over 24000 runes), `when_to_use` (empty, or either criterion over 240 runes —
the error is reported against `when_to_use` even when `when_not_to_use` is the
long one).

There is **no web UI for templates yet.** Everything above is MCP and REST; do
not send a user to a screen.
