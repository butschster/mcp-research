<template>
  <div class="graph-list">
    <section v-for="group in groups" :key="group.key" class="graph-list-group">
      <h2 class="graph-list-title">
        <span class="graph-list-dot" :style="{ background: group.color }" aria-hidden="true"></span>
        {{ group.label }}
        <span class="graph-list-count">{{ group.nodes.length }}</span>
      </h2>
      <p v-if="!group.nodes.length" class="list-empty">None in this project.</p>
      <ul v-else class="data-rows data-rows--bounded graph-list-items">
        <li v-for="n in group.nodes" :key="n.id" class="data-row graph-list-item">
          <NuxtLink v-if="n.href" :to="n.href" class="graph-list-link">
            <span v-if="n.code" class="short-code">{{ n.code }}</span>
            <span class="graph-list-label">{{ n.label }}</span>
          </NuxtLink>
          <span v-else class="graph-list-link graph-list-link--inert">
            <span v-if="n.code" class="short-code">{{ n.code }}</span>
            <span class="graph-list-label">{{ n.label }}</span>
          </span>
          <span class="graph-list-meta card-meta">
            <StatusBadge v-if="n.status" :status="n.status" />
            <span>{{ n.connections === 1 ? '1 connection' : `${n.connections} connections` }}</span>
          </span>
          <!-- What it connects to, by code: the list is the keyboard and
               screen-reader path, and a count alone says a node is busy
               without saying with what. -->
          <span v-if="n.neighbours.length" class="graph-list-neighbours card-meta">
            <template v-for="(m, i) in n.neighbours" :key="m.id">
              <NuxtLink v-if="m.href" :to="m.href" class="short-code">{{ m.code }}</NuxtLink>
              <span v-else class="short-code">{{ m.code }}</span>{{ i < n.neighbours.length - 1 ? ' ' : '' }}
            </template>
            <span v-if="n.more > 0">+{{ n.more }}</span>
          </span>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
/**
 * The knowledge graph as a list: the same nodes, grouped by type, each with the
 * number of edges that touch it.
 *
 * This is not a fallback for when the canvas fails. The canvas is mouse-only —
 * pan by drag, open by double-click — and a public visitor with a keyboard, a
 * screen reader or a phone gets a black rectangle out of it. The list is the
 * peer that can be read and operated, on the same screen, one switch away.
 */
import type { GraphNode, GraphEdge } from '~/composables/useResearchGraph'

interface NodeTypeFilter {
  key: string
  label: string
  color: string
}

const props = defineProps<{
  nodes: GraphNode[]
  edges: GraphEdge[]
  /** The rows to show, in order — the same list the sidebar draws. */
  nodeTypes: NodeTypeFilter[]
  visibleNodeTypes: Set<string>
  /** Which edge types count — the sidebar's edge checkboxes apply here too. */
  visibleEdgeTypes?: Set<string>
  /** Where a node leads, or '' for a node with no page of its own. */
  hrefFor: (node: GraphNode) => string
}>()

const NEIGHBOURS_SHOWN = 8

const byId = computed(() => new Map(props.nodes.map(n => [n.id, n])))

/** The edges the filters leave: both ends visible, type ticked. */
const activeEdges = computed(() => {
  const visible = new Set(props.nodes.filter(n => props.visibleNodeTypes.has(n.type)).map(n => n.id))
  return props.edges.filter((e) => {
    const src = typeof e.source === 'string' ? e.source : e.source?.id
    const tgt = typeof e.target === 'string' ? e.target : e.target?.id
    if (!src || !tgt || !visible.has(src) || !visible.has(tgt)) return false
    return !props.visibleEdgeTypes || props.visibleEdgeTypes.has(e.type)
  })
})

/** Every node's neighbours, in edge order, without duplicates. */
const neighbours = computed(() => {
  const adj = new Map<string, GraphNode[]>()
  const seen = new Map<string, Set<string>>()
  const add = (from: string, to: string) => {
    const target = byId.value.get(to)
    if (!target) return
    if (!seen.has(from)) seen.set(from, new Set())
    if (seen.get(from)!.has(to)) return
    seen.get(from)!.add(to)
    if (!adj.has(from)) adj.set(from, [])
    adj.get(from)!.push(target)
  }
  for (const e of activeEdges.value) {
    const src = typeof e.source === 'string' ? e.source : e.source?.id
    const tgt = typeof e.target === 'string' ? e.target : e.target?.id
    if (!src || !tgt) continue
    add(src, tgt)
    add(tgt, src)
  }
  return adj
})

const connectionCounts = computed(() => {
  const counts = new Map<string, number>()
  for (const e of activeEdges.value) {
    const src = typeof e.source === 'string' ? e.source : e.source?.id
    const tgt = typeof e.target === 'string' ? e.target : e.target?.id
    if (src) counts.set(src, (counts.get(src) || 0) + 1)
    if (tgt) counts.set(tgt, (counts.get(tgt) || 0) + 1)
  }
  return counts
})

const groups = computed(() =>
  props.nodeTypes
    .filter(nt => props.visibleNodeTypes.has(nt.key))
    .map(nt => ({
      key: nt.key,
      label: nt.label,
      color: nt.color,
      nodes: props.nodes
        .filter(n => n.type === nt.key)
        .map(n => {
          const all = (neighbours.value.get(n.id) ?? []).filter(m => m.code)
          return {
            id: n.id,
            code: n.code,
            label: n.label || n.code || n.id,
            status: n.status,
            href: props.hrefFor(n),
            connections: connectionCounts.value.get(n.id) || 0,
            neighbours: all.slice(0, NEIGHBOURS_SHOWN).map(m => ({ id: m.id, code: m.code, href: props.hrefFor(m) })),
            more: Math.max(0, all.length - NEIGHBOURS_SHOWN),
          }
        }),
    })),
)
</script>

<style scoped>
.graph-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-4);
  overflow-y: auto;
  flex: 1;
  min-width: 0;
}
.graph-list-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin: 0 0 var(--space-2);
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
}
.graph-list-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex: none;
}
.graph-list-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-weight: var(--weight-normal);
}
/* The rows are the product's `.data-row` (system.css): dividers between rows,
   outer rules from `--bounded`, the standard row height. Only the column
   split is local. */
.graph-list-items {
  list-style: none;
  margin: 0;
  padding: 0;
}
.graph-list-item {
  grid-template-columns: minmax(0, 1fr) auto;
  min-width: 0;
}
.graph-list-neighbours {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  font-size: var(--type-xs);
  margin-top: calc(-1 * var(--space-2));
}
.graph-list-neighbours a { text-decoration: none; }
.graph-list-neighbours a:hover { text-decoration: underline; }
.graph-list-link {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
  color: var(--color-text);
  text-decoration: none;
}
.graph-list-link:hover .graph-list-label { text-decoration: underline; }
.graph-list-link--inert { color: var(--color-text-muted); cursor: default; }
.graph-list-label { overflow-wrap: anywhere; }
.graph-list-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex: none;
  font-size: var(--type-xs);
  white-space: nowrap;
}
@media (max-width: 480px) {
  .graph-list-item { grid-template-columns: minmax(0, 1fr); gap: var(--space-1); }
}
</style>
