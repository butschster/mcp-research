<template>
  <div v-if="pending">
    <div class="skeleton-card skeleton-header"></div>
    <div class="skeleton-card skeleton-content"></div>
  </div>

  <EmptyState
    v-else-if="excluded"
    title="Not part of this link"
    description="The person who shared this research didn't include interview sessions. Ask them if you need them."
  >
    <NuxtLink class="btn btn-primary" :to="researchPath(slug)">Back to the research</NuxtLink>
  </EmptyState>

  <EmptyState
    v-else-if="!session"
    title="Couldn't load this session"
    description="It may have been removed, or the server didn't answer. The rest of the research is still here."
  >
    <button class="btn btn-primary" @click="load()">Try again</button>
    <NuxtLink class="btn" :to="researchPath(slug)">Back to the research</NuxtLink>
  </EmptyState>

  <div v-else>
    <div class="page-header">
      <Breadcrumbs :crumbs="[
        { label: researchName, to: researchPath(slug) },
        { label: session.title },
      ]" />
      <div class="session-header">
        <h1 class="page-title">{{ session.title }}</h1>
        <StatusBadge :status="session.status" />
      </div>
      <p v-if="session.focus" class="card-meta mt-2" v-html="'Focus: ' + renderRefs(session.focus, slug)"></p>
    </div>

    <div v-if="session.notes" class="card notes-card">
      <h3 class="card-section-title">Session notes</h3>
      <div class="notes-text markdown-content" v-html="renderedNotes"></div>
    </div>

    <div class="panel-progress">
      <ProgressBar :value="progress.answered" :total="progress.total" />
      <div class="progress-stats">
        <span class="stat-answered">{{ progress.answered }} answered</span>
        <span class="stat-pending">{{ progress.pending }} pending</span>
      </div>
    </div>

    <QuestionList :questions="grouped" :research-slug="slug" :session-id="sessionCode" />

    <div v-if="sessionEntries.length" class="session-entries">
      <h3 class="card-section-title">Entries from this session</h3>
      <div class="grid">
        <EntryCard
          v-for="entry in sessionEntries"
          :key="entry.id"
          :entry="entry"
          :research-slug="slug"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { parseMarkdown } from '~/composables/useSafeMarkdown'

const route = useRoute()
const sessionCode = computed(() => route.params.sessionId as string)

const { shareFetch, research, researchId, researchCode, include, slug } = useShare()
const researchName = computed(() => research.value?.name || 'Research')

const session = ref<any | null>(null)
// The API hands questions back already grouped by status, which is the shape
// QuestionList takes.
const grouped = ref<Record<string, any[]>>({})
const serverProgress = ref<any>(null)
const sessionEntries = ref<any[]>([])
const pending = ref(true)

// Sending somebody to ask for something they already have is the worst of the
// wrong answers here, so this screen is shown only when the link genuinely
// excludes sessions — not for a mistyped code or a dropped connection.
const excluded = computed(() => !include.value.sessions)

async function load() {
  if (excluded.value) {
    session.value = null
    pending.value = false
    return
  }
  pending.value = true
  try {
    const res = await shareFetch<{ data: any }>(
      `/researches/${researchId.value}/sessions/${sessionCode.value}`,
    )
    session.value = res.data?.session ?? null
    grouped.value = res.data?.questions ?? {}
    serverProgress.value = res.data?.progress ?? null
    sessionEntries.value = res.data?.entries ?? []
  } catch {
    session.value = null
  } finally {
    pending.value = false
  }
}

watch([sessionCode, researchId], () => void load(), { immediate: true })

const progress = computed(() => {
  if (serverProgress.value) return serverProgress.value
  const all = Object.values(grouped.value).flat()
  const answered = (grouped.value.answered ?? []).length
  return { total: all.length, answered, pending: all.length - answered }
})

const renderedNotes = computed(() => {
  if (!session.value?.notes) return ''
  const html = parseMarkdown(String(session.value.notes).replace(/\\n/g, '\n')) as string
  return linkRefs(html, slug.value)
})

useResearchRealtime(
  () => slug.value,
  (event) => { if (event.entity === 'session' || event.entity === 'question') void load() },
  { onResync: () => void load(), researchId: () => researchId.value },
)
</script>

<style scoped>
.session-header { display: flex; align-items: center; gap: var(--space-3); flex-wrap: wrap; }
.notes-card { margin-bottom: var(--space-5); }
.panel-progress { margin-bottom: var(--space-4); }
.progress-stats { display: flex; gap: var(--space-3); font-size: var(--type-xs); color: var(--color-text-muted); margin-top: var(--space-2); }
.session-entries { margin-top: var(--space-6); }
</style>
