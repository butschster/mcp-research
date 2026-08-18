---
slug: legacy-system-decision
name: Rewrite, refactor or leave it alone
description: Decide what to do about the system everyone complains about — before the rewrite decides itself.
when_to_use: Use when a system you own is blamed for slow delivery, incidents or hiring pain, and the choice among leaving it, containing it, refactoring, strangling or rewriting has to survive the people who will do the work.
when_not_to_use: Not for choosing between named external products, even as replacements — that is a comparison. Not for one incident's causes — that is a postmortem. Not after the rewrite has been promised: that needs a delivery plan, not a justification.
skills: [structured-interviewing, evidence-grading]
---

# Rewrite, refactor or leave it alone

This research ends in a decision about a system you already own, made against
its operational record rather than against the memory of its worst day.
Optimise for a defensible intervention, not for a new architecture.

## Before you propose anything

The kickoff already covered the decision, who makes it, what they believe, what
would change their mind and when they need it. **Do not ask any of it again.**
Ask these together, in one message, and propose no structure until you have the
first two.

1. **"Show me the system's last twelve months as records, not memory:
   incidents, change failure rate, lead time on a typical change, where the git
   churn concentrates. Name what is missing."**
   The kickoff captured what people believe about the system; this asks for its
   record. Where the metrics were never collected, three months of git log plus
   the incident tracker licenses a weaker but honest claim — say which you are
   working from.
2. **"Who are the last two people who landed a non-trivial change in this
   system successfully, and are they still here?"**
   The kickoff asked who decides; this asks who can execute. Whether
   refactoring is even on the menu is decided by this answer, not by the code.
3. **"What must keep working through any transition — the contracts, the
   compliance surface, the integration nobody documented but somebody bills
   through?"**
   Every option is priced against these invariants.
4. **"What has already been tried on this system — the past refactor, the
   rewrite branch that died — and why did each stop?"**
   The reason the last attempt died is usually still alive.

## Structure to propose

Adapt it. Create a section when you have something to put in it.

- **The system as it is** — what it does, for whom, and its operational record.
  The numbers before the architecture diagram.
- **The pain, traced** — each complaint tied to a measurable event. "The code
  is bad" is not a pain until it has a date and a cost.
- **Invariants** — what must survive any option.
- **Options** — leave-alone as candidate zero, then contain, refactor in place,
  strangle, rewrite. Each with its first ninety days and the named people who
  would do it.
- **The invisible system** — behavior only the old code knows: edge cases,
  implicit contracts, the bug that became a feature.
- **Prior attempts** — the graveyard, and why each stopped.
- **Decision** — ADR-shaped, closing with the tripwire that reopens it.

The tempting structure is *current architecture → problems → proposed new
architecture → migration plan*. It fails because it takes the rewrite as
concluded and researches its justification: the proposed architecture is a
document that cannot be wrong, so writing it teaches nothing; the problems
section is curated to motivate it; and nothing in the structure prices the
option of doing less.

## Working rules

**Cost every option in the currency of the complaint.** If the pain is delivery
speed, each option is scored on lead time six months out — not on architectural
merit. Engineers score options on engineering quality, but the decision was
summoned by a delivery or incident pain, and an option that does not move that
number is decoration, however clean.

**The rewrite is estimated against the invisible system, not the visible one.**
The obvious estimate prices rebuilding the code you can read; most of the real
cost is behavior no document states. A rewrite estimate must be accompanied by
at least ten undocumented behaviors the old system enforces, gathered by
interviewing the last two successful changers — and if ten cannot be produced,
the estimate is recorded as a guess, in those words.

**"Who does the work" is a criterion, not a staffing detail.** Each live option
names its people and whether they would stay for it. Every multi-quarter option
must say what remains valuable if it stops halfway, because the strangler's
classic failure is outliving the tenure of everyone who committed to it. Half a
strangle is progress; half a rewrite is a second system to maintain.

**Leave-it-alone is written by its strongest advocate.** If nobody in the room
will argue for it, write the entry as the person who built the system would —
then find them and check. Grade every claim about the system by the
evidence-grading tiers: the record beats the anecdote, and the anecdote beats
the architecture review.

## What a finished entry looks like

An option: what it is; its first ninety days concretely; the named people; its
cost and payoff in the complaint's currency; what remains if abandoned halfway;
and its failure story — how this option specifically goes wrong here, given the
graveyard. 150–400 words; detail goes into a linked sub-entry.

Refuse to write: an option scored on "clean architecture" or "best practices";
a rewrite estimate with no invisible-behavior list; a migration plan before the
decision entry exists; and a leave-alone entry written as a strawman.

## When it is done

Leave-it-alone has a serious entry, written as its strongest advocate would
write it. Every live option has named doers and a halfway-abandonment answer.
The pain has a number, or an entry says why it cannot and names the proxy used.
The graveyard section explains why this attempt will not die the same way. The
decision entry names its tripwire — the observable event that reopens the
question. If the recommendation is rewrite or strangle, done additionally
requires the invisible-behavior list — or, where the changers are gone and it
could not be built, an entry recording that and the estimate labelled a guess,
in those words — and the first replaced slice named with a date. If leave-alone
lost without a numbered pain, the research is not finished — say so rather than
closing sections.

Close with a task dated for the first replaced slice or the first tripwire
check, whichever the decision produced.

## The trap

**The old system is the spec — the only spec you have — and the comparison is
rigged by memory.** The rewrite's advocate compares their imagined system on
its best day against the real one on its worst; the honest comparison is the
reverse. And the organisation that produced the legacy system is the one that
will produce its replacement: unless the research can name what will be
different this time — different people, different constraints, different
incentives — the rewrite is the same system's second draft, minus twenty years
of bug fixes it does not remember making.
