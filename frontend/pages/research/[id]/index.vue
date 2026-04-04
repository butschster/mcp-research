<template>
  <div v-if="pending" class="empty-state">Loading...</div>
  <div v-else-if="!research" class="empty-state">Research not found.</div>
  <div v-else>
    <div class="page-header" style="display: flex; justify-content: space-between; align-items: center;">
      <div>
        <Breadcrumbs :crumbs="[
          { label: 'Research', to: '/' },
          { label: research.name },
          ...(selectedSection ? [{ label: selectedSection.display_name || selectedSection.name }] : []),
        ]" />
        <h1 class="page-title" style="margin-top: 0.25rem;">{{ research.name }}</h1>
        <p v-if="research.goal" class="card-meta" style="margin-top: 0.25rem;">{{ research.goal }}</p>
      </div>
      <div style="display: flex; gap: 0.5rem; align-items: center;">
        <StatusBadge :status="research.status" />
        <PrintButton />
      </div>
    </div>

    <!-- Active Session Widget -->
    <div v-if="activeSession" class="card" style="margin-bottom: 1.5rem;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem;">
        <div style="display: flex; align-items: center; gap: 0.5rem;">
          <h3 style="font-size: 1rem;">{{ activeSession.title }}</h3>
          <StatusBadge :status="activeSession.status" />
        </div>
        <NuxtLink :to="`/research/${research.id}/session/${activeSession.id}`" class="btn" style="font-size: 0.75rem;">
          Full session &rarr;
        </NuxtLink>
      </div>
      <p v-if="activeSession.focus" class="card-meta" style="margin-bottom: 0.75rem;">Focus: {{ activeSession.focus }}</p>

      <!-- Progress bar -->
      <div v-if="sessionData" style="margin-bottom: 0.75rem;">
        <div class="progress-bar" style="margin-bottom: 0.375rem;">
          <div class="progress-bar-fill" :style="{ width: sessionProgress + '%' }" />
        </div>
        <span class="card-meta">{{ sessionData.progress?.answered ?? 0 }} / {{ sessionData.progress?.total ?? 0 }} answered</span>
      </div>

      <!-- Question checklist -->
      <div v-if="sessionQuestions.length" class="todo-list">
        <div v-for="q in sessionQuestions" :key="q.id" class="todo-item">
          <span :class="['todo-check', `todo-${q.status}`]">
            <template v-if="q.status === 'answered'">&#10003;</template>
            <template v-else-if="q.status === 'skipped'">&times;</template>
            <template v-else-if="q.status === 'deferred'">&rarr;</template>
            <template v-else>&#9675;</template>
          </span>
          <span :class="['todo-text', { 'todo-done': q.status === 'answered' }]">{{ q.text }}</span>
          <StatusBadge v-if="q.priority === 'high'" status="high" />
        </div>
      </div>
    </div>

    <!-- Tasks Panel -->
    <div v-if="tasks.length" class="card" style="margin-bottom: 1.5rem;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem;">
        <h3 style="font-size: 1rem;">Todo</h3>
        <span class="card-meta">{{ tasksDone }} / {{ tasks.length }}</span>
      </div>
      <div class="progress-bar" style="margin-bottom: 0.75rem;">
        <div class="progress-bar-fill" :style="{ width: tasksProgress + '%' }" />
      </div>
      <div class="todo-list">
        <div v-for="t in tasks" :key="t.id" class="todo-item">
          <span :class="['todo-check', `todo-${t.status}`]">
            <template v-if="t.status === 'completed'">&#10003;</template>
            <template v-else-if="t.status === 'failed'">&times;</template>
            <template v-else-if="t.status === 'blocked'">&#9632;</template>
            <template v-else-if="t.status === 'deferred'">&rarr;</template>
            <template v-else-if="t.status === 'in_progress'">&#9654;</template>
            <template v-else>&#9675;</template>
          </span>
          <div style="flex: 1; min-width: 0;">
            <span :class="['todo-text', { 'todo-done': t.status === 'completed' }]">{{ t.title }}</span>
            <div v-if="t.result" class="card-meta" style="font-size: 0.75rem; margin-top: 0.125rem;">{{ t.result }}</div>
          </div>
          <StatusBadge v-if="t.priority === 'high'" status="high" />
          <StatusBadge :status="t.status" />
        </div>
      </div>
    </div>

    <div class="layout-sidebar">
      <!-- Section Sidebar -->
      <div class="sidebar">
        <h3 style="font-size: 0.875rem; color: var(--color-text-muted); margin-bottom: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em;">
          Sections
        </h3>
        <div v-for="section in sections" :key="section.id"
          :class="['sidebar-item', { active: selectedSectionId === section.id }]"
          @click="selectedSectionId = section.id">
          <span>{{ section.display_name || section.name }}</span>
          <span class="card-meta">{{ section.entries_count }}</span>
        </div>
      </div>

      <!-- Section Content -->
      <div>
        <div v-if="!selectedSection" class="empty-state">
          Select a section from the sidebar.
        </div>
        <div v-else>
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
            <h2>{{ selectedSection.display_name || selectedSection.name }}</h2>
            <StatusBadge :status="selectedSection.status" />
          </div>
          <p v-if="selectedSection.description" class="card-meta" style="margin-bottom: 1rem;">
            {{ selectedSection.description }}
          </p>

          <!-- Tags panel -->
          <div v-if="allEntryTags.length" class="tags-panel">
            <span
              v-for="tag in allEntryTags"
              :key="tag"
              :class="['tag', 'tag-clickable', { 'tag-active': entryTagFilter === tag }]"
              @click="entryTagFilter = entryTagFilter === tag ? '' : tag"
            >{{ tag }}</span>
          </div>

          <div v-if="entriesPending" class="empty-state">Loading entries...</div>
          <div v-else-if="!filteredEntries?.length" class="empty-state">
            No entries in this section yet.
          </div>
          <div v-else class="grid" style="grid-template-columns: 1fr;">
            <NuxtLink v-for="entry in filteredEntries" :key="entry.id"
              :to="`/research/${research.id}/entry/${entry.id}`"
              class="card" style="display: block; text-decoration: none; color: inherit;">
              <div style="display: flex; justify-content: space-between; align-items: start;">
                <h3 class="card-title">{{ entry.title }}</h3>
                <StatusBadge :status="entry.status" />
              </div>
              <p v-if="entry.description" class="card-meta" style="margin-top: 0.25rem;">{{ entry.description }}</p>
              <div v-if="entry.tags?.length" style="margin-top: 0.5rem; display: flex; gap: 0.375rem; flex-wrap: wrap;">
                <span
                  v-for="tag in entry.tags"
                  :key="tag"
                  class="tag tag-clickable"
                  @click.prevent.stop="entryTagFilter = tag"
                >{{ tag }}</span>
              </div>
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const { data, pending } = await useApi<{ data: { research: any; sections: any[]; active_session: any } }>(`/api/researches/${id}`)

