<template>
  <div class="graph-list">
    <section v-for="group in groups" :key="group.key" class="graph-list-group">
      <h2 class="graph-list-title">
        <span class="graph-list-dot" :style="{ background: group.color }" aria-hidden="true"></span>
        {{ group.label }}
        <span class="graph-list-count">{{ group.nodes.length }}</span>
      </h2>
      <p v-if="!group.nodes.length" class="card-meta graph-list-empty">None in this project.</p>
      <ul v-else class="graph-list-items">
        <li v-for="n in group.nodes" :key="n.id" class="graph-list-item">
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
  /** Where a node leads, or '' for a node with no page of its own. */
  hrefFor: (node: GraphNode) => string
}>()

const connectionCounts = computed(() => {
  const counts = new Map<string, number>()
  for (const e of props.edges) {
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
        .map(n => ({
          id: n.id,
          code: n.code,
          label: n.label || n.code || n.id,
          status: n.status,
          href: props.hrefFor(n),
          connections: connectionCounts.value.get(n.id) || 0,
        })),
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
.graph-list-empty { margin: 0; font-size: var(--type-xs); }
.graph-list-items {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
}
.graph-list-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-2) 0;
  border-top: 1px solid var(--color-border);
  min-width: 0;
}
.graph-list-link {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
  color: var(--color-text);
  text-decoration: none;
  /* A comfortable target for a phone, which is where this view is the default. */
  min-height: var(--control-h-sm);
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
  .graph-list-item { flex-direction: column; align-items: flex-start; gap: var(--space-1); }
}
</style>
