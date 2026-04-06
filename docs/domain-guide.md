# Domain Guide

This document describes every entity in the system, how they relate to each other, and how to use them effectively.

## Overview

The system is built around one core idea: **structured research through AI-assisted interviews**. An AI assistant designs a research structure, conducts interactive Q&A sessions to gather knowledge, and writes findings into organized entries. Everything connects through cross-references, forming a knowledge graph.

```
Research
  |-- Section (3-7 logical topics)
  |     |-- Entry (markdown documents with findings)
  |
  |-- Session (interview workflow)
  |     |-- Question (structured Q&A, can be nested)
  |
  |-- Task (todo items for tracking work)
```

## Entities

### Research

A research project is the top-level container. It defines what you're investigating, why, and how.

| Field | Description |
|-------|-------------|
| `name` | Project name (e.g., "Competitive Landscape Analysis") |
| `goal` | One declarative sentence describing success criteria |
| `description` | What this research covers |
| `instruction` | Working methodology — tone, depth, rules for how entries should be written |
| `memory` | Array of key insights that persist across sessions |
| `tags` | Categorization tags |
| `status` | `active` / `completed` / `archived` |
| `code` | Auto-assigned short code: `R1`, `R2`, `R3`... |

**When to use:** Create one research per distinct investigation topic. Don't mix unrelated topics in one research — create separate ones instead.

**Goal vs Description:** The goal is a success criterion ("Identify the top 3 competitive threats and their mitigation strategies"). The description is scope ("Analysis of direct and indirect competitors in the enterprise SaaS market").

**Instruction field:** This is crucial. Set it during initialization to define how all future entries should be written — tone (formal/casual), depth (executive summary vs deep-dive), domain-specific rules, and examples of high-quality output.

**Memory array:** Think of it as a shared notepad across sessions. When the AI discovers something important during a session, it appends to memory. Next session, it reads memory to avoid re-asking known things. Use `add_memory` to append without replacing existing entries.

---

### Section

Sections divide a research into logical topics. They define the structure of your investigation.

| Field | Description |
|-------|-------------|
| `name` | Slug identifier (e.g., `market-analysis`) |
| `display_name` | Human-readable name (e.g., "Market Analysis") |
| `description` | One sentence defining scope |
| `position` | Sort order (0-based) — determines investigation sequence |
| `status` | `draft` / `active` / `completed` / `archived` |
| `code` | Auto-assigned: `S1`, `S2`... |

**How many sections?** Typically 3-7. Too few means sections are too broad; too many means they overlap.

