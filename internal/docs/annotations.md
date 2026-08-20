# Annotations

A person reads a document, reaches a sentence they do not believe, and marks it.
The marks form a queue you work when they ask you to.

## What an annotation is, and is not

**An annotation is a request for work addressed to a place in the text.** It is
not a record of what we know.

That border matters because everything else this product records about how far a
claim can be trusted is written by the machine — about its own work:

| | Written by | Answers |
|---|---|---|
| entry provenance (`session_id`, revisions) | the system | who produced this document, and when |
| what a document says it rests on | the agent, in the prose | how well established the claim is |
| **annotation** | **a person, reading** | **what is wrong with this sentence** |

Two consequences you will meet directly:

- **You cannot create one.** There is no `annotation_create` tool. A mark is
  born from somebody reading and pointing at a sentence, in the web UI. If you
  want to record your own uncertainty about something you wrote, say it in the
  document — a callout, or plain prose naming what the claim rests on.
- **You cannot close one.** `annotation_answer` takes a mark as far as
  `answered`. A person decides whether the work is accepted.

## Tools

| Tool | Purpose |
|------|---------|
| `annotation_list` | The queue: what is marked, where, and what it asks for |
| `annotation_answer` | Record what you did about one mark |

Both follow the ordinary input-schema rule of this server — **send every
property**, using `null` for the ones you are skipping:

- `annotation_list(research_id, status, kind, entry_id, limit, offset)`.
  `research_id` is a plain string and must be the **UUID**: this tool does not
  resolve an `R1` code, and one is `not found`. The five filters are nullable;
  `status` defaults to `open`, `limit` to 15 — which is also the ceiling, so a
  larger number is clamped rather than refused. A full batch comes back with a
  `hint` saying there may be more; finish it, have it accepted, then continue
  with `offset`.
- `annotation_answer(annotation_id, resolution, task_id)`. The first two are
  plain strings, never `null` and never empty — a mark moved to `answered` with
  nothing recorded cannot be reviewed. `task_id` is nullable, and is only for a
  mark somebody promoted to a task.

Refusals arrive as ordinary tool results with `isError: true`:

| Text | Means |
|------|-------|
| `annotation A4 is already settled` | Somebody closed or dismissed it while you were working on it. Move on; do not reopen |
| `your role in this team does not allow this` | You are a `viewer` here. Answering is a write, like writing an entry |
| `not found` | Wrong id, or a research belonging to a team you are not in — deliberately indistinguishable |

There is no tool call that reaches `closed` or `dismissed`, so there is nothing
to retry: the refusal `only a person may close or dismiss an annotation` exists
on the REST route, for a machine credential attempting it there.

## Kinds

The kind is the instruction. They are different jobs:

| Kind | What to do |
|------|-----------|
| `verify` | Find a source and cite it — or say plainly that you could not confirm it. "I could not verify this" is a real answer; inventing a source is not |
| `dig` | Write a child entry and link it from the anchored block with `[[E19]]`. Depth goes in a new document, not in a paragraph that grows sideways |
| `disagree` | **Do not rewrite the text.** Record both positions and who holds each |

`disagree` is the one that goes wrong. The reflex on being handed an objection
is to edit until the objection no longer applies — which destroys the
disagreement instead of recording it. If somebody disagrees with a claim, the
document should end up saying that two positions exist.

## Anchor state: where the marked text is now

Every read resolves the anchor against the document as it currently stands. It
is never stored, because an agent may have rewritten the document since.

| State | Meaning | What to do |
|-------|---------|-----------|
| `anchored` | The text is where it was | Ordinary work |
| `drifted` | The block is there, the sentence under the mark changed | Read the block's current text before answering — the mark may already be moot, or the claim may have been softened rather than checked |
| `moved` | The sentence turned up elsewhere in the document | Check `confidence` before trusting the placement |
| `orphaned` | The text is gone | Do **not** guess what it said. `entry_history` and `entry_diff` from `anchored_revision` show what happened |

`confidence` is `1` for a mark still sitting on its own block, `0.9` where the
sentence was found elsewhere with the words that used to surround it, and `0.6`
where only the sentence itself matched. It is about the *placement*, not about
the claim — `0.6` means read the block before you trust that this is the text
they meant. Matching ignores case and collapses runs of whitespace, and each
occurrence is claimed once, so two marks on a repeated sentence never stack on
the same copy.

