<template>
  <div>
    <PageHeader title="Research Projects">
      <template #actions>
        <button class="btn btn-sm" @click="triggerImport" :disabled="importing">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="12" x2="12" y2="18"/><polyline points="9 15 12 12 15 15"/></svg>
            {{ importing ? 'Importing...' : 'Import JSON' }}
          </button>
          <input ref="fileInput" type="file" accept=".json" style="display:none" @change="handleImportFile" />
      </template>
    </PageHeader>

    <!-- Filters -->
    <div class="filter-bar">
      <select v-model="statusFilter" class="select">
        <option value="">All statuses</option>
        <option value="active">Active</option>
        <option value="completed">Completed</option>
        <option value="archived">Archived</option>
      </select>
      <!-- Only when there is a choice to make. A solo user never sees this. -->
      <select v-if="teams.length > 1" v-model="teamFilter" :disabled="teamsLoading" class="select" aria-label="Filter by team">
        <option value="">All teams</option>
        <option v-for="team in teams" :key="team.id" :value="team.id">{{ team.name }}</option>
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

    <!-- Empty because of the team filter, not because there is nothing -->
    <EmptyState
      v-else-if="teamFilter"
      icon="&#x1F4C1;"
      title="No researches in this team yet"
      description="Researches created in this team will appear here. An agent connected to your account can create one."
    >
      <button class="btn btn-sm" @click="teamFilter = ''">Show all teams</button>
    </EmptyState>

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

const route = useRoute()
const router = useRouter()
const { teams, loading: teamsLoading, load: loadTeams } = useTeams()

const statusFilter = ref('active')
const tagFilter = ref('')

// The team filter lives in the URL rather than in global state, so a scoped
// list is a link someone can send — which is the only real advantage a
// workspace switcher would have had.
const teamFilter = computed({
  get: () => (typeof route.query.team === 'string' ? route.query.team : ''),
  set: (value: string) => {
    const query = { ...route.query }
    if (value) query.team = value
    else delete query.team
    router.replace({ query })
  },
})

onMounted(() => loadTeams())
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
    useToasts().push({ variant: 'error', title: 'Import failed', message: e?.message || String(e), timeout: 0 })
  } finally {
    importing.value = false
    input.value = '' // reset file input
  }
}

const apiUrl = computed(() => {
  const params = new URLSearchParams()
  if (statusFilter.value) params.set('status', statusFilter.value)
  if (teamFilter.value) params.set('team', teamFilter.value)
  const query = params.toString()
  return query ? `/api/researches?${query}` : '/api/researches'
})

const { data, pending, refresh } = useApi<{ data: any[] }>(apiUrl.value)

async function refreshList() {
  const res = await authFetch<{ data: any[] }>(`${base}${apiUrl.value}`)
  data.value = res
}

watch([statusFilter, teamFilter], refreshList)

const researches = computed(() => data.value?.data ?? [])

const filtered = computed(() =>
  tagFilter.value
    ? researches.value.filter((r: any) => r.tags?.includes(tagFilter.value))
    : researches.value
)

// Real-time updates
async function reloadResearches() {
  data.value = await authFetch<{ data: any[] }>(`${base}${apiUrl.value}`)
}

useRealtimeUpdates(
  (event) => {
    // A transfer can take a research out of this list as easily as a create
    // puts one in, and access.revoked says a card here is no longer real.
    if (event.entity === 'research' || event.type === 'access.revoked') reloadResearches()
  },
  { onResync: reloadResearches },
)
</script>

<style scoped>
/* Moved out of the global stylesheet: these style this component and
   nothing else, and they were a directory away from the markup they
   describe. What stays global is what three unrelated components share. */
.filter-bar {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-6);
}
  @media (max-width: 768px) {
  .filter-bar { flex-wrap: wrap; }
}
.skeleton-list {
  /* This page's skeletons stand in for cards in a grid, not for rows. */
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
