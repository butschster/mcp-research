<template>
  <div class="mindmap-page">
    <!-- Toolbar -->
    <div class="mindmap-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="researchPath(researchSlug)" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back
        </NuxtLink>
        <span v-if="researchName" class="toolbar-title">{{ researchName }}</span>
      </div>
      <div class="toolbar-right">
        <!-- Filter chips -->
        <button
          v-for="group in filterGroups"
          :key="group.key"
          :class="['btn btn-sm filter-chip', { active: visibleGroups.has(group.key) }]"
          @click="toggleGroup(group.key)"
        >{{ group.label }}</button>

        <button
          :class="['btn btn-sm filter-chip crossref-chip', { active: showCrossrefs }]"
          @click="toggleCrossrefs"
        >
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          Crossrefs
        </button>

        <span class="toolbar-sep"></span>

        <!-- Layout toggle -->
        <button
          :class="['btn btn-sm', { active: layoutDirection === 'LR' }]"
          @click="setLayoutDirection('LR')"
          title="Left to right"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
        </button>
        <button
          :class="['btn btn-sm', { active: layoutDirection === 'TB' }]"
          @click="setLayoutDirection('TB')"
          title="Top to bottom"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg>
        </button>

        <span class="toolbar-sep"></span>

        <!-- Expand/collapse -->
        <button class="btn btn-sm" @click="expandAll">Expand all</button>
        <button class="btn btn-sm" @click="collapseAll">Collapse</button>

        <!-- Fit view -->
        <button class="btn btn-sm" @click="fitAll" title="Fit view">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="mindmap-loading">
      <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
      <p class="card-meta mt-4">Loading mindmap...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="mindmap-loading">
      <p class="card-meta">{{ error }}</p>
      <button class="btn btn-sm mt-4" @click="refresh">Retry</button>
    </div>

    <!-- Canvas -->
    <div v-else class="mindmap-canvas">
      <MindmapCanvas ref="canvas" :nodes="nodes" :edges="edges" @node-click="onNodeClick" />
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string

const {
  nodes,
  edges,
  loading,
  error,
  refresh,
  toggleCollapse,
  expandAll,
  collapseAll,
  layoutDirection,
  setLayoutDirection,
  visibleGroups,
  toggleGroup,
  showCrossrefs,
  toggleCrossrefs,
} = useResearchMindmap(id)

const filterGroups = [
  { key: 'entries', label: 'Documents' },
  { key: 'questions', label: 'Sessions' },
  { key: 'tasks', label: 'Tasks' },
]

// Research name for toolbar
const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? '')
const researchSlug = computed(() => researchData.value?.data?.research?.code || id)

const canvas = ref<{ fitAll: () => void } | null>(null)

function fitAll() {
  canvas.value?.fitAll()
}

function onNodeClick({ node }: { node: any }) {
  if (node.type === 'section' || node.type === 'group-label') {
    toggleCollapse(node.id)
    nextTick(() => canvas.value?.fitAll())
  }
}

// Initial load
onMounted(() => {
  refresh()
})

// Real-time updates
//
// The refit is deliberately not repeated. Fitting the whole map on every event
// discards the reader's pan and zoom, and an agent at work emits one every few
// seconds — so the page fought whoever was reading it, precisely when it was
// most worth reading.
async function reloadMindmap() {
  await refresh()
}

useResearchRealtime(() => id, reloadMindmap, {
  researchId: () => researchData.value?.data?.research?.id,
  onResync: reloadMindmap,
})
</script>

<style scoped>
.mindmap-page {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
  z-index: var(--z-overlay);
}

.mindmap-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-5);
  background: var(--color-surface-raised);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--color-border);
  gap: var(--space-4);
  flex-shrink: 0;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}
.toolbar-back { gap: var(--space-1); }
.toolbar-title {
  font-size: var(--type-sm);
  font-weight: var(--weight-semibold);
  color: var(--color-text);
  letter-spacing: -0.01em;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.toolbar-sep {
  width: 1px;
  height: 20px;
  background: var(--color-border-strong);
  margin: 0 var(--space-1);
}

.filter-chip {
  color: var(--color-text-muted);
  border-color: var(--color-border);
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

.mindmap-canvas {
  flex: 1;
  min-height: 0;
}

.mindmap-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* Responsive */
@media (max-width: 768px) {
  .mindmap-toolbar {
    flex-wrap: wrap;
    padding: var(--space-2) var(--space-3);
    gap: var(--space-2);
  }
  .toolbar-right {
    flex-wrap: wrap;
    gap: var(--space-1);
  }
  .toolbar-title { display: none; }
  .toolbar-sep { display: none; }
}
</style>
