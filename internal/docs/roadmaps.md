# Roadmaps

Roadmaps are visual directed graphs within a research — learning paths, strategy maps, decision trees, or step-by-step guides. Unlike the auto-generated mindmap (which visualizes all research data), roadmaps are deliberately designed and their nodes can track progress with custom statuses.

## MCP Tools

| Tool | Description |
|------|-------------|
| `roadmap_create` | Create a complete roadmap with nodes and edges in one call |
| `roadmap_get` | Get a roadmap with all its nodes and edges |
| `roadmap_list` | List roadmaps for a research (metadata only) |
| `roadmap_update` | Update roadmap title, description, statuses, or status |
| `roadmap_delete` | Delete a roadmap and all its nodes/edges |
| `roadmap_add_nodes` | Add new nodes and edges to an existing roadmap |
| `roadmap_update_node` | Update a single node (status, title, description, type, position) |
| `roadmap_remove_nodes` | Remove nodes by ID (connected edges auto-delete) |

## When to Create a Roadmap

| Situation | Example |
|-----------|---------|
| Topic has a clear learning sequence | "Vue 3 Learning Path": HTML → JS → Vue basics → Composition API → Testing → Deploy |
| Research uncovers a multi-step process | "Database Migration Plan": backup → schema changes → data migration → validation → cutover |
| User asks "how do I get from A to B?" | "Career Path to Senior Engineer": skills → projects → mentoring → leadership |
| Multiple alternatives exist | "Frontend Framework Decision": evaluate → benchmark → decide (React / Vue / Svelte branches) |
| System architecture mapping | "Microservices Dependencies": API Gateway → Auth → Users → Orders → Payments |
| Onboarding or process flow | "New Hire Onboarding": paperwork → tooling → codebase → first PR → first feature |

## When NOT to Use a Roadmap

Use sections and entries instead when:
- Content is purely textual with no inherent sequence
- A simple list or table would suffice
- The information doesn't have meaningful relationships between items
- You just need to organize findings by topic (that's what sections are for)

## Node Types

| Type | Purpose | Visual |
|------|---------|--------|
| `step` | Regular action item or learning step (default) | Green left accent |
| `milestone` | Key achievement or checkpoint | Purple left accent |
| `decision` | Fork in the path where a choice is needed | Amber left accent |
| `info` | Reference material, prerequisite, or note | Blue left accent |
| `group` | Container for related steps (visual grouping) | Gray left accent |

## Edge Types

| Type | Purpose | Example label |
|------|---------|---------------|
| `default` | Normal progression | "next", "then", "requires" |
| `success` | Positive outcome path | "if passed", "on success", "approved" |
| `warning` | Failure or risk path | "if failed", "if blocked", "rejected" |
| `optional` | Non-required alternative | "alternative", "skip", "advanced only" |

## Custom Statuses

Each roadmap defines its own status vocabulary in the `statuses` field. This allows the same system to model different domains:

| Domain | Statuses |
|--------|----------|
| Learning path | `["not_started", "learning", "practiced", "mastered"]` |
| Marketing strategy | `["planned", "approved", "launched"]` |
| Engineering plan | `["todo", "in_progress", "review", "done"]` |
| Hiring pipeline | `["open", "screening", "interview", "offer", "hired"]` |
| No tracking | `[]` (purely structural graph, nodes have no status) |

Node statuses are free-form strings but should match the roadmap's `statuses` list for consistent UI rendering (colors, progress bars).

## How to Build a Roadmap

### Step 1: Create the full graph in one call

Use `roadmap_create` with all nodes and edges. Assign `temp_id` to each node so edges can reference them before real IDs exist.

```
roadmap_create({
  research_id: "...",
  title: "Vue 3 Learning Path",
  statuses: ["not_started", "in_progress", "completed"],
  nodes: [
    { temp_id: "n1", title: "HTML & CSS Basics", node_type: "step", status: "completed" },
    { temp_id: "n2", title: "JavaScript ES6+", node_type: "step", status: "in_progress" },
    { temp_id: "n3", title: "Fundamentals Complete", node_type: "milestone", status: "not_started" },
    { temp_id: "n4", title: "Choose framework path", node_type: "decision", status: "not_started" },
  ],
  edges: [
    { source: "n1", target: "n2", label: "next" },
    { source: "n2", target: "n3", label: "next" },
    { source: "n3", target: "n4", label: "next" },
  ]
})
```

### Step 2: Extend as needed

Use `roadmap_add_nodes` to add new nodes and edges to the graph. New edges can reference both `temp_id` of new nodes and real IDs of existing nodes.

### Step 3: Track progress

Use `roadmap_update_node` to change node statuses as the user progresses through the roadmap.

### Step 4: Prune

Use `roadmap_remove_nodes` to remove nodes that are no longer relevant. Edges connected to removed nodes are deleted automatically.

## Best Practices

- **Build the full graph in one call** — `roadmap_create` with all nodes and edges is better than adding them one by one
- **Create roadmaps after gathering enough information** — typically after 1-2 research sessions when the structure is clear
- **Use milestone nodes** to mark key achievements or checkpoints that the user can celebrate
- **Use decision nodes** when the path genuinely branches — don't force a linear sequence when alternatives exist
- **Use info nodes** for prerequisites or context that isn't an action step
- **Keep node descriptions concise** — detailed content belongs in entries, roadmap nodes provide the overview
- **Choose domain-appropriate statuses** — "mastered" feels right for learning, "launched" for marketing, "deployed" for engineering
- **Leave statuses empty** for purely structural graphs (architecture diagrams, dependency maps) where progress tracking doesn't apply
