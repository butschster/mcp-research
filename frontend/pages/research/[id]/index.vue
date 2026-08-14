<template>
  <div v-if="pending" class="skeleton-page">
    <div class="skeleton-card skeleton-header"></div>
    <div class="layout-sidebar">
      <div class="skeleton-card skeleton-sidebar"></div>
      <div>
        <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
      </div>
    </div>
  </div>

  <div v-else-if="research">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[{ label: 'Research', to: '/' }, { label: research.name }]" />
      <div class="research-header">
        <div class="title-with-code">
          <span v-if="research.code" class="short-code">{{ research.code }}</span>
          <h1 class="page-title">{{ research.name }}</h1>
        </div>
        <div class="research-actions">
          <StatusBadge :status="research.status" />
          <TeamChip v-if="showTeamChip" :name="research.team_name" />
          <TeamViewerNotice v-if="isViewer" :team-name="research.team_name" />

          <!-- Icon nav buttons -->
          <NuxtLink :to="`/research/${researchSlug}/tasks`" class="btn btn-icon" title="Tasks">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
            <span v-if="tasks.length" class="btn-count">{{ tasks.length }}</span>
          </NuxtLink>
          <NuxtLink :to="`/research/${researchSlug}/roadmaps`" class="btn btn-icon" title="Roadmaps">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
            <span v-if="roadmaps.length" class="btn-count">{{ roadmaps.length }}</span>
          </NuxtLink>
          <NuxtLink :to="`/research/${researchSlug}/mindmap`" class="btn btn-icon" title="Mind map">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><circle cx="4" cy="6" r="2"/><circle cx="20" cy="6" r="2"/><circle cx="4" cy="18" r="2"/><circle cx="20" cy="18" r="2"/><path d="M9.5 10.5 5.5 7.5"/><path d="M14.5 10.5l4-3"/><path d="M9.5 13.5 5.5 16.5"/><path d="M14.5 13.5l4 3"/></svg>
          </NuxtLink>
          <NuxtLink :to="`/research/${researchSlug}/graph`" class="btn btn-icon" title="Knowledge graph">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="6" cy="6" r="3"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="18" r="3"/><path d="M8.5 8.5 15.5 15.5"/><path d="M15.5 8.5 8.5 15.5"/><path d="M6 9v6"/><path d="M18 9v6"/></svg>
          </NuxtLink>

          <!-- More menu -->
          <ActionMenu>
            <button class="action-menu-item" @click="detailsOpen = !detailsOpen">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
              {{ detailsOpen ? 'Hide details' : 'Details' }}
            </button>
            <NuxtLink :to="`/research/${researchSlug}/export`" class="action-menu-item">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              Export
            </NuxtLink>
            <button class="action-menu-item" :disabled="exporting" @click="downloadPortableJSON()">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><polyline points="9 15 12 18 15 15"/></svg>
              {{ exporting ? 'Saving...' : 'Download JSON' }}
            </button>
            <NuxtLink
              v-if="research.team_id && authEnabled"
              :to="`/teams/${research.team_id}`"
              class="action-menu-item"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
              Members
            </NuxtLink>
            <button v-if="canAdmin && authEnabled" class="action-menu-item" @click="openTransfer()">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="17 1 21 5 17 9"/><path d="M3 11V9a4 4 0 0 1 4-4h14"/><polyline points="7 23 3 19 7 15"/><path d="M21 13v2a4 4 0 0 1-4 4H3"/></svg>
              Move to team…
            </button>
            <div v-if="canWrite" class="action-menu-divider"></div>
            <button
              v-if="canWrite"
              class="action-menu-item"
              :class="{ 'action-menu-item--danger': research.status !== 'archived' }"
              @click="toggleArchive()"
            >
              <svg v-if="research.status === 'archived'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
              {{ research.status === 'archived' ? 'Restore' : 'Archive' }}
            </button>
          </ActionMenu>
        </div>
      </div>
      <p v-if="research.goal && !detailsOpen" class="card-meta mt-2">{{ research.goal }}</p>
    </div>

    <!-- Research details panel -->
    <ResearchDetailsPanel
      :research="research"
      :open="detailsOpen"
      @save="handleDetailsSave"
      @update:open="detailsOpen = $event"
    />

    <!-- Active sessions -->
    <ResearchActiveSessionsGrid :sessions="activeSessions" :research-slug="researchSlug" />

    <!-- Closed sessions (collapsed) -->
    <ResearchPastSessionsList :sessions="closedSessions" :research-slug="researchSlug" />

    <!-- Sidebar layout: sections + entries -->
    <div class="layout-sidebar">
      <!-- Sidebar -->
      <ResearchSidebar
        :sections="sections"
        :active-section="activeSection"
        :total-entry-count="totalEntryCount"
        :links-total="researchLinksTotal"
        @update:active-section="activeSection = $event"
      />

      <!-- Main: entries -->
      <div>
        <!-- All entries view -->
        <ResearchEntriesView
          v-if="isAllEntries"
          :entries="allEntries"
          :sections="sections"
          :research-slug="researchSlug"
          :loading="allEntriesPending"
          mode="all"
          :tags="globalTags"
        />

        <!-- Single section view -->
        <ResearchEntriesView
          v-else-if="currentSection"
          :entries="entries"
          :sections="sections"
          :research-slug="researchSlug"
          :loading="entriesPending"
          mode="section"
          :section-info="currentSection"
          :tags="[]"
        />

        <!-- External links view -->
        <ResearchExternalLinksView
          v-else-if="isLinksView"
          :groups="researchLinksGrouped"
          :loading="researchLinksLoading"
        />

        <EmptyState
          v-else
          icon="&#x1F448;"
          title="Select a section"
          description="Choose a section from the sidebar to view its entries."
        />
      </div>
    </div>
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Research not found" />
    <ResearchTransferModal
      :visible="transferOpen"
      :research="{ code: research?.code, name: research?.name || '' }"
      :current-team-id="research?.team_id || ''"
      :current-team-name="research?.team_name || ''"
      :teams="[...ownedTeams]"
      :busy="transferring"
      :error="transferError"
      @transfer="transfer"
      @close="transferOpen = false"
    />
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string

