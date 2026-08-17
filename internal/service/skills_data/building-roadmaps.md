---
slug: building-roadmaps
name: Building roadmaps
description: Use when the finding is a path, a dependency network or a decision tree rather than a list — a learning route, a migration plan, a strategy map.
ambient: true
---

# Building roadmaps

A roadmap is a directed graph of typed nodes. It earns its place when the
relationship between the pieces carries meaning that a list would lose — order,
dependency, branching. When the pieces are simply several things, they are
entries.

## When to build one

**After a session or two**, not at the start. The structure of a path is the
thing you are trying to learn; drawing it before you know it produces a diagram
of your first guess, and a diagram is much harder to abandon than a paragraph.

Signals that it is time: the human has said "before X you need Y" more than
once; the same three options keep recurring with different consequences; you
find yourself numbering entries.

## Building it

Work **backwards from the outcome**. Put the terminal node down first — the
thing that will be true when this is done — and ask what must hold immediately
before it. Repeat. A graph grown forwards from "fundamentals" sorts by
difficulty, which is not the same as sorting by dependency and is usually wrong.

- **Node types** carry meaning: a `milestone` is a checkpoint worth naming, a
  `step` is work, a `decision` is a genuine fork, `info` is context that is not
  work.
- **Edges are directed and labelled.** "requires", "leads to", "blocked by".
  An unlabelled edge in a graph of thirty nodes is a line nobody can interpret.
- **Statuses are per roadmap.** Pick a set that matches the domain rather than
  accepting a generic one — for learning, something like not-started → learning
  → practiced → applied; for a migration, planned → in-progress → verified.
- **`ref_type` / `ref_id`** point a node at a real entry, task, session or
  question. That is what turns a diagram into a live view of the work, and it is
  the difference between a roadmap that goes stale and one that does not.

## Keeping it honest

- Update node status from evidence, not optimism. Ask what was produced before
  advancing anything.
- A node that has not moved in three sessions is telling you something: either
  it is blocked, or it was never really on the path. Say which.
- Do not let the graph grow past what somebody can hold. Thirty nodes is a lot;
  sixty is a picture people stop opening.
- If the outcome changes, redraw rather than patch. A graph that grew by
  accretion around a moved goal encodes two plans and describes neither.
