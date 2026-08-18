<template>
  <div :class="['rm-stage-col', { 'is-leftover': leftover }]">
    <div class="rm-stage-head">
      <span class="rm-stage-name" :title="name">{{ name }}</span>
      <span class="rm-stage-count">{{ nodes.length }}</span>
    </div>
    <div class="rm-stage-body">
      <p v-if="nodes.length === 0" class="rm-stage-empty">{{ emptyText }}</p>
      <button
        v-for="n in nodes"
        :key="n.id"
        type="button"
        class="rm-card-btn"
        @click="emit('node-click', n.id)"
        @mouseenter="emit('hover', n.id)"
        @mouseleave="emit('hover', null)"
        @focus="emit('hover', n.id)"
        @blur="emit('hover', null)"
      >
        <RoadmapNodeCard
          :data="nodeCardData(n)"
          :deps="deps.get(n.id)"
          :highlighted="highlightIds?.has(n.id)"
          :dimmed="!!highlightIds && !highlightIds.has(n.id)"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nodeCardData, type RawRoadmapNode } from '~/utils/roadmap'

const props = defineProps<{
  name: string
  nodes: readonly RawRoadmapNode[]
  deps: Map<string, { code: string }[]>
  leftover?: boolean
  /** When set, ids in it are highlighted and everything else is dimmed. */
  highlightIds?: Set<string> | null
}>()

const emit = defineEmits<{
  'node-click': [string]
  hover: [string | null]
}>()

const emptyText = computed(() =>
  props.leftover ? 'No nodes here' : 'Nothing here yet',
)
</script>

<style scoped>
.rm-stage-col {
  flex: 0 0 300px;
  width: 300px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  scroll-snap-align: start;
}
.rm-stage-head {
  position: sticky;
  top: 0;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  border-radius: var(--radius) var(--radius) 0 0;
  z-index: 1;
}
.is-leftover .rm-stage-head {
  background: var(--color-surface-hover);
}
.rm-stage-name {
  font-size: var(--type-xs);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.rm-stage-count {
  margin-left: auto;
  flex-shrink: 0;
  font-size: var(--type-xs);
  font-variant-numeric: tabular-nums;
  color: var(--color-text-muted);
  background: var(--color-surface-hover);
  border-radius: var(--radius-xs);
  padding: 0.05rem 0.4rem;
}
.is-leftover .rm-stage-count {
  background: var(--color-surface);
}
.rm-stage-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.rm-stage-empty {
  font-size: var(--type-xs);
  color: var(--color-text-faint);
  text-align: center;
  padding: var(--space-4) 0;
}
.rm-card-btn {
  appearance: none;
  background: none;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  display: block;
  width: 100%;
  border-radius: var(--radius);
}
.rm-card-btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
