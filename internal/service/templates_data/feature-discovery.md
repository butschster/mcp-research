---
slug: feature-discovery
name: Feature discovery
description: Find what to build next from evidence the product already generates, and end with two or three candidates each carrying the test that could kill it.
when_to_use: Use when an existing product needs its next thing to build, the backlog is opinions, and the answer has to come from observed evidence — workarounds, ticket clusters, churn causes — rather than a brainstorm.
when_not_to_use: Not for whole-business ideas (startup idea, venture discovery), the fate of a shipped feature (feature kill-or-keep), who the product is for (audience definition), or ordering candidates you already hold — roadmap prioritisation.
skills: [structured-interviewing, evidence-grading]
---

# Feature discovery

This research ends with two or three candidates, each summoned by evidence the
product already produced. A feature nobody's behaviour asked for is not a
candidate here, however good it sounds.

## Before you propose anything

The kickoff already covered the decision, who makes it, what they believe and
when they need it. **Do not ask any of it again.** Ask these together, in one
message, and propose no structure until you have the first two.

1. **"What evidence does the product already throw off that nobody reads —
   support tickets, feature requests, usage data, sales-call notes, an old
   churn study? Name what exists and who can pull each of them this week."**
   The whole methodology runs on this inventory; without it there is only
   generation, and generation is the thing this research exists to avoid.
2. **"What are users doing with the product that it was not designed for — the
   exports into spreadsheets, the fields repurposed, the second account, the
   browser extension somebody wrote?"** Workarounds are demand with a receipt.
3. **"What was the last feature request you refused, and what did the
   requester do next?"** The answer shows both what demand looks like here and
   what the substitute is when it goes unmet.
4. **"If a candidate wins, who builds it and roughly how much can they build
   next quarter?"** A shortlist sized beyond the capacity that will receive it
   is a wish list with evidence attached.

## Structure to propose

Adapt it. Create a section when you have something to put in it.

- **Evidence inventory** — what exists, where it lives, who pulls it. Written
  first, because it bounds everything after it.
- **Workarounds** — one entry per workaround somebody built and maintains, with
  what it costs them to keep running.
- **Request clusters** — tickets and asks grouped by the problem behind them,
  never by the solution asked for.
- **The job it almost does** — where usage bends the product toward work it
  does not finish.
- **Candidates** — one entry each: the problem, the evidence that summoned it,
  the cheapest test.
- **Killed** — candidates dropped, and why, and what would reopen them.
- **Shortlist** — two or three. Written last.

The tempting structure is brainstorm → "what do you want?" survey → vote →
build. It fails three times. Users answer a want-question with solutions, and
the solution they name is rarely the cheapest fix for the problem they have.
The survey samples the vocal. And a generated list has no evidence attached to
any entry, so the argument that follows is settled by seniority.

## Working rules

**A candidate must name the evidence that summoned it.** The obvious version
generates ideas first and then goes looking for supporting evidence — that is
a lawyer's method, not a researcher's. A candidate with no workaround, ticket
cluster, churn cause or bent usage behind it goes to Killed on arrival, not to
the shortlist. Write where the evidence was found before writing what it
implies.

**Record the problem, not the requested solution.** When somebody asks for a
feature, write down what they were trying to do when they asked. Ten requests
for ten different features are often one problem wearing ten costumes, and the
cluster is counted in problems. "Users have asked for this" enters no entry
without the count and the problem behind the asks.

**A workaround is the strongest signal and your first competitor.** Somebody
pays for it in hours, every week — ask what it costs them to keep running,
because that number is the ceiling of what the feature is worth to them. But
the obvious reading — "they built a workaround, so build the feature" — misses
that the workaround already serves them. The test for a workaround-summoned
candidate must show the person would abandon what they built, not merely that
they would praise a replacement.

**Cheap tests before expensive certainty.** Every candidate carries the
cheapest observation that would validate or kill it — a concierge version, a
fake door with a named cohort, one account's workaround replaced by hand —
with its cost in money and days. Grade what the test produces by what it cost
the user: switched beats signed up beats clicked beats said.

## What a finished entry looks like

A candidate: the problem in one sentence, in the words of somebody who has it;
the evidence that summoned it ([[E12]]) with counts — how many accounts, worth
how much; the cheapest test with its cost and duration; and what argues
against building it. 150–400 words.

Refuse to write three things: a candidate whose only evidence is that a
competitor has it; "users have asked for this" without the count and the
underlying problem; and a shortlist longer than three — a fourth candidate is
a ranking somebody refused to finish.

## When it is done

The shortlist entry names two or three candidates, each citing its summoning
evidence and carrying a test with a cost and a date, and the Killed section
holds at least two candidates with the reason each died. If nothing has been
killed, the evidence has not been allowed to referee — say so rather than
closing sections. Close with a dated task: run the first candidate's test,
owned and dated.

## The trap

Every channel in the inventory was built to hear current, engaged customers —
tickets come from people who bothered to write, workarounds from power users,
sales notes from deals big enough to chase. What nobody files a ticket about
is the thing that made somebody quietly leave. **Before the shortlist closes,
check each candidate against the churn evidence: a roadmap assembled purely
from the engaged is a plan to serve the retained better while the leak
continues.**
