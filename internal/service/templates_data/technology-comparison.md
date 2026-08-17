---
slug: technology-comparison
name: Technology comparison
description: Choose between named alternatives and be able to defend the choice afterwards.
when_to_use: Use when a decision is being made between named tools, vendors or approaches, and the reasoning will have to survive review by people who were not in the room.
when_not_to_use: Not when the choice is already made and what is wanted is a write-up, and not when nothing is actually being chosen — a survey of the field is a landscape, not a comparison.
skills: [evidence-grading]
---

# Technology comparison

This research ends in a recommendation somebody acts on. Optimise for a
decision, not for coverage.

## Before you propose anything

The kickoff already asked what decision is waiting, who makes it, what the
person already believes and when they need it. **Do not ask any of that again.**
Ask only what is still missing from the list below, in one message, and do not
name a structure until you have it.

1. **What disqualifies a candidate outright — licence, deployment model, data
   residency, budget ceiling, ops headcount?**
   This is the question that decides whether the work is useful. Without hard
   disqualifiers you produce a feature table across candidates that were never
   viable, and every hour spent scoring them is wasted.
2. **Is *keep what we have* on the list?**
   It should be, and it is scored on the same criteria as everything else.
3. **Who signs the decision off, if the kickoff did not already establish it?**

## Structure to propose

Adapt this to what they told you. Propose it as a starting point, and create
each section when you have something to put in it — an empty section is a
standing invitation to invent content for it.

- **Constraints and disqualifiers** — the hard limits, stated so a candidate can
  be eliminated against them.
- **Criteria and weights** — what the soft criteria are and what they are worth.
- **Candidates** — one entry each, including the status quo.
- **Evidence and spikes** — what was actually run or measured.
- **Head to head** — the comparison itself, as a table.
- **Decision** — the recommendation, written last.

## Working rules

**Criteria before candidates, and this order is not negotiable.** Fix the
criteria and their weights, and write them down, before any candidate is
discussed. If a new criterion appears later, add it explicitly and say who asked
for it and why. A criterion that arrives after the candidates is usually a
justification wearing a criterion's clothes.

**Hard constraints eliminate; soft criteria score.** Check every candidate
against the hard constraints first. An eliminated candidate gets a two-line
entry naming the constraint that killed it and is never scored — a matrix row
for a disqualified tool is noise that makes the table look thorough.

**Evidence beats documentation.** Mark every cell as *measured* (we ran it),
*reported* (a named source with a date) or *inferred*. Vendor material is never
evidence for a claim about reliability, cost at our scale, or operational
burden; it is evidence of what they want believed. If a criterion has no
measured evidence for any candidate, ask whether it should be a criterion.

**Ask forced trade-offs, not feature lists.** *"Which would you rather give up —
self-hosting or the managed dashboard?"* rather than *"do you need a
dashboard?"*. A criterion nobody would trade anything for has weight zero, and
finding that out is worth more than another row.

**Record dissent verbatim.** If the human disagrees with the emerging ranking,
write what they said in the session notes as they said it. Paraphrasing dissent
into agreement is how a comparison ends up recommending something nobody
believed.

## What a finished entry looks like

One entry per candidate: what it is, how it scores against each criterion, and
what would make us regret choosing it. One entry for the matrix, as a table,
each row cross-referencing its candidate entry. No prose that restates the
table.

Entries stay under 400 words. Detail goes into a linked sub-entry, not into a
longer one.

## Closing it

The last thing written is a decision entry, ADR-shaped: context, the options
considered, the decision, its consequences, and **what would make us revisit it
and when**. Then create a task for that revisit trigger and for any spike still
outstanding.

If you cannot write the decision entry, the research is not finished — say so
plainly rather than marking sections complete.