// Research data
const { data: researchData, pending } = await useApi<{ data: any }>(`/api/researches/${id}`)

const research = computed(() => researchData.value?.data?.research)

// The role rides on the payload every research page already awaits, so no
// screen ever renders edit controls and then takes them away.
const { authEnabled } = useAuth()
const { canWrite, canAdmin, isViewer, setFromResearch } = useResearchRole()
watch(research, (r) => setFromResearch(r), { immediate: true })

const showTeamChip = computed(() => !!research.value?.team_name && !research.value?.team_is_personal)

// Moving a research is the only action that changes who can see it, so it gets
// a dialog that says so rather than a menu item that just does it.
const { ownedTeams, load: loadTeams } = useTeams()
const transferOpen = ref(false)
const transferring = ref(false)
const transferError = ref('')

async function openTransfer() {
  transferError.value = ''
  // Awaited: opening first showed "you own only this team" and then a picker
  // with nothing selected and Move disabled, on every deep-linked page load.
  await loadTeams()
  transferOpen.value = true
}

async function transfer(teamId: string) {
  transferring.value = true
  transferError.value = ''
  try {
    const { authFetch } = useAuth()
    const cfg = useRuntimeConfig()
    await authFetch(`${cfg.public.apiBase || ''}/api/researches/${research.value.id}/transfer`, {
      method: 'POST',
      body: { team_id: teamId },
    })
    transferOpen.value = false
    const target = ownedTeams.value.find((t) => t.id === teamId)
    useToasts().success(`Moved to ${target?.name || 'the other team'}`)
    // Permissions may have changed under the reader — refetch rather than
    // leaving the page showing what they could do a moment ago.
    await refreshNuxtData()
  } catch (e: any) {
    transferError.value = e?.data?.error || 'The server refused the move'
  } finally {
    transferring.value = false
  }
}
const researchSlug = computed(() => research.value?.code || id)
const sections = computed(() => researchData.value?.data?.sections ?? [])
const activeSession = computed(() => researchData.value?.data?.active_session)

const totalEntryCount = computed(() =>
  sections.value.reduce((sum: number, s: any) => sum + (s.entries_count || 0), 0)
)

// Active section (default: first, or '__all__' for all entries)
const activeSection = ref(route.query.section as string || '')
const isAllEntries = computed(() => activeSection.value === '__all__')
const isLinksView = computed(() => activeSection.value === '__links__')
const router = useRouter()

watch(activeSection, (val) => {
  router.replace({ query: { ...route.query, section: val || undefined } })
})

watch(sections, (secs) => {
  if (!activeSection.value && secs.length) activeSection.value = secs[0].id
}, { immediate: true })

const currentSection = computed(() =>
  isAllEntries.value ? null : sections.value.find((s: any) => s.id === activeSection.value) ?? null
)

// --- Section entries ---
const entriesUrl = computed(() =>
  !isAllEntries.value && activeSection.value
    ? `/api/researches/${id}/sections/${activeSection.value}/entries`
    : null
)
const { data: entriesData, pending: entriesPending } = useApi<{ data: any[] }>(
  computed(() => entriesUrl.value ?? '/api/researches/__none__/sections/__none__/entries')
)
const entries = computed(() =>
  !isAllEntries.value && activeSection.value ? (entriesData.value?.data ?? []) : []
)

