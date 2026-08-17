---
slug: managing-a-research
name: Managing a research
description: Use when deciding what to create next in a research — a section, an entry, a session, a task or a roadmap — or when a research feels disorganised.
ambient: true
---

# Managing a research

The structure is not paperwork. Each entity answers a different question, and
putting something in the wrong one is how a research becomes unsearchable.

## What each thing is for

**Section** — a topic within the research. Three to seven of them, in
investigation order (context → current state → gaps → options → recommendation),
never alphabetical. A section is a *place*, and the conductor is told to aim at
the least-covered ones — so **create a section when you have something to put in
it**, not in advance. An empty section is a standing instruction to invent
content for it.

**Entry** — one finding, written to be read on its own. Not a chat log, not a
running diary. If two findings need different titles, they are two entries.

**Session** — one working conversation with the human, holding the questions you
asked and what they answered. A research has many sessions over its life;
`session_get` resumes the open one rather than starting again.

**Question** — one thing you need from the human, recorded before you ask so the
answer has somewhere to land. Answer with `question_update`; follow-ups are new
questions, not edits to the old one.

**Task** — work to be done that is not a question: something to look up, verify,
or write. The `result` field is where the outcome goes, and a task closed with an
empty result has recorded nothing.

**Roadmap** — a directed graph, for when the shape of the thing is a path or a
dependency network rather than a list. Build it *after* a session or two, when
the structure is known. Built at minute zero it is a guess with a diagram.

## The order that works

1. Read the research: `research_get` returns the goal, the sections and the open
   session. Read the attached skills index in the same response and load the one
   whose trigger matches what you are about to do.
2. Look before writing: `entry_list` on the relevant section, and `entry_read` on
   anything you might be about to duplicate.
3. Ask one question at a time. Record each answer as it arrives, not in a batch
   at the end — a session that dies mid-conversation should lose nothing.
4. Write an entry when a topic has enough in it to stand alone. Do not wait for
   the section to be finished.
5. Mark a section `completed` when it is covered, and say so plainly if it is
   not.

## Rules that are easy to get wrong

- **Do not restart.** If a session is open, continue it. If an entry exists on
  your topic, extend it — check `entry_history` first, because an edit by a
  human is a correction to build on, not to undo.
- **Check you may write before you interview.** `research_list` marks a
  read-only research; a viewer's first write fails after the human has already
  answered five questions.
- **Cross-reference instead of repeating.** `[[E3]]` links an entry, `[[R2:E5]]`
  one in another research, `[[RM1]]` a roadmap. A fact copied into three entries
  is three facts to keep in step.
- **Record the decision, not just the conclusion.** What was considered and
  rejected is the part nobody can reconstruct later.
- **Say what you do not know.** An unanswered question left open is information.
  A gap papered over with plausible prose is a defect that reads like content.
