<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <div v-else-if="session">
    <!-- Header -->
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: 'Research', to: '/' },
        { label: researchName, to: `/research/${researchSlug}` },
        { label: session.title }
      ]" />
      <div class="session-header">
        <h1 class="page-title">{{ session.title }}</h1>
        <div class="session-header-actions">
          <StatusBadge :status="session.status" />
          <TeamViewerNotice v-if="isViewer" :team-name="researchData?.data?.research?.team_name" />
          <ActionMenu>
            <NuxtLink
              :to="`/research/${researchSlug}/session/${sessionSlug}/export`"
              class="action-menu-item"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              Export
            </NuxtLink>
          </ActionMenu>
        </div>
      </div>
      <p v-if="session.focus" class="card-meta mt-2" v-html="'Focus: ' + renderRefs(session.focus, researchSlug)"></p>
    </div>

    <!-- Notes -->
    <div v-if="session.notes" class="card notes-card">
      <h3 class="card-section-title">Session notes</h3>
      <div ref="notesEl" class="notes-text markdown-content" v-html="linkRefs(parseMarkdown(normalizeContent(session.notes)) as string, researchSlug)"></div>
    </div>

    <!-- Tabs -->
    <div class="content-tabs">
      <button
        :class="['content-tab', { active: activeTab === 'questions' }]"
        @click="activeTab = 'questions'"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>
        Questions
        <span class="tab-count">{{ progress.answered }} / {{ progress.total }}</span>
      </button>
      <button
        :class="['content-tab', { active: activeTab === 'entries' }]"
        @click="activeTab = 'entries'"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
        Entries
        <span v-if="sessionEntries.length" class="tab-count">{{ sessionEntries.length }}</span>
      </button>
      <button
        :class="['content-tab', { active: activeTab === 'changes' }]"
        @click="activeTab = 'changes'"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v5h5"/><path d="M3.05 13A9 9 0 1 0 6 5.3L3 8"/><path d="M12 7v5l4 2"/></svg>
        Changes
      </button>
    </div>

    <!-- Questions panel -->
    <div v-if="activeTab === 'questions'">
      <div class="panel-progress">
        <ProgressBar :value="progress.answered" :total="progress.total" />
        <div class="progress-stats">
          <span class="stat-answered">{{ progress.answered }} answered</span>
          <span class="stat-pending">{{ progress.pending }} pending</span>
          <span v-if="progress.deferred" class="stat-muted">{{ progress.deferred }} deferred</span>
          <span v-if="progress.skipped" class="stat-skipped">{{ progress.skipped }} skipped</span>
        </div>
      </div>

      <!-- Add question -->
      <button v-if="canWrite && !showAddQuestion" class="btn btn-sm add-btn" @click="showAddQuestion = true">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        Add question
      </button>
      <form v-else-if="canWrite" class="add-form" @submit.prevent="submitQuestion">
        <input v-model="newQuestion.text" class="add-input" placeholder="Question text..." required autofocus />
        <div class="add-form-row">
          <input v-model="newQuestion.area" class="add-input add-input-sm" placeholder="Area (optional)" />
          <select v-model="newQuestion.priority" class="add-select">
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="low">Low</option>
          </select>
          <button type="submit" class="btn btn-sm" :disabled="addingQuestion || !newQuestion.text.trim()">
            {{ addingQuestion ? 'Adding...' : 'Add' }}
          </button>
          <button type="button" class="btn btn-sm" @click="showAddQuestion = false">Cancel</button>
        </div>
      </form>

      <QuestionList :questions="questions" :research-slug="researchSlug" :session-id="sessionId" />
    </div>

    <!-- Entries panel -->
    <div v-if="activeTab === 'entries'">
      <div v-if="sessionEntries.length" class="grid entries-grid">
        <NuxtLink
          v-for="entry in sessionEntries"
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
      <EmptyState v-else icon="&#x1F4C4;" title="No entries" description="Entries linked to this session will appear here." />
    </div>

    <!-- Changes panel -->
    <SessionChangesList
      v-if="activeTab === 'changes' && session"
      :session-id="session.id"
      :research-slug="researchSlug"
    />
  </div>

  <EmptyState v-else icon="&#x1F50D;" title="Session not found" />
</template>

<script setup lang="ts">
import { parseMarkdown } from '~/composables/useSafeMarkdown'
import { renderMermaidBlocks } from '~/composables/useMermaid'
const route = useRoute()
const id = route.params.id as string
const sessionId = route.params.sessionId as string

const { data: researchData } = await useApi<{ data: any }>(`/api/researches/${id}`)
const researchName = computed(() => researchData.value?.data?.research?.name ?? 'Research')
const researchSlug = computed(() => researchData.value?.data?.research?.code || id)

// Every research-scoped page publishes the caller's role from the payload it
// already fetches, so the controls beneath it — down to a checkbox inside
// rendered content — know whether they may write.
const { canWrite, canAdmin, isViewer, setFromResearch } = useResearchRole()
watch(researchData, (d) => setFromResearch(d?.data?.research), { immediate: true })

const { data, pending } = await useApi<{ data: any }>(`/api/researches/${id}/sessions/${sessionId}`)

const session  = computed(() => data.value?.data?.session ?? data.value?.data?.Session)
const sessionSlug = computed(() => session.value?.code || sessionId)