const research = computed(() => data.value?.data?.research)
const sections = computed(() => data.value?.data?.sections ?? [])
const activeSession = computed(() => data.value?.data?.active_session)

// Fetch tasks
const { data: tasksResponse } = await useApi<{ data: any[] }>(`/api/researches/${id}/tasks`)
const tasks = computed(() => tasksResponse.value?.data ?? [])
const tasksDone = computed(() => tasks.value.filter((t: any) => t.status === 'completed').length)
const tasksProgress = computed(() => {
  if (!tasks.value.length) return 0
  return Math.round((tasksDone.value / tasks.value.length) * 100)
})

// Fetch active session questions
const { data: sessionResponse } = useAsyncData(
  'active-session',
  () => {
    if (!activeSession.value?.id) return Promise.resolve(null)
    const config = useRuntimeConfig()
    const baseURL = config.public.apiBase || ''
    return $fetch<{ data: any }>(`${baseURL}/api/sessions/${activeSession.value.id}`)
  },
  { watch: [activeSession] },
)

const sessionData = computed(() => sessionResponse.value?.data)
const sessionProgress = computed(() => {
  const p = sessionData.value?.progress ?? sessionData.value?.Progress
  if (!p?.total) return 0
  return Math.round((p.answered / p.total) * 100)
})
const sessionQuestions = computed(() => {
  const qs = sessionData.value?.questions ?? sessionData.value?.Questions ?? {}
  const order = ['pending', 'in_progress', 'answered', 'deferred', 'skipped']
  const all: any[] = []
  for (const status of order) {
    for (const q of qs[status] ?? []) {
      all.push({ ...q, status })
    }
  }
  return all
})

// Restore selected section from query param, fall back to first section
const initialSectionId = (route.query.section as string) || sections.value?.[0]?.id || ''
const selectedSectionId = ref<string>(initialSectionId)

// Sync to URL query param
watch(selectedSectionId, (val) => {
  router.replace({ query: { ...route.query, section: val || undefined } })
})

const selectedSection = computed(() =>
  sections.value.find((s: any) => s.id === selectedSectionId.value) ?? null
)

const entriesUrl = computed(() => {
  if (!selectedSectionId.value) return null
  return `/api/researches/${id}/sections/${selectedSectionId.value}/entries`
})

const { data: entriesResponse, pending: entriesPending } = useAsyncData(
  () => {
    if (!entriesUrl.value) return Promise.resolve(null)
    const config = useRuntimeConfig()
    const baseURL = config.public.apiBase || ''
    return $fetch<{ data: any[] }>(`${baseURL}${entriesUrl.value}`)
  },
  { watch: [entriesUrl] },
)

const entries = computed(() => entriesResponse.value?.data ?? [])

const entryTagFilter = ref('')

// Reset tag filter when switching sections
watch(selectedSectionId, () => { entryTagFilter.value = '' })

const allEntryTags = computed(() => {
  const tags = new Set<string>()
  for (const e of entries.value) {
    for (const t of e.tags ?? []) {
      tags.add(t)
    }
  }
  return [...tags].sort()
})

const filteredEntries = computed(() => {
  if (!entryTagFilter.value) return entries.value
  return entries.value.filter((e: any) => e.tags?.includes(entryTagFilter.value))
})
</script>

<style scoped>
.tags-panel {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  padding: 0.75rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  margin-bottom: 1rem;
}
.tag-clickable {
  cursor: pointer;
  transition: all 0.15s;
  user-select: none;
}
.tag-clickable:hover {
  background: rgba(56, 189, 248, 0.15);
  color: var(--color-primary);
}
.tag-active {
  background: rgba(56, 189, 248, 0.2);
  color: var(--color-primary);
  border: 1px solid rgba(56, 189, 248, 0.4);
}
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.todo-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0;
  font-size: 0.875rem;
}
.todo-check {
  flex-shrink: 0;
  width: 1.25rem;
  text-align: center;
  font-size: 0.8125rem;
}
.todo-pending { color: var(--color-text-muted); }
.todo-in_progress { color: var(--color-warning); }
.todo-completed, .todo-answered { color: var(--color-success); }
.todo-blocked { color: var(--color-error); }
.todo-failed { color: var(--color-error); }
.todo-deferred { color: var(--color-text-muted); }
.todo-skipped { color: var(--color-error); }
.todo-text {
  flex: 1;
}
.todo-done {
  text-decoration: line-through;
  color: var(--color-text-muted);
}
</style>
