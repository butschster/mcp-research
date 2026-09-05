<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card" style="height: 60px; margin-bottom: 2rem;"></div>
    <div class="skeleton-card" style="height: 600px;"></div>
  </div>

  <div v-else-if="exportData" class="export-page">
    <!-- Toolbar (hidden in print) -->
    <div class="export-toolbar no-print">
      <Breadcrumbs :crumbs="[
        { label: 'Projects', to: '/' },
        { label: exportData.research.name, to: `/research/${researchSlug}` },
        { label: 'Export' }
      ]" />
      <div class="toolbar-actions">
        <!-- "Take it with you" is one group; printing is a different kind of
             act, so a wider gap does the grouping rather than a divider. -->
        <div class="toolbar-group">
          <button class="btn btn-sm" @click="downloadMarkdown">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Download .md
          </button>
          <ExportSplitButton
            label="Obsidian .zip"
            busy-label="Preparing…"
            options-label="Obsidian export options"
            hint="Sections become folders, documents become notes, references become links. Opens in Obsidian or any markdown editor."
            :busy="downloading"
            :with-options="hasOptionalContent"
            :options-open="optionsOpen"
            @download="downloadVault()"
            @options="optionsOpen = true"
          >
            <template #icon>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
            </template>
          </ExportSplitButton>
        </div>
        <button class="btn btn-sm" @click="printPage">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/><rect x="6" y="14" width="12" height="8"/></svg>
          Print / PDF
        </button>
      </div>
    </div>

    <ResearchObsidianExportDialog
      :visible="optionsOpen"
      :counts="optionCounts"
      @close="optionsOpen = false"
      @download="onDialogDownload"
    />

    <ResearchExportDocument :data="exportData" :research-slug="researchSlug" />
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Project not found" />
</template>

<script setup lang="ts">
import type { ObsidianExportOptions } from '~/composables/useObsidianExport'

const route = useRoute()
const id = route.params.id as string

const { data: raw, pending } = await useApi<any>(`/api/researches/${id}/export`)
const exportData = computed(() => raw.value)
const researchSlug = computed(() => exportData.value?.research?.code || id)

function downloadMarkdown() {
  const md = exportData.value?.markdown
  if (!md) return
  const research = exportData.value.research
  // Same saver and the same filename rules as the vault: a colon in a research
  // name produced a file Windows refuses.
  saveBlob(
    new Blob([md], { type: 'text/markdown' }),
    vaultFallbackName(research.code || '', research.name || '').replace(/\.zip$/, '.md'),
  )
}

function printPage() {
  window.print()
}

// --- Obsidian vault ---

const {
  pending: downloading,
  slow: downloadSlow,
  error: downloadError,
  filename: downloadedName,
  start: startDownload,
} = useDownload()
const toasts = useToasts()
const optionsOpen = ref(false)
const lastOptions = ref<ObsidianExportOptions>(defaultObsidianOptions())

const optionCounts = computed(() => ({
  sessions: exportData.value?.sessions?.length ?? 0,
  tasks: exportData.value?.tasks?.length ?? 0,
  roadmaps: exportData.value?.roadmap_count ?? 0,
}))

// With nothing optional in the research, the only live toggles would be "HTML"
// and "Revisions" — not worth an affordance, so the caret disappears and the
// plain button remains.
const hasOptionalContent = computed(() => {
  const c = optionCounts.value
  return c.sessions > 0 || c.tasks > 0 || c.roadmaps > 0
})

async function downloadVault(opts: ObsidianExportOptions = defaultObsidianOptions()) {
  lastOptions.value = opts
  const research = exportData.value?.research
  const ok = await startDownload(
    obsidianExportPath(researchSlug.value, opts),
    vaultFallbackName(research?.code || '', research?.name || ''),
  )
  if (ok) {
    toasts.success(downloadedName.value || 'The archive was saved.', 'Downloaded')
    return
  }
  const err = downloadError.value
  if (!err) return
  toasts.error(err.message, 'Export failed', retryAction(err.status))
}

// Each failure carries the one act that can answer it.
function retryAction(status: number) {
  if (status === 404) return { label: 'Reload the page', onClick: () => window.location.reload() }
  if (status === 401 || status === 403) return { label: 'Sign in', onClick: () => navigateTo('/login') }
  return { label: 'Try again', onClick: () => { downloadVault(lastOptions.value) } }
}

// The button says "Preparing…" on its own; this says why it is taking so long,
// and only when it is.
watch(downloadSlow, (slow) => {
  if (slow) toasts.push({ message: 'Still working — a large project takes a moment.' })
})

function onDialogDownload(opts: ObsidianExportOptions) {
  // The dialog closes first so there is one place showing download state, and
  // focus lands beside the line that will speak.
  optionsOpen.value = false
  downloadVault(opts)
}

</script>

<style scoped>
/* Toolbar */
.export-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}
/* The wide gap between the group and Print is the only grouping mechanism —
   no segmented control, no divider, both of which would be a new idiom in a
   product that has none. */
.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-5);
  justify-content: flex-end;
}
.toolbar-group { display: flex; gap: var(--space-2); }

@media (max-width: 768px) {
  .export-toolbar { flex-direction: column; align-items: flex-start; gap: var(--space-3); }
  /* Once the row wraps, a 1.25rem gap no longer reads as grouping — the wrap
     does it instead. */
  .toolbar-actions { gap: var(--space-2); justify-content: flex-start; }
}

@media print {
  .no-print { display: none !important; }
}
</style>