// Mermaid rendering for session notes
const notesEl = ref<HTMLElement | null>(null)
watch(() => session.value?.notes, () => {
  nextTick(() => {
    if (notesEl.value) renderMermaidBlocks(notesEl.value)
  })
})
onMounted(() => {
  if (notesEl.value) renderMermaidBlocks(notesEl.value)
})
const questions = computed(() => data.value?.data?.questions ?? data.value?.data?.Questions ?? {})
const progress  = computed(() => ({
  total:    data.value?.data?.progress?.total    ?? 0,
  answered: data.value?.data?.progress?.answered ?? 0,
  pending:  data.value?.data?.progress?.pending  ?? 0,
  deferred: data.value?.data?.progress?.deferred ?? 0,
  skipped:  data.value?.data?.progress?.skipped  ?? 0,
}))

// Session entries (from API response)
const sessionEntries = computed(() => data.value?.data?.entries ?? [])

function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}

// Active tab
const activeTab = ref<'questions' | 'entries' | 'changes'>('questions')

const { authFetch } = useAuth()
const rtBase = useRuntimeConfig().public.apiBase || ''

// Add question
const showAddQuestion = ref(false)
const addingQuestion = ref(false)
const newQuestion = ref({ text: '', area: '', priority: 'medium' })
const realSessionId = computed(() => session.value?.id || sessionId)

async function submitQuestion() {
  if (!newQuestion.value.text.trim()) return
  addingQuestion.value = true
  try {
    await authFetch(`${rtBase}/api/sessions/${realSessionId.value}/questions`, {
      method: 'POST',
      body: JSON.stringify({
        questions: [{
          text: newQuestion.value.text,
          area: newQuestion.value.area,
          priority: newQuestion.value.priority,
        }],
      }),
      headers: { 'Content-Type': 'application/json' },
    })
    newQuestion.value = { text: '', area: '', priority: 'medium' }
    showAddQuestion.value = false
    data.value = await authFetch<any>(`${rtBase}/api/researches/${id}/sessions/${sessionId}`)
  } catch { /* ignore */ }
  addingQuestion.value = false
}

// Real-time updates
async function reloadSession() {
  data.value = await authFetch<any>(`${rtBase}/api/researches/${id}/sessions/${sessionId}`)
}

useResearchRealtime(
  () => id,
  (event) => { if (['question', 'session'].includes(event.entity)) reloadSession() },
  {
    researchId: () => researchData.value?.data?.research?.id,
    onResync: reloadSession,
  },
)
</script>

<style scoped>
.session-header { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); }
.session-header .page-title { min-width: 0; overflow-wrap: anywhere; }
.session-header-actions { display: flex; align-items: center; gap: var(--space-3); flex-shrink: 0; }
.notes-card { margin-bottom: var(--space-6); }
.notes-text { white-space: pre-wrap; color: var(--color-text-muted); font-size: var(--type-sm); line-height: 1.6; }

/* Panel progress */
.panel-progress { margin-bottom: var(--space-5); }
.progress-stats {
  display: flex; gap: var(--space-4); font-size: var(--type-xs);
  margin-top: var(--space-2); color: var(--color-text-muted);
}
.stat-skipped  { color: var(--color-error); font-weight: 500; }
.stat-muted    { color: var(--color-text-muted); }

/* Tabs */
.content-tabs {
  display: flex; gap: 0; margin-bottom: var(--space-5);
  border-bottom: 1px solid var(--color-border);
}
.content-tab {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-5);
  background: none; border: none; border-bottom: 2px solid transparent;
  color: var(--color-text-muted); font-size: var(--type-sm); font-weight: 500;
  font-family: inherit; cursor: pointer; transition: all var(--transition-fast);
  margin-bottom: -1px;
}
.content-tab:hover { color: var(--color-text); }
.content-tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.tab-count {
  font-size: var(--type-xs); color: var(--color-text-muted);
  background: var(--color-surface-hover); border-radius: 4px;
  padding: 0.15rem 0.4rem; font-variant-numeric: tabular-nums;
}
.content-tab.active .tab-count { background: var(--color-primary-muted); color: var(--color-primary); }

/* Add form */
.add-btn { margin-bottom: var(--space-4); }
.add-form { margin-bottom: var(--space-5); }
.add-input {
  width: 100%; padding: var(--space-2) var(--space-3);
  background: var(--color-surface); border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm); color: var(--color-text);
  font-size: var(--type-sm); font-family: inherit;
  margin-bottom: var(--space-2);
}
.add-input:focus { outline: 2px solid var(--color-primary); outline-offset: -1px; }
.add-input-sm { flex: 1; margin-bottom: 0; }
.add-form-row { display: flex; gap: var(--space-2); align-items: center; }
.add-select {
  padding: var(--space-2) var(--space-3);
  background: var(--color-surface); border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-sm); color: var(--color-text-muted);
  font-size: var(--type-sm); font-family: inherit; cursor: pointer;
}

/* Entries grid */
.entries-grid { grid-template-columns: 1fr; }
.entry-card { display: block; text-decoration: none; color: inherit; }
.entry-card-header { display: flex; justify-content: space-between; align-items: flex-start; gap: var(--space-2); }
.entry-title-row { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }

.skeleton-header { height: 60px; margin-bottom: var(--space-4); }

/* Responsive */
@media (max-width: 768px) {
  .session-header { flex-direction: column; align-items: flex-start; gap: var(--space-2); }
  .add-form-row { flex-wrap: wrap; }
  .add-input-sm { flex: 1 1 100%; }
  .progress-stats { flex-wrap: wrap; gap: var(--space-2); }
  .content-tab { padding: var(--space-2) var(--space-3); font-size: var(--type-xs); }
}
</style>
