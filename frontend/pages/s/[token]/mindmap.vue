<template>
  <div class="share-mindmap-page">
    <div class="view-toolbar">
      <div class="toolbar-left">
        <NuxtLink :to="researchPath(slug)" class="btn btn-sm toolbar-back">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m15 18-6-6 6-6"/></svg>
          Back to project
        </NuxtLink>
        <h1 class="toolbar-title">Mind map</h1>
        <span v-if="researchCode" class="toolbar-code">{{ researchCode }}</span>
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

    <p v-if="largeNotice" class="card-meta large-notice" role="status">
      Large project — showing documents. Add sessions and tasks from the toolbar.
      <button type="button" class="link-btn" @click="largeNotice = false">Dismiss</button>
    </p>

    <div class="view-panel">
      <!-- The skeleton is for the first load only. A refresh swaps the nodes
           under a mounted canvas, which keeps the visitor's pan and zoom; an
           unmount would have refitted the whole map on every event. -->
      <div v-if="!ready" class="view-state">
        <div class="skeleton-card" style="width: 200px; height: 80px;"></div>
        <p class="card-meta mt-4">Loading mind map…</p>
      </div>

      <div v-else-if="error && !nodes.length" class="view-state">
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

      <!-- The reader's own doing, as distinct from an empty project. -->
      <div v-else-if="nothingVisible" class="view-state">
        <EmptyState
          title="Nothing matches these filters"
          description="Every group is switched off. Turn one back on in the toolbar."
        >
          <button class="btn btn-primary" @click="showEverything">Show everything</button>
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
const nothingVisible = computed(() => filterGroups.value.every(g => !visibleGroups.value.has(g.key)))
function showEverything() {
  visibleGroups.value = new Set(filterGroups.value.map(g => g.key))
}
/** The first fetch has answered. */
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

async function load() {
  await refresh()
  ready.value = true
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
.toolbar-left { display: flex; align-items: center; gap: var(--space-2); min-width: 0; flex-wrap: wrap; }
.toolbar-title { margin: 0; font-size: var(--type-sm); font-weight: var(--weight-semibold); overflow-wrap: anywhere; }
.toolbar-code { font-family: 'JetBrains Mono', monospace; font-size: var(--type-xs); color: var(--color-text-muted); }
.large-notice { margin: 0 0 var(--space-3); display: flex; gap: var(--space-3); flex-wrap: wrap; }


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
  /* The panel clips; on a landscape phone the action under the copy would
     otherwise be cut off with nothing to scroll. */
  overflow: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
}

@media (max-width: 768px) {
  .view-panel { height: 60vh; }
  .toolbar-title { display: none; }
}
</style>
