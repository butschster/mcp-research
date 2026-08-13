---
name: ux-review
description: Runs a UX review of the mcp-research web UI and turns it into a staged redesign plan. Use when asked to review the design, improve usability, rethink the interface, plan a redesign, or get expert opinions on the product's UX.
---

# UX Review & Redesign Planning

You are leading a usability review of the mcp-research web UI and converting it
into a plan someone can execute.

The product is a **structured research tool**, not a dashboard product: a user
(often an AI agent acting on their behalf) creates a research, runs Q&A sessions,
writes cross-linked markdown entries, tracks tasks on a kanban, and builds
roadmap and knowledge graphs. Judge it as a thinking and writing tool — reading
comfort, navigation between linked notes, and not losing work — not as a metrics
dashboard.

## Phase 0 — Ground yourself in the actual UI

Never assume the structure; it changes. Build the map from the code:

1. `CLAUDE.md` — architecture and data model
2. `internal/docs/domain-guide.md` — what each entity means and how they relate
3. Walk `frontend/pages/**` — Nuxt file-based routes, this is the route map
4. `frontend/components/**` — the component catalog, each with a `.stories.ts`
5. `frontend/assets/css/main.css` — design tokens: color, type scale, spacing,
   the `--z-*` scale
6. `frontend/composables/` — `useApi`, `useAuth`, `useRealtimeUpdates`,
   `useCrossRefs`, `useKeyboardNav`

Constraints that bound every proposal:

- SPA (`ssr: false`), built by Nuxt and embedded into the Go binary
- Dark theme only — there is no light theme to design for
- No component framework: plain Vue components plus hand-written CSS on tokens
- Vue Flow drives the mindmap, roadmap and knowledge graph views
- `useRealtimeUpdates` can repaint a screen while the user is working on it

Write a short **Product Context Summary**: the core journeys, the current state
of the UI, and where it visibly strains.

For a findings pass over state coverage, keyboard access, edge-case data and
responsiveness, the `ux-tester` agent does exactly that in a fresh context. Use
its output as input here rather than redoing it inline.

## Phase 1 — Look outward

Compare against tools in the same category, not security dashboards: Obsidian,
Notion, Linear, Roam, Logseq, Craft, Bear. What to look at specifically:

- How a dense linked-note view stays readable
- How backlinks and reference graphs are surfaced without becoming decoration
- How an interview or checklist flow keeps a sense of progress
- How markdown editing and reading modes coexist

Search the web rather than relying on memory of these products. Bring back
patterns with a reason they apply here, not a list of screenshots.

## Phase 2 — Expert panel

Simulate a brainstorm across five perspectives. Each reads the Phase 0 findings
and Phase 1 patterns, then argues from its own angle. Keep every proposal tied
to a specific file or component — an idea nobody can locate cannot be built.

| Expert | Focus |
|---|---|
| Product Designer | Information architecture, navigation between research → session → entry, page hierarchy |
| Usability Researcher | Nielsen heuristics, cognitive load, error prevention, recoverability of lost work |
| Knowledge-Tool Specialist | Reading comfort, markdown rendering, cross-reference affordances, graph legibility |
| Accessibility | WCAG, keyboard paths, contrast against the dark palette, focus visibility |
| Interaction & Performance | Perceived speed, skeleton states, realtime repaints, optimistic updates |

Let them disagree. Record the disagreement instead of averaging it away — a
contested proposal is more useful than a bland consensus one.

## Phase 3 — The plan

Produce a written plan with this shape:

- **Executive summary** — the three changes that matter most, and why
- **Problem inventory** — each issue with `file:line`, severity, and who it hurts
- **Roadmap in streams** — Quick wins (hours) / Component overhaul (days) /
  Structural changes (weeks). Order by value over cost, not by area
- **Per-page specifications** — for each page: what changes, which components
  are added, changed, or removed, and what stays untouched
- **Token changes** — any addition to `main.css` variables, and why the existing
  scale cannot carry it
- **What we are not doing** — proposals considered and dropped, with the reason

Save it as a markdown file in the repository and tell the user the path. A plan
that lives only in the conversation is lost at the end of the session.

## Rules

- Complete Phase 0 before proposing anything; every claim cites a real file
- Respect the tokens in `main.css` — new hard-coded colors and spacings are a
  regression, not a design
- Dark theme only: do not spend effort on a light palette that does not exist
- Any component change carries its `.stories.ts` update; that invariant is
  enforced by a hook
- Mind stacking contexts: `animation ... both` keeps a transform and creates
  one. This already caused a dropdown to sink under the cards below it
- Do not propose a component framework or a CSS library; the project is
  deliberately on plain CSS
- Present the plan and get approval before changing any code