**Ordering matters:** Sections should follow a logical investigation sequence, not alphabetical. Example:
1. Context & Background (what we already know)
2. Current State (where things stand today)
3. Gaps & Challenges (what's missing)
4. Options & Alternatives (possible approaches)
5. Recommendations (what to do)

**When to mark completed:** A section must have at least one entry before it can be marked `completed`. Don't mark it completed until the topic is fully covered.

---

### Entry

Entries are the actual content — markdown documents containing research findings, analysis, or knowledge.

| Field | Description |
|-------|-------------|
| `title` | Entry title (auto-generated from content if omitted) |
| `content` | Full markdown content |
| `description` | Short summary (auto-generated from first few lines if omitted) |
| `section_id` | Which section this entry belongs to |
| `session_id` | Optional: session that produced this entry (tracks provenance) |
| `tags` | Categorization tags for filtering |
| `status` | `draft` / `active` / `completed` / `archived` |
| `code` | Auto-assigned per research: `E1`, `E2`, `E3`... |

**Content format:** Standard markdown — headers, lists, tables, code blocks, bold/italic. Use `[[E3]]` syntax to cross-reference other entries.

**Tags strategy:** Use consistent tags across entries to enable filtering. Examples:
- By type: `analysis`, `comparison`, `summary`, `data-point`
- By topic: `pricing`, `security`, `performance`
- By source: `interview`, `desk-research`, `survey`

Tags are displayed with counters on the research page. Clicking a tag filters entries across all sections.

**When to create entries:** After accumulating enough information from a session's Q&A. Don't create one entry per question — synthesize answers into coherent documents.

**Auto-generation:** If you omit `title`, it's extracted from the first line of content. If you omit `description`, it's built from lines 2-5 (max 200 chars).

---

### Session

A session is an interactive interview workflow. It's how knowledge enters the system — through structured Q&A between the AI and the user (or another AI).

| Field | Description |
|-------|-------------|
| `title` | Session title (e.g., "Market Positioning Deep Dive") |
| `focus` | Specific topic this session investigates |
| `notes` | Accumulated observations and pivot decisions |
| `status` | `active` / `completed` / `archived` |
| `code` | Auto-assigned: `SS1`, `SS2`... |

**One active session at a time.** A research can have multiple sessions, but only one should be `active`. Complete the current session before starting a new one.

**When to create a new session:**
- Starting a new line of investigation
- Focusing on an under-covered section
- Following up on insights from a previous session

**Session notes:** Use `add_note` to append observations as the session progresses. Notes capture decisions, pivots, and context that don't belong in entries. Example: "User clarified they're only interested in enterprise segment, pivoting away from SMB analysis."

**Multiple sessions are normal.** A research might have:
- Session 1: Initial broad survey (5 questions across all sections)
- Session 2: Deep dive into competitive landscape (8 focused questions)
- Session 3: Follow-up on pricing model specifics (3 targeted questions)

---

### Question

Questions are the building blocks of a session — structured prompts designed to extract knowledge.

| Field | Description |
|-------|-------------|
| `text` | The question itself |
| `area` | Topic area (matches section topics, e.g., `market-analysis`) |
| `rationale` | Why this question matters (helps prioritize) |
| `priority` | `high` / `medium` / `low` |
| `status` | `pending` / `in_progress` / `answered` / `deferred` / `skipped` |
| `answer` | Captured response (required when marking as `answered`) |
| `parent_id` | Optional: links to parent question for follow-ups |
| `code` | Auto-assigned: `Q1`, `Q2`... |

**Question statuses:**
- `pending` — Not yet asked. Starting state.
- `in_progress` — Currently being discussed.
- `answered` — Answer recorded. Requires non-empty `answer` field.
- `deferred` — Postponed for later. Use when the question is valid but premature.
- `skipped` — Deliberately skipped. Use when the question is no longer relevant.

**Follow-up questions:** Questions can be nested up to 3 levels deep. When an answer raises new questions, create child questions with `parent_id` pointing to the original.

**Areas:** Match areas to your section topics. This enables filtering questions by section focus. Example areas: `competitive-landscape`, `pricing`, `technology-stack`.

**How to write good questions:**
- Be specific, not vague ("What is the pricing model for enterprise tier?" not "Tell me about pricing")
- Include rationale so priority is clear
- One topic per question — don't combine multiple asks

---

### Task

Tasks are todo items for tracking work that needs to happen during or after research.

| Field | Description |
|-------|-------------|
| `title` | What needs to be done |
| `description` | Details about the task |
| `priority` | `high` / `medium` / `low` |
| `status` | `pending` / `in_progress` / `blocked` / `completed` / `failed` / `deferred` |
| `result` | Outcome or result (set when completing) |
| `code` | Auto-assigned: `T1`, `T2`... |

**Tasks vs Questions:** Questions capture knowledge through interview. Tasks track work items. Use questions for "What is X?" and tasks for "Do X."

**Task statuses:**
- `pending` — Not started
- `in_progress` — Being worked on
- `blocked` — Waiting for something
- `completed` — Done, `result` field describes the outcome
- `failed` — Could not be completed
- `deferred` — Postponed

**When to use tasks:**
- "Summarize all pricing data into an entry" (content creation)
- "Verify the claim about market share" (fact-checking)
- "Compare feature matrices across competitors" (analysis)
- "Rebuild cross-references after bulk entry updates" (maintenance)

**Results:** When completing a task, always fill the `result` field with what was accomplished. Results support markdown and cross-references (`[[E5]]`).

---

### Cross-Reference (CrossRef)

Cross-references are links between documents, extracted automatically from `[[...]]` patterns in content.

| Syntax | Meaning | Example |
|--------|---------|---------|
| `[[E3]]` | Entry E3 in the same research | "See pricing analysis in [[E3]]" |
| `[[R2:E5]]` | Entry E5 in research R2 | "Related to [[R2:E5]] from our prior research" |
| `[[R2]]` | Research R2 itself | "Building on [[R2]]" |

**Where cross-references work:** Entry content, question text/answers/rationale, task titles/results, session notes. All rendered as clickable links in the web UI.

**Resolution:** When you write `[[E3]]`, the system looks up entry with code E3 in the current research. If found, the reference is marked `resolved` and becomes a clickable link. If not found (e.g., the entry doesn't exist yet), it's stored as `unresolved` and will resolve when the target is created and cross-references are rebuilt.

**Rebuild:** If references become stale (entries deleted or codes changed), use the rebuild endpoint to re-scan all content and update link resolution.

**Visualization:** Cross-references appear on:
- Entry detail page — "Cross-references" block showing outgoing and incoming links
- Mindmap — dashed purple lines between connected nodes
- Export — preserved as `[[...]]` syntax in markdown output

---

## Typical Workflow

### 1. Initialize Research

Use the `research/initialize` prompt or create manually:

1. Define name, goal, and description
2. Design 3-7 sections with logical ordering
3. Set the `instruction` field with methodology guidelines
4. Optionally seed `memory` with known context

### 2. Conduct First Session

1. Create a session targeting the least-covered sections
2. Add 3-8 initial questions across priority areas
3. Work through questions one at a time
4. Record answers, create follow-up questions as needed
5. When enough material accumulates, create entries
6. Update session notes with key observations

### 3. Build Out Content

1. Create entries in appropriate sections
2. Use cross-references (`[[E1]]`) to link related findings
3. Tag entries consistently for filtering
4. Track remaining work as tasks

### 4. Iterate

1. Review coverage — which sections have entries? Which are thin?
2. Start a new session focused on gaps
3. Build on prior sessions (the AI reads `memory` and existing entries)
4. Cross-reference new findings with existing entries

### 5. Complete

1. Mark sections as `completed` when fully covered
2. Complete the final session
3. Review the full research via the export page
4. Mark research as `completed`

---

## Short Codes

Every entity gets an auto-assigned short code for easy reference:

| Entity | Pattern | Scope |
|--------|---------|-------|
| Research | `R1`, `R2` | Global |
| Section | `S1`, `S2` | Per research |
| Entry | `E1`, `E2` | Per research |
| Session | `SS1`, `SS2` | Per research |
| Question | `Q1`, `Q2` | Per session |
| Task | `T1`, `T2` | Per research |

Short codes are used in URLs (e.g., `/research/R2/entry/E3`) and cross-references (`[[E3]]`, `[[R2:E5]]`).

---

## Tips

- **Start narrow, expand later.** Begin with 3-4 sections. Add more if needed.
- **One entry per topic.** Don't dump everything in one giant entry. Split by subtopic.
- **Cross-reference aggressively.** Links between entries make the knowledge graph useful.
- **Use consistent tags.** Decide on a tag taxonomy early and stick to it.
- **Complete sessions properly.** Mark questions as answered/deferred/skipped. Don't leave them pending.
- **Fill task results.** When completing a task, describe what was done. Future sessions benefit from this context.
- **Use memory for surprises.** Don't store obvious facts. Store unexpected findings, pivots, and corrections that future sessions need to know.
