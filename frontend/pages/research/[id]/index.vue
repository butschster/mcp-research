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
          <button
            class="btn btn-sm"
            :class="research.status === 'archived' ? 'btn-primary' : 'btn-danger'"
            @click="toggleArchive"
          >
            <svg v-if="research.status === 'archived'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
            {{ research.status === 'archived' ? 'Restore' : 'Archive' }}
          </button>
          <NuxtLink :to="`/research/${researchSlug}/tasks`" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
            Tasks
            <span v-if="tasks.length" class="btn-count">{{ tasks.length }}</span>
          </NuxtLink>
          <NuxtLink :to="`/research/${researchSlug}/mindmap`" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><circle cx="4" cy="6" r="2"/><circle cx="20" cy="6" r="2"/><circle cx="4" cy="18" r="2"/><circle cx="20" cy="18" r="2"/><path d="M9.5 10.5 5.5 7.5"/><path d="M14.5 10.5l4-3"/><path d="M9.5 13.5 5.5 16.5"/><path d="M14.5 13.5l4 3"/></svg>
            Mind map
          </NuxtLink>
          <NuxtLink :to="`/research/${researchSlug}/export`" class="btn btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Export
          </NuxtLink>
          <button class="btn btn-sm" @click="detailsOpen = !detailsOpen" :title="detailsOpen ? 'Hide details' : 'Show details'">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </button>
        </div>
      </div>
      <p v-if="research.goal && !detailsOpen" class="card-meta mt-2">{{ research.goal }}</p>
    </div>

    <!-- Research details panel -->
    <div v-if="detailsOpen" class="card details-panel">
      <div class="details-grid">
        <!-- Goal -->
        <div class="detail-field" @dblclick="startEdit('goal')">
          <label class="detail-label">Goal</label>
          <div v-if="editingField !== 'goal'" class="detail-value" :class="{ 'detail-empty': !research.goal }">
            {{ research.goal || 'Not set — double-click to edit' }}
          </div>
          <div v-else class="detail-edit">
            <input v-model="editValue" class="detail-input" @keydown.enter="saveEdit('goal')" @keydown.escape="cancelEdit" ref="editInput" />
            <div class="detail-edit-actions">
              <button class="btn btn-sm btn-primary" @click="saveEdit('goal')">Save</button>
              <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
            </div>
          </div>
        </div>

        <!-- Description -->
        <div class="detail-field" @dblclick="startEdit('description')">
          <label class="detail-label">Description</label>
          <div v-if="editingField !== 'description'" class="detail-value" :class="{ 'detail-empty': !research.description }">
            {{ research.description || 'Not set — double-click to edit' }}
          </div>
          <div v-else class="detail-edit">
            <textarea v-model="editValue" class="detail-textarea" rows="3" @keydown.escape="cancelEdit" ref="editInput"></textarea>
            <div class="detail-edit-actions">
              <button class="btn btn-sm btn-primary" @click="saveEdit('description')">Save</button>
              <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
            </div>
          </div>
        </div>

        <!-- Instruction -->
        <div class="detail-field detail-field-wide" @dblclick="startEdit('instruction')">
          <label class="detail-label">Instruction</label>
          <div v-if="editingField !== 'instruction'" class="detail-value detail-value-pre" :class="{ 'detail-empty': !research.instruction }">
            {{ research.instruction || 'Not set — double-click to edit' }}
          </div>
          <div v-else class="detail-edit">
            <textarea v-model="editValue" class="detail-textarea" rows="6" @keydown.escape="cancelEdit" ref="editInput"></textarea>
            <div class="detail-edit-actions">
              <button class="btn btn-sm btn-primary" @click="saveEdit('instruction')">Save</button>
              <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
            </div>
          </div>
        </div>

        <!-- Memory -->
        <div class="detail-field detail-field-wide">
          <label class="detail-label">Memory <span class="detail-count">{{ research.memory?.length || 0 }}</span></label>
          <div v-if="research.memory?.length" class="memory-list">
            <div v-for="(item, i) in research.memory" :key="i" class="memory-item">
              <span class="memory-bullet">{{ i + 1 }}</span>
              <span>{{ item }}</span>
            </div>
          </div>
          <div v-else class="detail-value detail-empty">No memory entries yet</div>
        </div>

        <!-- Tags -->
        <div class="detail-field" @dblclick="startEdit('tags')">
          <label class="detail-label">Tags</label>
          <div v-if="editingField !== 'tags'">
            <div v-if="research.tags?.length" class="tags-row">
              <span v-for="tag in research.tags" :key="tag" :class="['tag', `tag-hue-${tagHue(tag)}`]">{{ tag }}</span>
            </div>
            <div v-else class="detail-value detail-empty">No tags — double-click to edit</div>
          </div>
          <div v-else class="detail-edit">
            <input v-model="editValue" class="detail-input" placeholder="tag1, tag2, tag3" @keydown.enter="saveEdit('tags')" @keydown.escape="cancelEdit" ref="editInput" />
            <span class="detail-hint">Comma-separated</span>
            <div class="detail-edit-actions">
              <button class="btn btn-sm btn-primary" @click="saveEdit('tags')">Save</button>
              <button class="btn btn-sm" @click="cancelEdit">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sessions list -->
    <div v-if="allSessions.length" class="sessions-list">
      <NuxtLink
        v-for="sess in allSessions"
        :key="sess.id"
        :to="`/research/${researchSlug}/session/${sess.code || sess.id}`"
        :class="['card session-widget', { 'session-active': sess.status === 'active' }]"
      >
        <div class="session-widget-header">
          <div class="flex items-center gap-2">
            <span v-if="sess.code" class="short-code">{{ sess.code }}</span>
            <StatusBadge :status="sess.status" />
          </div>
        </div>
        <h3 class="session-title">{{ sess.title }}</h3>
        <p v-if="sess.focus" class="session-focus">{{ sess.focus }}</p>
      </NuxtLink>
    </div>

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
            <svg v-if="t.status === 'completed'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            <svg v-else-if="t.status === 'failed'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            <svg v-else-if="t.status === 'blocked'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
            <svg v-else-if="t.status === 'deferred'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <svg v-else-if="t.status === 'in_progress'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4"/><path d="M12 18v4"/><path d="m4.93 4.93 2.83 2.83"/><path d="m16.24 16.24 2.83 2.83"/><path d="M2 12h4"/><path d="M18 12h4"/><path d="m4.93 19.07 2.83-2.83"/><path d="m16.24 7.76 2.83-2.83"/></svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/></svg>
          </span>
          <div class="todo-content">
            <span :class="['todo-text', { 'todo-done': t.status === 'completed' }]" v-html="renderRefs(t.title, researchSlug)"></span>
            <div v-if="t.result" class="card-meta todo-result" v-html="renderRefs(marked.parse(t.result) as string, researchSlug)"></div>
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

        <div class="sidebar-divider"></div>

        <!-- External links -->
        <div
          :class="['sidebar-item', { active: activeSection === '__links__' }]"
          @click="activeSection = '__links__'"
        >
          <div class="sidebar-item-content">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
            <span class="sidebar-item-name">External links</span>
            <span v-if="researchLinksTotal" class="sidebar-count">{{ researchLinksTotal }}</span>
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

        <!-- External links view -->
        <template v-else-if="isLinksView">
          <div class="section-header">
            <h2 class="section-title">External links</h2>
          </div>

          <div v-if="researchLinksLoading">
            <div v-for="i in 3" :key="i" class="skeleton-card skeleton-entry"></div>
          </div>

          <div v-else-if="researchLinksGrouped.length">
            <div v-for="group in researchLinksGrouped" :key="group.domain" class="links-domain-group">
              <h3 class="links-domain-title">{{ group.domain }}</h3>
              <div class="links-list">
                <a
                  v-for="link in group.links"
                  :key="link.url"
                  :href="link.url"
                  target="_blank"
                  rel="noopener"
                  class="card link-card"
                >
                  <div class="link-card-header">
                    <span class="link-title">{{ link.title || link.url }}</span>
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="link-external-icon"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                  </div>
                  <div v-if="link.entry_code" class="link-source">
                    <span class="short-code">{{ link.entry_code }}</span>
                    <span class="card-meta">{{ link.entry_title }}</span>
                  </div>
                </a>
              </div>
            </div>
          </div>

          <EmptyState v-else icon="&#x1F517;" title="No external links" description="Links from entry content will appear here." />
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
import { marked } from 'marked'
marked.setOptions({ gfm: true, breaks: true })

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

