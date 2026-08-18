---
slug: user-interview-study
name: User interview study
description: Learn what people actually do, from people, and end up with findings somebody can act on.
when_to_use: Use when the next step is talking to real users or customers and synthesising what they say into findings a team will act on.
when_not_to_use: Not when nobody can actually be reached — with no access this is desk research and needs a different shape. Not for a satisfaction survey: this is about behaviour, not scores.
skills: [structured-interviewing, evidence-grading]
---

# User interview study

Every entry here must be traceable to something a real person said or did. That
is the whole credibility of the work.

## Before you propose anything

The kickoff already asked what decision is waiting and what the person believes.
**Do not ask that again.** One thing decides whether this methodology applies at
all, and it is usually still missing:

**"Who will you actually talk to, roughly how many, and can you reach them?"**

If they cannot reach anybody, this is not an interview study. Say so and switch:
what they want is usually a **literature review** (if the evidence is published)
or a **competitive landscape** (if the question is about a market). Offer the
swap rather than running an interview study with no interviews.

## Structure to propose

Adapt it; create each section when there is something for it.

- **Decision brief** — what is being decided and by whom.
- **Discussion guide** — the questions, kept as a living document.
- **Participants and sampling** — who was spoken to and who was missed.
- **Evidence log** — what people said, before it is a finding.
- **Findings** — one entry per claim.
- **Disconfirming and unresolved** — what argued against, and what is still open.
- **Recommendations** — written last.

Note what is *not* here. "Study design" is not a section: it is the goal and the
working rules. "Contradictions" is not a section either — a contradiction is a
property of a finding, and as a destination it gets written last, which means
never.

## How the evidence actually gets in

This product interviews *you*, not the participants. So say plainly, at the
start, how what people said will reach the research — there is no transcript
import and no participant record, and pretending otherwise is how this
methodology fails quietly.

The shape that works: **one entry per participant in the evidence log**, written
right after each conversation, titled `P3 — <role or context>`. Quotes go in
that entry, marked as quotes. Everything later cites `[[E12]]` rather than
restating. The participants section holds one entry listing who was spoken to
and who was missed.

If the human pastes a transcript into a session answer, put it in that
participant's entry verbatim before anything else — a paraphrase written from
memory is not evidence, and by the next session the original is gone.

## Working rules

**Evidence before themes.** Write no finding entry until the evidence log holds
at least three participant entries. Count them with `entry_list` on that section
rather than from memory — three conversations you remember is not the same as
three that are written down. Themes emerge from evidence, never from the section
list.

**Two people make a finding. One makes a signal.** A single-source observation
is worth recording — title it *"Signal:"* and say it is one person. A finding
needs two participants who said it independently and unprompted.

**Every finding carries a verbatim quote** with the participant id (P3, P7).
Paraphrase is not evidence. Capture the words they used for things — what they
call the product, the problem, the workaround. That vocabulary is itself a
finding.

**Separate what you observed from what you concluded.** Each finding entry has
an evidence part and an interpretation part, and the interpretation may be
wrong. Say so.

**Never write "users want".** Write what the user did, and when. If they offered
a solution, record the problem it was solving instead.

**Disconfirmation is part of the job.** Every finding ends with what would make
it wrong, and names any participant who contradicted it. When new evidence
contradicts an existing finding, update that entry — do not create a rival one
and leave the reader to referee.

## What a finished entry looks like

A finding: the claim as a title, the evidence with quotes and participant ids,
then the interpretation. 150–400 words.

A recommendation names an owner, cross-references the findings it rests on
(`[[E12]]`), and carries a confidence: strong (five or more participants,
consistent), moderate, or speculative. A recommendation with no linked finding
is deleted rather than softened.

## When it is done

When two consecutive sessions produce no new theme, say saturation is likely and
ask whether to stop. Do not keep interviewing to fill a section — a thin section
is a result, and padding it produces something indistinguishable from a finding.

If a section is still empty after five sessions, say so and propose deleting it
rather than filling it.
