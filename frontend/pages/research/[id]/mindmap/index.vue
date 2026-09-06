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
      <MindmapToolbar
        :groups="filterGroups"
        :visible-groups="visibleGroups"
        :show-crossrefs="showCrossrefs"
        :layout-direction="layoutDirection"
        @toggle-group="toggleGroup"
        @toggle-crossrefs="toggleCrossrefs"
        @set-direction="setLayoutDirection"
        @expand-all="expandAll"
        @collapse-all="collapseAll"
        @fit="fitAll"
      />
    </div>

    <!-- Loading: first load only. Later refreshes swap the nodes under the
         mounted canvas so the reader keeps their pan and zoom. -->
    <div v-if="!ready" class="mindmap-loading">
      <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
      <p class="card-meta mt-4">Loading mindmap...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error && !nodes.length" class="mindmap-loading">
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
const ready = ref(false)

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
onMounted(async () => {
  await refresh()
  ready.value = true
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
  .toolbar-title { display: none; }
}
</style>
