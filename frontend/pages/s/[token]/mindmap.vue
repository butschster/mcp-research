<template>
  <div class="share-mindmap-page">
    <div class="view-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="researchPath(slug)" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back to project
        </NuxtLink>
        <span class="toolbar-title">Mind map</span>
        <span v-if="researchCode" class="toolbar-code">{{ researchCode }}</span>
      </div>
      <div class="toolbar-right">
        <!-- A chip for a part the link excludes is not rendered at all — not
             disabled. A greyed "Sessions" says there are sessions being kept
             from you, which is the one thing a share must not say. -->
        <button
          v-for="group in filterGroups"
          :key="group.key"
          type="button"
          :class="['btn btn-sm filter-chip', { active: visibleGroups.has(group.key) }]"
          :aria-pressed="visibleGroups.has(group.key)"
          @click="toggleGroup(group.key)"
        >{{ group.label }}</button>

        <button
          type="button"
          :class="['btn btn-sm filter-chip crossref-chip', { active: showCrossrefs }]"
          :aria-pressed="showCrossrefs"
          @click="toggleCrossrefs"
        >
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          Crossrefs
        </button>

        <span class="toolbar-sep"></span>

        <button
          type="button"
          :class="['btn btn-sm', { active: layoutDirection === 'LR' }]"
          title="Left to right"
          aria-label="Left to right"
          @click="setLayoutDirection('LR')"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
        </button>
        <button
          type="button"
          :class="['btn btn-sm', { active: layoutDirection === 'TB' }]"
          title="Top to bottom"
          aria-label="Top to bottom"
          @click="setLayoutDirection('TB')"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="m19 12-7 7-7-7"/></svg>
        </button>

        <span class="toolbar-sep"></span>

        <button type="button" class="btn btn-sm" @click="expandAll">Expand all</button>
        <button type="button" class="btn btn-sm" @click="collapseAll">Collapse</button>
        <button type="button" class="btn btn-sm" title="Fit view" aria-label="Fit view" @click="fitAll">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        </button>
      </div>
    </div>

    <p v-if="largeNotice" class="card-meta large-notice" role="status">
      Large project — showing documents. Add sessions and tasks from the toolbar.
      <button type="button" class="link-btn" @click="largeNotice = false">Dismiss</button>
    </p>

    <div class="view-panel">
      <div v-if="loading" class="view-state">
        <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
        <p class="card-meta mt-4">Loading mind map…</p>
      </div>

      <div v-else-if="error" class="view-state">
        <EmptyState
          title="Couldn't build the mind map"
          description="Part of this project didn't load. The link is fine — try again."
        >
          <button class="btn btn-primary" @click="refresh()">Try again</button>
        </EmptyState>
      </div>

      <div v-else-if="itemCount === 0" class="view-state">
        <EmptyState
          title="Nothing to map yet"
          description="This project has no documents, sessions or tasks to lay out. It fills in as the project grows — you can leave this page open."
        >
          <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to project</NuxtLink>
        </EmptyState>
      </div>

      <!-- Dragging is off: a canvas a stranger cannot save into has no reason
           to let them rearrange it, and a phone gets pan instead of drag. -->
      <MindmapCanvas
        v-else
        ref="canvas"
        :nodes="nodes"
        :edges="edges"
        :draggable="false"
        @node-click="onNodeClick"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * The shared mind map.
 *
 * Built from the same five reads the owner's page makes, through `shareFetch`
 * so the owner opening their own link sees what a visitor sees. The parts a
 * link excludes are not requested — `parts` says which — and the composable
 * treats every part but the research itself as optional, because a link that
 * excludes tasks answering 404 to the tasks list is the link working.
 */
const { shareFetch, researchId, researchCode, include, slug, markGone } = useShare()

const {
  nodes, edges, loading, error, errorStatus, itemCount,
  refresh, toggleCollapse, expandAll, collapseAll,
  layoutDirection, setLayoutDirection,
  visibleGroups, toggleGroup, showCrossrefs, toggleCrossrefs,
} = useResearchMindmap(researchId.value, {
  fetcher: (path) => shareFetch(path),
  parts: { sessions: include.value.sessions, tasks: include.value.tasks },
})

// A 404 on the research itself means the link has gone; the shell owns that.
watch(errorStatus, (status) => { if (status === 404) markGone() })

const filterGroups = computed(() => [
  { key: 'entries', label: 'Documents' },
  ...(include.value.sessions ? [{ key: 'questions', label: 'Sessions' }] : []),
  ...(include.value.tasks ? [{ key: 'tasks', label: 'Tasks' }] : []),
])

const canvas = ref<{ fitAll: () => void } | null>(null)
const largeNotice = ref(false)

function fitAll() {
  canvas.value?.fitAll()
}

function onNodeClick({ node }: { node: any }) {
  if (node.type === 'section' || node.type === 'group-label') {
    toggleCollapse(node.id)
    nextTick(() => canvas.value?.fitAll())
  }
}

async function load() {
  await refresh()
  // Every question of every session is its own node; past this the layout is
  // a wall. Start with documents and say so.
  if (nodes.value.length > 250) {
    visibleGroups.value = new Set(['entries'])
    largeNotice.value = true
  }
}

onMounted(() => { void load() })

// No refit on an event: an agent writing during a demo would yank the
// visitor's zoom every few seconds.
useResearchRealtime(
  () => slug.value,
  () => void refresh(),
  { onResync: () => void refresh(), researchId: () => researchId.value },
)

useHead({ title: () => (researchCode.value ? `${researchCode.value} — mind map` : 'Mind map') })
</script>

<style scoped>
.share-mindmap-page { display: flex; flex-direction: column; }
.view-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  margin-bottom: var(--space-4);
}
.toolbar-left, .toolbar-right { display: flex; align-items: center; gap: var(--space-2); min-width: 0; flex-wrap: wrap; }
.toolbar-title { font-size: var(--type-sm); font-weight: var(--weight-semibold); overflow-wrap: anywhere; }
.toolbar-code { font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: var(--color-text-muted); }
.toolbar-sep {
  width: 1px;
  height: 20px;
  background: var(--color-border-strong);
  margin: 0 var(--space-1);
}
.large-notice { margin: 0 0 var(--space-3); display: flex; gap: var(--space-3); flex-wrap: wrap; }

/* The owner's chip treatment, copied rather than reinvented. */
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

.view-panel {
  display: flex;
  height: 70vh;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--color-bg);
}
.view-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
}

@media (max-width: 768px) {
  .view-panel { height: 60vh; }
  .toolbar-title { display: none; }
  .toolbar-sep { display: none; }
}
</style>
