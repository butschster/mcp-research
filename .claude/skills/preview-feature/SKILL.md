---
name: preview-feature
description: Stand up the app locally with data built for the feature you just wrote — fresh build, throwaway database, seeded researches covering every branch and edge case, and a map of URLs saying what to look at on each. Use to see a change working end to end, to hand the user something to click through, or before opening a PR on anything with a visible surface.
---

# Previewing a feature locally

The point is not "the server starts". The point is that the user opens one URL
per branch of what you built and sees it — the happy path, the empty state, the
edge case that breaks layouts, the error.

## Step 1 — Work out what to show

From `git status --short`, `git diff` and the issue, list **every branch the
feature has**: each state a screen can be in, each shape of data the code treats
differently, each new code path. That list is the seed plan; a preview that only
shows the happy path hides exactly the cases worth looking at.

For anything with a diff, a list, or a document, always include: empty, one
item, many items, very long text with no break opportunity, Cyrillic (this
product is used in Russian), and whatever the code caps or truncates.

## Step 2 — Build

```bash
make build-all      # frontend + embed + binary; `make build` alone ships a stale UI
```

The embedded frontend is what the browser gets. Skipping `frontend-embed` after
a frontend change is the most common way a preview shows the old interface and
everyone concludes the feature does not work.

## Step 3 — Run it on a throwaway database

```bash
./bin/mcp-research --transport sse \
  --db "$SCRATCH/preview.db" \
  --web-port 8090 --mcp-port 8091 \
  --auth-enabled --default-user dev@local.dev \
  > "$SCRATCH/server.log" 2>&1 &
```

- **Never point it at `research.db` or `test.db`.** Those hold the user's own
  work; a seed script writes into them and they are not yours to fill with
  fixtures.
- Non-default ports keep it off whatever the user already has on 8088.
- `--default-user` hands out an auto-login token, so the browser opens straight
  into the UI and curl can use the same token:

```bash
TOKEN=$(curl -s http://localhost:8090/api/auth/info \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['auto_login_token'])")
```

Check the log for the migration you added before seeding anything.

## Step 4 — Seed the branches

Write one script (in the scratchpad, not the repo) that builds a research per
scenario, using the REST API with the token. Route shapes and where each id
lives in the response are in the `local-api-testing` skill — read it rather than
guessing.

Rules that make a preview worth opening:

- **One research per scenario**, named for what it demonstrates: "R2 — long
  documents", "R3 — empty states". The user should pick from the list without
  being told which is which.
- **Seed through the API, not SQL.** The API runs the service layer, which is
  the code under test; SQL inserts prove nothing and can create states the
  product cannot.
- **Include the ugly cases.** The 4000-line document, the entry with no title,
  the session with zero questions, the tag that is one character, the title that
  is 300 characters of Russian.
- **Leave one scenario mid-flight** — an active session, an unanswered question,
  a task in progress — so the live WebSocket updates have something to move.

Print the ids and codes as you go; you need them for the map.

## Step 5 — Hand over a map

Report a table: URL, what it demonstrates, what to look for. Not a list of
features — a list of things to click.

```
http://localhost:8090/research/R1/entry/E1   History → 5 revisions, agent vs human, restore r2
http://localhost:8090/research/R2/session/SS1  Changes tab → 3 created, 1 modified, expand a diff
http://localhost:8090/research/R3             empty research — the getting-started state
```

Include the curl one-liner for anything only visible over the API, and say
explicitly what you could **not** demonstrate and why.

## Step 6 — Say how to stop it

Give the kill command and where the database and log live. Leave the server
running — the user is about to click through it.

```bash
kill $(cat "$SCRATCH/server.pid")
```

## What ruins a preview

- **A stale frontend.** `make build-all`, every time.
- **Seeding the user's real database.** Irreversible, and it looks like the
  product corrupted itself.
- **Only the happy path.** The states that break are the reason to look.
- **A wall of URLs with no annotations.** Say what each one is for, or nobody
  opens the third one.
