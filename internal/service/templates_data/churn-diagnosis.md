---
slug: churn-diagnosis
name: Churn diagnosis
description: Find out why paying customers actually leave, and end with two or three fixable causes — separated from the churn you should keep.
when_to_use: Use when paying customers are leaving and the reasons are folklore — the work has to end in ranked, fixable causes with named ex-customers behind them, not a retention dashboard.
when_not_to_use: Not for interviewing current users — a user interview study. Not when the cause is price (monetisation readiness) or wrong-fit signings (audience definition). Not for usage fading before anyone cancels — retention diagnosis.
skills: [structured-interviewing, evidence-grading]
---

# Churn diagnosis

This research ends with at most three causes, each claiming named departed
accounts, and a statement of the churn you decide to keep. A churn rate
explains nothing; accounts do.

## Before you propose anything

The kickoff already covered the decision, who makes it, what they believe and
when they need it. **Do not ask any of it again.** Ask these together, in one
message, and propose no structure until you have the first.

1. **"Pull the last ten accounts that left, by name, with what they paid and
   how long they stayed. Can you get that list — and for how many could
   someone still get a conversation?"**
   This list decides whether the study runs on interviews or on archaeology,
   and that fork changes everything downstream.
2. **"Which of those ten are you glad are gone? Which departure genuinely
   surprised someone?"** The regretted and the unregretted have different
   causes and different fixes, and a general theory of churn survives neither.
3. **"What does your system record at cancellation — a dropdown, free text,
   nothing? Who fills it in, and what were their options?"** The answer says
   how much of the existing "data" has to be distrusted before anything is
   built on it.
4. **"For each departure, what can you see in the prior thirty days — logins,
   tickets, an invoice change, a champion leaving?"** The trail is the only
   evidence for the accounts that will never talk to you.

## Structure to propose

Adapt it. Create a section when you have something to put in it.

- **Departed roll** — one entry per churned account: what they paid, tenure,
  stated versus suspected reason, regretted or not. Written first, because it
  is the sample everything else draws on.
- **The last thirty days** — what the trail shows before each exit.
- **Exit conversations** — evidence log, one entry per ex-customer actually
  reached.
- **Candidate causes** — one entry per cause, naming the accounts it claims.
- **Kept churn** — the departures we accept, and what signed those accounts in
  the first place.
- **Fixes** — written last: cause, intervention, and the named account it would
  have saved.

The tempting structure is churn-rate trend → segment breakdown → exit survey →
reasons ranked by dropdown frequency. It fails because the dropdown taxonomy
was written by whoever built the cancellation form — "missing feature" and "too
expensive" are the polite exits — so ranking by frequency reproduces the form,
not the market. And a rate trend cannot tell a cause getting worse from the
customer mix shifting.

## Working rules

**The stated reason is the start of the conversation, not the finding.** The
obvious version records "too expensive" and files it under pricing. "Too
expensive" means not-worth-it at that price — so ask what they used the product
for in their last month. If the answer is "nothing", price did not kill the
account; disuse did, months earlier. Every cause entry records the stated
reason and the last-month usage, side by side.

**A cause must name its accounts, and an account has one primary cause.** The
obvious version writes "onboarding is weak" and lets it claim everybody.
Forcing single attribution makes causes compete for the same accounts, which is
what ranking means. If only two ex-customers can be reached, two is a signal,
not a ranking — say the ranking rests on trail evidence and mark each cause
accordingly.

**Diagnose the save, not the exit.** For each regretted account, name the last
moment a plausible intervention existed and what someone would have had to see
to trigger it. If no such moment exists, the fix lives in sales or targeting,
not in the product — hand it to an audience definition rather than inventing a
retention feature.

## What a finished entry looks like

A candidate cause: the mechanism in one sentence as its title, the departed
accounts it claims with tenure and revenue, the quote or trail evidence for
each ([[E12]]), the fraction of regretted revenue it explains with the
arithmetic shown, and the intervention that would have saved a named account.
150–400 words.

Refuse to write three things: "poor product-market fit" as a cause — it
restates the churn; a cause sourced only to the cancellation dropdown; and a
fix with no named account it would have saved.

## When it is done

The fixes entry ranks at most three causes, each claiming named regretted
accounts that together cover the majority of regretted revenue, with the
arithmetic shown — and the kept-churn entry says which departures we accept and
what upstream sign identifies them at signing. With fewer than five reachable
ex-customers, close on trail evidence and label every cause trail-inferred —
do not hold sessions open for interviews that will not happen. Close with a
dated task: the first intervention from the top-ranked cause, with an owner and
a date.

## The trap

The ex-customers who answer your email are the ones who left on good terms —
the polite churn. The angriest and the most indifferent both ignore you, and
they are where the money went. So the interviews will say the product is good
and circumstances changed. **Before believing any cause built on
conversations, check it against the accounts that would not talk: does the
trail of the silent ones match? A cause confirmed only by the friendly
departures is a compliment, not a diagnosis.**
