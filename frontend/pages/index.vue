<template>
  <div>
    <div class="page-header">
      <div class="page-header-row">
        <h1 class="page-title">Research Projects</h1>
        <div class="page-header-actions">
          <button class="btn btn-sm" @click="triggerImport" :disabled="importing">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="12" x2="12" y2="18"/><polyline points="9 15 12 12 15 15"/></svg>
            {{ importing ? 'Importing...' : 'Import JSON' }}
          </button>
          <input ref="fileInput" type="file" accept=".json" style="display:none" @change="handleImportFile" />
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <select v-model="statusFilter">
        <option value="">All statuses</option>
        <option value="active">Active</option>
        <option value="completed">Completed</option>
        <option value="archived">Archived</option>
      </select>
      <span v-if="tagFilter" class="active-tag-filter">
        Tag: <strong>{{ tagFilter }}</strong>
        <button class="tag-clear" @click="tagFilter = ''">&times;</button>
      </span>
    </div>

    <!-- Onboarding -->
    <GettingStartedBanner :has-researches="researches.length > 0" />

    <!-- Loading -->
    <div v-if="pending" class="skeleton-list">
      <div v-for="i in 4" :key="i" class="skeleton-card"></div>
    </div>

    <!-- List -->
    <div v-else-if="filtered.length" class="grid grid-2">
      <ResearchCard
        v-for="r in filtered"
        :key="r.id"
        :research="r"
        @tag-click="tagFilter = $event"
        @status-changed="refreshList"
      />
    </div>

    <!-- Empty -->
    <EmptyState
      v-else
      icon="&#x1F52C;"
      title="No research projects yet"
      description="Type this into Claude to start your first research session:"
      command="Use the research/initialize prompt to create a new research project"
    />
  </div>
</template>

<script setup lang="ts">
const { authFetch } = useAuth()
const config = useRuntimeConfig()
const base = config.public.apiBase || ''

const statusFilter = ref('active')
const tagFilter = ref('')
const importing = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

function triggerImport() {
  fileInput.value?.click()
}

async function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  importing.value = true
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    const result = await authFetch<{ research_id: string; code: string; name: string }>(
      `${base}/api/researches/import`,
      { method: 'POST', body: data }
    )
    await navigateTo(`/research/${result.code}`)
  } catch (e: any) {
    alert('Import failed: ' + (e.message || e))
  } finally {
    importing.value = false
    input.value = '' // reset file input
  }
}

const apiUrl = computed(() =>
  statusFilter.value ? `/api/researches?status=${statusFilter.value}` : '/api/researches'
)

const { data, pending, refresh } = useApi<{ data: any[] }>(apiUrl.value)

async function refreshList() {
  const res = await authFetch<{ data: any[] }>(`${base}${apiUrl.value}`)
  data.value = res
}

watch(statusFilter, refreshList)

const researches = computed(() => data.value?.data ?? [])

const filtered = computed(() =>
  tagFilter.value
    ? researches.value.filter((r: any) => r.tags?.includes(tagFilter.value))
    : researches.value
)

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.entity === 'research') {
    const res = await authFetch<{ data: any[] }>(`${base}${apiUrl.value}`)
    data.value = res
  }
})
</script>

<style scoped>
.page-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-4);
}
.page-header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}
.skeleton-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: var(--space-6);
}
.skeleton-card {
  height: 120px;
  border-radius: var(--radius);
}
.active-tag-filter {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  background: var(--color-primary-muted);
  border: 1px solid rgba(108, 197, 224, 0.15);
  border-radius: var(--radius-sm);
  font-size: var(--type-xs);
  color: var(--color-primary);
}
.tag-clear {
  background: none;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  font-size: var(--type-sm);
  padding: 0;
  line-height: 1;
  opacity: 0.7;
  transition: opacity var(--transition-fast);
}
.tag-clear:hover { opacity: 1; }

/* Responsive */
@media (max-width: 768px) {
  .skeleton-list { grid-template-columns: 1fr; }
}
</style>
