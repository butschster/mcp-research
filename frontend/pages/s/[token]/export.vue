<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card" style="height: 60px; margin-bottom: 2rem;"></div>
    <div class="skeleton-card" style="height: 600px;"></div>
  </div>

  <EmptyState
    v-else-if="excluded"
    title="Not part of this link"
    description="The person who shared this project didn't include downloading. Ask them if you need it."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to project</NuxtLink>
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
        <!-- "Take it with you" is one group; printing is a different kind of
             act, so a wider gap does the grouping rather than a divider. The
             same arrangement as the owner's toolbar, because it is the same
             three acts. -->
        <div class="toolbar-group">
          <button class="btn btn-sm" @click="downloadMarkdown">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Download .md
          </button>
          <!-- No options caret: the link already decided what the vault
               contains, so the only thing a dialog could offer is taking less
               than was shared. -->
          <button class="btn btn-sm" :disabled="downloading" @click="downloadVault">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
            {{ downloading ? 'Preparing…' : 'Obsidian .zip' }}
          </button>
        </div>
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
 * Markdown, the vault, and print — the same three the owner has, on a link that
 * includes downloading. The vault was refused here for a while on the grounds
 * that it builds its own payload rather than reusing this page's; it builds it
 * from `ResearchService.Get`, which is where the redaction lives, so what was
 * actually missing was narrowing its *options* to the link's include flags.
 * `service.clampForShare` does that on the server, which is the only place it
 * can be trusted — this page's request is a request, not the decision.
 *
 * The portable (.json) dump stays off: it is a re-importable copy of the
 * record rather than a reading of it, and it is not mounted under the share
 * prefix at all.
 */
const { shareFetch, researchId, researchCode, include, slug, token, unlock } = useShare()

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

// --- Obsidian vault ---

const {
  pending: downloading,
  slow: downloadSlow,
  error: downloadError,
  filename: downloadedName,
  start: startDownload,
} = useDownload()
const toasts = useToasts()

async function downloadVault() {
  const research = exportData.value?.research
  const ok = await startDownload(
    sharedObsidianExportPath(token.value, researchId.value, {
      // Asked for exactly what the link says it carries. The server narrows it
      // again regardless — a visitor can edit a URL — but sending more than the
      // link publishes would make the request a lie about its own intent.
      sessions: include.value.sessions,
      tasks: include.value.tasks,
      roadmaps: include.value.roadmaps,
      html: true,
      revisions: false,
    }),
    vaultFallbackName(research?.code || '', research?.name || ''),
    // A visitor's request, even when the person making it happens to be signed
    // in — an owner checking their own link must see what they published.
    { anonymous: true, headers: unlock.value ? { 'X-Share-Unlock': unlock.value } : {} },
  )
  if (ok) {
    toasts.success(downloadedName.value || 'The archive was saved.', 'Downloaded')
    return
  }
  const err = downloadError.value
  if (!err) return
  toasts.error(err.message, 'Export failed', {
    label: 'Try again',
    onClick: () => { void downloadVault() },
  })
}

// The button says "Preparing…" on its own; this says why it is taking so long,
// and only when it is.
watch(downloadSlow, (slow) => {
  if (slow) toasts.push({ message: 'Still working — a large project takes a moment.' })
})

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
/* The wide gap between the group and Print is the only grouping mechanism —
   the same one the owner's toolbar uses, and for the same reason: no segmented
   control, no divider, both of which would be a new idiom in a product with
   none. */
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
