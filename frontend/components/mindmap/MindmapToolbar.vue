<template>
  <div class="mindmap-toolbar-controls">
    <!-- One chip per group the caller offers. A share page offers fewer: a
         chip for a part the link excludes would advertise it. -->
    <button
      v-for="group in groups"
      :key="group.key"
      type="button"
      :class="['btn btn-sm filter-chip', { active: visibleGroups.has(group.key) }]"
      :aria-pressed="visibleGroups.has(group.key)"
      @click="emit('toggle-group', group.key)"
    >{{ group.label }}</button>

    <button
      type="button"
      :class="['btn btn-sm filter-chip crossref-chip', { active: showCrossrefs }]"
      :aria-pressed="showCrossrefs"
      @click="emit('toggle-crossrefs')"
    >
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
      Crossrefs
    </button>

    <span class="toolbar-sep" aria-hidden="true"></span>

    <button
      type="button"
      :class="['btn btn-sm', { active: layoutDirection === 'LR' }]"
      :aria-pressed="layoutDirection === 'LR'"
      title="Left to right"
      aria-label="Left to right"
      @click="emit('set-direction', 'LR')"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
    </button>
    <button
      type="button"
      :class="['btn btn-sm', { active: layoutDirection === 'TB' }]"
      :aria-pressed="layoutDirection === 'TB'"
      title="Top to bottom"
      aria-label="Top to bottom"
      @click="emit('set-direction', 'TB')"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg>
    </button>

    <span class="toolbar-sep" aria-hidden="true"></span>

    <button type="button" class="btn btn-sm" @click="emit('expand-all')">Expand all</button>
    <button type="button" class="btn btn-sm" @click="emit('collapse-all')">Collapse</button>
    <button type="button" class="btn btn-sm" title="Fit view" aria-label="Fit view" @click="emit('fit')">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
    </button>
  </div>
</template>

<script setup lang="ts">
/**
 * The mind map's controls: group chips, the crossref chip, layout direction,
 * expand/collapse and fit.
 *
 * One component for the owner's page and the shared page. The two carried
 * verbatim copies of this markup and had already drifted — one had
 * `aria-pressed` on the chips and the other did not — which is the whole
 * argument for a component that owns the ARIA once.
 */
defineProps<{
  groups: { key: string; label: string }[]
  visibleGroups: Set<string>
  showCrossrefs: boolean
  layoutDirection: 'LR' | 'TB'
}>()

const emit = defineEmits<{
  'toggle-group': [key: string]
  'toggle-crossrefs': []
  'set-direction': [direction: 'LR' | 'TB']
  'expand-all': []
  'collapse-all': []
  fit: []
}>()
</script>

<style scoped>
.mindmap-toolbar-controls {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
  min-width: 0;
}
.toolbar-sep {
  width: 1px;
  height: 20px;
  background: var(--color-border-strong);
  margin: 0 var(--space-1);
}
/* Chips keep .btn's border: a fainter one made them read as a weaker kind of
   control than the labelled buttons beside them. */
.filter-chip { color: var(--color-text-muted); }
/* The pressed layout direction was announced (aria-pressed) but not drawn. */
.btn.active {
  color: var(--color-primary);
  border-color: rgba(var(--color-primary-rgb), 0.3);
  background: var(--color-primary-muted);
}
.filter-chip.active {
  color: var(--color-primary);
  border-color: rgba(var(--color-primary-rgb), 0.3);
  background: var(--color-primary-muted);
}
.crossref-chip.active {
  color: var(--hue-5);
  border-color: rgba(var(--hue-5-rgb), 0.3);
  background: rgba(var(--hue-5-rgb), 0.1);
}
@media (max-width: 768px) {
  .mindmap-toolbar-controls { gap: var(--space-1); }
  .toolbar-sep { display: none; }
}
</style>
