---
name: new-feature
description: The delivery loop for a feature in this repository — scope it from the issue and the code, build it, verify it yourself, then run the full agent fleet over the diff and open a PR. Use when starting a feature, taking a GitHub issue into work, implementing something that crosses layers, or when the work is finished and needs reviewing before it ships.
---

# Delivering a feature

One loop, six steps, in order. The fleet in step 4 is not optional and not a
judgement call — every agent in this repository exists because something got
through without it.

## Step 1 — Scope

Gather before writing anything.

- If there is an issue, read it whole: `gh issue view <n>`. The specs in this
  repository carry a **"The decision this issue encodes"** section — that is the
  part that binds. Deviating from it is a decision to state out loud, not a
  detail to quietly reinterpret.
- Read the code the change lands in, not just the code it calls. Trace one
  existing example of the same shape end to end: storage → service → api/mcp →
  frontend.
- Check the constraints in `CLAUDE.md` that the change touches — route
  registration, access control, the pure-Go/no-CGo build, the MCP tool contract.
- Ask the user only where two readings would produce materially different work.
  Everything else is a judgement call you make and state.

Done when you can name every layer the change touches and every file you expect
to create.

## Step 2 — Plan the work, and start the designer

Create tasks with `TaskCreate`, one per layer or per shippable slice, and keep
them current as you go. A feature here reliably spans six or more files across
four layers; the list is what keeps the last layer from being forgotten.

**If the feature has any user-facing surface, dispatch `ui-designer` now**, in
the background, before you write a line of storage code. It convenes a UX/UI
panel over the feature and returns a build-ready specification: screens, states,
which components to reuse versus create, interaction rules, and any field the UI
needs that the API does not yet return.

The point is the overlap — it designs while you build the backend, so the
specification is waiting when you reach the frontend instead of being invented
at the keyboard. Give it the issue number, the surface it covers, and the
endpoints the feature will expose.

Read its data-contract gaps as soon as they arrive: a field the design needs and
the API does not return is backend work, and you are still in the backend.

## Step 3 — Build

Work bottom-up: migration and storage, then service, then API and MCP, then
frontend, then docs. Each layer compiles and its tests pass before the next one
starts.

Build the frontend from the designer's specification. Where you deviate from it,
say which part and why — in the PR body, not silently.

- New migration: take the next free number in `internal/storage/migrations/`.
- New MCP tool: register it in `internal/mcp/tools.go`, optional fields are
  pointers, never return a Go error.
- New route: `wrap` for writes, `wrapRead` for reads.

## Step 4 — Verify it yourself

Before any agent looks at it:

- `make test` — the whole suite, not the package you touched.
- `npx nuxt typecheck` in `frontend/` when the frontend changed. Pre-existing
  errors are noise; filter to your own files.
- Run it for real. The `local-api-testing` skill drives the live API; use it to
  exercise the new paths, the error codes, and — when there is a migration — an
  upgrade of a database that already holds data.
- Then stand it up for the user with the **`preview-feature`** skill: a fresh
  build, a throwaway database, one seeded research per branch of what you built,
  and a map of URLs. Anything with a visible surface gets this before the fleet
  runs — an agent reviewing a screen nobody has looked at reviews the code, not
  the product.

Report what you ran and what it said. A test you did not run is a test that
failed.

## Step 5 — Run the fleet

Dispatch **every** agent in the table below in one message so they run in
parallel, each with: the branch, that changes are uncommitted, the issue number,
the list of new and modified files, and the specific risks you already know
about. A vague brief gets a vague review.

| Agent | Reviews | Skip only when |
|---|---|---|
| `code-reviewer` | Correctness of the diff, across layers | never |
| `access-auditor` | Every read and write path for cross-user leaks | never |
| `llms-docs` | `llms.txt` and `internal/docs/` against the code | never |
| `usability-reviewer` | Whether the feature is workable end to end, and what it still cannot do | never |
| `ux-tester` | UI states, keyboard, focus, responsiveness, edge-case data | no frontend change |
| `story-writer` | Storybook stories for new and changed components | no component added or changed |
| `component-curator` | Duplication against the existing component library | no component added |

`ui-designer` is not in this table — it runs in step 2, before the code exists.
Reviewing a design after the UI is built is how a redesign gets scheduled.

The three conditional rows are skipped only when their subject genuinely is not
in the diff, and the skip is stated in your report with the reason. Everything
else runs on every feature.

Then: read every finding and act on it. Fix what is real. For anything you
reject, say which finding and why — a dismissed finding that is never mentioned
is indistinguishable from one you never read.

Agents that write (`story-writer`, `llms-docs`) change files. Re-run `make test`
and the typecheck after the fleet lands.

## Step 6 — Ship

Use the `committing-and-creating-prs` skill: branch, commit, push, PR. The PR
body states what the feature does, what the fleet found, and what you decided
not to do.

If the frontend changed, run `make frontend-embed` before committing so the
embedded assets match the source.

## What makes this loop fail

- **Reviewing before verifying.** An agent finding a bug your own test run would
  have caught is a wasted review cycle.
- **A brief with no risks in it.** You know where the change is thin; say so in
  the dispatch and the review lands there.
- **Asking whether to run the fleet.** It runs. This document is the answer.
