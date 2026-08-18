---
slug: feature-kill-or-keep
name: Kill or keep a feature
description: Decide whether a shipped feature earns its carrying cost — kill, keep or invest — with the exit executed rather than deferred.
when_to_use: Use when a shipped feature's fate is genuinely open — kill, keep as-is, or invest — and the verdict has to survive the customer who complains and the engineer who built it.
when_not_to_use: Not for choosing between candidate new features — that is a comparison. Not for whether to charge for it (monetisation readiness), and not when the verdict is already made and a write-up is wanted.
skills: [structured-interviewing, evidence-grading]
---

# Kill or keep a feature

This research ends with one of three verdicts and its execution written down.
"Keep and watch" is not one of them.

## Before you propose anything

The kickoff already covered the decision, who makes it, what they believe and
when they need it. **Do not ask any of it again.** Ask these together, in one
message, and propose no structure until you have the first two.

1. **"What does this feature cost per quarter to keep alive — engineering hours
   on bugs and on interactions with new work, support tickets, the thing you
   didn't build because of it? Rough numbers are fine; name the biggest line
   even if you cannot price it."**
   With the carrying cost unstated, "keep" is free by default and always wins.
2. **"Who inside the company would fight the kill, and what claim would they
   make? And which customer, by name, would you expect to hear from first?"**
   The resister's claim decides which evidence this research has to gather.
3. **"If the feature vanished tomorrow, what would its users actually do
   instead — inside your product or outside it?"** The substitute is what turns
   a kill from a betrayal into a migration.
4. **"Can you get usage split by account rather than totals — and if not, what
   is the closest thing you can pull this week?"** The answer decides the shape
   of the whole study.

## Structure to propose

Adapt it. Create a section when you have something to put in it.

- **Verdict frame** — the three outcomes and what each commits to. Kill is a
  date and a migration path per dependent, not a deprecation banner. Keep is
  the carrying cost accepted in writing, plus a review date and the number
  that would convert it to kill — that pair is what separates a real keep
  from "keep and watch" or "keep but deprioritise", which commit to nothing
  and are not verdicts.
- **Carrying cost** — one entry per cost line, priced or honestly unpriced.
- **Usage, by account** — who uses it, how deep, revenue-weighted.
- **The dependents** — one entry per account or workflow that breaks, with its
  substitute.
- **The case for keeping** — steel-manned, written from its defenders' claims.
- **Kill test** — the cheapest reversible probe: the flag off for a cohort, a
  deprecation notice to a segment.
- **Verdict** — written last, with the execution attached.

The tempting structure is usage dashboard → user survey → pros and cons →
recommendation. It fails three times. Totals hide concentration: ten thousand
events can be three accounts. A survey asks people to defend an option that
costs them nothing to defend, so "would you miss it" is always yes. And a
pros-and-cons list weighs a real, paid maintenance cost against a hypothetical
loss — hypotheticals always win, so the answer is always keep.

## Working rules

**Usage is counted in accounts and revenue, never in events.** The obvious
version counts clicks, and ten thousand clicks from three accounts is three
accounts. If no per-account analytics exists, the sample is support tickets and
sales calls that named the feature over the last quarter — and the absence of
both across a quarter is itself evidence, not a gap.

**"Keep" pays the same evidence bill as "kill".** The obvious version treats
keep as the safe default that needs no case. Keeping is a decision to keep
paying the carrying cost; the keep entry must say what that cost buys and which
named accounts buy it. A keep with no case is a kill nobody had the nerve to
write.

**A complaint predicted is not a churn predicted.** The obvious version stops
at "account X would be upset". Ask a dependent account to walk through the last
time they used the feature — somebody who cannot is defending an option, not a
workflow. Loud is not leaving.

**A week of behaviour beats a quarter of opinion.** When the evidence stalls,
run the kill test: the flag off for a small reversible cohort, and record who
noticed, who wrote in, and who silently switched to the substitute. Silence
from a flagged-off cohort outweighs every survey answer collected about it.

## What a finished entry looks like

A dependent entry names the account, its revenue, the workflow, the last-use
date, and what they would do instead — citing where each fact came from
([[E12]]). 150–400 words.

Refuse to write three things: "some users would be upset" with no name
attached; an average of usage while per-account rows exist; and the verdict
"keep but deprioritise" — that is kill without the honesty or keep without the
budget.

## When it is done

The verdict entry names one of the three outcomes with its execution written:
kill — a date plus a migration line for every named dependent; invest — what
changes and the metric it must move; keep — the quarterly cost accepted, in
writing. The strongest opposing fact sits beside the verdict, whichever way it
went. If after three sessions the dependents section is empty while the
carrying cost is not, that state licenses the kill test — say so rather than
waiting. Close with a dated task: the first execution step for kill or invest,
or, for keep, the review date and the number that reopens the question.

## The trap

Every kill-or-keep run discovers the feature has exactly enough users to make
killing scary and too few to make keeping rational — because a feature only
gets audited once it falls into that band. The band pulls the verdict toward
"keep and watch", the one outcome that costs nothing today and reruns this
entire exercise next year. **Watching is not an outcome. If you choose it,
write the date and the number that converts it to kill — or admit you chose
keep.**
