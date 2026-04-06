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
          <NuxtLink :to="`/research/${researchSlug}/mindmap`" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><circle cx="4" cy="6" r="2"/><circle cx="20" cy="6" r="2"/><circle cx="4" cy="18" r="2"/><circle cx="20" cy="18" r="2"/><path d="M9.5 10.5 5.5 7.5"/><path d="M14.5 10.5l4-3"/><path d="M9.5 13.5 5.5 16.5"/><path d="M14.5 13.5l4 3"/></svg>
            Mind map
          </NuxtLink>
        </div>
      </div>
      <p v-if="research.goal" class="card-meta mt-2">{{ research.goal }}</p>
    </div>

    <!-- Active session summary -->
    <NuxtLink v-if="activeSession" :to="`/research/${researchSlug}/session/${activeSession.id}`" class="card session-widget">
      <div class="session-widget-header">
        <div class="flex items-center gap-2">
          <span class="session-label">{{ activeSession?.status === 'active' ? 'Active session' : 'Session' }}</span>
          <StatusBadge :status="activeSession.status" />
        </div>
      </div>
      <h3 class="session-title">{{ activeSession.title }}</h3>
      <p v-if="activeSession.focus" class="card-meta mt-2">{{ activeSession.focus }}</p>

      <div class="session-stats">
        <div v-if="sessionProgress" class="session-stat">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
          <span>{{ sessionProgress.answered }} / {{ sessionProgress.total }} questions</span>
          <ProgressBar :value="sessionProgress.answered" :total="sessionProgress.total" class="stat-progress" />
        </div>
        <div v-if="tasks.length" class="session-stat">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
          <span>{{ completedTasks }} / {{ tasks.length }} tasks</span>
          <ProgressBar :value="completedTasks" :total="tasks.length" class="stat-progress" />
        </div>
      </div>
    </NuxtLink>

    <!-- Tasks widget (shown only when no active session but tasks exist) -->
    <div v-else-if="tasks.length" class="card task-widget">
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
            <div v-if="t.result" class="card-meta todo-result" v-html="renderRefs(t.result, researchSlug)"></div>
          </div>
          <StatusBadge v-if="t.priority === 'high'" :status="t.priority" />
          <StatusBadge :status="t.status" />
        </div>
      </div>
    </div>

    <!-- Sidebar layout: sections + entries -->
    <div class="layout-sidebar">
      <!-- Sidebar -->
      <nav class="sidebar">
        <!-- All entries -->
        <div
          :class="['sidebar-item', { active: isAllEntries }]"
          @click="activeSection = '__all__'"
        >
          <div class="sidebar-item-content">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
            <span class="sidebar-item-name">All entries</span>
            <span class="sidebar-count">{{ totalEntryCount }}</span>
          </div>
        </div>

        <div class="sidebar-divider"></div>

        <!-- Per-section -->
        <div
          v-for="section in sections"
          :key="section.id"
          :class="['sidebar-item', { active: activeSection === section.id }]"
          @click="activeSection = section.id"
        >
          <div class="sidebar-item-content">
            <span class="sidebar-item-name">{{ section.display_name || section.name }}</span>
            <span class="sidebar-count">{{ section.entries_count }}</span>
          </div>
          <div v-if="section.entries_count > 0" class="sidebar-progress">
            <div class="sidebar-progress-fill" :style="{ width: sectionProgressWidth(section) }"></div>
          </div>
        </div>
      </nav>

      <!-- Main: entries -->
      <div>
        <!-- All entries view -->
        <template v-if="isAllEntries">
          <div class="section-header">
            <h2 class="section-title">All entries</h2>
          </div>

          <!-- Global tags with counters -->
          <div v-if="globalTags.length" class="tags-panel mb-4">
            <span
              v-for="tc in globalTags"
              :key="tc.tag"
              :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tc.tag)}`, { 'tag-active': activeTag === tc.tag }]"
              @click="activeTag = activeTag === tc.tag ? '' : tc.tag"
            >{{ tc.tag }}<span v-if="tc.count > 1" class="tag-count">{{ tc.count }}</span></span>
          </div>

          <!-- Loading -->
          <div v-if="allEntriesPending">
            <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
          </div>

          <template v-else-if="filteredAllEntries.length">
            <!-- Group by section -->
            <template v-for="group in groupedEntries" :key="group.section.id">
              <h3 class="group-section-title">{{ group.section.display_name || group.section.name }}</h3>
              <div class="grid entries-grid mb-4">
                <NuxtLink
                  v-for="entry in group.entries"
                  :key="entry.id"
                  :to="`/research/${researchSlug}/entry/${entry.code || entry.id}`"
                  class="card entry-card"
                >
                  <div class="entry-card-header">
                    <div class="entry-title-row">
                      <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
                      <h3 class="card-title">{{ entry.title }}</h3>
                    </div>
                    <StatusBadge :status="entry.status" />
                  </div>
                  <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
                  <div v-if="entry.tags?.length" class="entry-tags">
                    <span v-for="tag in entry.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
                  </div>
                </NuxtLink>
              </div>
            </template>
          </template>

          <EmptyState
            v-else
            icon="&#x1F4C4;"
            title="No entries yet"
            description="Claude will populate this research with entries."
          />
        </template>

        <!-- Single section view -->
        <template v-else-if="currentSection">
          <div class="section-header">
            <h2 class="section-title">{{ currentSection.display_name || currentSection.name }}</h2>
            <StatusBadge :status="currentSection.status" />
          </div>
          <p v-if="currentSection.description" class="card-meta mb-4">
            {{ currentSection.description }}
          </p>

          <!-- Tag filter for entries -->
          <div v-if="entryTagCounts.length" class="tags-panel mb-4">
            <span
              v-for="tc in entryTagCounts"
              :key="tc.tag"
              :class="['tag', 'tag-clickable', `tag-hue-${tagHue(tc.tag)}`, { 'tag-active': activeTag === tc.tag }]"
              @click="activeTag = activeTag === tc.tag ? '' : tc.tag"
            >{{ tc.tag }}<span v-if="tc.count > 1" class="tag-count">{{ tc.count }}</span></span>
          </div>

          <!-- Entries loading -->
          <div v-if="entriesPending">
            <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
          </div>

          <div v-else-if="filteredEntries.length" class="grid entries-grid">
            <NuxtLink
              v-for="entry in filteredEntries"
              :key="entry.id"
              :to="`/research/${researchSlug}/entry/${entry.code || entry.id}`"
              class="card entry-card"
            >
              <div class="entry-card-header">
                <div class="entry-title-row">
                  <span v-if="entry.code" class="short-code">{{ entry.code }}</span>
                  <h3 class="card-title">{{ entry.title }}</h3>
                </div>
                <StatusBadge :status="entry.status" />
              </div>
              <p v-if="entry.description" class="card-meta mt-2" v-html="renderRefs(entry.description, researchSlug)"></p>
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
const researchSlug = computed(() => research.value?.code || id)
const sections = computed(() => researchData.value?.data?.sections ?? [])
const activeSession = computed(() => researchData.value?.data?.active_session)

const totalEntryCount = computed(() =>
  sections.value.reduce((sum: number, s: any) => sum + (s.entries_count || 0), 0)
)

// Active section (default: first, or '__all__' for all entries)
const activeSection = ref(route.query.section as string || '')
const isAllEntries = computed(() => activeSection.value === '__all__')
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

// Entry tag filter
const activeTag = ref('')
watch(activeSection, () => { activeTag.value = '' })

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

const entryTagCounts = computed(() => {
  const map = new Map<string, number>()
  for (const e of entries.value) for (const t of (e.tags ?? [])) map.set(t, (map.get(t) || 0) + 1)
  return [...map.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .map(([tag, count]) => ({ tag, count }))
})
const entryTags = computed(() => entryTagCounts.value.map(tc => tc.tag))

const filteredEntries = computed(() =>
  activeTag.value ? entries.value.filter((e: any) => e.tags?.includes(activeTag.value)) : entries.value
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

const filteredAllEntries = computed(() =>
  activeTag.value
    ? allEntries.value.filter((e: any) => e.tags?.includes(activeTag.value))
    : allEntries.value
)

// Group entries by section for the all-entries view
const groupedEntries = computed(() => {
  const groups: { section: any; entries: any[] }[] = []
  const sectionMap = new Map<string, any[]>()

  for (const entry of filteredAllEntries.value) {
    const list = sectionMap.get(entry.section_id) ?? []
    list.push(entry)
    sectionMap.set(entry.section_id, list)
  }

  for (const section of sections.value) {
    const sectionEntries = sectionMap.get(section.id)
    if (sectionEntries?.length) {
      groups.push({ section, entries: sectionEntries })
    }
  }

  return groups
})

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
const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''
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
  }
  if (event.entity === 'task') {
    tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  }
  if (['question', 'session'].includes(event.entity) && activeSession.value?.id) {
    sessionData.value = await authFetch<any>(`${rtBase}/api/sessions/${activeSession.value.id}`)
  }
})
</script>

<style scoped>
/* Header */
.research-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.research-actions { display: flex; align-items: center; gap: var(--space-3); }
.title-with-code { display: flex; align-items: center; gap: var(--space-3); }
.entry-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
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

/* Sidebar */
.sidebar-item-content {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
.sidebar-item-content svg {
  flex-shrink: 0;
  opacity: 0.5;
}
.sidebar-item.active .sidebar-item-content svg { opacity: 1; }
.sidebar-item-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--type-sm);
}
.sidebar-count {
  font-size: var(--type-xs);
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  min-width: 1.2em;
  text-align: right;
}
.sidebar-divider {
  height: 1px;
  background: var(--color-border);
  margin: var(--space-1) var(--space-3);
}

/* Session widget */
.session-widget {
  display: block;
  text-decoration: none;
  color: inherit;
  margin-bottom: var(--space-6);
  border-color: rgba(108, 197, 224, 0.15);
  position: relative;
  overflow: hidden;
}
.session-widget:hover { text-decoration: none; }
.session-widget::after {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 1px;
  background: linear-gradient(90deg, transparent, rgba(108, 197, 224, 0.3), transparent);
}
.session-widget-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: var(--space-2);
}
.session-label {
  font-size: var(--type-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-primary);
}
.session-title { font-size: var(--type-lg); font-weight: 600; letter-spacing: -0.01em; }

/* Session stats (questions + tasks summary) */
.session-stats {
  display: flex;
  gap: var(--space-6);
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border);
}
.session-stat {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-sm);
  color: var(--color-text-muted);
  flex: 1;
  min-width: 0;
}
.session-stat svg { flex-shrink: 0; opacity: 0.6; }
.session-stat span { white-space: nowrap; }
.stat-progress { flex: 1; min-width: 60px; }

/* Task widget */
.task-widget { margin-bottom: var(--space-6); }
.task-header {
  justify-content: space-between;
  width: 100%;
  margin-bottom: var(--space-3);
}
.task-header:hover .task-header-title { color: var(--color-primary); }
.task-header-title { font-size: var(--type-base); font-weight: 600; letter-spacing: -0.01em; }
.task-header-right { display: flex; align-items: center; gap: var(--space-3); }
.task-chevron {
  font-size: var(--type-lg); color: var(--color-text-muted);
  transition: transform var(--transition-base); display: inline-block; line-height: 1;
}
.task-chevron.open { transform: rotate(90deg); }

/* Todo list */
.todo-list { display: flex; flex-direction: column; gap: var(--space-1); margin-top: var(--space-3); }
.todo-item {
  display: flex; align-items: center; gap: var(--space-3); font-size: var(--type-sm);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}
.todo-item:hover { background: var(--color-surface-hover); }
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
.section-title { font-size: var(--type-xl); font-weight: 600; letter-spacing: -0.02em; }
.tags-panel { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.tag-active { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-clickable { cursor: pointer; transition: all var(--transition-fast); }
.tag-clickable:hover { background: var(--color-primary-muted); color: var(--color-primary); }
.tag-count {
  font-size: 0.75em;
  opacity: 0.7;
  margin-left: 0.15em;
}

.group-section-title {
  font-size: var(--type-base);
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
}

.entries-grid { grid-template-columns: 1fr; }
.entry-card { display: block; text-decoration: none; color: inherit; }
.entry-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-2); }
.entry-tags { display: flex; gap: var(--space-2); flex-wrap: wrap; margin-top: var(--space-3); }

/* Skeleton */
.skeleton-header { height: 60px; margin-bottom: var(--space-6); }
.skeleton-sidebar { height: 300px; }
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }
</style>