An orphan is a finding, not a broken row: it means the paragraph somebody
doubted was rewritten or deleted. Either the doubt was answered by the rewrite
or an inconvenient claim was buried, and only a person can say which.

**Your own writes are checked against this.** Every `entry_update` and
`entry_patch` compares the anchors before and after, and produces
`annotation_report` naming the marks it drifted or orphaned:

```json
{ "annotation_report": { "annotations_drifted": ["A4"], "annotations_orphaned": ["A7"] } }
```

Only a change of state is reported — a mark already orphaned is not repeated at
you — and a `moved` mark is not reported at all. A patch of pure `set_state` ops
produces none of it: a tick moves no prose.

**Where you can read it.** In the result of the write itself — `entry_update`
and `entry_patch` both return it, alongside the block report. If you see it, you
have just changed text somebody had flagged: say so when you answer those marks,
rather than leaving them to find out from the queue.

## Working the queue — a pass

Not "go through all of them". A pass is scoped, worked and accepted as a unit.

1. **Read the queue.** `annotation_list(research_id, status: "open")`. The
   server caps a batch at 15 — a pass is only as large as the diff a person will
   actually read.
2. **Say what you will take.** Group by document: the marks come back ordered by
   entry, with `by_entry` and `by_kind` counts beside `count`, because a batch
   that is one document costs one read of it and a batch scattered over six
   costs six.
3. **Work each one** according to its kind.
4. **Answer** with `annotation_answer`, naming what you produced: `[[E19]]` for
   an entry you wrote, `[[Q7]]` for a question you asked. The resolution is
   indexed for `[[...]]` exactly as a task's is, so these become real links.
   `[[A4]]` is **not** one of them — a mark's own code resolves to nothing;
   name it in plain text when you report to the person.
5. **Leave it at `answered`.** A person accepts or sends it back.

Write one revision per document, not one per mark. A pass somebody rejects is
then undone with a single restore.

## When you cannot answer

Say so. Do not invent a source, and do not answer a question that needs the
person.

Create a question in the active session with `question_create` and name it in
the resolution: *"needs you — asked [[Q9]]"*. That is what stops a mark going
round in circles.

`attempts` counts how many times an answer was sent back, `previous_resolution`
is the answer that was refused, and `rejections` carries the objections
themselves — one entry per refusal, with the revision it was made against.

**Read the rejections before trying again.** Repeating a refused answer is the
commonest way a pass wastes itself, and the objection is the only thing that
says what would be different. Two rejections means nobody can settle this from
inside the system: escalate it to a question instead of attempting a third.

## What a mark carries

```json
{
  "annotation_id": "…",
  "code": "A4",
  "entry_id": "…",
  "entry_code": "E12",
  "entry_title": "Latency budget",
  "block_id": "b3k9x1qz",
  "kind": "verify",
  "status": "open",
  "quote": "costs fall by 40 percent",
  "note": "where is this from?",
  "attempts": 1,
  "previous_resolution": "found it on their blog",
  "rejections": [{ "reason": "a blog is not the benchmark itself", "revision": 8 }],
  "entry_type": "blocks",
  "anchored_revision": 7,
  "anchor": {
    "state": "anchored",
    "strategy": "block_id",
    "confidence": 1,
    "block_text": "…the current text of the block…"
  }
}
```

`block_text` is why you can triage the queue without reading the documents. Read
the entry for the ones you are about to work, not for the ones you are sorting.
`block_id`, `anchored_revision`, `anchor`, `previous_resolution` and the anchor's
`block_text` are each omitted when there is nothing to say, so treat an absent
`anchor` as *unknown*, never as `anchored`.

An item with **no `block_id`** is a mark on a **markdown** document, which has no
addressable blocks: it is placed by its text alone — `strategy: "quote_in_block"`
where the sentence is still there, `"none"` where it is not — and it drifts more
easily than one pinned to a block. The tool response does not name `entry_type`
(the REST payload does), so read the strategy rather than assuming a block was
lost.

## Boundaries

- No share route reaches a mark. Who doubted which sentence is working process,
  like revision history.
- **No export carries them** — not markdown, not the Obsidian vault, and not the
  portable dump either. A research moved to another server arrives with an empty
  queue, so a mark is answered where it was made.
- Creating a mark, answering one, editing or deleting one are all writes: a
  `viewer` reads the queue and changes nothing in it.
- Deleting the entry deletes its marks with it, as it does its revisions.