// --- All entries ---
const { data: allEntriesData, pending: allEntriesPending } = useApi<{ data: any[] }>(
  computed(() => isAllEntries.value ? `/api/researches/${id}/entries` : '/api/researches/__none__/entries')
)
const allEntries = computed(() => isAllEntries.value ? (allEntriesData.value?.data ?? []) : [])

// Global tags from API
const { data: tagsData } = useApi<{ data: any[] }>(
  computed(() => isAllEntries.value ? `/api/researches/${id}/tags` : '/api/researches/__none__/tags')
)
const globalTags = computed(() => isAllEntries.value ? (tagsData.value?.data ?? []) : [])

// Tasks (count only, for header button)
const { data: tasksData } = await useApi<{ data: any[] }>(`/api/researches/${id}/tasks`)
const tasks = computed(() => tasksData.value?.data ?? [])

const { data: roadmapsData } = await useApi<{ data: any[] }>(`/api/researches/${id}/roadmaps`)
const roadmaps = computed(() => roadmapsData.value?.data ?? [])

// All sessions
const { data: sessionsData } = await useApi<{ data: any[] }>(`/api/researches/${id}/sessions`)
const allSessions = computed(() => sessionsData.value?.data ?? [])
const activeSessions = computed(() => allSessions.value.filter((s: any) => s.status === 'active'))
const closedSessions = computed(() => allSessions.value.filter((s: any) => s.status !== 'active'))

// External links
const { data: researchLinksData, pending: researchLinksLoading } = useApi<{ data: any[]; total: number }>(
  computed(() => isLinksView.value ? `/api/researches/${id}/links` : '/api/researches/__none__/links')
)
const researchLinksGrouped = computed(() => researchLinksData.value?.data ?? [])
const researchLinksTotal = computed(() => researchLinksData.value?.total ?? 0)

// Auth & API
const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''

// Details panel
const detailsOpen = ref(false)

async function handleDetailsSave(field: string, value: any) {
  const body: Record<string, any> = {}
  if (field === 'tags') {
    body.tags = value
  } else {
    body[field] = value
  }
  await authFetch(`${rtBase}/api/researches/${id}`, { method: 'PUT', body })
  researchData.value = await authFetch<any>(`${rtBase}/api/researches/${id}`)
}

// Download portable JSON
const exporting = ref(false)
async function downloadPortableJSON() {
  exporting.value = true
  try {
    const data = await authFetch<any>(`${rtBase}/api/researches/${id}/export/portable`)
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${research.value?.name || 'research'}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) {
    alert('Export failed: ' + (e.message || e))
  } finally {
    exporting.value = false
  }
}

// Archive toggle
async function toggleArchive() {
  const newStatus = research.value.status === 'archived' ? 'active' : 'archived'
  await authFetch(`${rtBase}/api/researches/${id}`, {
    method: 'PUT',
    body: { status: newStatus },
  })
  researchData.value = await authFetch<any>(`${rtBase}/api/researches/${id}`)
}

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return

  if (['research', 'section', 'session'].includes(event.entity)) {
    researchData.value = await authFetch<any>(`${rtBase}/api/researches/${id}`)
  }
  if (event.entity === 'entry') {
    if (entriesUrl.value) {
      entriesData.value = await authFetch<any>(`${rtBase}${entriesUrl.value}`)
    }
    if (isAllEntries.value) {
      allEntriesData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/entries`)
      tagsData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tags`)
    }
    if (isLinksView.value) {
      researchLinksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/links`)
    }
  }
  if (['question', 'session'].includes(event.entity)) {
    sessionsData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/sessions`)
  }
  if (event.entity === 'roadmap') {
    roadmapsData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/roadmaps`)
  }
})
</script>

<style scoped>
/* Header */
.research-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.research-actions { display: flex; align-items: center; gap: var(--space-2); }
.title-with-code { display: flex; align-items: center; gap: var(--space-3); }
.short-code {
  font-size: var(--type-xs);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-muted);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  flex-shrink: 0;
  line-height: 1;
}

/* Icon buttons */
.btn-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.35rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color var(--transition-fast), background var(--transition-fast), border-color var(--transition-fast);
  text-decoration: none;
  line-height: 1;
}
.btn-icon:hover {
  color: var(--color-text);
  background: var(--color-surface-hover);
  border-color: var(--color-border-hover);
}
.btn-count {
  font-size: 0.65rem;
  background: var(--color-surface-hover);
  padding: 0.05rem 0.3rem;
  border-radius: 3px;
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  line-height: 1.2;
}

/* Skeleton */
.skeleton-header { height: 60px; margin-bottom: var(--space-6); }
.skeleton-sidebar { height: 300px; }
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }

/* Responsive */
@media (max-width: 768px) {
  .research-header { flex-direction: column; align-items: flex-start; gap: var(--space-3); }
  .research-actions { flex-wrap: wrap; gap: var(--space-2); }
  .title-with-code { flex-wrap: wrap; }
}
</style>
