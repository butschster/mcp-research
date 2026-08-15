<template>
  <div v-if="research">
    <div class="page-header">
      <div class="research-header">
        <div class="title-with-code">
          <span v-if="research.code" class="short-code">{{ research.code }}</span>
          <h1 class="page-title">{{ research.name }}</h1>
        </div>
        <div class="research-actions">
          <StatusBadge :status="research.status" />

          <NuxtLink v-if="include.tasks" :to="tasksPath(slug)" class="btn btn-icon" title="Tasks">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
          </NuxtLink>
          <NuxtLink v-if="include.roadmaps" :to="roadmapsPath(slug)" class="btn btn-icon" title="Roadmaps">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
            <span v-if="roadmaps.length" class="btn-count">{{ roadmaps.length }}</span>
          </NuxtLink>
          <NuxtLink v-if="include.export" :to="exportPath(slug)" class="btn btn-icon" title="Export">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </NuxtLink>
        </div>
      </div>
      <p v-if="research.goal" class="card-meta mt-2">{{ research.goal }}</p>
      <p v-if="research.description" class="card-meta mt-2">{{ research.description }}</p>
      <TagList v-if="research.tags?.length" :tags="research.tags" class="mt-2" />
    </div>

    <ResearchActiveSessionsGrid v-if="include.sessions" :sessions="activeSessions" :research-slug="slug" />
    <ResearchPastSessionsList v-if="include.sessions" :sessions="closedSessions" :research-slug="slug" />

    <EmptyState
      v-if="!sections.length"
      icon="&#x1F4C4;"
      title="Nothing here yet"
      :description="emptyDescription"
    />

    <div v-else class="layout-sidebar">
      <ResearchSidebar
        :sections="sections"
        :active-section="activeSection"
        :total-entry-count="totalEntryCount"
        :links-total="linksTotal"
        @update:active-section="activeSection = $event"
      />

      <div>
        <ResearchEntriesView
          v-if="isAllEntries"
          :entries="allEntries"
          :sections="sections"
          :research-slug="slug"
          :loading="loadingEntries"
          mode="all"
          :tags="tags"
        />
        <ResearchEntriesView
          v-else-if="currentSection"
          :entries="entries"
          :sections="sections"
          :research-slug="slug"
          :loading="loadingEntries"
          mode="section"
          :section-info="currentSection"
          :tags="[]"
        />
        <ResearchExternalLinksView
          v-else-if="isLinksView"
          :groups="linkGroups"
          :loading="loadingLinks"
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
</template>

<script setup lang="ts">
/**
 * The shared research overview — the owner's page with everything that writes,
 * navigates away, or names the workspace behind the link taken out.
 *
 * It fetches nothing about permission and nothing about the share: the shell
 * above it already resolved both, and a child that re-decided would be a second
 * place to get it wrong.
 */
const { shareFetch, research, sections, include, ownerName, researchId, researchCode, slug } = useShare()

const route = useRoute()
const router = useRouter()

const activeSection = ref((route.query.section as string) || '')
const isAllEntries = computed(() => activeSection.value === '__all__')
const isLinksView = computed(() => activeSection.value === '__links__')
const currentSection = computed(() =>
  isAllEntries.value ? null : sections.value.find((s: any) => s.id === activeSection.value) ?? null,
)
const totalEntryCount = computed(() =>
  sections.value.reduce((sum: number, s: any) => sum + (s.entries_count || 0), 0),
)

watch(activeSection, (val) => {
  router.replace({ query: { ...route.query, section: val || undefined } })
})
watch(sections, (secs) => {
  if (!activeSection.value && secs.length) activeSection.value = secs[0].id
}, { immediate: true })

const emptyDescription = computed(() =>
  ownerName.value
    ? `${ownerName.value} hasn't added anything to this research yet. This page updates by itself when they do — you can leave it open.`
    : 'Nothing has been added to this research yet.',
)

// --- Content ---

const entries = ref<any[]>([])
const allEntries = ref<any[]>([])
const tags = ref<any[]>([])
const loadingEntries = ref(false)

async function loadEntries() {
  if (isLinksView.value) return
  loadingEntries.value = true
  try {
    if (isAllEntries.value) {
      const [e, t] = await Promise.all([
        shareFetch<{ data: any[] }>(`/researches/${researchId.value}/entries`),
        shareFetch<{ data: any[] }>(`/researches/${researchId.value}/tags`),
      ])
      allEntries.value = e.data ?? []
      tags.value = t.data ?? []
    } else if (activeSection.value) {
      const e = await shareFetch<{ data: any[] }>(
        `/researches/${researchId.value}/sections/${activeSection.value}/entries`,
      )
      entries.value = e.data ?? []
    }
  } catch {
    // The shell owns every failure that means something. A list that could not
    // be fetched renders as an empty list and the next event refetches it.
  } finally {
    loadingEntries.value = false
  }
}

const linkGroups = ref<any[]>([])
const linksTotal = ref(0)
const loadingLinks = ref(false)

async function loadLinks() {
  if (!isLinksView.value && linksTotal.value) return
  loadingLinks.value = true
  try {
    const res = await shareFetch<{ data: any[]; total: number }>(`/researches/${researchId.value}/links`)
    linkGroups.value = res.data ?? []
    linksTotal.value = res.total ?? 0
  } catch {
    linkGroups.value = []
  } finally {
    loadingLinks.value = false
  }
}

const roadmaps = ref<any[]>([])
const sessions = ref<any[]>([])
const activeSessions = computed(() => sessions.value.filter((s: any) => s.status === 'active'))
const closedSessions = computed(() => sessions.value.filter((s: any) => s.status !== 'active'))

async function loadExtras() {
  if (include.value.roadmaps) {
    try {
      const res = await shareFetch<{ data: any[] }>(`/researches/${researchId.value}/roadmaps`)
      roadmaps.value = res.data ?? []
    } catch { roadmaps.value = [] }
  }
  if (include.value.sessions) {
    try {
      const res = await shareFetch<{ data: any[] }>(`/researches/${researchId.value}/sessions`)
      sessions.value = res.data ?? []
    } catch { sessions.value = [] }
  }
}

watch([activeSection, researchId], () => void loadEntries(), { immediate: true })
watch(isLinksView, (on) => { if (on) void loadLinks() })
// Not awaited: awaiting here makes the child a Suspense boundary too, so the
// overview waited on the shell and then on these before painting anything.
void loadExtras()
void loadLinks()

// The shell refetches the research payload on every event; the lists live here,
// so they subscribe too. Same burst coalescing, same scope.
useResearchRealtime(
  () => slug.value,
  (event) => {
    if (event.entity === 'entry' || event.entity === 'section') {
      void loadEntries()
      void loadLinks()
    }
    if (event.entity === 'session' || event.entity === 'question' || event.entity === 'roadmap') {
      void loadExtras()
    }
  },
  {
    onResync: () => { void loadEntries(); void loadExtras(); void loadLinks() },
    researchId: () => researchId.value,
  },
)
</script>
