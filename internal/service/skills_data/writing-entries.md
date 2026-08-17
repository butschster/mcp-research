---
slug: writing-entries
name: Writing entries
description: Use when about to create or rewrite an entry, or when choosing between a markdown entry and a blocks document.
ambient: true
---

# Writing entries

An entry is a finding somebody will read months from now without the
conversation that produced it. Write for that reader.

## Shape

- **Title** is the claim, not the topic. "Egress costs rule out Vendor B above
  40TB" beats "Vendor B costs".
- **Open with the finding.** Then the evidence, then what it implies. The reader
  who stops after one sentence should still have the answer.
- **150–400 words** for most entries. Longer than that is usually two entries
  and a cross-reference.
- **One entry, one claim.** If the title needs an "and", split it.

## Which type

`entry_type: markdown` is the default and the right answer most of the time.

`entry_type: blocks` is a JSON document of typed blocks — paragraph, heading,
list, table, quote, code, mermaid, checklist, callout, divider, image, html. Use
it when the content genuinely is structured: a comparison table, a checklist a
human will tick, a diagram beside its explanation. Its advantages are real but
narrow, and they cost you the ability to edit the entry as plain text.

A `checklist` block keeps the ticks a human made even when an agent rewrites the
document around them. `entry_patch` edits one block by id, atomically — use it
rather than resending a whole document to change one line.

The `html` block holds a self-contained document rendered in a sandboxed iframe:
charts, interactive layouts, anything the block catalog cannot express. Write it
in **one theme**, hard-coded. A document carrying both a light and a dark palette
flips to the light one when printed and prints dark text on the dark surface
behind the frame.

## Cross-references

`[[E3]]`, `[[R2:E5]]`, `[[RM1]]`, `[[RM1:N3]]`. They are extracted on write and
become a real graph, so they are worth more than prose links.

Reference rather than repeat. When you find yourself restating a finding from
another entry, link it and state only what is new here.

## What not to write

- **No claim without its source.** Where a source is weak, say so in the same
  sentence as the claim, not in a caveat further down.
- **No "studies show", "it is widely accepted", "users want".** Name the study,
  the person, or the observation. If you counted, give the denominator.
- **No filling.** A thin section is a result. Padding it with plausible general
  knowledge produces something indistinguishable from a finding, which is worse
  than an obvious gap.
- **No silent overwrite.** Before rewriting an entry another session wrote,
  read its history. Building on a human's correction is the point; reverting it
  is a bug that leaves no trace in the text.
