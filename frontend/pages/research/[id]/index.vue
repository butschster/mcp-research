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
        <h1 class="page-title">{{ research.name }}</h1>
        <StatusBadge :status="research.status" />
      </div>
      <p v-if="research.goal" class="card-meta mt-2">{{ research.goal }}</p>
    </div>

    <!-- Active session widget -->
    <div v-if="activeSession" class="card session-widget">
      <div class="session-widget-header">
        <div class="flex items-center gap-2">
          <span class="card-meta">Active Session</span>
          <StatusBadge :status="activeSession.status" />
        </div>
        <NuxtLink :to="`/research/${id}/session/${activeSession.id}`" class="btn btn-sm">
          View questions &rarr;
        </NuxtLink>
      </div>
      <h3 class="session-title">{{ activeSession.title }}</h3>
      <p v-if="activeSession.focus" class="card-meta mt-2">{{ activeSession.focus }}</p>
      <div v-if="sessionProgress" class="session-progress">
        <ProgressBar :value="sessionProgress.answered" :total="sessionProgress.total" />
        <span class="card-meta">
          {{ sessionProgress.answered }} / {{ sessionProgress.total }} questions answered
        </span>
      </div>
    </div>

    <!-- Tasks widget -->
    <div v-if="tasks.length" class="card task-widget">
      <button class="btn-ghost task-header" @click="tasksOpen = !tasksOpen">
        <h3 class="task-header-title">Tasks</h3>
        <div class="task-header-right">
          <span class="card-meta">{{ completedTasks }} / {{ tasks.length }}</span>
          <span class="task-chevron" :class="{ open: tasksOpen }">&rsaquo;</span>
        </div>
      </button>
      <ProgressBar :value="completedTasks" :total="tasks.length" />
      <div v-show="tasksOpen" class="todo-list">
        <div v-for="t in tasks" :key="t.id" class="todo-item">
          <span :class="['todo-check', `todo-${t.status}`]">
            <template v-if="t.status === 'completed'">&check;</template>
            <template v-else-if="t.status === 'failed'">&times;</template>
            <template v-else-if="t.status === 'blocked'">&block;</template>
            <template v-else-if="t.status === 'deferred'">&rarr;</template>
            <template v-else-if="t.status === 'in_progress'">&triangleright;</template>
            <template v-else>&cir;</template>
          </span>
          <div class="todo-content">
            <span :class="['todo-text', { 'todo-done': t.status === 'completed' }]">{{ t.title }}</span>
            <div v-if="t.result" class="card-meta todo-result">{{ t.result }}</div>
          </div>
          <StatusBadge v-if="t.priority === 'high'" :status="t.priority" />
          <StatusBadge :status="t.status" />
        </div>
      </div>
    </div>

    <!-- Sidebar layout: sections + entries -->
    <div class="layout-sidebar">
      <!-- Sidebar -->
      <div class="sidebar">
        <h3 class="sidebar-label">Sections</h3>
        <div
          v-for="section in sections"
          :key="section.id"
          :class="['sidebar-item', { active: activeSection === section.id }]"
          @click="activeSection = section.id"
        >
          <div class="sidebar-item-content">
            <span class="sidebar-item-name">{{ section.display_name || section.name }}</span>
            <StatusBadge :status="section.status" />
          </div>
          <div class="sidebar-item-meta">
            <span class="card-meta">{{ section.entries_count }} entries</span>
          </div>
          <div v-if="section.entries_count > 0" class="sidebar-progress">
            <div class="sidebar-progress-fill" :style="{ width: sectionProgressWidth(section) }"></div>
          </div>
        </div>
      </div>

      <!-- Main: entries -->
      <div>
        <template v-if="currentSection">
          <div class="section-header">
            <h2 class="section-title">{{ currentSection.display_name || currentSection.name }}</h2>
            <StatusBadge :status="currentSection.status" />
          </div>
          <p v-if="currentSection.description" class="card-meta mb-4">
            {{ currentSection.description }}
          </p>

          <!-- Tag filter for entries -->
          <div v-if="entryTags.length" class="tags-panel mb-4">
            <span
              v-for="tag in entryTags"
              :key="tag"
              :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tag)}`, { 'tag-active': activeTag === tag }]"
              @click="activeTag = activeTag === tag ? '' : tag"
            >{{ tag }}</span>
          </div>

          <!-- Entries loading -->
          <div v-if="entriesPending">
            <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
          </div>

          <div v-else-if="filteredEntries.length" class="grid entries-grid">
            <NuxtLink
              v-for="entry in filteredEntries"
              :key="entry.id"
              :to="`/research/${id}/entry/${entry.id}`"
              class="card entry-card"
            >
              <div class="entry-card-header">
                <h3 class="card-title">{{ entry.title }}</h3>
                <StatusBadge :status="entry.status" />
              </div>
              <p v-if="entry.description" class="card-meta mt-2">{{ entry.description }}</p>
              <div v-if="entry.tags?.length" class="entry-tags">
                <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
              </div>
            </NuxtLink>
          </div>

          <EmptyState
            v-else
            icon="&#x1F4C4;"
            title="No entries yet"
            description="Claude will populate this section with research entries."
          />
        </template>

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
</template>

<script setup lang="ts">
const route = useRoute()
const id = route.params.id as string

// Research data
const { data: researchData, pending } = await useApi<{ data: any }>(`/api/researches/${id}`)

const research = computed(() => researchData.value?.data?.research)
const sections = computed(() => researchData.value?.data?.sections ?? [])
const activeSession = computed(() => researchData.value?.data?.active_session)

// Active section (default: first)
const activeSection = ref(route.query.section as string || '')
const router = useRouter()

watch(activeSection, (val) => {
  router.replace({ query: { ...route.query, section: val || undefined } })
})

watch(sections, (secs) => {
  if (!activeSection.value && secs.length) activeSection.value = secs[0].id
}, { immediate: true })

const currentSection = computed(() => sections.value.find((s: any) => s.id === activeSection.value) ?? null)

// Entries
const entriesUrl = computed(() =>
  activeSection.value ? `/api/researches/${id}/sections/${activeSection.value}/entries` : null
)
const { data: entriesData, pending: entriesPending } = useApi<{ data: any[] }>(
  computed(() => entriesUrl.value ?? '/api/researches/__none__/sections/__none__/entries')
)
const entries = computed(() => activeSection.value ? (entriesData.value?.data ?? []) : [])

// Entry tag filter
const activeTag = ref('')
watch(activeSection, () => { activeTag.value = '' })

const entryTags = computed(() => {
  const set = new Set<string>()
  for (const e of entries.value) for (const t of (e.tags ?? [])) set.add(t)
  return [...set].sort()
})

const filteredEntries = computed(() =>
  activeTag.value ? entries.value.filter((e: any) => e.tags?.includes(activeTag.value)) : entries.value
)

// Tag color
function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}

// Section progress
function sectionProgressWidth(section: any): string {
  if (section.status === 'completed') return '100%'
  if (section.status === 'active') return '50%'
  if (section.status === 'draft') return '10%'
  return '0%'
}

// Tasks
const { data: tasksData } = await useApi<{ data: any[] }>(`/api/researches/${id}/tasks`)
const tasks = computed(() => tasksData.value?.data ?? [])
const completedTasks = computed(() => tasks.value.filter((t: any) => t.status === 'completed').length)
const tasksOpen = ref(true)

// Active session progress
const { data: sessionData } = useApi<{ data: any }>(
  computed(() => activeSession.value?.id ? `/api/sessions/${activeSession.value.id}` : '/api/sessions/__none__')
)
const sessionProgress = computed(() => sessionData.value?.data?.progress ?? null)

// Real-time updates
useRealtimeUpdates(async (event) => {
  if (event.research_id && event.research_id !== id) return
  const config = useRuntimeConfig()
  const base = config.public.apiBase || ''

  if (['research', 'section', 'session'].includes(event.entity)) {
    const res = await $fetch<any>(`${base}/api/researches/${id}`)
    researchData.value = res
  }
  if (event.entity === 'entry' && entriesUrl.value) {
    const res = await $fetch<any>(`${base}${entriesUrl.value}`)
    entriesData.value = res
  }
  if (event.entity === 'task') {
    const res = await $fetch<any>(`${base}/api/researches/${id}/tasks`)
    tasksData.value = res
  }
  if (['question', 'session'].includes(event.entity) && activeSession.value?.id) {
    const res = await $fetch<any>(`${base}/api/sessions/${activeSession.value.id}`)
    sessionData.value = res
  }
})
</script>

<style scoped>
/* Header */
.research-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }

/* Sidebar */
.sidebar-label { font-size: var(--type-xs); font-weight: 700; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.08em; margin-bottom: var(--space-4); }
.sidebar-item-content { display: flex; align-items: center; justify-content: space-between; gap: var(--space-2); }
.sidebar-item-name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: var(--type-sm); }
.sidebar-item-meta { font-size: var(--type-xs); margin-top: var(--space-1); }

/* Session widget */
.session-widget { margin-bottom: var(--space-6); border-color: rgba(56,189,248,0.2); }
.session-widget-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-2); }
.session-title { font-size: var(--type-base); font-weight: 600; }
.session-progress { margin-top: var(--space-3); display: flex; flex-direction: column; gap: var(--space-2); }

/* Task widget */
.task-widget { margin-bottom: var(--space-6); }
.task-header {
  justify-content: space-between;
  width: 100%;
  margin-bottom: var(--space-3);
}
.task-header:hover .task-header-title { color: var(--color-primary); }
.task-header-title { font-size: var(--type-base); font-weight: 600; }
.task-header-right { display: flex; align-items: center; gap: var(--space-3); }
.task-chevron {
  font-size: var(--type-lg); color: var(--color-text-muted);
  transition: transform var(--transition-base); display: inline-block; line-height: 1;
}
.task-chevron.open { transform: rotate(90deg); }

/* Todo list */
.todo-list { display: flex; flex-direction: column; gap: var(--space-2); margin-top: var(--space-3); }
.todo-item { display: flex; align-items: center; gap: var(--space-3); font-size: var(--type-sm); }
.todo-check { width: var(--space-5); text-align: center; flex-shrink: 0; color: var(--color-text-muted); }
.todo-content { flex: 1; min-width: 0; }
.todo-done { text-decoration: line-through; color: var(--color-text-muted); }
.todo-result { font-size: var(--type-xs); margin-top: var(--space-1); }
.todo-completed .todo-check { color: var(--color-success); }
.todo-failed .todo-check { color: var(--color-error); }
.todo-blocked .todo-check { color: var(--color-error); }
.todo-in_progress .todo-check { color: var(--color-warning); }

/* Sections + Entries */
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--space-4); }
.section-title { font-size: var(--type-xl); font-weight: 600; }
.tags-panel { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.tag-active { background: rgba(56,189,248,0.15); color: var(--color-primary); }
.tag-clickable { cursor: pointer; transition: all var(--transition-fast); }
.tag-clickable:hover { background: rgba(56,189,248,0.12); color: var(--color-primary); }

.entries-grid { grid-template-columns: 1fr; }
.entry-card { display: block; text-decoration: none; color: inherit; }
.entry-card:hover { border-color: var(--color-primary); }
.entry-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-2); }
.entry-tags { display: flex; gap: var(--space-2); flex-wrap: wrap; margin-top: var(--space-3); }

/* Skeleton */
.skeleton-header { height: 60px; margin-bottom: var(--space-6); }
.skeleton-sidebar { height: 300px; }
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }
</style>
