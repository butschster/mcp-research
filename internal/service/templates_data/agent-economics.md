---
slug: agent-economics
name: What an AI agent can be sold to do
description: Work out which real work an agent can be paid to do, what a run costs, what it can be charged, and at what error rate it stops being worth buying.
when_to_use: Use when deciding whether an AI agent or model capability can be sold as a product or a service, and the answer has to survive real customer inputs and a real invoice.
when_not_to_use: Not for choosing a model, a framework or a vendor — that is a comparison. Not when which business to start is undecided, which is venture discovery. Not for pricing proven value — that is monetisation readiness.
skills: [evidence-grading]
---

# What an AI agent can be sold to do

The unit of analysis is **a job somebody already pays for**, never a capability.
"The model can do X" is the beginning of the question, not an answer to it.

## Before you propose anything

The kickoff established the decision, who makes it, what they believe and by
when. **Do not ask that again.** Ask these together, in one message, and **create
nothing until the first is answered** — if nobody does this work today, this is
the wrong methodology and a research made in the meantime has to be abandoned.

1. **"Whose job is this today, and what does that person cost?"**
   A named role, an hourly rate or a retainer, and roughly how many hours a week.
   Without it there is nothing to price against and nothing to compare a run to.
   If nobody does this work today, that is the finding: you are not replacing a
   cost, you are creating a category, and the research is a different one.
2. **"What have you already run it on, and how did you choose those inputs?"**
   The second half matters more. Inputs you picked are a demo; inputs from the
   customer's actual queue are evidence, and the gap between them is the subject
   of this whole research. **If there is no customer yet**, say so and take the
   best rung of this ladder that is actually reachable: a colleague's real
   backlog, a year of historic tickets from the job as it is done today, one
   friendly prospect's unfiltered sample, or a public corpus of the same
   artefacts. Getting a worse sample is the work; waiting for a perfect one is
   how this research never finishes.
3. **"When it gets one wrong, who finds out, what does it cost them, and how
   often would they put up with it?"**
   A wrong answer somebody spots in ten seconds and a wrong answer that reaches a
   client are different businesses with different prices and different delivery.
4. **What is off the table?** Holding customer data, human review, being in the
   critical path, a per-seat price. Ask now.

## Structure to propose

Adapt it. Create a section when you have something to put in it.

- **The job as it is done today** — steps, who, hours, cost. In money.
- **Real runs** — one entry per batch against inputs you did not choose, dated,
  with every failure recorded verbatim.
- **Failure budget** — the tolerable error rate, who catches a miss, what a miss
  costs, and who pays for it.
- **Unit economics** — cost per run against price per run, in the customer's own
  unit.
- **Delivery shape** — product, service or hybrid, as a *consequence* of the
  failure budget rather than a preference.
- **Evidence** — graded.
- **Decision** — written last.

The tempting structure is *capability → demo → pricing → go to market*. It fails
twice. The demo is built on inputs you chose, so it measures your taste in
examples. And pricing derived from other AI products imports their billing metric
along with it — per seat, per message, per credit — none of which may be how this
job creates value.

## Working rules

**Run it on inputs you did not choose.** Thirty, sampled without filtering, is
the number that supports a claim about a rate. Ten supports a claim about *kinds*
of failure and nothing about how often — which is still worth having, so write
the ten and say which claim it licenses rather than waiting for thirty you cannot
get. Record every failure verbatim, including the ones that look like bad inputs:
the customer will send those too, and "the user asked it wrong" is not a defence
that survives an invoice.

**The error rate is not the number. The cost of an error is.** Ninety-five
percent is excellent for drafting a first pass and catastrophic for anything sent
to a client unread. Write the two together or neither means anything.

**Cost per run is COGS, and it grows with success.** Inference, retries, the
long-context calls, the human minutes of review. Never state a margin without all
four in it. A flat subscription over a metered cost is a bet on usage; say out
loud that you are making it.

**Who checks the output decides what you are selling.** If the customer must
check every run, you are selling a draft and should price like a tool. If nobody
checks, you have taken on the liability and should price like a service — and
find out what happens when you are wrong.

**Price against the labour it replaces, not against software.** The comparison in
the buyer's head is the hours, the agency, the offshore team. A price anchored to
other SaaS quietly caps the value at what software is "supposed" to cost.

**Model capability is a moving floor, not a moat.** Anything that becomes easy
for you becomes easy for everyone on the next release. Say, in one entry, what
you would still have if the underlying model got twice as good tomorrow — the
data, the distribution, the integration, the accountability. If the answer is
nothing, that is the finding.

## What a finished entry looks like

A batch of real runs: how the inputs were sampled, how many, the failures
classified with two or three quoted in full — the rest go in a linked sub-entry —
the cost of the batch, and the wall-clock time. A unit
economics entry: cost per run built up from its parts, the price, the customer's
current cost for the same work, and the usage level at which the margin turns
negative. Under 400 words.

Refuse to write: an accuracy figure with no denominator and no error cost; a
margin that omits inference; a price justified by what a competitor charges for a
different metric; and a run on inputs you chose reported as a rate — it is
evidence about kinds of failure, and the entry has to say which it is.

## When it is done

You can state the job, the cost per run, the price, **the error rate the buyer
would tolerate** and the rate measured on inputs you did not choose, and what
happens when it is wrong. Where the sample is smaller than thirty, the entry says
so and claims kinds of failure rather than a rate — that is a finished research
with a stated limit, not an unfinished one. If you cannot say what a miss costs
and who absorbs it, it *is* unfinished; say so rather than closing sections.
Close with a task for the next batch of real runs.

## The trap

**The whole business lives in the fraction it gets wrong.** Everyone measures the
95%; the price, the delivery shape, the support cost and the churn are all set by
the rest. And the worst failures are the ones the customer cannot detect — a
confident wrong answer destroys trust in the correct ones too, retroactively,
which is a loss no accuracy number contains.

Before believing a result, ask how the customer would find out it was wrong. If
the answer is "they would not", that is not a strength.