// All sessions
const { data: sessionsData } = await useApi<{ data: any[] }>(`/api/researches/${id}/sessions`)
const allSessions = computed(() => sessionsData.value?.data ?? [])

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
const editingField = ref<string | null>(null)
const editValue = ref('')
const editInput = ref<HTMLElement | null>(null)

function startEdit(field: string) {
  if (field === 'tags') {
    editValue.value = (research.value?.tags ?? []).join(', ')
  } else {
    editValue.value = research.value?.[field] ?? ''
  }
  editingField.value = field
  nextTick(() => editInput.value?.focus?.())
}

function cancelEdit() {
  editingField.value = null
  editValue.value = ''
}

async function saveEdit(field: string) {
  const body: Record<string, any> = {}
  if (field === 'tags') {
    body.tags = editValue.value.split(',').map((t: string) => t.trim()).filter(Boolean)
  } else {
    body[field] = editValue.value
  }
  await authFetch(`${rtBase}/api/researches/${id}`, { method: 'PUT', body })
  researchData.value = await authFetch<any>(`${rtBase}/api/researches/${id}`)
  editingField.value = null
  editValue.value = ''
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
  if (event.entity === 'task') {
    tasksData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/tasks`)
  }
  if (['question', 'session'].includes(event.entity)) {
    sessionsData.value = await authFetch<any>(`${rtBase}/api/researches/${id}/sessions`)
  }
})
</script>

<style scoped>
/* Header */
.research-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.research-actions { display: flex; align-items: center; gap: var(--space-3); }
.btn-count {
  font-size: 0.75em;
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  font-variant-numeric: tabular-nums;
  opacity: 0.7;
}
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

/* Details panel */
.details-panel { margin-bottom: var(--space-6); }
.details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-5);
}
.detail-field-wide { grid-column: 1 / -1; }
.detail-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--type-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
}
.detail-count {
  font-size: 0.625rem;
  background: var(--color-surface-hover);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  font-variant-numeric: tabular-nums;
}
.detail-value {
  font-size: var(--type-sm);
  color: var(--color-text);
  line-height: 1.6;
  cursor: default;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  transition: border-color var(--transition-fast);
}
.detail-value:hover { border-color: var(--color-border); }
.detail-value-pre { white-space: pre-wrap; }
.detail-empty {
  color: var(--color-text-muted);
  font-style: italic;
  opacity: 0.6;
}
.detail-edit { display: flex; flex-direction: column; gap: var(--space-2); }
.detail-input, .detail-textarea {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: var(--type-sm);
  font-family: inherit;
  line-height: 1.5;
}
.detail-textarea { resize: vertical; min-height: 60px; }
.detail-input:focus, .detail-textarea:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.detail-edit-actions { display: flex; gap: var(--space-2); }
.detail-hint { font-size: var(--type-xs); color: var(--color-text-muted); }
.memory-list { display: flex; flex-direction: column; gap: var(--space-2); }
.memory-item {
  display: flex;
  gap: var(--space-3);
  font-size: var(--type-sm);
  line-height: 1.5;
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface-hover);
  border-radius: var(--radius-sm);
}
.memory-bullet {
  font-size: var(--type-xs);
  font-weight: 700;
  color: var(--color-text-muted);
  min-width: 1.2em;
  text-align: right;
  flex-shrink: 0;
  margin-top: 2px;
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

/* Sessions list */
.sessions-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

/* Session widget */
.session-widget {
  display: block;
  text-decoration: none;
  color: inherit;
  margin-bottom: 0;
  border-color: var(--color-border);
  position: relative;
  overflow: hidden;
}
.session-active {
  border-color: rgba(108, 197, 224, 0.15);
}
.session-widget:hover { text-decoration: none; }
.session-widget::after {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 1px;
  background: linear-gradient(90deg, transparent, rgba(108, 197, 224, 0.3), transparent);
}
.session-widget-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: var(--space-1);
}
.session-title { font-size: var(--type-sm); font-weight: 600; letter-spacing: -0.01em; }
.session-focus { font-size: var(--type-xs); color: var(--color-text-muted); margin-top: var(--space-1); line-height: 1.4; }

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
.todo-check { display: flex; align-items: center; justify-content: center; width: 16px; flex-shrink: 0; color: var(--color-text-muted); }
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

/* External links */
.links-domain-group { margin-bottom: var(--space-5); }
.links-domain-title {
  font-size: var(--type-sm);
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: var(--space-2);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
}
.links-list { display: flex; flex-direction: column; gap: var(--space-2); }
.link-card {
  display: block;
  text-decoration: none;
  color: inherit;
  transition: border-color var(--transition-fast);
}
.link-card:hover { border-color: var(--color-border-strong); }
.link-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-2);
}
.link-title {
  font-size: var(--type-sm);
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.link-external-icon { opacity: 0.3; flex-shrink: 0; }
.link-card:hover .link-external-icon { opacity: 0.7; }
.link-source {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-2);
  font-size: var(--type-xs);
}

/* Skeleton */
.skeleton-header { height: 60px; margin-bottom: var(--space-6); }
.skeleton-sidebar { height: 300px; }
.skeleton-entry { height: 90px; margin-bottom: var(--space-3); }
</style>
