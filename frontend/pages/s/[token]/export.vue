<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card" style="height: 60px; margin-bottom: 2rem;"></div>
    <div class="skeleton-card" style="height: 600px;"></div>
  </div>

  <EmptyState
    v-else-if="excluded"
    title="Not part of this link"
    description="The person who shared this research didn't include downloading. Ask them if you need it."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to the research</NuxtLink>
  </EmptyState>

  <EmptyState
    v-else-if="!exportData"
    title="Couldn't build the document"
    description="The server didn't answer. Try again in a moment."
  >
    <button class="btn btn-primary" @click="load()">Try again</button>
  </EmptyState>

  <div v-else class="export-page">
    <div class="export-toolbar no-print">
      <Breadcrumbs :crumbs="[
        { label: exportData.research.name, to: researchPath(slug) },
        { label: 'Export' },
      ]" />
      <div class="toolbar-actions">
        <button class="btn btn-sm" @click="downloadMarkdown">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          Download .md
        </button>
        <button class="btn btn-sm" @click="printPage">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/><rect x="6" y="14" width="12" height="8"/></svg>
          Print / PDF
        </button>
      </div>
    </div>

    <!-- The same document the owner prints, from the same component. The
         payload it is handed is the redacted one: the server left out
         `instruction`, `memory`, and any part the link does not include. -->
    <ResearchExportDocument :data="exportData" :research-slug="slug" />
  </div>
</template>

<script setup lang="ts">
/**
 * A shared research, as a document.
 *
 * The vault (.zip) and portable (.json) formats are not offered: both build
 * their payload from the repository rather than from the redacted read, so
 * neither is safe to hand to a visitor and neither is mounted under the share
 * prefix.
 */
const { shareFetch, researchId, researchCode, include, slug } = useShare()

const exportData = ref<any | null>(null)
const pending = ref(true)

const excluded = computed(() => !include.value.export)

async function load() {
  if (excluded.value) {
    exportData.value = null
    pending.value = false
    return
  }
  try {
    exportData.value = await shareFetch<any>(`/researches/${researchId.value}/export`)
  } catch {
    exportData.value = null
  } finally {
    pending.value = false
  }
}

void load()

function downloadMarkdown() {
  const md = exportData.value?.markdown
  if (!md) return
  const research = exportData.value.research
  saveBlob(
    new Blob([md], { type: 'text/markdown' }),
    vaultFallbackName(research.code || '', research.name || '').replace(/\.zip$/, '.md'),
  )
}

function printPage() {
  window.print()
}

useResearchRealtime(
  () => slug.value,
  () => void load(),
  { onResync: () => void load(), researchId: () => researchId.value },
)
</script>

<style scoped>
.export-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}
.toolbar-actions { display: flex; flex-wrap: wrap; gap: var(--space-2); justify-content: flex-end; }

@media (max-width: 768px) {
  .export-toolbar { flex-direction: column; align-items: flex-start; gap: var(--space-3); }
  .toolbar-actions { justify-content: flex-start; }
}
@media print {
  .no-print { display: none !important; }
}
</style>
