---
slug: incident-postmortem
name: Incident postmortem
description: Turn an incident into a causal account people sign and at most five corrective actions that actually close.
when_to_use: Use when a production incident or serious near-miss has just happened, and the organisation needs a causal account and corrective actions that survive the quarter — not a document that closes the ticket.
when_not_to_use: Not for choosing the failed component's replacement — that is a comparison. Not when a pattern of incidents puts a whole system's future in question — that is the legacy-system decision. Not for a customer-facing apology document.
skills: [structured-interviewing, evidence-grading]
---

# Incident postmortem

This research ends with a causal account every responder has signed or dissented
from on the record, and at most five corrective actions with owners and dates.
Its real product is the next incident's response, not this incident's
explanation.

## Before you propose anything

The kickoff already covered the decision, who makes it, what they believe, what
would change their mind and when they need it. **Do not ask any of it again.**
Ask these together, in one message, and propose no structure until you have the
first two. Evidence decays by the hour, so the first question outranks
everything else in this document.

1. **"What is the raw record — chat logs, alert timeline, deploy log,
   dashboards — and what has already been restarted, rotated or cleaned up so
   that evidence is gone?"**
   The kickoff asked what people believe; this asks what still exists to check
   beliefs against. Anything perishable gets captured before a single interview
   happens.
2. **"Who was hands-on during the incident, can each of them talk before memory
   rewrites itself — and is anyone afraid of what this document will say?"**
   A scared responder produces a fictional timeline. If the answer to the last
   part is yes, that fear is the first finding, and the account cannot be
   trusted until it is addressed.
3. **"What did the system and the responders do right — which defenses fired,
   which call made it smaller than it could have been?"**
   Without this the account is a catalogue of failure, and the defenses that
   held get deleted in the next cost cut.
4. **"Has this class of incident happened before, and where is that
   postmortem?"**
   A repeat incident is a different investigation, because the first document's
   failure — its unread account, its unclosed actions — is now a contributing
   factor.

## Structure to propose

Adapt it. Create a section when you have something to put in it.

- **Timeline** — a single entry whose body is one line per event, UTC, every
  line naming its source: an alert, a chat message, a deploy record. It is
  exempt from the word range below, and narrative is forbidden until it
  exists.
- **What was believed at the time** — what each responder saw and concluded at
  each decision point, verbatim, before hindsight rewrites it.
- **Contributing factors** — plural by construction. There is no "root cause"
  section.
- **What held** — the defenses and decisions that made it smaller.
- **Prior signals** — near-misses, ignored tickets, the last postmortem's
  unclosed actions.
- **Corrective actions** — capped at five, each with an owner, a date, and the
  incident class it prevents.
- **The account** — the narrative, written last, answering "why did this make
  sense at the time" for every decision in it.

The tempting structure is *summary → root cause → fix → lessons learned*. It
fails three times. "Root cause", singular, selects one causal chain by walking
backwards until it hits a person or a component somebody already wanted to
replace. The "fix" is written before the timeline is established, so the
timeline gets curated to support it. And "lessons learned" is where insight goes
to die, because a lesson has no owner and no date.

## Working rules

**Interview responders individually, before any group review; the meeting is
for corrections, not recollection.** The obvious version convenes everyone and
writes down the emerging consensus — which is the most senior, most fluent
narrative, not the truest one. Individual accounts first, verbatim, with the
responder's name attached. A contradiction between two accounts is a finding to
record, not an embarrassment to smooth over.

**"Human error" is where the investigation starts, never where it ends.** Each
time a person's mistake appears, the entry names the condition that made the
mistake easy: what the screen showed, what the runbook said, what the deadline
was, what the last hundred identical actions did. Blameless does not mean names
removed — it means a name is a pointer to a system condition, and removing the
name removes the pointer.

**Mark every counterfactual.** "The alert would have caught it if…" is graded
like any other claim: **demonstrated** (replayed against the record),
**reported** (someone tested something similar, named and dated), or
**asserted**. The obvious postmortem is built from asserted counterfactuals,
which is why its actions do not prevent anything. The tiers in the
evidence-grading skill govern every other claim in the research.

**The incident's cost gets a number or an honest refusal.** Derive it as
duration × affected traffic × unit value; where any factor is unknown, write
"unknown" into the account with which factor was missing — an unknown cost is
itself a finding about the organisation's observability, not a blank to skip.

## What a finished entry looks like

A contributing factor: the condition in one sentence; the timeline lines where
it shows up; who described it, quoted verbatim; and whether the condition is
still true today. 150–400 words; detail goes into a linked sub-entry.

Refuse to write: a single root cause; a timeline line without a source; a
corrective action without an owner, a date and the class it prevents; "human
error" or "be more careful" as a factor or an action; and a sixth action.

## When it is done

The timeline is fully sourced. Every hands-on responder has read the account
and either agreed or has their objection recorded verbatim. There are at most
five actions, each with an owner, a date and a prevented class. Prior signals
are listed, or an entry says the search was done and found none. The account
answers "why did this make sense at the time" for every decision in it. If
responders will neither sign nor dissent on the record, the research is not
finished — say so rather than closing sections.

Close with one task per corrective action, its deadline in the title, and
one more whose title carries a date a month out — *"Count closed actions as of
<date>"* — whose work is to do that count and record the number in the research.

## The trap

**Root cause is a decision, not a discovery.** The causal chain behind any
incident has no natural end — keep asking "why" and you reach the hiring plan,
the funding round, the founding of the company. Where you stop is a choice, and
organisations reliably stop at the cheapest fixable point, or at the most
junior person present. Practitioners know the document's real product is the
next incident's response — which is why the "what held" section and the
closed-action rate matter more than the causal elegance of the account.
